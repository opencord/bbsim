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

package dhcp

import (
	"gotest.tools/assert"
	"net"
	"testing"
)

func TestCreateIpFromMacAddress(t *testing.T) {
	dhcpServer := NewDHCPServer()

	mac1 := net.HardwareAddr{0x2e, 0x60, 0x00, 0x0c, 0x0f, 0x02}
	ip1 := dhcpServer.createIpFromMacAddress(mac1)
	assert.Equal(t, "10.12.15.2", ip1.String())

	mac2 := net.HardwareAddr{0x2e, 0x60, 0x00, 0x00, 0x00, 0x00}
	ip2 := dhcpServer.createIpFromMacAddress(mac2)
	assert.Equal(t, "10.0.0.0", ip2.String())
}
