/*
 * Portions copyright 2019-present Open Networking Foundation
 * Original copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
)

const (
	DEFAULT_OLT_DEVICE_HEADER_FORMAT = "table{{ .ID }}\t{{ .SerialNumber }}\t{{ .OperState }}\t{{ .InternalState }}"
	DEFAULT_PORT_HEADER_FORMAT       = "table{{ .ID }}\t{{ .OperState }}"
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
	tableFormat := format.Format(DEFAULT_OLT_DEVICE_HEADER_FORMAT)
	tableFormat.Execute(os.Stdout, true, olt)

	return nil
}

func (o *OltNNIs) Execute(args []string) error {
	olt := getOLT()

	printOltHeader("NNI Ports for", olt)

	tableFormat := format.Format(DEFAULT_PORT_HEADER_FORMAT)
	tableFormat.Execute(os.Stdout, true, olt.NNIPorts)

	return nil
}

func (o *OltPONs) Execute(args []string) error {
	olt := getOLT()

	printOltHeader("PON Ports for", olt)

	tableFormat := format.Format(DEFAULT_PORT_HEADER_FORMAT)
	tableFormat.Execute(os.Stdout, true, olt.PONPorts)

	return nil
}
