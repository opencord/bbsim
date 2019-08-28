package eapol

import (
	bbsim "gerrit.opencord.org/bbsim/internal/bbsim/types"
	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/go/openolt"
	"google.golang.org/grpc"
	"gotest.tools/assert"
	"testing"
)

type fakeStream struct {
	calledSend int
	grpc.ServerStream
}

func (s fakeStream) Send(flow *openolt.Indication) error {
	s.calledSend++
	return nil
}

func TestCreateWPASupplicant(t *testing.T) {
	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber *openolt.SerialNumber = new(openolt.SerialNumber)

	//sn = new(openolt.SerialNumber)
	serialNumber.VendorId = []byte("BBSM")
	serialNumber.VendorSpecific = []byte{0, byte(0 % 256), byte(ponPortId), byte(onuId)}

	eapolStateMachine := fsm.NewFSM(
	"start_auth",
	fsm.Events{
		{Name: "eap_start_sent", Src: []string{"auth_started"}, Dst: "eap_start_sent"},
		{Name: "eap_resonse_identity_sent", Src: []string{"eap_start_sent"}, Dst: "eap_resonse_identity_sent"},
		{Name: "eap_resonse_challenge_sent", Src: []string{"eap_resonse_identity_sent"}, Dst: "eap_resonse_challenge_sent"},
		{Name: "eap_resonse_success_received", Src: []string{"eap_resonse_challenge_sent"}, Dst: "eap_resonse_success_received"},
	},
	fsm.Callbacks{},
	)

	pktOutCh := make(chan *bbsim.ByteMsg, 1024)

	stream := fakeStream{}


	// FIXME we don't care that the stream is called, but only that sendEapStart is called
	// and that the status is transitioned to eap_start_sent

	// NOTE the test is stuck waiting for EAPCodeSuccess to stop the thread
	// close the channel to terminate the thread
	CreateWPASupplicant(onuId, ponPortId, serialNumber, eapolStateMachine, stream, pktOutCh)

	assert.Equal(t, stream.calledSend, 1)
}