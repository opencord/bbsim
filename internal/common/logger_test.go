/*
 * Copyright 2018-2024 Open Networking Foundation (ONF) and the ONF Contributors

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

package common_test

import (
	"github.com/opencord/bbsim/internal/common"
	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
	"testing"
)

func init() {
	common.SetLogLevel(logrus.StandardLogger(), "error", false)
}

func Test_SetLogLevel(t *testing.T) {
	log := logrus.New()

	common.SetLogLevel(log, "trace", false)
	assert.Equal(t, log.Level, logrus.TraceLevel)

	common.SetLogLevel(log, "debug", false)
	assert.Equal(t, log.Level, logrus.DebugLevel)

	common.SetLogLevel(log, "info", false)
	assert.Equal(t, log.Level, logrus.InfoLevel)

	common.SetLogLevel(log, "warn", false)
	assert.Equal(t, log.Level, logrus.WarnLevel)

	common.SetLogLevel(log, "error", false)
	assert.Equal(t, log.Level, logrus.ErrorLevel)

	common.SetLogLevel(log, "foobar", false)
	assert.Equal(t, log.Level, logrus.DebugLevel)
}

func Test_SetLogLevelCaller(t *testing.T) {
	log := logrus.New()

	common.SetLogLevel(log, "debug", true)
	assert.Equal(t, log.ReportCaller, true)

	common.SetLogLevel(log, "debug", false)
	assert.Equal(t, log.ReportCaller, false)
}
