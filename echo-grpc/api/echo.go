// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative echo.proto

package api

import (
	"context"
	"github.com/googlecloudplatform/grpc-gke-nlb-tutorial/reverse-grpc/api"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Server for the Echo gRPC API
type Server struct {
	UnimplementedEchoServer
	ReverseAddress string
}

// Echo the content of the request
func (s *Server) Echo(ctx context.Context, in *EchoRequest) (*EchoResponse, error) {
	reverse := in.GetReverse()
	sleep := in.GetSleep()
	content := in.GetContent()

	log.Printf("Handling Echo request [%v] with context %v, content %v, sleep %v, reverse: %v",
		in, ctx, content, sleep, reverse)
	if sleep > 0 {
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	if reverse == true {
		conn, err := grpc.Dial(s.ReverseAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to address %v", s.ReverseAddress)
		}
		defer conn.Close()
		c := api.NewReverseClient(conn)
		r, err := c.Reverse(ctx, &api.ReverseRequest{Content: content})
		if err != nil {
			log.Fatalf("could not reverse: %v", err)
		}
		content = r.GetContent()
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Unable to get hostname %v", err)
		hostname = ""
	}
	grpc.SendHeader(ctx, metadata.Pairs("hostname", hostname))
	return &EchoResponse{Content: content}, nil
}
