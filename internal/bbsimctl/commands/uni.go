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
	"strings"

	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
	log "github.com/sirupsen/logrus"
)

const (
	DEFAULT_UNI_HEADER_FORMAT               = "table{{ .OnuSn }}\t{{ .OnuID }}\t{{ .ID }}\t{{ .MeID }}\t{{ .PortNo }}\t{{ .OperState }}\t{{ .Type }}"
	DEFAULT_UNI_HEADER_FORMAT_WITH_SERVICES = "table{{ .OnuSn }}\t{{ .OnuID }}\t{{ .ID }}\t{{ .MeID }}\t{{ .PortNo }}\t{{ .OperState }}\t{{ .Type }}\t{{ .Services }}"
)

type UniIdInt string

type UNIList struct {
	Verbose bool `short:"v" long:"verbose" description:"Print all the Unis from all the ONUs"`
}

type UNIGet struct {
	Verbose bool `short:"v" long:"verbose" description:"Print all the informations we have about UNIs"`
	Args    struct {
		OnuSn OnuSnString
	} `positional-args:"yes" required:"yes"`
}

type UNIServices struct {
	Verbose bool `short:"v" long:"verbose" description:"Print all the Services from the specified UNI"`
	Args    struct {
		OnuSn OnuSnString
		UniId UniIdInt
	} `positional-args:"yes"`
}

type UNIEapolRestart struct {
	Args struct {
		OnuSn OnuSnString
		UniId UniIdInt
	} `positional-args:"yes" required:"yes"`
}

type UNIDhcpRestart struct {
	Args struct {
		OnuSn OnuSnString
		UniId UniIdInt
	} `positional-args:"yes" required:"yes"`
}

type UNIOptions struct {
	List         UNIList         `command:"list"`
	Get          UNIGet          `command:"get"`
	Services     UNIServices     `command:"services"`
	RestartEapol UNIEapolRestart `command:"auth_restart"`
	RestartDchp  UNIDhcpRestart  `command:"dhcp_restart"`
}

func RegisterUNICommands(parser *flags.Parser) {
	_, _ = parser.AddCommand("uni", "UNI Commands", "Commands to query and manipulate UNIs", &UNIOptions{})
}

func (options *UNIList) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	unis, err := client.GetUnis(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get UNIs: %v", err)
		return err
	}

	var tableFormat format.Format
	if options.Verbose {
		tableFormat = format.Format(DEFAULT_UNI_HEADER_FORMAT_WITH_SERVICES)
	} else {
		tableFormat = format.Format(DEFAULT_UNI_HEADER_FORMAT)
	}
	if err := tableFormat.Execute(os.Stdout, true, unis.Items); err != nil {
		log.Fatalf("Error while formatting Unis table: %s", err)
	}

	return nil
}

func (options *UNIGet) Execute(args []string) error {

	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()
	req := pb.ONURequest{
		SerialNumber: string(options.Args.OnuSn),
	}
	res, err := client.GetOnuUnis(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot not get unis for ONU %s: %v", options.Args.OnuSn, err)
		return err
	}

	var tableFormat format.Format
	if options.Verbose {
		tableFormat = format.Format(DEFAULT_UNI_HEADER_FORMAT_WITH_SERVICES)
	} else {
		tableFormat = format.Format(DEFAULT_UNI_HEADER_FORMAT)
	}
	if err := tableFormat.Execute(os.Stdout, true, res.Items); err != nil {
		log.Fatalf("Error while formatting Unis table: %s", err)
	}

	return nil
}

//Get Services for specified UNI
//First get UNIs from specified ONU SN
//Then get Services from appropriate UNI
func (options *UNIServices) Execute(args []string) error {
	services, err := getServices(string(options.Args.OnuSn), string(options.Args.UniId))

	if err != nil {
		log.Fatalf("Cannot not get services for ONU %s and UNI %s: %v", options.Args.OnuSn, options.Args.UniId, err)
		return err
	}

	tableFormat := format.Format(DEFAULT_SERVICE_HEADER_FORMAT)
	if err := tableFormat.Execute(os.Stdout, true, services.Items); err != nil {
		log.Fatalf("Error while formatting ONUs table: %s", err)
	}

	return nil
}

func (options *UNIEapolRestart) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.UNIRequest{
		OnuSerialNumber: string(options.Args.OnuSn),
		UniID:           string(options.Args.UniId),
	}
	res, err := client.RestartEapol(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot restart EAPOL for ONU %s and UNI %s: %v", options.Args.OnuSn, options.Args.UniId, err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))

	return nil
}

func (options *UNIDhcpRestart) Execute(args []string) error {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.UNIRequest{
		OnuSerialNumber: string(options.Args.OnuSn),
		UniID:           string(options.Args.UniId),
	}
	res, err := client.RestartDhcp(ctx, &req)

	if err != nil {
		log.Fatalf("Cannot restart DHCP for ONU %s and UNI %s: %v", options.Args.OnuSn, options.Args.UniId, err)
		return err
	}

	fmt.Println(fmt.Sprintf("[Status: %d] %s", res.StatusCode, res.Message))

	return nil
}

func (uniId *UniIdInt) Complete(match string) []flags.Completion {
	client, conn := connect()
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	onus, err := client.GetONUs(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get ONUs: %v", err)
		return nil
	}

	// go-flag doesn't allow us to read the previous parameters to the command so we can't get the ONU Serial Number,
	// but since all the ONUs have the same number of UNIs thus we can re-use the UNIs belonging to the first ONU in the list
	// pending issue here https://github.com/jessevdk/go-flags/issues/305
	unis := onus.Items[0].Unis

	list := make([]flags.Completion, 0)
	for _, uni := range unis {
		strID := strconv.Itoa(int(uni.ID))
		if strings.HasPrefix(strID, match) {
			list = append(list, flags.Completion{Item: strID})
		}
	}

	return list
}
