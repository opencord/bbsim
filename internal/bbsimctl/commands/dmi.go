/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

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
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/opencord/bbsim/api/bbsim"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	DEFAULT_TRANSCEIVER_HEADER_FORMAT = "table{{ .ID }}\t{{ .UUID }}\t{{ .Name }}\t{{ .Technology }}\t{{ .PluggedIn }}\t{{ .PonIds }}"
)

type DMIOptions struct {
	Events      DmiEventOptions       `command:"events"`
	Transceiver DmiTransceiverOptions `command:"transceiver"`
}

type DmiEventOptions struct {
	Create DmiEventCreate `command:"create"`
}

type DmiEventCreate struct {
	Args struct {
		Name string
	} `positional-args:"yes" required:"yes"`
}

type DmiTransceiverOptions struct {
	PlugIn  DmiTransceiverPlugIn  `command:"plug_in"`
	PlugOut DmiTransceiverPlugOut `command:"plug_out"`
	List    DmiTransceiversList   `command:"list"`
}

type DmiTransceiversList struct {
}

type DmiTransceiverPlugIn struct {
	Args struct {
		TransceiverId uint32
	} `positional-args:"yes" required:"yes"`
}

type DmiTransceiverPlugOut struct {
	Args struct {
		TransceiverId uint32
	} `positional-args:"yes" required:"yes"`
}

func RegisterDMICommands(parser *flags.Parser) {
	_, _ = parser.AddCommand("dmi", "DMI Commands", "Commands to create events", &DMIOptions{})
}

func dmiEventGrpcClient() (bbsim.BBsimDmiClient, *grpc.ClientConn) {
	conn, err := grpc.Dial(config.DmiConfig.Server, grpc.WithInsecure())
	if err != nil {
		log.Errorf("BBsimDmiClient connection failed  : %v", err)
		return nil, conn
	}
	return bbsim.NewBBsimDmiClient(conn), conn
}

// Execute create event
func (o *DmiEventCreate) Execute(args []string) error {
	client, conn := dmiEventGrpcClient()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := bbsim.DmiEvent{EventName: o.Args.Name}
	res, err := client.CreateEvent(ctx, &req)
	if err != nil {
		log.Errorf("Cannot create DMI event: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

//Print a list of the transceivers and their state
func (pon *DmiTransceiversList) Execute(args []string) error {
	client, conn := dmiEventGrpcClient()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	transceivers, err := client.GetTransceivers(ctx, &pb.DmiEmpty{})
	if err != nil {
		log.Errorf("Cannot get transceivers list: %v", err)
		return err
	}

	// print out
	tableFormat := format.Format(DEFAULT_TRANSCEIVER_HEADER_FORMAT)

	if err := tableFormat.Execute(os.Stdout, true, transceivers.Items); err != nil {
		log.Fatalf("Error while formatting transceivers table: %s", err)
	}

	return nil
}

//Plug in the specified transceiver
func (pon *DmiTransceiverPlugIn) Execute(args []string) error {
	client, conn := dmiEventGrpcClient()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.TransceiverRequest{
		TransceiverId: uint32(pon.Args.TransceiverId),
	}

	res, err := client.PlugInTransceiver(ctx, &req)
	if err != nil {
		log.Errorf("Cannot plug in PON transceiver: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

//Plug out the specified transceiver
func (pon *DmiTransceiverPlugOut) Execute(args []string) error {
	client, conn := dmiEventGrpcClient()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.TransceiverRequest{
		TransceiverId: uint32(pon.Args.TransceiverId),
	}

	res, err := client.PlugOutTransceiver(ctx, &req)
	if err != nil {
		log.Errorf("Cannot plug out PON transceiver: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}
