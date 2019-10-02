/*
 * Copyright 2018-present Open Networking Foundation

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

package bbr

import (
	"fmt"
	bbsimPb "github.com/opencord/bbsim/api/bbsim"
	pb "github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Connect(ip string, port string) (pb.OpenoltClient, *grpc.ClientConn) {
	server := fmt.Sprintf("%s:%s", ip, port)
	conn, err := grpc.Dial(server, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, conn
	}
	return pb.NewOpenoltClient(conn), conn
}

func ApiConnect(ip string, port string) (bbsimPb.BBSimClient, *grpc.ClientConn) {
	server := fmt.Sprintf("%s:%s", ip, port)
	conn, err := grpc.Dial(server, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
		return nil, conn
	}
	return bbsimPb.NewBBSimClient(conn), conn
}

//func getONUs() *pb.ONUs {
//
//	client, conn := connect()
//	defer conn.Close()
//
//	// Contact the server and print out its response.
//	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Grpc.Timeout)
//	defer cancel()
//
//	onus, err := client.GetONUs(ctx, &pb.Empty{})
//	if err != nil {
//		log.Fatalf("could not get OLT: %v", err)
//		return nil
//	}
//	return onus
//}
