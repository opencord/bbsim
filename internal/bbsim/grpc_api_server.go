package main

import (
	"context"
	"gerrit.opencord.org/bbsim/api/bbsim"
)

type BBSimServer struct {
}

func (s BBSimServer) ListOlts(ctx context.Context, req *bbsim.Empty) (*bbsim.Olts, error) {
	olts := bbsim.Olts{}
	return &olts, nil
}