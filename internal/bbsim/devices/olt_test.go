// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package devices

import (
	"gotest.tools/assert"
	"net"
	"testing"
)

func createMockOlt(numPon int, numOnu int) OltDevice {
	olt := OltDevice{
		ID: 0,
	}

	for i := 0; i < numPon; i++ {
		pon := PonPort{
			ID: uint32(i),
		}

		for j := 0; j < numOnu; j++ {
			onuId := uint32(i + j)
			onu := Onu{
				ID:        onuId,
				PonPort:   pon,
				PonPortID: pon.ID,
				HwAddress: net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(pon.ID), byte(onuId)},
			}
			onu.SerialNumber = onu.NewSN(olt.ID, pon.ID, onu.ID)
			pon.Onus = append(pon.Onus, &onu)
		}
		olt.Pons = append(olt.Pons, &pon)
	}
	return olt
}

func Test_Olt_FindOnuBySn_Success(t *testing.T) {

	numPon := 4
	numOnu := 4

	olt := createMockOlt(numPon, numOnu)

	onu, err := olt.FindOnuBySn("BBSM00000303")

	assert.Equal(t, err, nil)
	assert.Equal(t, onu.Sn(), "BBSM00000303")
	assert.Equal(t, onu.ID, uint32(3))
	assert.Equal(t, onu.PonPortID, uint32(3))
}

func Test_Olt_FindOnuBySn_Error(t *testing.T) {

	numPon := 1
	numOnu := 4

	olt := createMockOlt(numPon, numOnu)

	_, err := olt.FindOnuBySn("BBSM00000303")

	assert.Equal(t, err.Error(), "cannot-find-onu-by-serial-number-BBSM00000303")
}

func Test_Olt_FindOnuByMacAddress_Success(t *testing.T) {

	numPon := 4
	numOnu := 4

	olt := createMockOlt(numPon, numOnu)

	mac := net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(3), byte(3)}

	onu, err := olt.FindOnuByMacAddress(mac)

	assert.Equal(t, err, nil)
	assert.Equal(t, onu.Sn(), "BBSM00000303")
	assert.Equal(t, onu.ID, uint32(3))
	assert.Equal(t, onu.PonPortID, uint32(3))
}

func Test_Olt_FindOnuByMacAddress_Error(t *testing.T) {

	numPon := 1
	numOnu := 4

	olt := createMockOlt(numPon, numOnu)

	mac := net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(3), byte(3)}

	_, err := olt.FindOnuByMacAddress(mac)

	assert.Equal(t, err.Error(), "cannot-find-onu-by-mac-address-2e:60:70:13:03:03")
}
