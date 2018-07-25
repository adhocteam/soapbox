// Code generated by protoc-gen-go. DO NOT EDIT.
// source: version.proto

package proto

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

type GetVersionResponse struct {
	Version              string   `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	GitCommit            string   `protobuf:"bytes,2,opt,name=git_commit,json=gitCommit,proto3" json:"git_commit,omitempty"`
	BuildTime            string   `protobuf:"bytes,3,opt,name=build_time,json=buildTime,proto3" json:"build_time,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetVersionResponse) Reset()         { *m = GetVersionResponse{} }
func (m *GetVersionResponse) String() string { return proto.CompactTextString(m) }
func (*GetVersionResponse) ProtoMessage()    {}
func (*GetVersionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_version_05ac553e5bec2404, []int{0}
}
func (m *GetVersionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetVersionResponse.Unmarshal(m, b)
}
func (m *GetVersionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetVersionResponse.Marshal(b, m, deterministic)
}
func (dst *GetVersionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetVersionResponse.Merge(dst, src)
}
func (m *GetVersionResponse) XXX_Size() int {
	return xxx_messageInfo_GetVersionResponse.Size(m)
}
func (m *GetVersionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetVersionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetVersionResponse proto.InternalMessageInfo

func (m *GetVersionResponse) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *GetVersionResponse) GetGitCommit() string {
	if m != nil {
		return m.GitCommit
	}
	return ""
}

func (m *GetVersionResponse) GetBuildTime() string {
	if m != nil {
		return m.BuildTime
	}
	return ""
}

func init() {
	proto.RegisterType((*GetVersionResponse)(nil), "soapbox.GetVersionResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// VersionClient is the client API for Version service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type VersionClient interface {
	GetVersion(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*GetVersionResponse, error)
}

type versionClient struct {
	cc *grpc.ClientConn
}

func NewVersionClient(cc *grpc.ClientConn) VersionClient {
	return &versionClient{cc}
}

func (c *versionClient) GetVersion(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*GetVersionResponse, error) {
	out := new(GetVersionResponse)
	err := c.cc.Invoke(ctx, "/soapbox.Version/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VersionServer is the server API for Version service.
type VersionServer interface {
	GetVersion(context.Context, *Empty) (*GetVersionResponse, error)
}

func RegisterVersionServer(s *grpc.Server, srv VersionServer) {
	s.RegisterService(&_Version_serviceDesc, srv)
}

func _Version_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VersionServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Version/GetVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VersionServer).GetVersion(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _Version_serviceDesc = grpc.ServiceDesc{
	ServiceName: "soapbox.Version",
	HandlerType: (*VersionServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetVersion",
			Handler:    _Version_GetVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "version.proto",
}

func init() { proto.RegisterFile("version.proto", fileDescriptor_version_05ac553e5bec2404) }

var fileDescriptor_version_05ac553e5bec2404 = []byte{
	// 174 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x4b, 0x2d, 0x2a,
	0xce, 0xcc, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2f, 0xce, 0x4f, 0x2c, 0x48,
	0xca, 0xaf, 0x90, 0xe2, 0x85, 0x32, 0x20, 0xe2, 0x4a, 0x39, 0x5c, 0x42, 0xee, 0xa9, 0x25, 0x61,
	0x10, 0xb5, 0x41, 0xa9, 0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9, 0x42, 0x12, 0x5c, 0xec, 0x50, 0xed,
	0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x30, 0xae, 0x90, 0x2c, 0x17, 0x57, 0x7a, 0x66, 0x49,
	0x7c, 0x72, 0x7e, 0x6e, 0x6e, 0x66, 0x89, 0x04, 0x13, 0x58, 0x92, 0x33, 0x3d, 0xb3, 0xc4, 0x19,
	0x2c, 0x00, 0x92, 0x4e, 0x2a, 0xcd, 0xcc, 0x49, 0x89, 0x2f, 0xc9, 0xcc, 0x4d, 0x95, 0x60, 0x86,
	0x48, 0x83, 0x45, 0x42, 0x32, 0x73, 0x53, 0x8d, 0xdc, 0xb8, 0xd8, 0xa1, 0x56, 0x09, 0x59, 0x73,
	0x71, 0x21, 0x2c, 0x16, 0xe2, 0xd3, 0x83, 0x39, 0xcb, 0x35, 0xb7, 0xa0, 0xa4, 0x52, 0x4a, 0x1a,
	0xce, 0xc7, 0x74, 0x9d, 0x12, 0x83, 0x13, 0x7b, 0x14, 0x2b, 0xd8, 0xf9, 0x49, 0x6c, 0x60, 0xca,
	0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0xc4, 0xa9, 0x1a, 0x96, 0xee, 0x00, 0x00, 0x00,
}
