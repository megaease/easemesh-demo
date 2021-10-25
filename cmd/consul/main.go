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
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/consul/api"
)

var (
	consulAddress = os.Getenv("CONSUL_ADDRESS")
	podIP         = os.Getenv("POD_IP")
	podPort       = os.Getenv("POD_PORT")
	serviceName   = os.Getenv("SERVICE_NAME")
	instanceID    = os.Getenv("INSTANCE_ID")

	_port int

	restyClient = resty.New()
)

const (
	orderSerice       = "order-consul"
	restaurantService = "restaurant-consul"
	awardService      = "award-mesh"
	deliveryService   = "delivery-consul"
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
	if consulAddress == "" {
		exitf("empty consul address")
	}
	if instanceID == "" {
		exitf("empty instance id")
	}
	if serviceName == "" {
		exitf("empty serviceName")
	}

	switch serviceName {
	case orderSerice, restaurantService, deliveryService:
	default:
		exitf("unsupport service name: %s", serviceName)
	}

	if podIP == "" {
		exitf("empty pod ip")
	}

	if podPort == "" {
		exitf("empty pod port")
	}
	port, err := strconv.ParseUint(podPort, 10, 16)
	if err != nil {
		exitf("parse port %s failed: %v", podPort, err)
	}
	_port = int(port)
}

func main() {
	fmt.Println("preflight...")
	prefligt()

	client, err := buildClient()
	if err != nil {
		exitf("build consul client failed: %s\n", err)
	}

	exit, done := make(chan struct{}), make(chan struct{})
	go runForConsul(client, exit, done)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", _port),
		Handler: newServiceHandler(client),
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			exitf("%v", err)
		}
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch

	close(exit)
	server.Shutdown(context.TODO())
	<-done
}

func exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

type serviceHandler struct {
	client *api.Client

	urlMutex      sync.Mutex
	restaurantURL string
	awardURL      string
	deliveryURL   string
}

func newServiceHandler(client *api.Client) *serviceHandler {
	h := &serviceHandler{client: client}

	go func() {
		for {
			var (
				restaurantURL string
				deliveryURL   string
				awardURL      string
				err           error
			)

			restaurantURL, err = h.getServiceURLFromConsul(restaurantService)
			if err != nil {
				log.Printf("get restaurant service url failed: %v", err)
			}

			awardURL, err = h.getServiceURLFromConsul(awardService)
			if err != nil {
				log.Printf("get award service url failed: %v", err)
			}

			deliveryURL, err = h.getServiceURLFromConsul(deliveryService)
			if err != nil {
				log.Printf("get delivery service url failed: %v", err)
			}

			h.urlMutex.Lock()
			h.restaurantURL = restaurantURL
			h.awardURL = awardURL
			h.deliveryURL = deliveryURL
			h.urlMutex.Unlock()

			<-time.After(5 * time.Second)
		}
	}()

	return h
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
		resp, err = h.handleRestaurant(body)
	case deliveryService:
		resp, err = h.handleDelivery(body)
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
	req := &OrderRequest{}
	err := json.Unmarshal(body, req)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	restaurantURL, err := h.getServiceURLFromCache(restaurantService)
	if err != nil {
		panic(fmt.Errorf("get restaurant url failed: %v", err))
	}

	restaurantResp, err := restyClient.R().SetHeader("Content-Type", "application/json").SetBody(RestaurantRequest{
		OrderID: req.OrderID,
		Food:    req.Food,
	}).SetResult(&RestaurantResponse{}).Post(restaurantURL)
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

	awardURL, err := h.getServiceURLFromCache(awardService)
	if err == nil {
		awardResp, err := restyClient.R().SetHeader("Content-Type", "application/json").SetBody(AwardRequest{
			OrderID: req.OrderID,
		}).SetResult(&AwardResponse{}).Post(awardURL)

		if err == nil {
			resp.Award = &OrderResponseItem{
				Item:         awardResp.Result().(*AwardResponse).Award,
				DeliveryTime: awardResp.Result().(*AwardResponse).DeliveryTime,
			}
		} else {
			log.Printf("call award %s failed: %v", awardURL, err)
		}
	}

	return resp, nil
}

func (h *serviceHandler) handleRestaurant(body []byte) (interface{}, error) {
	req := &RestaurantRequest{}
	err := json.Unmarshal(body, req)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	deliveryURL, err := h.getServiceURLFromCache(deliveryService)
	if err != nil {
		panic(fmt.Errorf("get delivery url failed: %v", err))
	}

	deliveryResp, err := restyClient.R().SetHeader("Content-Type", "application/json").SetBody(DeliveryRequest{
		OrderID: req.OrderID,
		Item:    req.Food,
	}).SetResult(&DeliveryResponse{}).Post(deliveryURL)
	if err != nil {
		panic(fmt.Errorf("call delivery service failed: %v", err))
	}

	return &RestaurantResponse{
		OrderID:      req.OrderID,
		Food:         req.Food,
		DeliveryTime: deliveryResp.Result().(*DeliveryResponse).DeliveryTime,
	}, nil
}

func (h *serviceHandler) handleDelivery(body []byte) (interface{}, error) {
	req := &DeliveryRequest{}
	err := json.Unmarshal(body, req)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %v", err)
	}

	deliveryTime := time.Now().Add(10 * time.Minute)

	return &DeliveryResponse{
		OrderID:      req.OrderID,
		Item:         req.Item,
		DeliveryTime: deliveryTime.Local().Format(timeFormat),
	}, nil
}

func (h *serviceHandler) getServiceURLFromCache(serviceName string) (string, error) {
	h.urlMutex.Lock()
	defer h.urlMutex.Unlock()

	var url string

	switch serviceName {
	case restaurantService:
		url = h.restaurantURL
	case awardService:
		url = h.awardURL
	case deliveryService:
		url = h.deliveryURL
	default:
		return "", fmt.Errorf("unsupport service %s", serviceName)
	}

	if url == "" {
		return "", fmt.Errorf("url not found")
	}

	return url, nil
}

func (h *serviceHandler) getServiceURLFromConsul(serviceName string) (string, error) {
	services, _, err := h.client.Catalog().Service(serviceName, "", nil)
	if err != nil {
		return "", nil
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no service available")
	}

	return fmt.Sprintf("http://%s:%d", services[0].ServiceAddress, services[0].ServicePort), nil
}

func deregister(client *api.Client, serviceID string) {
	err := client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		log.Printf("deregister %s failed: %v", serviceID, err)
	} else {
		log.Printf("deregister %s", serviceID)
	}
}

func runForConsul(client *api.Client, exit, done chan struct{}) {
	defer func() {
		deregister(client, instanceID)
		close(done)
	}()

	firstCleanDone := false
	for {
		services, _, err := client.Catalog().Service(serviceName, "", nil)
		if err != nil {
			log.Printf("get catalog service %s failed: %v", serviceName, err)
			select {
			case <-time.After(5 * time.Second):
				continue
			case <-exit:
				return
			}
		}

		needRegister := true
		for _, service := range services {
			if service.ServiceID != instanceID {
				if !firstCleanDone {
					deregister(client, service.ServiceID)
				}
			}
			if service.ServiceID == instanceID &&
				service.ServiceAddress == podIP &&
				service.ServicePort == _port {
				needRegister = false
			}
		}

		firstCleanDone = true

		if needRegister {
			registration := &api.AgentServiceRegistration{
				Kind:    api.ServiceKindTypical,
				ID:      instanceID,
				Name:    serviceName,
				Address: podIP,
				Port:    _port,
			}

			err := client.Agent().ServiceRegister(registration)
			if err != nil {
				log.Printf("register %s/%s failed: %v", serviceName, instanceID, err)
			}
		}

		select {
		case <-time.After(5 * time.Second):
		case <-exit:
			return
		}
	}
}

func buildClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = consulAddress
	config.Scheme = "http"

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil

}
