// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"github.com/opencord/bbsim/internal/common"
	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
	"testing"
)

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
