// Code generated by protoc-gen-go. DO NOT EDIT.
// source: dmi/sw_management_service.proto

package dmi

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
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

type SoftwareVersionInformation struct {
	ActiveVersions       []*ImageVersion `protobuf:"bytes,1,rep,name=active_versions,json=activeVersions,proto3" json:"active_versions,omitempty"`
	StandbyVersions      []*ImageVersion `protobuf:"bytes,2,rep,name=standby_versions,json=standbyVersions,proto3" json:"standby_versions,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *SoftwareVersionInformation) Reset()         { *m = SoftwareVersionInformation{} }
func (m *SoftwareVersionInformation) String() string { return proto.CompactTextString(m) }
func (*SoftwareVersionInformation) ProtoMessage()    {}
func (*SoftwareVersionInformation) Descriptor() ([]byte, []int) {
	return fileDescriptor_000929e4bec891d7, []int{0}
}

func (m *SoftwareVersionInformation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SoftwareVersionInformation.Unmarshal(m, b)
}
func (m *SoftwareVersionInformation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SoftwareVersionInformation.Marshal(b, m, deterministic)
}
func (m *SoftwareVersionInformation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SoftwareVersionInformation.Merge(m, src)
}
func (m *SoftwareVersionInformation) XXX_Size() int {
	return xxx_messageInfo_SoftwareVersionInformation.Size(m)
}
func (m *SoftwareVersionInformation) XXX_DiscardUnknown() {
	xxx_messageInfo_SoftwareVersionInformation.DiscardUnknown(m)
}

var xxx_messageInfo_SoftwareVersionInformation proto.InternalMessageInfo

func (m *SoftwareVersionInformation) GetActiveVersions() []*ImageVersion {
	if m != nil {
		return m.ActiveVersions
	}
	return nil
}

func (m *SoftwareVersionInformation) GetStandbyVersions() []*ImageVersion {
	if m != nil {
		return m.StandbyVersions
	}
	return nil
}

type GetSoftwareVersionInformationResponse struct {
	Status               Status                      `protobuf:"varint,1,opt,name=status,proto3,enum=dmi.Status" json:"status,omitempty"`
	Reason               Reason                      `protobuf:"varint,2,opt,name=reason,proto3,enum=dmi.Reason" json:"reason,omitempty"`
	Info                 *SoftwareVersionInformation `protobuf:"bytes,3,opt,name=info,proto3" json:"info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                    `json:"-"`
	XXX_unrecognized     []byte                      `json:"-"`
	XXX_sizecache        int32                       `json:"-"`
}

func (m *GetSoftwareVersionInformationResponse) Reset()         { *m = GetSoftwareVersionInformationResponse{} }
func (m *GetSoftwareVersionInformationResponse) String() string { return proto.CompactTextString(m) }
func (*GetSoftwareVersionInformationResponse) ProtoMessage()    {}
func (*GetSoftwareVersionInformationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_000929e4bec891d7, []int{1}
}

func (m *GetSoftwareVersionInformationResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetSoftwareVersionInformationResponse.Unmarshal(m, b)
}
func (m *GetSoftwareVersionInformationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetSoftwareVersionInformationResponse.Marshal(b, m, deterministic)
}
func (m *GetSoftwareVersionInformationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetSoftwareVersionInformationResponse.Merge(m, src)
}
func (m *GetSoftwareVersionInformationResponse) XXX_Size() int {
	return xxx_messageInfo_GetSoftwareVersionInformationResponse.Size(m)
}
func (m *GetSoftwareVersionInformationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetSoftwareVersionInformationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetSoftwareVersionInformationResponse proto.InternalMessageInfo

func (m *GetSoftwareVersionInformationResponse) GetStatus() Status {
	if m != nil {
		return m.Status
	}
	return Status_UNDEFINED_STATUS
}

func (m *GetSoftwareVersionInformationResponse) GetReason() Reason {
	if m != nil {
		return m.Reason
	}
	return Reason_UNDEFINED_REASON
}

func (m *GetSoftwareVersionInformationResponse) GetInfo() *SoftwareVersionInformation {
	if m != nil {
		return m.Info
	}
	return nil
}

type DownloadImageRequest struct {
	DeviceUuid           *Uuid             `protobuf:"bytes,1,opt,name=device_uuid,json=deviceUuid,proto3" json:"device_uuid,omitempty"`
	ImageInfo            *ImageInformation `protobuf:"bytes,2,opt,name=image_info,json=imageInfo,proto3" json:"image_info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *DownloadImageRequest) Reset()         { *m = DownloadImageRequest{} }
func (m *DownloadImageRequest) String() string { return proto.CompactTextString(m) }
func (*DownloadImageRequest) ProtoMessage()    {}
func (*DownloadImageRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_000929e4bec891d7, []int{2}
}

func (m *DownloadImageRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DownloadImageRequest.Unmarshal(m, b)
}
func (m *DownloadImageRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DownloadImageRequest.Marshal(b, m, deterministic)
}
func (m *DownloadImageRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DownloadImageRequest.Merge(m, src)
}
func (m *DownloadImageRequest) XXX_Size() int {
	return xxx_messageInfo_DownloadImageRequest.Size(m)
}
func (m *DownloadImageRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DownloadImageRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DownloadImageRequest proto.InternalMessageInfo

func (m *DownloadImageRequest) GetDeviceUuid() *Uuid {
	if m != nil {
		return m.DeviceUuid
	}
	return nil
}

func (m *DownloadImageRequest) GetImageInfo() *ImageInformation {
	if m != nil {
		return m.ImageInfo
	}
	return nil
}

func init() {
	proto.RegisterType((*SoftwareVersionInformation)(nil), "dmi.SoftwareVersionInformation")
	proto.RegisterType((*GetSoftwareVersionInformationResponse)(nil), "dmi.GetSoftwareVersionInformationResponse")
	proto.RegisterType((*DownloadImageRequest)(nil), "dmi.DownloadImageRequest")
}

func init() { proto.RegisterFile("dmi/sw_management_service.proto", fileDescriptor_000929e4bec891d7) }

var fileDescriptor_000929e4bec891d7 = []byte{
	// 458 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0x5d, 0x6b, 0x13, 0x41,
	0x14, 0x65, 0x13, 0x29, 0x74, 0xd6, 0x36, 0xed, 0x50, 0x21, 0xee, 0x4b, 0x4a, 0x44, 0x08, 0x85,
	0x66, 0x65, 0xd3, 0x17, 0xad, 0x08, 0x4a, 0x41, 0xf3, 0xa0, 0xe0, 0x46, 0x7d, 0xf0, 0x65, 0x99,
	0xec, 0xdc, 0xa4, 0x03, 0xce, 0xdc, 0x38, 0x33, 0xbb, 0xd1, 0x3f, 0xe2, 0x4f, 0xd0, 0xbf, 0x29,
	0xf3, 0x11, 0x53, 0x69, 0x03, 0x7d, 0xdb, 0x3d, 0xf7, 0x9c, 0x33, 0x67, 0xee, 0xbd, 0x43, 0x06,
	0x5c, 0x8a, 0xdc, 0xac, 0x2b, 0xc9, 0x14, 0x5b, 0x82, 0x04, 0x65, 0x2b, 0x03, 0xba, 0x15, 0x35,
	0x8c, 0x57, 0x1a, 0x2d, 0xd2, 0x2e, 0x97, 0x22, 0x3b, 0x76, 0xac, 0x1a, 0xa5, 0x44, 0x65, 0x02,
	0x9e, 0x3d, 0x74, 0xd0, 0xf5, 0x3a, 0xfe, 0xd1, 0x68, 0x23, 0x24, 0x5b, 0x46, 0xe5, 0xf0, 0x57,
	0x42, 0xb2, 0x19, 0x2e, 0xec, 0x9a, 0x69, 0xf8, 0x02, 0xda, 0x08, 0x54, 0x53, 0xb5, 0x40, 0x2d,
	0x99, 0x15, 0xa8, 0xe8, 0x0b, 0xd2, 0x63, 0xb5, 0x15, 0x2d, 0x54, 0x6d, 0x28, 0x9a, 0x7e, 0x72,
	0xda, 0x1d, 0xa5, 0xc5, 0xf1, 0x98, 0x4b, 0x31, 0x9e, 0x3a, 0xa7, 0x28, 0x2b, 0x0f, 0x03, 0x33,
	0xfe, 0x1a, 0xfa, 0x92, 0x1c, 0x19, 0xcb, 0x14, 0x9f, 0xff, 0xdc, 0x8a, 0x3b, 0xbb, 0xc4, 0xbd,
	0x48, 0xdd, 0xa8, 0x87, 0xbf, 0x13, 0xf2, 0xf4, 0x2d, 0xd8, 0xdd, 0xd9, 0x4a, 0x30, 0x2b, 0x54,
	0x06, 0xe8, 0x13, 0xb2, 0x67, 0x2c, 0xb3, 0x8d, 0x8b, 0x96, 0x8c, 0x0e, 0x8b, 0xd4, 0xbb, 0xcf,
	0x3c, 0x54, 0xc6, 0x92, 0x23, 0x69, 0x60, 0x06, 0x55, 0xbf, 0x73, 0x83, 0x54, 0x7a, 0xa8, 0x8c,
	0x25, 0x3a, 0x21, 0x0f, 0x84, 0x5a, 0x60, 0xbf, 0x7b, 0x9a, 0x8c, 0xd2, 0x62, 0x10, 0x7c, 0x76,
	0x07, 0xf0, 0xe4, 0xe1, 0x0f, 0x72, 0x72, 0x85, 0x6b, 0xf5, 0x0d, 0x19, 0xf7, 0x37, 0x2a, 0xe1,
	0x7b, 0x03, 0xc6, 0xd2, 0x33, 0x92, 0x72, 0x70, 0x33, 0xaa, 0x9a, 0x46, 0x70, 0x9f, 0x2d, 0x2d,
	0xf6, 0xbd, 0xe7, 0xe7, 0x46, 0xf0, 0x92, 0x84, 0xaa, 0xfb, 0xa6, 0x17, 0x84, 0xf8, 0xa1, 0x54,
	0xfe, 0xf8, 0x8e, 0xa7, 0x3e, 0xda, 0x36, 0xe9, 0xe6, 0xa1, 0xfb, 0x62, 0x83, 0x14, 0x7f, 0x3a,
	0x64, 0xf0, 0x81, 0xb9, 0x9e, 0x6f, 0x42, 0xbe, 0xff, 0xb7, 0x20, 0xb3, 0xb0, 0x1f, 0xf4, 0x23,
	0xa1, 0xb7, 0xbb, 0x48, 0x7b, 0xde, 0xfb, 0x1d, 0xd3, 0xdc, 0xa1, 0xd3, 0xab, 0xec, 0xcc, 0x03,
	0xf7, 0xeb, 0xf7, 0x2b, 0x72, 0xf0, 0xdf, 0x85, 0xe9, 0x63, 0x2f, 0xbe, 0xab, 0x09, 0xd9, 0xd1,
	0xf6, 0x12, 0x61, 0x20, 0xcf, 0x12, 0x7a, 0x41, 0x0e, 0x5e, 0xbb, 0x4d, 0x61, 0x16, 0x82, 0xfe,
	0x56, 0x9a, 0xbb, 0x54, 0x97, 0xe4, 0xa4, 0x84, 0x16, 0xb4, 0xfd, 0x84, 0xb3, 0xb0, 0x2a, 0xf7,
	0x17, 0xbf, 0xb9, 0xfc, 0xfa, 0x7c, 0x29, 0xec, 0x75, 0x33, 0x1f, 0xd7, 0x28, 0x73, 0x5c, 0x81,
	0xaa, 0x51, 0xf3, 0x3c, 0x4c, 0xe0, 0x7c, 0xfb, 0xac, 0xce, 0x85, 0xb2, 0xa0, 0x17, 0xac, 0x86,
	0xbc, 0x9d, 0xe4, 0x4b, 0xcc, 0xb9, 0x14, 0xf3, 0x3d, 0xff, 0x52, 0x26, 0x7f, 0x03, 0x00, 0x00,
	0xff, 0xff, 0x6d, 0x7d, 0x3b, 0xc8, 0x86, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NativeSoftwareManagementServiceClient is the client API for NativeSoftwareManagementService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NativeSoftwareManagementServiceClient interface {
	// Get the software version information of the Active and Standby images
	GetSoftwareVersion(ctx context.Context, in *HardwareID, opts ...grpc.CallOption) (*GetSoftwareVersionInformationResponse, error)
	// Downloads and installs the image in the standby partition, returns the status/progress of the Install
	DownloadImage(ctx context.Context, in *DownloadImageRequest, opts ...grpc.CallOption) (NativeSoftwareManagementService_DownloadImageClient, error)
	// Activates and runs the OLT with the image in the standby partition. If things are fine this image will
	// henceforth be marked as the Active Partition. The old working image would remain on the Standby partition.
	// Any possibly required (sub-)steps like "commit" are left to the "Device Manager"
	ActivateImage(ctx context.Context, in *HardwareID, opts ...grpc.CallOption) (NativeSoftwareManagementService_ActivateImageClient, error)
	// Marks the image in the Standby as Active and reboots the device, so that it boots from that image which was in the standby.
	// This API is to be used if operator wants to go back to the previous software
	RevertToStandbyImage(ctx context.Context, in *HardwareID, opts ...grpc.CallOption) (NativeSoftwareManagementService_RevertToStandbyImageClient, error)
}

type nativeSoftwareManagementServiceClient struct {
	cc *grpc.ClientConn
}

func NewNativeSoftwareManagementServiceClient(cc *grpc.ClientConn) NativeSoftwareManagementServiceClient {
	return &nativeSoftwareManagementServiceClient{cc}
}

func (c *nativeSoftwareManagementServiceClient) GetSoftwareVersion(ctx context.Context, in *HardwareID, opts ...grpc.CallOption) (*GetSoftwareVersionInformationResponse, error) {
	out := new(GetSoftwareVersionInformationResponse)
	err := c.cc.Invoke(ctx, "/dmi.NativeSoftwareManagementService/GetSoftwareVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nativeSoftwareManagementServiceClient) DownloadImage(ctx context.Context, in *DownloadImageRequest, opts ...grpc.CallOption) (NativeSoftwareManagementService_DownloadImageClient, error) {
	stream, err := c.cc.NewStream(ctx, &_NativeSoftwareManagementService_serviceDesc.Streams[0], "/dmi.NativeSoftwareManagementService/DownloadImage", opts...)
	if err != nil {
		return nil, err
	}
	x := &nativeSoftwareManagementServiceDownloadImageClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type NativeSoftwareManagementService_DownloadImageClient interface {
	Recv() (*ImageStatus, error)
	grpc.ClientStream
}

type nativeSoftwareManagementServiceDownloadImageClient struct {
	grpc.ClientStream
}

func (x *nativeSoftwareManagementServiceDownloadImageClient) Recv() (*ImageStatus, error) {
	m := new(ImageStatus)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *nativeSoftwareManagementServiceClient) ActivateImage(ctx context.Context, in *HardwareID, opts ...grpc.CallOption) (NativeSoftwareManagementService_ActivateImageClient, error) {
	stream, err := c.cc.NewStream(ctx, &_NativeSoftwareManagementService_serviceDesc.Streams[1], "/dmi.NativeSoftwareManagementService/ActivateImage", opts...)
	if err != nil {
		return nil, err
	}
	x := &nativeSoftwareManagementServiceActivateImageClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type NativeSoftwareManagementService_ActivateImageClient interface {
	Recv() (*ImageStatus, error)
	grpc.ClientStream
}

type nativeSoftwareManagementServiceActivateImageClient struct {
	grpc.ClientStream
}

func (x *nativeSoftwareManagementServiceActivateImageClient) Recv() (*ImageStatus, error) {
	m := new(ImageStatus)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *nativeSoftwareManagementServiceClient) RevertToStandbyImage(ctx context.Context, in *HardwareID, opts ...grpc.CallOption) (NativeSoftwareManagementService_RevertToStandbyImageClient, error) {
	stream, err := c.cc.NewStream(ctx, &_NativeSoftwareManagementService_serviceDesc.Streams[2], "/dmi.NativeSoftwareManagementService/RevertToStandbyImage", opts...)
	if err != nil {
		return nil, err
	}
	x := &nativeSoftwareManagementServiceRevertToStandbyImageClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type NativeSoftwareManagementService_RevertToStandbyImageClient interface {
	Recv() (*ImageStatus, error)
	grpc.ClientStream
}

type nativeSoftwareManagementServiceRevertToStandbyImageClient struct {
	grpc.ClientStream
}

func (x *nativeSoftwareManagementServiceRevertToStandbyImageClient) Recv() (*ImageStatus, error) {
	m := new(ImageStatus)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// NativeSoftwareManagementServiceServer is the server API for NativeSoftwareManagementService service.
type NativeSoftwareManagementServiceServer interface {
	// Get the software version information of the Active and Standby images
	GetSoftwareVersion(context.Context, *HardwareID) (*GetSoftwareVersionInformationResponse, error)
	// Downloads and installs the image in the standby partition, returns the status/progress of the Install
	DownloadImage(*DownloadImageRequest, NativeSoftwareManagementService_DownloadImageServer) error
	// Activates and runs the OLT with the image in the standby partition. If things are fine this image will
	// henceforth be marked as the Active Partition. The old working image would remain on the Standby partition.
	// Any possibly required (sub-)steps like "commit" are left to the "Device Manager"
	ActivateImage(*HardwareID, NativeSoftwareManagementService_ActivateImageServer) error
	// Marks the image in the Standby as Active and reboots the device, so that it boots from that image which was in the standby.
	// This API is to be used if operator wants to go back to the previous software
	RevertToStandbyImage(*HardwareID, NativeSoftwareManagementService_RevertToStandbyImageServer) error
}

func RegisterNativeSoftwareManagementServiceServer(s *grpc.Server, srv NativeSoftwareManagementServiceServer) {
	s.RegisterService(&_NativeSoftwareManagementService_serviceDesc, srv)
}

func _NativeSoftwareManagementService_GetSoftwareVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HardwareID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NativeSoftwareManagementServiceServer).GetSoftwareVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dmi.NativeSoftwareManagementService/GetSoftwareVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NativeSoftwareManagementServiceServer).GetSoftwareVersion(ctx, req.(*HardwareID))
	}
	return interceptor(ctx, in, info, handler)
}

func _NativeSoftwareManagementService_DownloadImage_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DownloadImageRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(NativeSoftwareManagementServiceServer).DownloadImage(m, &nativeSoftwareManagementServiceDownloadImageServer{stream})
}

type NativeSoftwareManagementService_DownloadImageServer interface {
	Send(*ImageStatus) error
	grpc.ServerStream
}

type nativeSoftwareManagementServiceDownloadImageServer struct {
	grpc.ServerStream
}

func (x *nativeSoftwareManagementServiceDownloadImageServer) Send(m *ImageStatus) error {
	return x.ServerStream.SendMsg(m)
}

func _NativeSoftwareManagementService_ActivateImage_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(HardwareID)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(NativeSoftwareManagementServiceServer).ActivateImage(m, &nativeSoftwareManagementServiceActivateImageServer{stream})
}

type NativeSoftwareManagementService_ActivateImageServer interface {
	Send(*ImageStatus) error
	grpc.ServerStream
}

type nativeSoftwareManagementServiceActivateImageServer struct {
	grpc.ServerStream
}

func (x *nativeSoftwareManagementServiceActivateImageServer) Send(m *ImageStatus) error {
	return x.ServerStream.SendMsg(m)
}

func _NativeSoftwareManagementService_RevertToStandbyImage_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(HardwareID)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(NativeSoftwareManagementServiceServer).RevertToStandbyImage(m, &nativeSoftwareManagementServiceRevertToStandbyImageServer{stream})
}

type NativeSoftwareManagementService_RevertToStandbyImageServer interface {
	Send(*ImageStatus) error
	grpc.ServerStream
}

type nativeSoftwareManagementServiceRevertToStandbyImageServer struct {
	grpc.ServerStream
}

func (x *nativeSoftwareManagementServiceRevertToStandbyImageServer) Send(m *ImageStatus) error {
	return x.ServerStream.SendMsg(m)
}

var _NativeSoftwareManagementService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "dmi.NativeSoftwareManagementService",
	HandlerType: (*NativeSoftwareManagementServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSoftwareVersion",
			Handler:    _NativeSoftwareManagementService_GetSoftwareVersion_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "DownloadImage",
			Handler:       _NativeSoftwareManagementService_DownloadImage_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "ActivateImage",
			Handler:       _NativeSoftwareManagementService_ActivateImage_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "RevertToStandbyImage",
			Handler:       _NativeSoftwareManagementService_RevertToStandbyImage_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "dmi/sw_management_service.proto",
}