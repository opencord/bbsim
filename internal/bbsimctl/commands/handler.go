/*
 * Copyright 2019-present Ciena Corporation
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
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
)

type RpcEventHandler struct {
	Response proto.Message
	Status   *status.Status
	Data     []byte
	Fields   map[string]map[string]interface{}
}

func (h *RpcEventHandler) OnResolveMethod(*desc.MethodDescriptor) {
}

func (h *RpcEventHandler) OnSendHeaders(metadata.MD) {
}

func (h *RpcEventHandler) OnReceiveHeaders(metadata.MD) {
}

func (h *RpcEventHandler) OnReceiveResponse(m proto.Message) {
	h.Response = m
}

func (h *RpcEventHandler) OnReceiveTrailers(s *status.Status, m metadata.MD) {
	h.Status = s
}

func (h *RpcEventHandler) GetParams(msg proto.Message) error {
	dmsg, err := dynamic.AsDynamicMessage(msg)
	if err != nil {
		return err
	}

	if h.Fields == nil || len(h.Fields) == 0 {
		return io.EOF
	}

	fields, ok := h.Fields[dmsg.XXX_MessageName()]
	if !ok {
		return nil
	}

	for k, v := range fields {
		dmsg.TrySetFieldByName(k, v)
	}
	delete(h.Fields, dmsg.XXX_MessageName())

	return nil
}
