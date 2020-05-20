// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/opencord/voltha-protos/v2/go/openolt"
)

// represent a gRPC stream
type Stream interface {
	Send(*openolt.Indication) error
}
