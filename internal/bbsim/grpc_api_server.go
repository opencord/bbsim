package main

import (
	"context"
	"fmt"
	"gerrit.opencord.org/bbsim/api/bbsim"
	"gerrit.opencord.org/bbsim/internal/bbsim/devices"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{
	"module": "GrpcApiServer",
})

type BBSimServer struct {
}

func (s BBSimServer) GetOlt(ctx context.Context, req *bbsim.Empty) (*bbsim.Olt, error) {
	olt := devices.GetOLT()
	nnis := []*bbsim.NNIPort{}
	pons := []*bbsim.PONPort{}
	res := bbsim.Olt{
		ID: int32(olt.ID),
		OperState: fmt.Sprintf("%s", olt.OperState),
		AdminState: fmt.Sprintf("%s", olt.AdminState),
		NNIPorts: nnis,
		PONPorts: pons,
	}
	logger.Debugf("Returning OLT: ", res)
	return &res, nil
}