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

package dmiserver

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

var metricsProducer sarama.AsyncProducer

// InitializeDMKafkaPublishers initializes  metrics kafka publisher
func InitializeDMKafkaPublishers(NewAsyncProducer func([]string, *sarama.Config) (sarama.AsyncProducer, error), oltID int, msgBusEndPoint string) error {
	var err error
	sarama.Logger = log.New()
	// producer config
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Metadata.Retry.Max = 10
	config.Metadata.Retry.Backoff = 10 * time.Second
	config.ClientID = "BBSim-OLT-Metrics-" + strconv.Itoa(oltID)

	metricsProducer, err = NewAsyncProducer([]string{msgBusEndPoint}, config)
	return err
}

// DMKafkaPublisher receives messages on ch and publish them to kafka on topic
func DMKafkaPublisher(ctx context.Context, ch chan interface{}, topic string) {
	defer log.Debugf("DMKafkaPublisher stopped")
loop:
	for {
		select {
		case metric := <-ch:
			log.Tracef("Writing to kafka topic(%s): %v", topic, metric)
			jsonMet, err := json.Marshal(metric)
			if err != nil {
				log.Errorf("Failed to get json metric %v", err)
				continue
			}
			metricsProducer.Input() <- &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.ByteEncoder(jsonMet),
			}
		case <-ctx.Done():
			log.Infof("Stopping DM Kafka Publisher for topic %s", topic)
			break loop
		}
	}
}
