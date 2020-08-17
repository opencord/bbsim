/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"encoding/json"
	"strconv"
	"time"

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
func InitializePublisher(NewAsyncProducer func([]string, *sarama.Config) (sarama.AsyncProducer, error), oltID int) error {

	var err error
	sarama.Logger = log.New()
	// producer config
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Metadata.Retry.Max = 10
	config.Metadata.Retry.Backoff = 10 * time.Second
	config.ClientID = "BBSim-OLT-" + strconv.Itoa(oltID)
	if len(Config.BBSim.KafkaEventTopic) > 0 {
		topic = Config.BBSim.KafkaEventTopic
	} else {
		topic = "BBSim-OLT-" + strconv.Itoa(oltID) + "-Events"
	}

	producer, err = NewAsyncProducer([]string{Config.BBSim.KafkaAddress}, config)
	return err
}

// KafkaPublisher receives messages on eventChannel and publish them to kafka
func KafkaPublisher(eventChannel chan Event) {
	defer log.Debugf("KafkaPublisher stopped")
	for {
		event := <-eventChannel
		log.WithFields(log.Fields{
			"EventType": event.EventType,
			"OnuSerial": event.OnuSerial,
			"OltID":     event.OltID,
			"IntfID":    event.IntfID,
			"OnuID":     event.OnuID,
			"EpochTime": event.EpochTime,
			"Timestamp": event.Timestamp,
		}).Trace("Received event on channel")
		jsonEvent, err := json.Marshal(event)
		if err != nil {
			log.Errorf("Failed to get json event %v", err)
			continue
		}
		producer.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(jsonEvent),
		}
		log.WithFields(log.Fields{
			"EventType": event.EventType,
			"OnuSerial": event.OnuSerial,
			"OltID":     event.OltID,
			"IntfID":    event.IntfID,
			"OnuID":     event.OnuID,
			"EpochTime": event.EpochTime,
			"Timestamp": event.Timestamp,
		}).Debug("Event sent on kafka")
	}

}
