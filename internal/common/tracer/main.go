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

package tracer

import (
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

// initJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func InitJaeger(service string) (opentracing.Tracer, io.Closer, error) {
	cfg, err := config.FromEnv()
	if err != nil {
		// parsing errors might happen here, such as when we get a string where we expect a number
		log.Errorf("Could not parse Jaeger env vars: %s", err.Error())
		return nil, nil, err
	}

	cfg.Sampler = &config.SamplerConfig{
		Type:  "const",
		Param: 1,
	}

	cfg.Reporter = &config.ReporterConfig{
		LogSpans: true,
	}

	cfg.ServiceName = service

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		log.Errorf("Could not initialize jaeger tracer: %s", err.Error())
		return nil, nil, err
	}
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)

	return tracer, closer, nil
}
