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

var (
	version 	string
	buildTime  	string
	commitHash 	string
	gitStatus	string
)

type BBSimServer struct {
}

func (s BBSimServer) Version(ctx context.Context, req *bbsim.Empty) (*bbsim.VersionNumber, error)  {
	// TODO add a flag to specofy whether the tree was clean at this commit or not
	return &bbsim.VersionNumber{
		Version: version,
		BuildTime: buildTime,
		CommitHash: commitHash,
		GitStatus: gitStatus,
	}, nil
}

func (s BBSimServer) GetOlt(ctx context.Context, req *bbsim.Empty) (*bbsim.Olt, error) {
	olt := devices.GetOLT()
	nnis := []*bbsim.NNIPort{}
	pons := []*bbsim.PONPort{}

	for _, nni := range olt.Nnis {
		n := bbsim.NNIPort{
			ID: int32(nni.ID),
			OperState: fmt.Sprintf("%s", nni.OperState),
		}
		nnis = append(nnis, &n)
	}

	for _, pon := range olt.Pons {
		p := bbsim.PONPort{
			ID: int32(pon.ID),
			OperState: fmt.Sprintf("%s", pon.OperState),
		}
		pons = append(pons, &p)
	}

	res := bbsim.Olt{
		ID: int32(olt.ID),
		OperState: fmt.Sprintf("%s", olt.OperState),
		NNIPorts: nnis,
		PONPorts: pons,
	}
	return &res, nil
}