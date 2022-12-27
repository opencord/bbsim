/*
 * Copyright 2019-2023 Open Networking Foundation (ONF) and the ONF Contributors
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
	"os"

	"github.com/jessevdk/go-flags"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsimctl/config"
	"github.com/opencord/cordctl/pkg/format"
	log "github.com/sirupsen/logrus"
)

const (
	DEFAULT_SERVICE_HEADER_FORMAT = "table{{ .OnuSn }}\t{{ .UniId }}\t{{ .InternalState }}\t{{ .Name }}\t{{ .HwAddress }}\t{{ .STag }}\t{{ .UsSTagPriority }}\t{{ .DsSTagPriority }}\t{{ .CTag }}\t{{ .UsCTagPriority }}\t{{ .DsCTagPriority }}\t{{ .UniTagMatch }}\t{{ .NeedsEapol }}\t{{ .NeedsDhcp }}\t{{ .NeedsIgmp }}\t{{ .NeedsPPPoE }}\t{{ .ConfigureMacAddress }}\t{{ .EnableMacLearning }}\t{{ .GemPort }}\t{{ .EapolState }}\t{{ .DhcpState }}\t{{ .IGMPState }}"
)

type ServiceList struct{}

type ServiceOptions struct {
	List ServiceList `command:"list"`
}

func RegisterServiceCommands(parser *flags.Parser) {
	_, _ = parser.AddCommand("service", "Service Commands", "Commands to interact with ONU Services", &ServiceOptions{})
}

func getServices(OnuSn string, UniID string) (*pb.Services, error) {

	client, conn := connect()
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
	defer cancel()

	req := pb.UNIRequest{
		OnuSerialNumber: OnuSn,
		UniID:           UniID,
	}

	services, err := client.GetServices(ctx, &req)

	return services, err
}

func (options *ServiceList) Execute(args []string) error {
	services, err := getServices("", "")

	if err != nil {
		log.Fatalf("could not get OLT: %v", err)
		return nil
	}
	// print out
	tableFormat := format.Format(DEFAULT_SERVICE_HEADER_FORMAT)
	if err := tableFormat.Execute(os.Stdout, true, services.Items); err != nil {
		log.Fatalf("Error while formatting ONUs table: %s", err)
	}

	return nil
}
