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
	"os"
	"strconv"

	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	DEFAULT_OLT_DEVICE_HEADER_FORMAT = "table{{ .ID }}\t{{ .SerialNumber }}\t{{ .OperState }}\t{{ .InternalState }}"
	DEFAULT_PORT_HEADER_FORMAT       = "table{{ .ID }}\t{{ .OperState }}"
)

type OltGet struct{}

type OltNNIs struct{}

type OltPONs struct{}

type OltShutdown struct{}

type OltPoweron struct{}

type OltReboot struct{}

type OltFlows struct{}

type oltOptions struct {
	Get      OltGet          `command:"get"`
	NNI      OltNNIs         `command:"nnis"`
	PON      OltPONs         `command:"pons"`
	Shutdown OltShutdown     `command:"shutdown"`
	Poweron  OltPoweron      `command:"poweron"`
	Reboot   OltReboot       `command:"reboot"`
	Alarms   OltAlarmOptions `command:"alarms"`
	Flows    OltFlows        `command:"flows"`
}

func RegisterOltCommands(parser *flags.Parser) {
	parser.AddCommand("olt", "OLT Commands", "Commands to query and manipulate the OLT device", &oltOptions{})
}

func getOLT() *pb.Olt {
	conn, err := grpc.Dial(config.GlobalConfig.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil
	}
	defer conn.Close()
	c := pb.NewBBSimClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	olt, err := c.GetOlt(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get OLT: %v", err)
		return nil
	}
	return olt
}

func printOltHeader(prefix string, o *pb.Olt) {
	fmt.Println(fmt.Sprintf("%s : %s", prefix, o.SerialNumber))
	fmt.Println()
}

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

func (o *OltShutdown) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	res, err := client.ShutdownOlt(ctx, &pb.Empty{})

	if err != nil {
		log.Fatalf("Cannot shut down OLT: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

func (o *OltPoweron) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	res, err := client.PoweronOlt(ctx, &pb.Empty{})

	if err != nil {
		log.Fatalf("Cannot power on OLT: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

func (o *OltReboot) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	res, err := client.RebootOlt(ctx, &pb.Empty{})

	if err != nil {
		log.Fatalf("Cannot reboot OLT: %v", err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))
	return nil
}

func (o *OltFlows) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.ONURequest{}
	res, err := client.GetFlows(ctx, &req)
	if err != nil {
		log.Errorf("Cannot get flows for OLT: %v", err)
		return err
	}

	if res.Flows == nil {
		fmt.Println("OLT has no flows")
		return nil
	}

	flowHeader := []string{
		"access_intf_id",
		"onu_id",
		"uni_id",
		"flow_id",
		"flow_type",
		"eth_type",
		"alloc_id",
		"network_intf_id",
		"gemport_id",
		"classifier",
		"action",
		"priority",
		"cookie",
		"port_no",
	}

	tableFlow := tablewriter.NewWriter(os.Stdout)
	tableFlow.SetRowLine(true)
	fmt.Fprintf(os.Stdout, "OLT Flows:\n")
	tableFlow.SetHeader(flowHeader)

	for _, flow := range res.Flows {
		flowInfo := []string{}
		flowInfo = append(flowInfo,
			strconv.Itoa(int(flow.AccessIntfId)),
			strconv.Itoa(int(flow.OnuId)),
			strconv.Itoa(int(flow.UniId)),
			strconv.Itoa(int(flow.FlowId)),
			flow.FlowType,
			fmt.Sprintf("%x", flow.Classifier.EthType),
			strconv.Itoa(int(flow.AllocId)),
			strconv.Itoa(int(flow.NetworkIntfId)),
			strconv.Itoa(int(flow.GemportId)),
			flow.Classifier.String(),
			flow.Action.String(),
			strconv.Itoa(int(flow.Priority)),
			strconv.Itoa(int(flow.Cookie)),
			strconv.Itoa(int(flow.PortNo)),
		)
		tableFlow.Append(flowInfo)
	}
	tableFlow.Render()
	tableFlow.SetNewLine("")
	return nil
}
