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
	orderSerice       = "order-mesh"
	restaurantService = "restaurant-mesh"
	awardService      = "award-mesh"
	deliveryService   = "delivery-mesh"
	timeFormat        = "2006-01-02T15:04:05"
)

type (
	// Order
	OrderRequest struct {
		OrderID string `json:"order_id"`
		Food    string `json:"food"`
	}

	OrderResponse struct {
		OrderID string             `json:"order_id"`
		Food    *OrderResponseItem `json:"food"`
		Award   *OrderResponseItem `json:"award"`
	}

	OrderResponseItem struct {
		DeliveryTime string `json:"delivery_time"`
		Item         string `json:"item"`
	}

	// Restaurant
	RestaurantRequest struct {
		OrderID string `json:"order_id"`
		Food    string `json:"food"`
	}

	RestaurantResponse struct {
		OrderID      string `json:"order_id"`
		Food         string `json:"food"`
		DeliveryTime string `json:"delivery_time"`
	}

	// Award
	AwardRequest struct {
		OrderID string `json:"order_id"`
	}

	AwardResponse struct {
		OrderID      string `json:"order_id"`
		Award        string `json:"award"`
		DeliveryTime string `json:"delivery_time"`
	}

	// Delivery
	DeliveryRequest struct {
		OrderID string `json:"order_id"`
		Item    string `json:"item"`
	}

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
	case orderSerice, restaurantService, deliveryService:
	default:
		exitf("unsupport service name: %s", serviceName)
	}

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
		resp, err = h.handleOrder(body)
	case restaurantService:
		resp, err = h.handleRestaurant(r.Header, body)
	case deliveryService:
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
	w.WriteHeader(200)
	w.Write(buff)
}

func (h *serviceHandler) handleOrder(body []byte) (interface{}, error) {
	span := tracing.NewSpan(tracer, serviceName)
	defer span.Finish()

	req := &OrderRequest{}
	err := json.Unmarshal(body, req)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	restaurantURL := fmt.Sprintf("http://%s:%d", restaurantService, podEgressPort)
	restaurantReq := restyClient.R()
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

	resp := &OrderResponse{
		OrderID: req.OrderID,
		Food: &OrderResponseItem{
			Item:         req.Food,
			DeliveryTime: restaurantResp.Result().(*RestaurantResponse).DeliveryTime,
		},
	}

	// NOTE: Allow failure of the award service.

	awardURL := fmt.Sprintf("http://%s:%d", awardService, podEgressPort)
	awardReq := restyClient.R()
	awardReq.SetHeader("Content-Type", "application/json").SetBody(AwardRequest{
		OrderID: req.OrderID,
	})
	awardReq.SetResult(&AwardResponse{})

	tracer.Inject(span.Context(),
		opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(awardReq.Header))

	awardResp, err := awardReq.Post(awardURL)
	if err != nil {
		log.Printf("call award %s failed: %v", awardURL, err)
	} else {
		resp.Award = &OrderResponseItem{
			Item:         awardResp.Result().(*AwardResponse).Award,
			DeliveryTime: awardResp.Result().(*AwardResponse).DeliveryTime,
		}
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
	deliveryReq.SetResult(&DeliveryResponse{})

	deliveryResp, err := deliveryReq.Post(deliveryURL)
	if err != nil {
		panic(fmt.Errorf("call delivery service failed: %v", err))
	}

	return &RestaurantResponse{
		OrderID:      req.OrderID,
		Food:         req.Food,
		DeliveryTime: deliveryResp.Result().(*DeliveryResponse).DeliveryTime,
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

	deliveryTime := time.Now().Add(10 * time.Minute)

	// NOTE: Make tracing more readable
	time.Sleep(10 * time.Millisecond)

	return &DeliveryResponse{
		OrderID:      req.OrderID,
		Item:         req.Item,
		DeliveryTime: deliveryTime.Local().Format(timeFormat),
	}, nil
}
