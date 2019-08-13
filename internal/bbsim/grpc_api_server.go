package main

import (
	"context"
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
			OperState: nni.OperState.Current(),
		}
		nnis = append(nnis, &n)
	}

	for _, pon := range olt.Pons {
		p := bbsim.PONPort{
			ID: int32(pon.ID),
			OperState: pon.OperState.Current(),
		}
		pons = append(pons, &p)
	}

	res := bbsim.Olt{
		ID: int32(olt.ID),
		OperState: olt.OperState.Current(),
		InternalState: olt.InternalState.Current(),
		NNIPorts: nnis,
		PONPorts: pons,
	}
	return &res, nil
}

func (s BBSimServer) GetONUs(ctx context.Context, req *bbsim.Empty) (*bbsim.ONUs, error){
	olt := devices.GetOLT()
	onus := bbsim.ONUs{
		Items: []*bbsim.ONU{},
	}

	for _, pon := range olt.Pons {
		for _, o := range pon.Onus {
			onu := bbsim.ONU{
				ID: int32(o.ID),
				SerialNumber: o.SerialNumber.String(),
				OperState: o.OperState.Current(),
				InternalState: o.InternalState.Current(),
				PonPortID: int32(o.PonPortID),
			}
			onus.Items = append(onus.Items, &onu)
		}
	}
	return &onus, nil
}