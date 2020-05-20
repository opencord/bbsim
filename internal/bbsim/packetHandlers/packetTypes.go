// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package packetHandlers

type PacketType int

const (
	UNKNOWN PacketType = iota
	EAPOL
	DHCP
)

func (t PacketType) String() string {
	names := [...]string{
		"UNKNOWN",
		"EAPOL",
		"DHCP",
	}
	return names[t]
}
