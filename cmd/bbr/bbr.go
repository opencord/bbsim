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

package main

import (
	"github.com/opencord/bbsim/internal/bbr"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
)

func main() {
	options := common.GetBBROpts()

	common.SetLogLevel(log.StandardLogger(), options.LogLevel, options.LogCaller)

	log.WithFields(log.Fields{
		"OltID":        options.OltID,
		"NumNniPerOlt": options.NumNniPerOlt,
		"NumPonPerOlt": options.NumPonPerOlt,
		"NumOnuPerPon": options.NumOnuPerPon,
		"BBSimIp":      options.BBSimIp,
		"BBSimPort":    options.BBSimPort,
	}).Info("BroadBand Reflector is on")

	olt := bbr.OLTMock{
		//OnuIndicationsChannel: make(chan interface{}, 1024),
		Onus:          make(map[string]bbr.ONUMock),
		TargetOnus:    options.NumOnuPerPon * options.NumPonPerOlt,
		TargetIp:      options.BBSimIp,
		TargetPort:    options.BBSimPort,
		TargetApiPort: options.BBSimApiPort,
		STag:          options.STag,
		NextCtag:      options.CTagInit,
	}

	olt.Enable()

}
