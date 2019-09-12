package commands

import (
	"github.com/jessevdk/go-flags"
	"context"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"google.golang.org/grpc"
	pb "github.com/opencord/bbsim/api/bbsim"
	log "github.com/sirupsen/logrus"
)

type OltStatus struct {}

type oltOptions struct {
	Status    OltStatus     `command:"status"`
}

func RegisterOltCommands(parser *flags.Parser) {
	parser.AddCommand("olt", "OLT Commands", "Commands to query and manipulate the OLT device", &oltOptions{})
}

func (o *OltStatus) Execute(args []string) error {
	conn, err := grpc.Dial(config.GlobalConfig.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBBSimClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	r, err := c.GetOlt(ctx, &pb.Empty{})
	if err != nil {
		log.Errorf("could not get OLT states: %v", err)
		return err
	}
	log.Printf("OLT: %s", r.InternalState)
	return nil
}
