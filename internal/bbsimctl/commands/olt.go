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

const (
	DEFAULT_OLT_DEVICE_HEADER = "ID\tSerialNumber\tOperState\tInternalState\tNNIPorts\tPONPorts"
	DEFAULT_PORT_HEADER       = "ID\tOperState"
)

type OltGet struct{}

type OltNNIs struct{}

type OltPONs struct{}

// TODO add autocomplete
type oltOptions struct {
	Get OltGet  `command:"get"`
	NNI OltNNIs `command:"nnis"`
	PON OltPONs `command:"pons"`
}

func RegisterOltCommands(parser *flags.Parser) {
	parser.AddCommand("olt", "OLT Commands", "Commands to query and manipulate the OLT device", &oltOptions{})
}

func getOLT() *pb.Olt {
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
	olt, err := c.GetOlt(ctx, &pb.Empty{})
	if err != nil {
		log.Fatal("could not get OLT: %v", err)
		return nil
	}
	return olt
}

func printOltHeader(prefix string, o *pb.Olt) {
	fmt.Println(fmt.Sprintf("%s : %s", prefix, o.SerialNumber))
	fmt.Println()
}

// TODO use voltctl or cordctl parser to print tables (this needs to be moved out of the internals package)
func (o *OltGet) Execute(args []string) error {
	olt := getOLT()

	// print out
	w := GetWriter()
	fmt.Fprintln(w, DEFAULT_OLT_DEVICE_HEADER)
	formattedData := fmt.Sprintf("%d\t%s\t%s\t%s\t%d\t%d", olt.ID, olt.SerialNumber, olt.OperState, olt.InternalState, len(olt.NNIPorts), len(olt.PONPorts))
	fmt.Fprintln(w, formattedData)
	fmt.Fprintln(w)
	w.Flush()

	return nil
}

func (o *OltNNIs) Execute(args []string) error {
	olt := getOLT()

	printOltHeader("NNI Ports for", olt)

	w := GetWriter()
	fmt.Fprintln(w, DEFAULT_PORT_HEADER)

	for _, nni := range olt.NNIPorts {
		fmt.Fprintln(w, fmt.Sprintf("%d\t%s", nni.ID, nni.OperState))
	}
	w.Flush()

	return nil
}

func (o *OltPONs) Execute(args []string) error {
	olt := getOLT()

	printOltHeader("PON Ports for", olt)

	w := GetWriter()
	fmt.Fprintln(w, DEFAULT_PORT_HEADER)

	for _, nni := range olt.PONPorts {
		fmt.Fprintln(w, fmt.Sprintf("%d\t%s", nni.ID, nni.OperState))
	}
	w.Flush()

	return nil
}
