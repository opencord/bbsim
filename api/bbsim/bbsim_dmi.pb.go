// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/bbsim/bbsim_dmi.proto

package bbsim

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type DmiEvent struct {
	EventName            string   `protobuf:"bytes,1,opt,name=event_name,json=eventName,proto3" json:"event_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DmiEvent) Reset()         { *m = DmiEvent{} }
func (m *DmiEvent) String() string { return proto.CompactTextString(m) }
func (*DmiEvent) ProtoMessage()    {}
func (*DmiEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_49e784b4938902cc, []int{0}
}

func (m *DmiEvent) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DmiEvent.Unmarshal(m, b)
}
func (m *DmiEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DmiEvent.Marshal(b, m, deterministic)
}
func (m *DmiEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DmiEvent.Merge(m, src)
}
func (m *DmiEvent) XXX_Size() int {
	return xxx_messageInfo_DmiEvent.Size(m)
}
func (m *DmiEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_DmiEvent.DiscardUnknown(m)
}

var xxx_messageInfo_DmiEvent proto.InternalMessageInfo

func (m *DmiEvent) GetEventName() string {
	if m != nil {
		return m.EventName
	}
	return ""
}

type DmiResponse struct {
	StatusCode           int32    `protobuf:"varint,1,opt,name=status_code,json=statusCode,proto3" json:"status_code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DmiResponse) Reset()         { *m = DmiResponse{} }
func (m *DmiResponse) String() string { return proto.CompactTextString(m) }
func (*DmiResponse) ProtoMessage()    {}
func (*DmiResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_49e784b4938902cc, []int{1}
}

func (m *DmiResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DmiResponse.Unmarshal(m, b)
}
func (m *DmiResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DmiResponse.Marshal(b, m, deterministic)
}
func (m *DmiResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DmiResponse.Merge(m, src)
}
func (m *DmiResponse) XXX_Size() int {
	return xxx_messageInfo_DmiResponse.Size(m)
}
func (m *DmiResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DmiResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DmiResponse proto.InternalMessageInfo

func (m *DmiResponse) GetStatusCode() int32 {
	if m != nil {
		return m.StatusCode
	}
	return 0
}

func (m *DmiResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type DmiEmpty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DmiEmpty) Reset()         { *m = DmiEmpty{} }
func (m *DmiEmpty) String() string { return proto.CompactTextString(m) }
func (*DmiEmpty) ProtoMessage()    {}
func (*DmiEmpty) Descriptor() ([]byte, []int) {
	return fileDescriptor_49e784b4938902cc, []int{2}
}

func (m *DmiEmpty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DmiEmpty.Unmarshal(m, b)
}
func (m *DmiEmpty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DmiEmpty.Marshal(b, m, deterministic)
}
func (m *DmiEmpty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DmiEmpty.Merge(m, src)
}
func (m *DmiEmpty) XXX_Size() int {
	return xxx_messageInfo_DmiEmpty.Size(m)
}
func (m *DmiEmpty) XXX_DiscardUnknown() {
	xxx_messageInfo_DmiEmpty.DiscardUnknown(m)
}

var xxx_messageInfo_DmiEmpty proto.InternalMessageInfo

type TransceiverRequest struct {
	TransceiverId        uint32   `protobuf:"varint,1,opt,name=TransceiverId,proto3" json:"TransceiverId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TransceiverRequest) Reset()         { *m = TransceiverRequest{} }
func (m *TransceiverRequest) String() string { return proto.CompactTextString(m) }
func (*TransceiverRequest) ProtoMessage()    {}
func (*TransceiverRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_49e784b4938902cc, []int{3}
}

func (m *TransceiverRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TransceiverRequest.Unmarshal(m, b)
}
func (m *TransceiverRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TransceiverRequest.Marshal(b, m, deterministic)
}
func (m *TransceiverRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TransceiverRequest.Merge(m, src)
}
func (m *TransceiverRequest) XXX_Size() int {
	return xxx_messageInfo_TransceiverRequest.Size(m)
}
func (m *TransceiverRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TransceiverRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TransceiverRequest proto.InternalMessageInfo

func (m *TransceiverRequest) GetTransceiverId() uint32 {
	if m != nil {
		return m.TransceiverId
	}
	return 0
}

type Transceiver struct {
	ID                   uint32   `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	UUID                 string   `protobuf:"bytes,2,opt,name=UUID,proto3" json:"UUID,omitempty"`
	Name                 string   `protobuf:"bytes,3,opt,name=Name,proto3" json:"Name,omitempty"`
	Technology           string   `protobuf:"bytes,4,opt,name=Technology,proto3" json:"Technology,omitempty"`
	PluggedIn            bool     `protobuf:"varint,5,opt,name=PluggedIn,proto3" json:"PluggedIn,omitempty"`
	PonIds               []uint32 `protobuf:"varint,6,rep,packed,name=PonIds,proto3" json:"PonIds,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Transceiver) Reset()         { *m = Transceiver{} }
func (m *Transceiver) String() string { return proto.CompactTextString(m) }
func (*Transceiver) ProtoMessage()    {}
func (*Transceiver) Descriptor() ([]byte, []int) {
	return fileDescriptor_49e784b4938902cc, []int{4}
}

func (m *Transceiver) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transceiver.Unmarshal(m, b)
}
func (m *Transceiver) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transceiver.Marshal(b, m, deterministic)
}
func (m *Transceiver) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transceiver.Merge(m, src)
}
func (m *Transceiver) XXX_Size() int {
	return xxx_messageInfo_Transceiver.Size(m)
}
func (m *Transceiver) XXX_DiscardUnknown() {
	xxx_messageInfo_Transceiver.DiscardUnknown(m)
}

var xxx_messageInfo_Transceiver proto.InternalMessageInfo

func (m *Transceiver) GetID() uint32 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *Transceiver) GetUUID() string {
	if m != nil {
		return m.UUID
	}
	return ""
}

func (m *Transceiver) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Transceiver) GetTechnology() string {
	if m != nil {
		return m.Technology
	}
	return ""
}

func (m *Transceiver) GetPluggedIn() bool {
	if m != nil {
		return m.PluggedIn
	}
	return false
}

func (m *Transceiver) GetPonIds() []uint32 {
	if m != nil {
		return m.PonIds
	}
	return nil
}

type Transceivers struct {
	Items                []*Transceiver `protobuf:"bytes,1,rep,name=Items,proto3" json:"Items,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Transceivers) Reset()         { *m = Transceivers{} }
func (m *Transceivers) String() string { return proto.CompactTextString(m) }
func (*Transceivers) ProtoMessage()    {}
func (*Transceivers) Descriptor() ([]byte, []int) {
	return fileDescriptor_49e784b4938902cc, []int{5}
}

func (m *Transceivers) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transceivers.Unmarshal(m, b)
}
func (m *Transceivers) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transceivers.Marshal(b, m, deterministic)
}
func (m *Transceivers) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transceivers.Merge(m, src)
}
func (m *Transceivers) XXX_Size() int {
	return xxx_messageInfo_Transceivers.Size(m)
}
func (m *Transceivers) XXX_DiscardUnknown() {
	xxx_messageInfo_Transceivers.DiscardUnknown(m)
}

var xxx_messageInfo_Transceivers proto.InternalMessageInfo

func (m *Transceivers) GetItems() []*Transceiver {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
	proto.RegisterType((*DmiEvent)(nil), "bbsim.DmiEvent")
	proto.RegisterType((*DmiResponse)(nil), "bbsim.DmiResponse")
	proto.RegisterType((*DmiEmpty)(nil), "bbsim.DmiEmpty")
	proto.RegisterType((*TransceiverRequest)(nil), "bbsim.TransceiverRequest")
	proto.RegisterType((*Transceiver)(nil), "bbsim.Transceiver")
	proto.RegisterType((*Transceivers)(nil), "bbsim.Transceivers")
}

func init() { proto.RegisterFile("api/bbsim/bbsim_dmi.proto", fileDescriptor_49e784b4938902cc) }

var fileDescriptor_49e784b4938902cc = []byte{
	// 414 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x92, 0xcd, 0x6e, 0xd4, 0x30,
	0x10, 0xc7, 0x9b, 0xdd, 0xee, 0xd2, 0x9d, 0xb0, 0x54, 0x0c, 0x12, 0x72, 0x2b, 0x3e, 0x56, 0x06,
	0xa4, 0x70, 0x49, 0xa5, 0xc2, 0x01, 0x38, 0x6e, 0x53, 0x81, 0x2f, 0x50, 0x45, 0xed, 0x85, 0xcb,
	0x2a, 0x1f, 0xa3, 0x34, 0x52, 0x6d, 0x87, 0xd8, 0xa9, 0xd4, 0x47, 0xe0, 0x19, 0x78, 0x59, 0x14,
	0x3b, 0x4b, 0x53, 0x01, 0xa7, 0x5e, 0xa2, 0xf1, 0x2f, 0x33, 0xe3, 0xff, 0xf8, 0x3f, 0x70, 0x90,
	0x35, 0xf5, 0x51, 0x9e, 0x9b, 0x5a, 0xfa, 0xef, 0xa6, 0x94, 0x75, 0xdc, 0xb4, 0xda, 0x6a, 0x9c,
	0x39, 0xc0, 0xdf, 0xc2, 0x5e, 0x22, 0xeb, 0xd3, 0x6b, 0x52, 0x16, 0x9f, 0x03, 0x50, 0x1f, 0x6c,
	0x54, 0x26, 0x89, 0x05, 0xab, 0x20, 0x5a, 0xa4, 0x0b, 0x47, 0xbe, 0x66, 0x92, 0xf8, 0x17, 0x08,
	0x13, 0x59, 0xa7, 0x64, 0x1a, 0xad, 0x0c, 0xe1, 0x4b, 0x08, 0x8d, 0xcd, 0x6c, 0x67, 0x36, 0x85,
	0x2e, 0x7d, 0xfa, 0x2c, 0x05, 0x8f, 0x4e, 0x74, 0x49, 0xc8, 0xe0, 0x81, 0x24, 0x63, 0xb2, 0x8a,
	0xd8, 0xc4, 0xf5, 0xda, 0x1e, 0x39, 0xf8, 0x4b, 0x65, 0x63, 0x6f, 0xf8, 0x27, 0xc0, 0xf3, 0x36,
	0x53, 0xa6, 0xa0, 0xfa, 0x9a, 0xda, 0x94, 0x7e, 0x74, 0x64, 0x2c, 0xbe, 0x86, 0xe5, 0x88, 0x8a,
	0xd2, 0xb5, 0x5f, 0xa6, 0x77, 0x21, 0xff, 0x15, 0x40, 0x38, 0x22, 0xf8, 0x08, 0x26, 0x22, 0x19,
	0x52, 0x27, 0x22, 0x41, 0x84, 0xdd, 0x8b, 0x0b, 0x91, 0x0c, 0xd7, 0xbb, 0xb8, 0x67, 0xfd, 0x34,
	0x6c, 0xea, 0x59, 0x1f, 0xe3, 0x0b, 0x80, 0x73, 0x2a, 0x2e, 0x95, 0xbe, 0xd2, 0xd5, 0x0d, 0xdb,
	0x75, 0x7f, 0x46, 0x04, 0x9f, 0xc1, 0xe2, 0xec, 0xaa, 0xab, 0x2a, 0x2a, 0x85, 0x62, 0xb3, 0x55,
	0x10, 0xed, 0xa5, 0xb7, 0x00, 0x9f, 0xc2, 0xfc, 0x4c, 0x2b, 0x51, 0x1a, 0x36, 0x5f, 0x4d, 0xa3,
	0x65, 0x3a, 0x9c, 0xf8, 0x07, 0x78, 0x38, 0x12, 0x67, 0x30, 0x82, 0x99, 0xb0, 0x24, 0x0d, 0x0b,
	0x56, 0xd3, 0x28, 0x3c, 0xc6, 0xd8, 0x39, 0x10, 0x8f, 0xa7, 0xf7, 0x09, 0xc7, 0x3f, 0x27, 0xb0,
	0x58, 0xaf, 0x07, 0xbf, 0xf0, 0x3d, 0x84, 0x27, 0x2d, 0x65, 0x96, 0xbc, 0x4b, 0xfb, 0x43, 0xdd,
	0xd6, 0xb6, 0x43, 0xbc, 0x05, 0x5b, 0x73, 0xf8, 0x0e, 0x7e, 0x84, 0xfd, 0xcf, 0x64, 0xef, 0x08,
	0x18, 0x57, 0xf6, 0x6f, 0x7f, 0xf8, 0xe4, 0x6f, 0x09, 0x86, 0xef, 0xe0, 0x29, 0x60, 0x3f, 0xdd,
	0xb7, 0x6e, 0x5c, 0x8e, 0x07, 0xff, 0xd0, 0xeb, 0xdd, 0xfa, 0x8f, 0x82, 0x04, 0x1e, 0xf7, 0x6d,
	0x84, 0xba, 0x4f, 0x97, 0xf5, 0x9b, 0xef, 0xaf, 0xaa, 0xda, 0x5e, 0x76, 0x79, 0x5c, 0x68, 0x79,
	0xa4, 0x1b, 0x52, 0x85, 0x6e, 0xcb, 0x61, 0xa9, 0xff, 0xac, 0x77, 0x3e, 0x77, 0x5b, 0xfd, 0xee,
	0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x62, 0xd2, 0xd2, 0x82, 0xf2, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// BBsimDmiClient is the client API for BBsimDmi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BBsimDmiClient interface {
	// Ask the DMI Server to create an event
	CreateEvent(ctx context.Context, in *DmiEvent, opts ...grpc.CallOption) (*DmiResponse, error)
	GetTransceivers(ctx context.Context, in *DmiEmpty, opts ...grpc.CallOption) (*Transceivers, error)
	// Plug out the transceiver by transceiverId
	PlugOutTransceiver(ctx context.Context, in *TransceiverRequest, opts ...grpc.CallOption) (*DmiResponse, error)
	// Plug in the transceiver of a PON by pon-port-ID
	PlugInTransceiver(ctx context.Context, in *TransceiverRequest, opts ...grpc.CallOption) (*DmiResponse, error)
}

type bBsimDmiClient struct {
	cc *grpc.ClientConn
}

func NewBBsimDmiClient(cc *grpc.ClientConn) BBsimDmiClient {
	return &bBsimDmiClient{cc}
}

func (c *bBsimDmiClient) CreateEvent(ctx context.Context, in *DmiEvent, opts ...grpc.CallOption) (*DmiResponse, error) {
	out := new(DmiResponse)
	err := c.cc.Invoke(ctx, "/bbsim.BBsim_dmi/CreateEvent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bBsimDmiClient) GetTransceivers(ctx context.Context, in *DmiEmpty, opts ...grpc.CallOption) (*Transceivers, error) {
	out := new(Transceivers)
	err := c.cc.Invoke(ctx, "/bbsim.BBsim_dmi/GetTransceivers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bBsimDmiClient) PlugOutTransceiver(ctx context.Context, in *TransceiverRequest, opts ...grpc.CallOption) (*DmiResponse, error) {
	out := new(DmiResponse)
	err := c.cc.Invoke(ctx, "/bbsim.BBsim_dmi/PlugOutTransceiver", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bBsimDmiClient) PlugInTransceiver(ctx context.Context, in *TransceiverRequest, opts ...grpc.CallOption) (*DmiResponse, error) {
	out := new(DmiResponse)
	err := c.cc.Invoke(ctx, "/bbsim.BBsim_dmi/PlugInTransceiver", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BBsimDmiServer is the server API for BBsimDmi service.
type BBsimDmiServer interface {
	// Ask the DMI Server to create an event
	CreateEvent(context.Context, *DmiEvent) (*DmiResponse, error)
	GetTransceivers(context.Context, *DmiEmpty) (*Transceivers, error)
	// Plug out the transceiver by transceiverId
	PlugOutTransceiver(context.Context, *TransceiverRequest) (*DmiResponse, error)
	// Plug in the transceiver of a PON by pon-port-ID
	PlugInTransceiver(context.Context, *TransceiverRequest) (*DmiResponse, error)
}

// UnimplementedBBsimDmiServer can be embedded to have forward compatible implementations.
type UnimplementedBBsimDmiServer struct {
}

func (*UnimplementedBBsimDmiServer) CreateEvent(ctx context.Context, req *DmiEvent) (*DmiResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateEvent not implemented")
}
func (*UnimplementedBBsimDmiServer) GetTransceivers(ctx context.Context, req *DmiEmpty) (*Transceivers, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransceivers not implemented")
}
func (*UnimplementedBBsimDmiServer) PlugOutTransceiver(ctx context.Context, req *TransceiverRequest) (*DmiResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PlugOutTransceiver not implemented")
}
func (*UnimplementedBBsimDmiServer) PlugInTransceiver(ctx context.Context, req *TransceiverRequest) (*DmiResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PlugInTransceiver not implemented")
}

func RegisterBBsimDmiServer(s *grpc.Server, srv BBsimDmiServer) {
	s.RegisterService(&_BBsimDmi_serviceDesc, srv)
}

func _BBsimDmi_CreateEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DmiEvent)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BBsimDmiServer).CreateEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bbsim.BBsim_dmi/CreateEvent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BBsimDmiServer).CreateEvent(ctx, req.(*DmiEvent))
	}
	return interceptor(ctx, in, info, handler)
}

func _BBsimDmi_GetTransceivers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DmiEmpty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BBsimDmiServer).GetTransceivers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bbsim.BBsim_dmi/GetTransceivers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BBsimDmiServer).GetTransceivers(ctx, req.(*DmiEmpty))
	}
	return interceptor(ctx, in, info, handler)
}

func _BBsimDmi_PlugOutTransceiver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransceiverRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BBsimDmiServer).PlugOutTransceiver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bbsim.BBsim_dmi/PlugOutTransceiver",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BBsimDmiServer).PlugOutTransceiver(ctx, req.(*TransceiverRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BBsimDmi_PlugInTransceiver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransceiverRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BBsimDmiServer).PlugInTransceiver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bbsim.BBsim_dmi/PlugInTransceiver",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BBsimDmiServer).PlugInTransceiver(ctx, req.(*TransceiverRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _BBsimDmi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bbsim.BBsim_dmi",
	HandlerType: (*BBsimDmiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateEvent",
			Handler:    _BBsimDmi_CreateEvent_Handler,
		},
		{
			MethodName: "GetTransceivers",
			Handler:    _BBsimDmi_GetTransceivers_Handler,
		},
		{
			MethodName: "PlugOutTransceiver",
			Handler:    _BBsimDmi_PlugOutTransceiver_Handler,
		},
		{
			MethodName: "PlugInTransceiver",
			Handler:    _BBsimDmi_PlugInTransceiver_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/bbsim/bbsim_dmi.proto",
}
