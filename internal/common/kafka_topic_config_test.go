/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
	"testing"

	"github.com/Shopify/sarama"
	"gotest.tools/assert"
)

type mockAsyncProducer struct {
	input     chan *sarama.ProducerMessage
	successes chan *sarama.ProducerMessage
	errors    chan *sarama.ProducerError
}

func (p mockAsyncProducer) AsyncClose() {

}

func (p mockAsyncProducer) Close() error {
	return nil
}

func (p mockAsyncProducer) Input() chan<- *sarama.ProducerMessage {
	return p.input
}

func (p mockAsyncProducer) Successes() <-chan *sarama.ProducerMessage {
	return p.successes
}

func (p mockAsyncProducer) Errors() <-chan *sarama.ProducerError {
	return p.errors
}

type mockSarama struct{}

func (m mockSarama) NewAsyncProducer(addrs []string, conf *sarama.Config) (sarama.AsyncProducer, error) {

	producer := &mockAsyncProducer{
		errors:    make(chan *sarama.ProducerError),
		input:     make(chan *sarama.ProducerMessage),
		successes: make(chan *sarama.ProducerMessage),
	}

	return producer, nil
}

func TestInitializePublisher(t *testing.T) {
	mockLib := mockSarama{}

	Config = &GlobalConfig{}

	err := InitializePublisher(mockLib.NewAsyncProducer, 0)

	assert.Equal(t, err, nil)
	assert.Equal(t, topic, "BBSim-OLT-0-Events")

	Config.BBSim.KafkaEventTopic = "Testing-Topic"
	err = InitializePublisher(mockLib.NewAsyncProducer, 0)
	assert.Equal(t, topic, "Testing-Topic")
	assert.Equal(t, err, nil)
}
