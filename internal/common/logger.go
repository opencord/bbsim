// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package common

import log "github.com/sirupsen/logrus"

func SetLogLevel(logger *log.Logger, level string, caller bool) {

	logger.SetReportCaller(caller)

	switch level {
	case "trace":
		logger.SetLevel(log.TraceLevel)
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.DebugLevel)
		logger.WithFields(log.Fields{
			"level": level,
		}).Warn("The provided level is unknown. Defaulting to 'debug'")
	}

}
