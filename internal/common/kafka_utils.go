package common

import (
	"encoding/json"
	"strconv"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

var producer sarama.AsyncProducer
var topic string

// Event defines structure for bbsim events
type Event struct {
	EventType string
	OnuSerial string
	OltID     int
	IntfID    int32
	OnuID     int32
	EpochTime int64
	Timestamp string
}

// InitializePublisher initalizes kafka publisher
func InitializePublisher(oltID int) error {

	var err error
	sarama.Logger = log.New()
	// producer config
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.ClientID = "BBSim-OLT-" + strconv.Itoa(oltID)
	topic = "BBSim-OLT-" + strconv.Itoa(oltID) + "-Events"

	producer, err = sarama.NewAsyncProducer([]string{Options.BBSim.KafkaAddress}, config)
	return err
}

// KafkaPublisher receives messages on eventChannel and publish them to kafka
func KafkaPublisher(eventChannel chan Event) {
	defer log.Debugf("KafkaPublisher stopped")
	for {
		select {
		case event := <-eventChannel:
			log.Debugf("Received event on channel %v", event)
			jsonEvent, err := json.Marshal(event)
			if err != nil {
				log.Errorf("Failed to get json event %v", err)
				continue
			}
			producer.Input() <- &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.ByteEncoder(jsonEvent),
			}
			log.Debugf("Event sent on kafka")
		}
	}
}
