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
func InitializePublisher(oltID int) error {

	var err error
	sarama.Logger = log.New()
	// producer config
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Metadata.Retry.Max = 10
	config.Metadata.Retry.Backoff = 10 * time.Second
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
