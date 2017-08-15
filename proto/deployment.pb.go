// Code generated by protoc-gen-go. DO NOT EDIT.
// source: deployment.proto

package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type DeploymentState int32

const (
	DeploymentState_DEPLOYMENT_ROLLOUT_WAIT  DeploymentState = 0
	DeploymentState_DEPLOYMENT_EVALUATE_WAIT DeploymentState = 1
	DeploymentState_DEPLOYMENT_ROLL_FORWARD  DeploymentState = 2
	DeploymentState_DEPLOYMENT_SUCCEEDED     DeploymentState = 3
	DeploymentState_DEPLOYMENT_FAILED        DeploymentState = 4
)

var DeploymentState_name = map[int32]string{
	0: "DEPLOYMENT_ROLLOUT_WAIT",
	1: "DEPLOYMENT_EVALUATE_WAIT",
	2: "DEPLOYMENT_ROLL_FORWARD",
	3: "DEPLOYMENT_SUCCEEDED",
	4: "DEPLOYMENT_FAILED",
}
var DeploymentState_value = map[string]int32{
	"DEPLOYMENT_ROLLOUT_WAIT":  0,
	"DEPLOYMENT_EVALUATE_WAIT": 1,
	"DEPLOYMENT_ROLL_FORWARD":  2,
	"DEPLOYMENT_SUCCEEDED":     3,
	"DEPLOYMENT_FAILED":        4,
}

func (x DeploymentState) String() string {
	return proto1.EnumName(DeploymentState_name, int32(x))
}
func (DeploymentState) EnumDescriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

type ListDeploymentRequest struct {
	ApplicationId int32 `protobuf:"varint,1,opt,name=application_id,json=applicationId" json:"application_id,omitempty"`
}

func (m *ListDeploymentRequest) Reset()                    { *m = ListDeploymentRequest{} }
func (m *ListDeploymentRequest) String() string            { return proto1.CompactTextString(m) }
func (*ListDeploymentRequest) ProtoMessage()               {}
func (*ListDeploymentRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *ListDeploymentRequest) GetApplicationId() int32 {
	if m != nil {
		return m.ApplicationId
	}
	return 0
}

type ListDeploymentResponse struct {
	Deployments []*Deployment `protobuf:"bytes,1,rep,name=deployments" json:"deployments,omitempty"`
}

func (m *ListDeploymentResponse) Reset()                    { *m = ListDeploymentResponse{} }
func (m *ListDeploymentResponse) String() string            { return proto1.CompactTextString(m) }
func (*ListDeploymentResponse) ProtoMessage()               {}
func (*ListDeploymentResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *ListDeploymentResponse) GetDeployments() []*Deployment {
	if m != nil {
		return m.Deployments
	}
	return nil
}

type GetDeploymentRequest struct {
	Id int32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *GetDeploymentRequest) Reset()                    { *m = GetDeploymentRequest{} }
func (m *GetDeploymentRequest) String() string            { return proto1.CompactTextString(m) }
func (*GetDeploymentRequest) ProtoMessage()               {}
func (*GetDeploymentRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{2} }

func (m *GetDeploymentRequest) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

type GetLatestDeploymentRequest struct {
	ApplicationId int32 `protobuf:"varint,1,opt,name=application_id,json=applicationId" json:"application_id,omitempty"`
	EnvironmentId int32 `protobuf:"varint,2,opt,name=environment_id,json=environmentId" json:"environment_id,omitempty"`
}

func (m *GetLatestDeploymentRequest) Reset()                    { *m = GetLatestDeploymentRequest{} }
func (m *GetLatestDeploymentRequest) String() string            { return proto1.CompactTextString(m) }
func (*GetLatestDeploymentRequest) ProtoMessage()               {}
func (*GetLatestDeploymentRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{3} }

func (m *GetLatestDeploymentRequest) GetApplicationId() int32 {
	if m != nil {
		return m.ApplicationId
	}
	return 0
}

func (m *GetLatestDeploymentRequest) GetEnvironmentId() int32 {
	if m != nil {
		return m.EnvironmentId
	}
	return 0
}

type Deployment struct {
	Id          int32                      `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Application *Application               `protobuf:"bytes,2,opt,name=application" json:"application,omitempty"`
	Env         *Environment               `protobuf:"bytes,3,opt,name=env" json:"env,omitempty"`
	Committish  string                     `protobuf:"bytes,4,opt,name=committish" json:"committish,omitempty"`
	State       DeploymentState            `protobuf:"varint,5,opt,name=state,enum=soapbox.DeploymentState" json:"state,omitempty"`
	CreatedAt   *google_protobuf.Timestamp `protobuf:"bytes,6,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
}

func (m *Deployment) Reset()                    { *m = Deployment{} }
func (m *Deployment) String() string            { return proto1.CompactTextString(m) }
func (*Deployment) ProtoMessage()               {}
func (*Deployment) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{4} }

func (m *Deployment) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Deployment) GetApplication() *Application {
	if m != nil {
		return m.Application
	}
	return nil
}

func (m *Deployment) GetEnv() *Environment {
	if m != nil {
		return m.Env
	}
	return nil
}

func (m *Deployment) GetCommittish() string {
	if m != nil {
		return m.Committish
	}
	return ""
}

func (m *Deployment) GetState() DeploymentState {
	if m != nil {
		return m.State
	}
	return DeploymentState_DEPLOYMENT_ROLLOUT_WAIT
}

func (m *Deployment) GetCreatedAt() *google_protobuf.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

type StartDeploymentResponse struct {
	Id int32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *StartDeploymentResponse) Reset()                    { *m = StartDeploymentResponse{} }
func (m *StartDeploymentResponse) String() string            { return proto1.CompactTextString(m) }
func (*StartDeploymentResponse) ProtoMessage()               {}
func (*StartDeploymentResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{5} }

func (m *StartDeploymentResponse) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

type GetDeploymentStatusRequest struct {
	Id int32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *GetDeploymentStatusRequest) Reset()                    { *m = GetDeploymentStatusRequest{} }
func (m *GetDeploymentStatusRequest) String() string            { return proto1.CompactTextString(m) }
func (*GetDeploymentStatusRequest) ProtoMessage()               {}
func (*GetDeploymentStatusRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{6} }

func (m *GetDeploymentStatusRequest) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

type GetDeploymentStatusResponse struct {
	State string `protobuf:"bytes,1,opt,name=state" json:"state,omitempty"`
}

func (m *GetDeploymentStatusResponse) Reset()                    { *m = GetDeploymentStatusResponse{} }
func (m *GetDeploymentStatusResponse) String() string            { return proto1.CompactTextString(m) }
func (*GetDeploymentStatusResponse) ProtoMessage()               {}
func (*GetDeploymentStatusResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{7} }

func (m *GetDeploymentStatusResponse) GetState() string {
	if m != nil {
		return m.State
	}
	return ""
}

type TeardownDeploymentRequest struct {
	Id int32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *TeardownDeploymentRequest) Reset()                    { *m = TeardownDeploymentRequest{} }
func (m *TeardownDeploymentRequest) String() string            { return proto1.CompactTextString(m) }
func (*TeardownDeploymentRequest) ProtoMessage()               {}
func (*TeardownDeploymentRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{8} }

func (m *TeardownDeploymentRequest) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func init() {
	proto1.RegisterType((*ListDeploymentRequest)(nil), "soapbox.ListDeploymentRequest")
	proto1.RegisterType((*ListDeploymentResponse)(nil), "soapbox.ListDeploymentResponse")
	proto1.RegisterType((*GetDeploymentRequest)(nil), "soapbox.GetDeploymentRequest")
	proto1.RegisterType((*GetLatestDeploymentRequest)(nil), "soapbox.GetLatestDeploymentRequest")
	proto1.RegisterType((*Deployment)(nil), "soapbox.Deployment")
	proto1.RegisterType((*StartDeploymentResponse)(nil), "soapbox.StartDeploymentResponse")
	proto1.RegisterType((*GetDeploymentStatusRequest)(nil), "soapbox.GetDeploymentStatusRequest")
	proto1.RegisterType((*GetDeploymentStatusResponse)(nil), "soapbox.GetDeploymentStatusResponse")
	proto1.RegisterType((*TeardownDeploymentRequest)(nil), "soapbox.TeardownDeploymentRequest")
	proto1.RegisterEnum("soapbox.DeploymentState", DeploymentState_name, DeploymentState_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Deployments service

type DeploymentsClient interface {
	ListDeployments(ctx context.Context, in *ListDeploymentRequest, opts ...grpc.CallOption) (*ListDeploymentResponse, error)
	GetDeployment(ctx context.Context, in *GetDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error)
	GetLatestDeployment(ctx context.Context, in *GetLatestDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error)
	StartDeployment(ctx context.Context, in *Deployment, opts ...grpc.CallOption) (*StartDeploymentResponse, error)
	GetDeploymentStatus(ctx context.Context, in *GetDeploymentStatusRequest, opts ...grpc.CallOption) (*GetDeploymentStatusResponse, error)
	TeardownDeployment(ctx context.Context, in *TeardownDeploymentRequest, opts ...grpc.CallOption) (*Empty, error)
}

type deploymentsClient struct {
	cc *grpc.ClientConn
}

func NewDeploymentsClient(cc *grpc.ClientConn) DeploymentsClient {
	return &deploymentsClient{cc}
}

func (c *deploymentsClient) ListDeployments(ctx context.Context, in *ListDeploymentRequest, opts ...grpc.CallOption) (*ListDeploymentResponse, error) {
	out := new(ListDeploymentResponse)
	err := grpc.Invoke(ctx, "/soapbox.Deployments/ListDeployments", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentsClient) GetDeployment(ctx context.Context, in *GetDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error) {
	out := new(Deployment)
	err := grpc.Invoke(ctx, "/soapbox.Deployments/GetDeployment", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentsClient) GetLatestDeployment(ctx context.Context, in *GetLatestDeploymentRequest, opts ...grpc.CallOption) (*Deployment, error) {
	out := new(Deployment)
	err := grpc.Invoke(ctx, "/soapbox.Deployments/GetLatestDeployment", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentsClient) StartDeployment(ctx context.Context, in *Deployment, opts ...grpc.CallOption) (*StartDeploymentResponse, error) {
	out := new(StartDeploymentResponse)
	err := grpc.Invoke(ctx, "/soapbox.Deployments/StartDeployment", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentsClient) GetDeploymentStatus(ctx context.Context, in *GetDeploymentStatusRequest, opts ...grpc.CallOption) (*GetDeploymentStatusResponse, error) {
	out := new(GetDeploymentStatusResponse)
	err := grpc.Invoke(ctx, "/soapbox.Deployments/GetDeploymentStatus", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deploymentsClient) TeardownDeployment(ctx context.Context, in *TeardownDeploymentRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/soapbox.Deployments/TeardownDeployment", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Deployments service

type DeploymentsServer interface {
	ListDeployments(context.Context, *ListDeploymentRequest) (*ListDeploymentResponse, error)
	GetDeployment(context.Context, *GetDeploymentRequest) (*Deployment, error)
	GetLatestDeployment(context.Context, *GetLatestDeploymentRequest) (*Deployment, error)
	StartDeployment(context.Context, *Deployment) (*StartDeploymentResponse, error)
	GetDeploymentStatus(context.Context, *GetDeploymentStatusRequest) (*GetDeploymentStatusResponse, error)
	TeardownDeployment(context.Context, *TeardownDeploymentRequest) (*Empty, error)
}

func RegisterDeploymentsServer(s *grpc.Server, srv DeploymentsServer) {
	s.RegisterService(&_Deployments_serviceDesc, srv)
}

func _Deployments_ListDeployments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListDeploymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentsServer).ListDeployments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Deployments/ListDeployments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentsServer).ListDeployments(ctx, req.(*ListDeploymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Deployments_GetDeployment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeploymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentsServer).GetDeployment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Deployments/GetDeployment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentsServer).GetDeployment(ctx, req.(*GetDeploymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Deployments_GetLatestDeployment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLatestDeploymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentsServer).GetLatestDeployment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Deployments/GetLatestDeployment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentsServer).GetLatestDeployment(ctx, req.(*GetLatestDeploymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Deployments_StartDeployment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Deployment)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentsServer).StartDeployment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Deployments/StartDeployment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentsServer).StartDeployment(ctx, req.(*Deployment))
	}
	return interceptor(ctx, in, info, handler)
}

func _Deployments_GetDeploymentStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeploymentStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentsServer).GetDeploymentStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Deployments/GetDeploymentStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentsServer).GetDeploymentStatus(ctx, req.(*GetDeploymentStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Deployments_TeardownDeployment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TeardownDeploymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeploymentsServer).TeardownDeployment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Deployments/TeardownDeployment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeploymentsServer).TeardownDeployment(ctx, req.(*TeardownDeploymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Deployments_serviceDesc = grpc.ServiceDesc{
	ServiceName: "soapbox.Deployments",
	HandlerType: (*DeploymentsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListDeployments",
			Handler:    _Deployments_ListDeployments_Handler,
		},
		{
			MethodName: "GetDeployment",
			Handler:    _Deployments_GetDeployment_Handler,
		},
		{
			MethodName: "GetLatestDeployment",
			Handler:    _Deployments_GetLatestDeployment_Handler,
		},
		{
			MethodName: "StartDeployment",
			Handler:    _Deployments_StartDeployment_Handler,
		},
		{
			MethodName: "GetDeploymentStatus",
			Handler:    _Deployments_GetDeploymentStatus_Handler,
		},
		{
			MethodName: "TeardownDeployment",
			Handler:    _Deployments_TeardownDeployment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "deployment.proto",
}

func init() { proto1.RegisterFile("deployment.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 601 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0xcd, 0x6e, 0xd3, 0x4c,
	0x14, 0x8d, 0x93, 0xa6, 0x55, 0x6f, 0xd4, 0x34, 0x9d, 0xa6, 0x5f, 0xfd, 0xb9, 0xd0, 0x46, 0x03,
	0x54, 0xe1, 0x47, 0xae, 0x94, 0x0a, 0x24, 0x36, 0x48, 0xa6, 0x76, 0xab, 0x54, 0x86, 0x80, 0xe3,
	0x50, 0xc1, 0x26, 0x72, 0xe2, 0xa1, 0x18, 0xc5, 0x1e, 0x93, 0x99, 0x14, 0xfa, 0x04, 0xbc, 0x03,
	0x6f, 0xca, 0x0e, 0xc5, 0x4e, 0xec, 0x89, 0xb1, 0x41, 0x62, 0x65, 0xcf, 0xbd, 0xe7, 0xfe, 0x9f,
	0x03, 0x0d, 0x97, 0x84, 0x13, 0x7a, 0xeb, 0x93, 0x80, 0xab, 0xe1, 0x94, 0x72, 0x8a, 0x36, 0x18,
	0x75, 0xc2, 0x11, 0xfd, 0xa6, 0x6c, 0x2d, 0x7e, 0x62, 0xbb, 0xb2, 0xe3, 0x84, 0xe1, 0xc4, 0x1b,
	0x3b, 0xdc, 0xa3, 0xc1, 0xd2, 0x44, 0x82, 0x1b, 0x6f, 0x4a, 0x83, 0x34, 0x5a, 0x39, 0xba, 0xa6,
	0xf4, 0x7a, 0x42, 0x4e, 0xa2, 0xd7, 0x68, 0xf6, 0xf1, 0x84, 0x7b, 0x3e, 0x61, 0xdc, 0xf1, 0xc3,
	0x18, 0x80, 0x5f, 0xc0, 0x9e, 0xe9, 0x31, 0xae, 0x27, 0x65, 0x2d, 0xf2, 0x65, 0x46, 0x18, 0x47,
	0x0f, 0xa0, 0x2e, 0x54, 0x18, 0x7a, 0xae, 0x2c, 0xb5, 0xa4, 0x76, 0xd5, 0xda, 0x12, 0xac, 0x5d,
	0x17, 0xf7, 0xe0, 0xbf, 0x6c, 0x3c, 0x0b, 0x69, 0xc0, 0x08, 0x7a, 0x0a, 0xb5, 0x74, 0x18, 0x26,
	0x4b, 0xad, 0x4a, 0xbb, 0xd6, 0xd9, 0x55, 0x97, 0x53, 0x08, 0x11, 0x22, 0x0e, 0x1f, 0x43, 0xf3,
	0x82, 0xe4, 0xf4, 0x53, 0x87, 0x72, 0xd2, 0x43, 0xd9, 0x73, 0xf1, 0x67, 0x50, 0x2e, 0x08, 0x37,
	0x1d, 0x4e, 0xfe, 0xbd, 0xfb, 0x39, 0x4c, 0xd8, 0xd9, 0x1c, 0x56, 0x8e, 0x61, 0x82, 0xb5, 0xeb,
	0xe2, 0xef, 0x65, 0x80, 0xb4, 0x46, 0xb6, 0x15, 0xf4, 0x0c, 0x6a, 0x42, 0xda, 0x28, 0x45, 0xad,
	0xd3, 0x4c, 0x26, 0xd5, 0x52, 0x9f, 0x25, 0x02, 0xd1, 0x31, 0x54, 0x48, 0x70, 0x23, 0x57, 0x32,
	0x78, 0x23, 0xad, 0x6d, 0xcd, 0x01, 0xe8, 0x10, 0x60, 0x4c, 0x7d, 0xdf, 0xe3, 0xdc, 0x63, 0x9f,
	0xe4, 0xb5, 0x96, 0xd4, 0xde, 0xb4, 0x04, 0x0b, 0x52, 0xa1, 0xca, 0xb8, 0xc3, 0x89, 0x5c, 0x6d,
	0x49, 0xed, 0x7a, 0x47, 0xce, 0xd9, 0x71, 0x7f, 0xee, 0xb7, 0x62, 0x18, 0x7a, 0x0e, 0x30, 0x9e,
	0x12, 0x87, 0x13, 0x77, 0xe8, 0x70, 0x79, 0x3d, 0x2a, 0xaf, 0xa8, 0x31, 0x53, 0xd4, 0x25, 0x53,
	0x54, 0x7b, 0xc9, 0x14, 0x6b, 0x73, 0x81, 0xd6, 0x38, 0x7e, 0x08, 0xfb, 0x7d, 0xee, 0x4c, 0xf3,
	0xee, 0x9d, 0x3d, 0xd0, 0x93, 0xe8, 0x40, 0xab, 0x2d, 0xcc, 0x58, 0xd1, 0x39, 0x4f, 0xe1, 0x20,
	0x17, 0xbd, 0x48, 0xde, 0x5c, 0x8e, 0x28, 0x45, 0xd3, 0xc7, 0x0f, 0xfc, 0x18, 0xfe, 0xb7, 0x89,
	0x33, 0x75, 0xe9, 0xd7, 0xe0, 0xaf, 0x84, 0x79, 0xf4, 0x43, 0x82, 0xed, 0xcc, 0x42, 0xd0, 0x01,
	0xec, 0xeb, 0xc6, 0x1b, 0xb3, 0xf7, 0xfe, 0x95, 0xf1, 0xda, 0x1e, 0x5a, 0x3d, 0xd3, 0xec, 0x0d,
	0xec, 0xe1, 0x95, 0xd6, 0xb5, 0x1b, 0x25, 0x74, 0x07, 0x64, 0xc1, 0x69, 0xbc, 0xd3, 0xcc, 0x81,
	0x66, 0x1b, 0xb1, 0x57, 0xca, 0x09, 0x1d, 0x9e, 0xf7, 0xac, 0x2b, 0xcd, 0xd2, 0x1b, 0x65, 0x24,
	0x43, 0x53, 0x70, 0xf6, 0x07, 0x67, 0x67, 0x86, 0xa1, 0x1b, 0x7a, 0xa3, 0x82, 0xf6, 0x60, 0x47,
	0xf0, 0x9c, 0x6b, 0x5d, 0xd3, 0xd0, 0x1b, 0x6b, 0x9d, 0x9f, 0x15, 0xa8, 0xa5, 0xcd, 0x31, 0x64,
	0xc3, 0xf6, 0xaa, 0xac, 0x18, 0x3a, 0x4c, 0xce, 0x9a, 0x2b, 0x58, 0xe5, 0xa8, 0xd0, 0x1f, 0xef,
	0x10, 0x97, 0x90, 0x01, 0x5b, 0x2b, 0x4b, 0x46, 0x77, 0x93, 0x98, 0x3c, 0xcd, 0x29, 0x79, 0x6a,
	0xc5, 0x25, 0xf4, 0x16, 0x76, 0x73, 0xa4, 0x87, 0xee, 0x89, 0xc9, 0x0a, 0x84, 0x59, 0x94, 0xf2,
	0x12, 0xb6, 0x33, 0xbc, 0x42, 0x79, 0x48, 0xa5, 0x95, 0x18, 0x0b, 0x68, 0x88, 0x4b, 0x68, 0x14,
	0xb5, 0x97, 0xa5, 0xd2, 0x6a, 0x7b, 0x05, 0xb4, 0x54, 0xee, 0xff, 0x19, 0x94, 0xd4, 0xb8, 0x04,
	0xf4, 0x3b, 0xf3, 0x10, 0x4e, 0xa2, 0x0b, 0x69, 0xa9, 0xd4, 0x53, 0x9d, 0xfb, 0x21, 0xbf, 0xc5,
	0xa5, 0x97, 0x1b, 0x1f, 0xaa, 0xb1, 0xe8, 0xd6, 0xa3, 0xcf, 0xe9, 0xaf, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x69, 0x5f, 0xc1, 0x6a, 0x05, 0x06, 0x00, 0x00,
}
