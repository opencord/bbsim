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
	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
)

const (
	DEFAULT_ONU_DEVICE_HEADER_FORMAT = "table{{ .PonPortID }}\t{{ .ID }}\t{{ .SerialNumber }}\t{{ .STag }}\t{{ .CTag }}\t{{ .OperState }}\t{{ .InternalState }}"
)

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
	tableFormat := format.Format(DEFAULT_ONU_DEVICE_HEADER_FORMAT)
	if err := tableFormat.Execute(os.Stdout, true, onus.Items); err != nil {
		log.Fatalf("Error while formatting ONUs table: %s", err)
	}

	return nil
}
