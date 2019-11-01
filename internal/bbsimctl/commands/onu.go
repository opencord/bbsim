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
	"strings"

	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	DEFAULT_ONU_DEVICE_HEADER_FORMAT = "table{{ .PonPortID }}\t{{ .ID }}\t{{ .PortNo }}\t{{ .SerialNumber }}\t{{ .HwAddress }}\t{{ .STag }}\t{{ .CTag }}\t{{ .OperState }}\t{{ .InternalState }}"
)

type OnuSnString string
type IgmpSubAction string

const IgmpJoinKey string = "join"
const IgmpLeaveKey string = "leave"

type ONUList struct{}

type ONUGet struct {
	Args struct {
		OnuSn OnuSnString
	} `positional-args:"yes" required:"yes"`
}

type ONUShutDown struct {
	Args struct {
		OnuSn OnuSnString
	} `positional-args:"yes" required:"yes"`
}

type ONUPowerOn struct {
	Args struct {
		OnuSn OnuSnString
	} `positional-args:"yes" required:"yes"`
}

type ONUEapolRestart struct {
	Args struct {
		OnuSn OnuSnString
	} `positional-args:"yes" required:"yes"`
}

type ONUDhcpRestart struct {
	Args struct {
		OnuSn OnuSnString
	} `positional-args:"yes" required:"yes"`
}

type ONUIgmp struct {
	Args struct {
		OnuSn OnuSnString
		SubAction IgmpSubAction
	} `positional-args:"yes" required:"yes"`
}

type ONUOptions struct {
	List         ONUList         `command:"list"`
	Get          ONUGet          `command:"get"`
	ShutDown     ONUShutDown     `command:"shutdown"`
	PowerOn      ONUPowerOn      `command:"poweron"`
	RestartEapol ONUEapolRestart `command:"auth_restart"`
	RestartDchp  ONUDhcpRestart  `command:"dhcp_restart"`
	Igmp         ONUIgmp         `command:"igmp"`
}

func RegisterONUCommands(parser *flags.Parser) {
	parser.AddCommand("onu", "ONU Commands", "Commands to query and manipulate ONU devices", &ONUOptions{})
}

func connect() (pb.BBSimClient, *grpc.ClientConn) {
	conn, err := grpc.Dial(config.GlobalConfig.Server, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, conn
	}
	return pb.NewBBSimClient(conn), conn
}

func getONUs() *pb.ONUs {

	client, conn := connect()
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	onus, err := client.GetONUs(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get OLT: %v", err)
		return nil
	}
	return onus
}

func (options *ONUList) Execute(args []string) error {
	onus := getONUs()

	// print out
	tableFormat := format.Format(DEFAULT_ONU_DEVICE_HEADER_FORMAT)
	if err := tableFormat.Execute(os.Stdout, true, onus.Items); err != nil {
		log.Fatalf("Error while formatting ONUs table: %s", err)
	}

	return nil
}

func (options *ONUGet) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	req := pb.ONURequest{
		SerialNumber: string(options.Args.OnuSn),
	}
	res, err := client.GetONU(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot not shutdown ONU %s: %v", options.Args.OnuSn, err)
		return err
	}

	tableFormat := format.Format(DEFAULT_ONU_DEVICE_HEADER_FORMAT)
	if err := tableFormat.Execute(os.Stdout, true, []*pb.ONU{res}); err != nil {
		log.Fatalf("Error while formatting ONUs table: %s", err)
	}

	return nil
}

func (options *ONUShutDown) Execute(args []string) error {

	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	req := pb.ONURequest{
		SerialNumber: string(options.Args.OnuSn),
	}
	res, err := client.ShutdownONU(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot shutdown ONU %s: %v", options.Args.OnuSn, err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))

	return nil
}

func (options *ONUPowerOn) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	req := pb.ONURequest{
		SerialNumber: string(options.Args.OnuSn),
	}
	res, err := client.PoweronONU(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot power on ONU %s: %v", options.Args.OnuSn, err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))

	return nil
}

func (options *ONUEapolRestart) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	req := pb.ONURequest{
		SerialNumber: string(options.Args.OnuSn),
	}
	res, err := client.RestartEapol(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot restart EAPOL for ONU %s: %v", options.Args.OnuSn, err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))

	return nil
}

func (options *ONUDhcpRestart) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	req := pb.ONURequest{
		SerialNumber: string(options.Args.OnuSn),
	}
	res, err := client.RestartDhcp(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot restart DHCP for ONU %s: %v", options.Args.OnuSn, err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))

	return nil
}

func (options *ONUIgmp) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.ONURequest{
		SerialNumber: string(options.Args.OnuSn),
	}

	var subActionVal pb.SubActionTypes
	if string(options.Args.SubAction) == IgmpJoinKey {
		subActionVal = pb.SubActionTypes_JOIN
	} else if string(options.Args.SubAction) == IgmpLeaveKey {
		subActionVal = pb.SubActionTypes_LEAVE
	}

	igmpReq := pb.IgmpRequest{
		OnuReq:       &req,
		SubActionVal: subActionVal,
	}
	res, err := client.GetONU(ctx, igmpReq.OnuReq)
	if err != nil {
		log.WithFields(log.Fields{
			"SerialNumber": options.Args.OnuSn,
		}).Fatalf("Cannot not get details on ONU error: %v", err)
		return err
	}
	log.WithFields(log.Fields{
    		"SerialNumber": igmpReq.OnuReq.SerialNumber,
	}).Debugf("ONU has indentified : %s", res)

	igmpRes, igmpErr := client.ChangeIgmpState(ctx, &igmpReq)
	if igmpErr != nil {
		log.WithFields(log.Fields{
                        "SubAction": options.Args.SubAction,
                }).Fatalf("Could not process Action: error: %v", igmpErr)
		return igmpErr
	} else {
		fmt.Sprintf("igmp state has been changed with response: %s",
			igmpRes.Message)
	}

	return nil
}

func (onuSn *OnuSnString) Complete(match string) []flags.Completion {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	onus, err := client.GetONUs(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get ONUs: %v", err)
		return nil
	}

	list := make([]flags.Completion, 0)
	for _, k := range onus.Items {
		if strings.HasPrefix(k.SerialNumber, match) {
			list = append(list, flags.Completion{Item: k.SerialNumber})
		}
	}

	return list
}
