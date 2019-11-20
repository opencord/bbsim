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
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
	"io"
)

func NewTracer(serviceName string) (opentracing.Tracer, io.Closer, error) {
	// load config from environment variables
	cfg, err := config.FromEnv()
	if err != nil {
		return nil, nil, err
	}
	//cfg := config.Configuration{
	//	ServiceName: serviceName,
	//}
	cfg.ServiceName = serviceName

	// create tracer from config
	return cfg.NewTracer(
		config.Metrics(jprom.New()),
	)
}
