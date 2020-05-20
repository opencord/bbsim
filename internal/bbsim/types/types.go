// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/google/gopacket"
)

// deprecated, use OnuPacketOutMessage instead
type ByteMsg struct {
	IntfId uint32
	OnuId  uint32
	Bytes  []byte
}

type PacketMsg struct {
	Pkt gopacket.Packet
}
