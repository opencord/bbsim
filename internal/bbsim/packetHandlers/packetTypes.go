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

package packetHandlers

type PacketType int

const (
	UNKNOWN PacketType = iota
	EAPOL
	DHCP
	IGMP
	ICMP
)

func (t PacketType) String() string {
	names := [...]string{
		"UNKNOWN",
		"EAPOL",
		"DHCP",
		"IGMP",
		"ICMP",
	}
	return names[t]
}
