// Code generated by protoc-gen-go. DO NOT EDIT.
// source: im_service.proto

/*
Package im_service is a generated protocol buffer package.

It is generated from these files:
	im_service.proto

It has these top-level messages:
	MsgModel
*/
package im_service

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type MsgModel struct {
	GateId     int32  `protobuf:"varint,1,opt,name=gate_id,json=gateId" json:"gate_id,omitempty"`
	Uid        uint32 `protobuf:"varint,2,opt,name=uid" json:"uid,omitempty"`
	MsgType    int32  `protobuf:"varint,3,opt,name=msg_type,json=msgType" json:"msg_type,omitempty"`
	MsgContent string `protobuf:"bytes,4,opt,name=msg_content,json=msgContent" json:"msg_content,omitempty"`
}

func (m *MsgModel) Reset()                    { *m = MsgModel{} }
func (m *MsgModel) String() string            { return proto.CompactTextString(m) }
func (*MsgModel) ProtoMessage()               {}
func (*MsgModel) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *MsgModel) GetGateId() int32 {
	if m != nil {
		return m.GateId
	}
	return 0
}

func (m *MsgModel) GetUid() uint32 {
	if m != nil {
		return m.Uid
	}
	return 0
}

func (m *MsgModel) GetMsgType() int32 {
	if m != nil {
		return m.MsgType
	}
	return 0
}

func (m *MsgModel) GetMsgContent() string {
	if m != nil {
		return m.MsgContent
	}
	return ""
}

func init() {
	proto.RegisterType((*MsgModel)(nil), "im_service.MsgModel")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Im service

type ImClient interface {
	NewMsg(ctx context.Context, opts ...grpc.CallOption) (Im_NewMsgClient, error)
}

type imClient struct {
	cc *grpc.ClientConn
}

func NewImClient(cc *grpc.ClientConn) ImClient {
	return &imClient{cc}
}

func (c *imClient) NewMsg(ctx context.Context, opts ...grpc.CallOption) (Im_NewMsgClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Im_serviceDesc.Streams[0], c.cc, "/im_service.im/NewMsg", opts...)
	if err != nil {
		return nil, err
	}
	x := &imNewMsgClient{stream}
	return x, nil
}

type Im_NewMsgClient interface {
	Send(*MsgModel) error
	Recv() (*MsgModel, error)
	grpc.ClientStream
}

type imNewMsgClient struct {
	grpc.ClientStream
}

func (x *imNewMsgClient) Send(m *MsgModel) error {
	return x.ClientStream.SendMsg(m)
}

func (x *imNewMsgClient) Recv() (*MsgModel, error) {
	m := new(MsgModel)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Im service

type ImServer interface {
	NewMsg(Im_NewMsgServer) error
}

func RegisterImServer(s *grpc.Server, srv ImServer) {
	s.RegisterService(&_Im_serviceDesc, srv)
}

func _Im_NewMsg_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ImServer).NewMsg(&imNewMsgServer{stream})
}

type Im_NewMsgServer interface {
	Send(*MsgModel) error
	Recv() (*MsgModel, error)
	grpc.ServerStream
}

type imNewMsgServer struct {
	grpc.ServerStream
}

func (x *imNewMsgServer) Send(m *MsgModel) error {
	return x.ServerStream.SendMsg(m)
}

func (x *imNewMsgServer) Recv() (*MsgModel, error) {
	m := new(MsgModel)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Im_serviceDesc = grpc.ServiceDesc{
	ServiceName: "im_service.im",
	HandlerType: (*ImServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "NewMsg",
			Handler:       _Im_NewMsg_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "im_service.proto",
}

func init() { proto.RegisterFile("im_service.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 183 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xc8, 0xcc, 0x8d, 0x2f,
	0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x42, 0x88,
	0x28, 0x15, 0x72, 0x71, 0xf8, 0x16, 0xa7, 0xfb, 0xe6, 0xa7, 0xa4, 0xe6, 0x08, 0x89, 0x73, 0xb1,
	0xa7, 0x27, 0x96, 0xa4, 0xc6, 0x67, 0xa6, 0x48, 0x30, 0x2a, 0x30, 0x6a, 0xb0, 0x06, 0xb1, 0x81,
	0xb8, 0x9e, 0x29, 0x42, 0x02, 0x5c, 0xcc, 0xa5, 0x99, 0x29, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0xbc,
	0x41, 0x20, 0xa6, 0x90, 0x24, 0x17, 0x47, 0x6e, 0x71, 0x7a, 0x7c, 0x49, 0x65, 0x41, 0xaa, 0x04,
	0x33, 0x58, 0x2d, 0x7b, 0x6e, 0x71, 0x7a, 0x48, 0x65, 0x41, 0xaa, 0x90, 0x3c, 0x17, 0x37, 0x48,
	0x2a, 0x39, 0x3f, 0xaf, 0x24, 0x35, 0xaf, 0x44, 0x82, 0x45, 0x81, 0x51, 0x83, 0x33, 0x88, 0x2b,
	0xb7, 0x38, 0xdd, 0x19, 0x22, 0x62, 0x64, 0xc7, 0xc5, 0x94, 0x99, 0x2b, 0x64, 0xc1, 0xc5, 0xe6,
	0x97, 0x5a, 0xee, 0x5b, 0x9c, 0x2e, 0x24, 0xa2, 0x87, 0xe4, 0x42, 0x98, 0x63, 0xa4, 0xb0, 0x8a,
	0x6a, 0x30, 0x1a, 0x30, 0x26, 0xb1, 0x81, 0x7d, 0x61, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x35,
	0xd1, 0x76, 0x9a, 0xd9, 0x00, 0x00, 0x00,
}