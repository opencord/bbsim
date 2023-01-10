/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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

package common

import (
	"net"
	"strconv"

	"github.com/opencord/voltha-protos/v5/go/openolt"
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
