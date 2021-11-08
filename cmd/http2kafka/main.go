package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Shopify/sarama"

	"github.com/megaease/consuldemo/pkg/jsontool"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	httpPort     = "28080"
	kafkaBrokers = "kafka-0.kafka-hs:9093,kafka-1.kafka-hs:9093,kafka-2.kafka-hs:9093"
	kafkaTopic   = "log-tracing"

	producer sarama.AsyncProducer
)

func preflight() {
	v := os.Getenv("HTTP_PORT")
	if v != "" {
		httpPort = v
	}
	log.Printf("http port: %s", httpPort)

	v = os.Getenv("KAFKA_BROKERS")
	if v != "" {
		kafkaBrokers = v
	}

	brokers := strings.Split(kafkaBrokers, ",")
	log.Printf("kafka brokers: %v", brokers)

	v = os.Getenv("KAFKA_TOPIC")
	if v != "" {
		kafkaTopic = v
	}
	log.Printf("kafka topic: %s", kafkaTopic)

	for {
		config := sarama.NewConfig()
		config.Version = sarama.MinVersion

		var err error
		producer, err = sarama.NewAsyncProducer(brokers, config)
		if err != nil {
			log.Printf("new producer failed: %v", err)
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("producer built successfully")
			break
		}
	}

	go func() {
		for err := range producer.Errors() {
			log.Printf("producer errors output: %v", err)
		}
	}()
}

func main() {
	preflight()

	zipkinServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: newZipkinHandler(),
	}

	go func() {
		err := zipkinServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			exitf("%v", err)
		}
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch

	zipkinServer.Shutdown(context.TODO())
	producer.Close()
}

type zipkinHandler struct {
	messageCount uint64
}

func newZipkinHandler() *zipkinHandler {
	return &zipkinHandler{}
}

func (h *zipkinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("read body failed: %v", err)))
		return
	}
	r.Body.Close()

	objects := jsontool.GetObjects(body, func(object []byte) bool {
		if strings.Contains(string(object), "/v1/catalog/services") ||
			strings.Contains(string(object), "/v1/health/service") {
			return false
		}
		return true
	})

	for _, object := range objects {
		object, _ = sjson.SetBytes(object, "type", "log-tracing")

		name := gjson.GetBytes(object, "name").String()
		if strings.Contains(name, "restaurant") {
			object, _ = sjson.SetBytes(object, "kind", "SERVER")
		}

		log.Printf("message: %s\n", object)

		producer.Input() <- &sarama.ProducerMessage{
			Topic: kafkaTopic,
			Key:   nil,
			Value: sarama.ByteEncoder(object),
		}

		messageCount := atomic.AddUint64(&h.messageCount, 1)
		log.Printf("message count: %d", messageCount)
	}
}

func exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
