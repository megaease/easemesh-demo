package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-resty/resty/v2"
	"github.com/opentracing/opentracing-go"

	"github.com/megaease/consuldemo/pkg/tracing"
	"github.com/megaease/consuldemo/pkg/tracing/zipkin"
)

var (
	podServicePort = 80
	podHealthPort  = 9900
	podEgressPort  = 13002

	serviceName     = os.Getenv("SERVICE_NAME")
	zipkinServerURL = os.Getenv("ZIPKIN_SERVER_URL")

	restyClient = resty.New()

	tracer *tracing.Tracing
)

const (
	orderSerice             = "order-mesh"
	restaurantService       = "restaurant-mesh"
	restaurantCanaryService = "restaurant-mesh-canary"
	awardService            = "award-mesh"
	deliveryService         = "delivery-mesh"
	devliveryCanaryService  = "delivery-mesh-canary"

	timeFormat = "2006-01-02T15:04:05"
)

type (
	// OrderRequest is the request of order.
	OrderRequest struct {
		OrderID string `json:"order_id"`
		Food    string `json:"food"`
	}

	// OrderResponse is the response of order.
	OrderResponse struct {
		OrderID    string              `json:"order_id"`
		Restuarant *RestaurantResponse `json:"restuarant"`
		Award      *AwardResponse      `json:"award,omitempty"`
	}

	// RestaurantRequest is the request of restuarant.
	RestaurantRequest struct {
		OrderID string `json:"order_id"`
		Food    string `json:"food"`
	}

	// RestaurantResponse is the response of restaurant.
	RestaurantResponse struct {
		OrderID      string `json:"order_id"`
		Food         string `json:"food"`
		DeliveryTime string `json:"delivery_time"`
	}

	// AwardRequest is the request of award.
	AwardRequest struct {
		OrderID string `json:"order_id"`
	}

	// AwardResponse is the response of award.
	AwardResponse struct {
		OrderID      string `json:"order_id"`
		Award        string `json:"award"`
		DeliveryTime string `json:"delivery_time"`
	}

	// DeliveryRequest is the request of delivery.
	DeliveryRequest struct {
		OrderID string `json:"order_id"`
		Item    string `json:"item"`
	}

	// DeliveryResponse is the response of delivery.
	DeliveryResponse struct {
		OrderID      string `json:"order_id"`
		Item         string `json:"item"`
		DeliveryTime string `json:"delivery_time"`
	}
)

func prefligt() {
	if serviceName == "" {
		exitf("empty serviceName")
	}

	switch serviceName {
	case orderSerice, restaurantService, deliveryService,
		restaurantCanaryService, devliveryCanaryService:
	default:
		exitf("unsupport service name: %s", serviceName)
	}
	log.Printf("service: %s", serviceName)

	serverURL := "http://localhost:9411/api/v2/spans"
	if zipkinServerURL != "" {
		serverURL = zipkinServerURL
	}

	var err error
	tracer, err = tracing.New(&tracing.Spec{
		ServiceName: serviceName,
		Zipkin: &zipkin.Spec{
			Hostport:   fmt.Sprintf("%s:%d", serviceName, podServicePort),
			ServerURL:  serverURL,
			SampleRate: 1,
			SameSpan:   true,
			ID128Bit:   false,
		},
	})

	if err != nil {
		exitf("create tracing failed: %v", err)
	}
}

func main() {
	log.Println("preflight...")
	prefligt()

	healthServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", podHealthPort),
		Handler: newHealthHandler(),
	}

	serviceServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", podServicePort),
		Handler: newServiceHandler(),
	}

	go func() {
		err := serviceServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			exitf("%v", err)
		}
	}()

	go func() {
		log.Println("listen health port:", podHealthPort)
		err := healthServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			exitf("%v", err)
		}
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch

	healthServer.Shutdown(context.TODO())
	serviceServer.Shutdown(context.TODO())
}

func exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

type healthHandler struct{}

func newHealthHandler() *healthHandler {
	return &healthHandler{}
}

func (h *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

type serviceHandler struct {
	urlMutex      sync.Mutex
	restaurantURL string
	awardURL      string
	deliveryURL   string
}

func newServiceHandler() *serviceHandler {
	return &serviceHandler{}
}

func (h *serviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("%v", r)
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("%v", r)))
		}
	}()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("read body failed: %v", err)))
		return
	}

	log.Printf("receive %s %s %s", r.Method, r.URL.Path, body)

	defer r.Body.Close()

	var resp interface{}

	switch serviceName {
	case orderSerice:
		resp, err = h.handleOrder(r.Header, body)
	case restaurantService, restaurantCanaryService:
		resp, err = h.handleRestaurant(r.Header, body)
	case deliveryService, devliveryCanaryService:
		resp, err = h.handleDelivery(r.Header, body)
	default:
		panic(fmt.Errorf("BUG: no correct service"))
	}

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}

	buff, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")

	// NOTE: For human-readable in the first service.
	if serviceName == orderSerice {
		buff, err = yaml.JSONToYAML(buff)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/yaml")
	}

	log.Printf("response: %s", buff)

	w.WriteHeader(200)
	w.Write(buff)
}

func (h *serviceHandler) handleOrder(header http.Header, body []byte) (interface{}, error) {
	span := tracing.NewSpan(tracer, serviceName)
	defer span.Finish()

	req := &OrderRequest{}
	err := json.Unmarshal(body, req)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	restaurantURL := fmt.Sprintf("http://%s:%d", restaurantService, podEgressPort)
	restaurantReq := restyClient.R()
	restaurantReq.Header = header.Clone()
	restaurantReq.SetHeader("Content-Type", "application/json").SetBody(RestaurantRequest{
		OrderID: req.OrderID,
		Food:    req.Food,
	})
	restaurantReq.SetResult(&RestaurantResponse{})

	tracer.Inject(span.Context(),
		opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(restaurantReq.Header))

	restaurantResp, err := restaurantReq.Post(restaurantURL)
	if err != nil {
		panic(fmt.Errorf("call restaurant service failed: %v", err))
	}

	if restaurantResp.StatusCode() != 200 {
		panic(fmt.Errorf("call restaurant %s failed: status code: %d",
			restaurantURL, restaurantResp.StatusCode()))
	}

	resp := &OrderResponse{
		OrderID:    req.OrderID,
		Restuarant: restaurantResp.Result().(*RestaurantResponse),
	}

	// NOTE: Allow failure of the award service.

	awardURL := fmt.Sprintf("http://%s:%d", awardService, podEgressPort)
	awardReq := restyClient.R()
	awardReq.Header = header.Clone()
	awardReq.SetHeader("Content-Type", "application/json").SetBody(AwardRequest{
		OrderID: req.OrderID,
	})
	awardReq.SetResult(&AwardResponse{})

	tracer.Inject(span.Context(),
		opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(awardReq.Header))

	awardResp, err := awardReq.Post(awardURL)
	if err != nil {
		log.Printf("call award %s failed: %v", awardURL, err)
	} else if awardResp.StatusCode() != 200 {
		log.Printf("call award %s failed: status code: %d", awardURL, awardResp.StatusCode())
	} else {
		resp.Award = awardResp.Result().(*AwardResponse)
	}

	return resp, nil
}

func (h *serviceHandler) handleRestaurant(header http.Header, body []byte) (interface{}, error) {
	deliveryReq := restyClient.R()
	parentCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))

	if err != nil {
		log.Printf("extract zipkin header %+v failed: %v", header, err)
	} else {
		span := tracing.NewSpanWithContext(tracer, restaurantService, parentCtx)
		defer span.Finish()
		tracer.Inject(span.Context(),
			opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(deliveryReq.Header))
	}

	req := &RestaurantRequest{}
	err = json.Unmarshal(body, req)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	deliveryURL := fmt.Sprintf("http://%s:%d", deliveryService, podEgressPort)
	deliveryReq.SetHeader("Content-Type", "application/json").SetBody(DeliveryRequest{
		OrderID: req.OrderID,
		Item:    req.Food,
	})
	deliveryReq.Header = header.Clone()
	deliveryReq.SetResult(&DeliveryResponse{})

	deliveryResp, err := deliveryReq.Post(deliveryURL)
	if err != nil {
		panic(fmt.Errorf("call delivery %s failed: %v", deliveryURL, err))
	} else if deliveryResp.StatusCode() != 200 {
		log.Printf("call delivery %s failed: status code: %d", deliveryURL, deliveryResp.StatusCode())
	}

	result := deliveryResp.Result().(*DeliveryResponse)
	deliveryTime := result.DeliveryTime

	if serviceName == restaurantCanaryService {
		deliveryTime += " (cook duration: 5m)"
	}

	return &RestaurantResponse{
		OrderID:      req.OrderID,
		Food:         req.Food,
		DeliveryTime: deliveryTime,
	}, nil
}

func (h *serviceHandler) handleDelivery(header http.Header, body []byte) (interface{}, error) {
	log.Printf("header: %+v, body: %s", header, body)

	parentCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))
	if err != nil {
		log.Printf("extract zipkin header %+v failed: %v", header, err)
	} else {
		span := tracing.NewSpanWithContext(tracer, deliveryService, parentCtx)
		defer span.Finish()
	}

	req := &DeliveryRequest{}
	err = json.Unmarshal(body, req)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	deliveryTime := time.Now().Add(10 * time.Minute).Local().Format(timeFormat)

	if serviceName == devliveryCanaryService {
		deliveryTime += " (road duration: 7m)"
	}

	// NOTE: Make tracing more readable
	time.Sleep(10 * time.Millisecond)

	return &DeliveryResponse{
		OrderID:      req.OrderID,
		Item:         req.Item,
		DeliveryTime: deliveryTime,
	}, nil
}
