package commands

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const DEFAULT_ONU_DEVICE_HEADER = "ID\tSerialNumber\tOperState\tInternalState\tPONPortID"

type ONUOptions struct{}

func RegisterONUCommands(parser *flags.Parser) {
	parser.AddCommand("onus", "List ONU Devices", "Commands to list the ONU devices and their internal state", &ONUOptions{})
}

func getONUs() *pb.ONUs {
	conn, err := grpc.Dial(config.GlobalConfig.Server, grpc.WithInsecure())

	if err != nil {
		log.Fatal("did not connect: %v", err)
		return nil
	}
	defer conn.Close()
	c := pb.NewBBSimClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	onus, err := c.GetONUs(ctx, &pb.Empty{})
	if err != nil {
		log.Fatal("could not get OLT: %v", err)
		return nil
	}
	return onus
}

func (options *ONUOptions) Execute(args []string) error {
	onus := getONUs()

	// print out
	w := GetWriter()

	fmt.Fprintln(w, DEFAULT_ONU_DEVICE_HEADER)

	for _, onu := range onus.Items {
		fmt.Fprintln(w, fmt.Sprintf("%d\t%s\t%s\t%s\t%d", onu.ID, onu.SerialNumber, onu.OperState, onu.InternalState, onu.PonPortID))
	}
	w.Flush()

	return nil
}
