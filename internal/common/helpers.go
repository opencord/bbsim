// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"net"
	"strconv"

	"github.com/opencord/voltha-protos/v2/go/openolt"
)

func OnuSnToString(sn *openolt.SerialNumber) string {
	s := string(sn.VendorId)
	for _, i := range sn.VendorSpecific {
		s = s + strconv.FormatInt(int64(i/16), 16) + strconv.FormatInt(int64(i%16), 16)
	}
	return s
}

// GetIPAddr returns the IPv4 address of an interface. 0.0.0.0 is returned if the IP cannot be determined.
func GetIPAddr(ifname string) (string, error) {
	ip := "0.0.0.0"

	intf, err := net.InterfaceByName(ifname)
	if err != nil {
		return ip, err
	}

	addrs, err := intf.Addrs()
	if err != nil {
		return ip, err
	}

	for _, addr := range addrs {
		// get first IPv4 address
		switch v := addr.(type) {
		case *net.IPNet:
			if v.IP.To4() != nil {
				ip = v.IP.String()
				break
			}
		}
	}

	return ip, nil
}
