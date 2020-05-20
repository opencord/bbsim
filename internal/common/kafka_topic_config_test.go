// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

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
	err := InitializePublisher(mockLib.NewAsyncProducer, 0)

	assert.Equal(t, err, nil)
	assert.Equal(t, topic, "BBSim-OLT-0-Events")

	Options.BBSim.KafkaEventTopic = "Testing-Topic"
	err = InitializePublisher(mockLib.NewAsyncProducer, 0)
	assert.Equal(t, topic, "Testing-Topic")
	assert.Equal(t, err, nil)
}
