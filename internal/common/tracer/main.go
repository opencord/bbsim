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
	"github.com/opentracing/opentracing-go"

	"github.com/uber/jaeger-client-go"
	jc "github.com/uber/jaeger-client-go/config"

	"io"

	"fmt"
)

func InitJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg, err := jc.FromEnv()
	if err != nil {
		fmt.Printf("%v", err)
	}

	tracer, closer, err := cfg.New(service, jc.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer

}
