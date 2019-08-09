package devices

import (
	"gerrit.opencord.org/bbsim/api/openolt"
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
)

func CreateONU(olt OltDevice, pon PonPort, id uint32) Onu {
		o := Onu{
			ID: id,
			OperState: DOWN,
			PonPortID: pon.ID,
			PonPort: pon,
		}
		o.SerialNumber = o.NewSN(olt.ID, pon.ID, o.ID)

		o.InternalState = fsm.NewFSM(
			"created",
			fsm.Events{
				{Name: "discover", Src: []string{"created"}, Dst: "discovered"},
				{Name: "enable", Src: []string{"discovered"}, Dst: "enabled"},
			},
			fsm.Callbacks{
				"enter_state": func(e *fsm.Event) {
					olt.stateChange(e)
				},
			},
		)
		return o
}

func (o Onu) stateChange(e *fsm.Event) {
	log.WithFields(log.Fields{
		"onuID": o.ID,
		"onuSN": o.SerialNumber,
		"dstState": e.Dst,
		"srcState": e.Src,
	}).Debugf("ONU state has changed")
}

func (o Onu) NewSN(oltid int, intfid uint32, onuid uint32) *openolt.SerialNumber {

	sn := new(openolt.SerialNumber)

	sn = new(openolt.SerialNumber)
	sn.VendorId = []byte("BBSM")
	sn.VendorSpecific = []byte{0, byte(oltid % 256), byte(intfid), byte(onuid)}

	return sn
}

func (o Onu) sendOnuDiscIndication(msg OnuDiscIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	discoverData := &openolt.Indication_OnuDiscInd{OnuDiscInd: &openolt.OnuDiscIndication{
		IntfId: msg.Onu.PonPortID,
		SerialNumber: msg.Onu.SerialNumber,
	}}
	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		log.Error("Failed to send Indication_OnuDiscInd: %v", err)
	}
	log.WithFields(log.Fields{
		"IntfId": msg.Onu.PonPortID,
		"SerialNumber": msg.Onu.SerialNumber,
	}).Debug("Sent Indication_OnuDiscInd")
}

func (o Onu) sendOnuIndication(msg OnuIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	// NOTE voltha returns an ID, but if we use that ID then it complains:
	// expected_onu_id: 1, received_onu_id: 1024, event: ONU-id-mismatch, can happen if both voltha and the olt rebooted
	// so we're using the internal ID that is 1
	// o.ID = msg.OnuID
	o.OperState = msg.OperState

	indData := &openolt.Indication_OnuInd{OnuInd: &openolt.OnuIndication{
		IntfId: o.PonPortID,
		OnuId: o.ID,
		OperState: o.OperState.String(),
		AdminState: o.OperState.String(),
		SerialNumber: o.SerialNumber,
	}}
	if err := stream.Send(&openolt.Indication{Data: indData}); err != nil {
		log.Error("Failed to send Indication_OnuInd: %v", err)
	}
	log.WithFields(log.Fields{
		"IntfId": o.PonPortID,
		"OnuId": o.ID,
		"OperState": msg.OperState.String(),
		"AdminState": msg.OperState.String(),
		"SerialNumber": o.SerialNumber,
	}).Debug("Sent Indication_OnuInd")
}