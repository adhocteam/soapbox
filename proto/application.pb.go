// Code generated by protoc-gen-go. DO NOT EDIT.
// source: application.proto

package proto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import timestamp "github.com/golang/protobuf/ptypes/timestamp"

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

type ApplicationType int32

const (
	ApplicationType_SERVER  ApplicationType = 0
	ApplicationType_CRONJOB ApplicationType = 1
)

var ApplicationType_name = map[int32]string{
	0: "SERVER",
	1: "CRONJOB",
}
var ApplicationType_value = map[string]int32{
	"SERVER":  0,
	"CRONJOB": 1,
}

func (x ApplicationType) String() string {
	return proto.EnumName(ApplicationType_name, int32(x))
}
func (ApplicationType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{0}
}

type MetricType int32

const (
	MetricType_REQUEST_COUNT  MetricType = 0
	MetricType_LATENCY        MetricType = 1
	MetricType_HTTP_5XX_COUNT MetricType = 2
	MetricType_HTTP_4XX_COUNT MetricType = 3
	MetricType_HTTP_2XX_COUNT MetricType = 4
)

var MetricType_name = map[int32]string{
	0: "REQUEST_COUNT",
	1: "LATENCY",
	2: "HTTP_5XX_COUNT",
	3: "HTTP_4XX_COUNT",
	4: "HTTP_2XX_COUNT",
}
var MetricType_value = map[string]int32{
	"REQUEST_COUNT":  0,
	"LATENCY":        1,
	"HTTP_5XX_COUNT": 2,
	"HTTP_4XX_COUNT": 3,
	"HTTP_2XX_COUNT": 4,
}

func (x MetricType) String() string {
	return proto.EnumName(MetricType_name, int32(x))
}
func (MetricType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{1}
}

type CreationState int32

const (
	CreationState_CREATE_INFRASTRUCTURE_WAIT      CreationState = 0
	CreationState_CREATE_INFRASTRUCTURE_SUCCEEDED CreationState = 1
	CreationState_CREATE_INFRASTRUCTURE_FAILED    CreationState = 2
)

var CreationState_name = map[int32]string{
	0: "CREATE_INFRASTRUCTURE_WAIT",
	1: "CREATE_INFRASTRUCTURE_SUCCEEDED",
	2: "CREATE_INFRASTRUCTURE_FAILED",
}
var CreationState_value = map[string]int32{
	"CREATE_INFRASTRUCTURE_WAIT":      0,
	"CREATE_INFRASTRUCTURE_SUCCEEDED": 1,
	"CREATE_INFRASTRUCTURE_FAILED":    2,
}

func (x CreationState) String() string {
	return proto.EnumName(CreationState_name, int32(x))
}
func (CreationState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{2}
}

type DeletionState int32

const (
	DeletionState_NOT_DELETED                     DeletionState = 0
	DeletionState_DELETE_INFRASTRUCTURE_WAIT      DeletionState = 1
	DeletionState_DELETE_INFRASTRUCTURE_SUCCEEDED DeletionState = 2
	DeletionState_DELETE_INFRASTRUCTURE_FAILED    DeletionState = 3
)

var DeletionState_name = map[int32]string{
	0: "NOT_DELETED",
	1: "DELETE_INFRASTRUCTURE_WAIT",
	2: "DELETE_INFRASTRUCTURE_SUCCEEDED",
	3: "DELETE_INFRASTRUCTURE_FAILED",
}
var DeletionState_value = map[string]int32{
	"NOT_DELETED":                     0,
	"DELETE_INFRASTRUCTURE_WAIT":      1,
	"DELETE_INFRASTRUCTURE_SUCCEEDED": 2,
	"DELETE_INFRASTRUCTURE_FAILED":    3,
}

func (x DeletionState) String() string {
	return proto.EnumName(DeletionState_name, int32(x))
}
func (DeletionState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{3}
}

type Application struct {
	Id                   int32                `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	UserId               int32                `protobuf:"varint,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Name                 string               `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Description          string               `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	ExternalDns          string               `protobuf:"bytes,5,opt,name=external_dns,json=externalDns,proto3" json:"external_dns,omitempty"`
	GithubRepoUrl        string               `protobuf:"bytes,6,opt,name=github_repo_url,json=githubRepoUrl,proto3" json:"github_repo_url,omitempty"`
	DockerfilePath       string               `protobuf:"bytes,7,opt,name=dockerfile_path,json=dockerfilePath,proto3" json:"dockerfile_path,omitempty"`
	EntrypointOverride   string               `protobuf:"bytes,8,opt,name=entrypoint_override,json=entrypointOverride,proto3" json:"entrypoint_override,omitempty"`
	Type                 ApplicationType      `protobuf:"varint,9,opt,name=type,proto3,enum=soapbox.ApplicationType" json:"type,omitempty"`
	CreatedAt            *timestamp.Timestamp `protobuf:"bytes,10,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Slug                 string               `protobuf:"bytes,11,opt,name=slug,proto3" json:"slug,omitempty"`
	InternalDns          string               `protobuf:"bytes,12,opt,name=internal_dns,json=internalDns,proto3" json:"internal_dns,omitempty"`
	CreationState        CreationState        `protobuf:"varint,13,opt,name=creation_state,json=creationState,proto3,enum=soapbox.CreationState" json:"creation_state,omitempty"`
	DeletionState        DeletionState        `protobuf:"varint,14,opt,name=deletion_state,json=deletionState,proto3,enum=soapbox.DeletionState" json:"deletion_state,omitempty"`
	AwsEncryptionKeyArn  string               `protobuf:"bytes,15,opt,name=aws_encryption_key_arn,json=awsEncryptionKeyArn,proto3" json:"aws_encryption_key_arn,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Application) Reset()         { *m = Application{} }
func (m *Application) String() string { return proto.CompactTextString(m) }
func (*Application) ProtoMessage()    {}
func (*Application) Descriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{0}
}
func (m *Application) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Application.Unmarshal(m, b)
}
func (m *Application) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Application.Marshal(b, m, deterministic)
}
func (dst *Application) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Application.Merge(dst, src)
}
func (m *Application) XXX_Size() int {
	return xxx_messageInfo_Application.Size(m)
}
func (m *Application) XXX_DiscardUnknown() {
	xxx_messageInfo_Application.DiscardUnknown(m)
}

var xxx_messageInfo_Application proto.InternalMessageInfo

func (m *Application) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Application) GetUserId() int32 {
	if m != nil {
		return m.UserId
	}
	return 0
}

func (m *Application) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Application) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Application) GetExternalDns() string {
	if m != nil {
		return m.ExternalDns
	}
	return ""
}

func (m *Application) GetGithubRepoUrl() string {
	if m != nil {
		return m.GithubRepoUrl
	}
	return ""
}

func (m *Application) GetDockerfilePath() string {
	if m != nil {
		return m.DockerfilePath
	}
	return ""
}

func (m *Application) GetEntrypointOverride() string {
	if m != nil {
		return m.EntrypointOverride
	}
	return ""
}

func (m *Application) GetType() ApplicationType {
	if m != nil {
		return m.Type
	}
	return ApplicationType_SERVER
}

func (m *Application) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *Application) GetSlug() string {
	if m != nil {
		return m.Slug
	}
	return ""
}

func (m *Application) GetInternalDns() string {
	if m != nil {
		return m.InternalDns
	}
	return ""
}

func (m *Application) GetCreationState() CreationState {
	if m != nil {
		return m.CreationState
	}
	return CreationState_CREATE_INFRASTRUCTURE_WAIT
}

func (m *Application) GetDeletionState() DeletionState {
	if m != nil {
		return m.DeletionState
	}
	return DeletionState_NOT_DELETED
}

func (m *Application) GetAwsEncryptionKeyArn() string {
	if m != nil {
		return m.AwsEncryptionKeyArn
	}
	return ""
}

type ListApplicationRequest struct {
	UserId               int32    `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListApplicationRequest) Reset()         { *m = ListApplicationRequest{} }
func (m *ListApplicationRequest) String() string { return proto.CompactTextString(m) }
func (*ListApplicationRequest) ProtoMessage()    {}
func (*ListApplicationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{1}
}
func (m *ListApplicationRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListApplicationRequest.Unmarshal(m, b)
}
func (m *ListApplicationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListApplicationRequest.Marshal(b, m, deterministic)
}
func (dst *ListApplicationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListApplicationRequest.Merge(dst, src)
}
func (m *ListApplicationRequest) XXX_Size() int {
	return xxx_messageInfo_ListApplicationRequest.Size(m)
}
func (m *ListApplicationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListApplicationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListApplicationRequest proto.InternalMessageInfo

func (m *ListApplicationRequest) GetUserId() int32 {
	if m != nil {
		return m.UserId
	}
	return 0
}

type ListApplicationResponse struct {
	Applications         []*Application `protobuf:"bytes,1,rep,name=applications,proto3" json:"applications,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ListApplicationResponse) Reset()         { *m = ListApplicationResponse{} }
func (m *ListApplicationResponse) String() string { return proto.CompactTextString(m) }
func (*ListApplicationResponse) ProtoMessage()    {}
func (*ListApplicationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{2}
}
func (m *ListApplicationResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListApplicationResponse.Unmarshal(m, b)
}
func (m *ListApplicationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListApplicationResponse.Marshal(b, m, deterministic)
}
func (dst *ListApplicationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListApplicationResponse.Merge(dst, src)
}
func (m *ListApplicationResponse) XXX_Size() int {
	return xxx_messageInfo_ListApplicationResponse.Size(m)
}
func (m *ListApplicationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListApplicationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListApplicationResponse proto.InternalMessageInfo

func (m *ListApplicationResponse) GetApplications() []*Application {
	if m != nil {
		return m.Applications
	}
	return nil
}

type GetApplicationRequest struct {
	Id                   int32    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetApplicationRequest) Reset()         { *m = GetApplicationRequest{} }
func (m *GetApplicationRequest) String() string { return proto.CompactTextString(m) }
func (*GetApplicationRequest) ProtoMessage()    {}
func (*GetApplicationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{3}
}
func (m *GetApplicationRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetApplicationRequest.Unmarshal(m, b)
}
func (m *GetApplicationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetApplicationRequest.Marshal(b, m, deterministic)
}
func (dst *GetApplicationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetApplicationRequest.Merge(dst, src)
}
func (m *GetApplicationRequest) XXX_Size() int {
	return xxx_messageInfo_GetApplicationRequest.Size(m)
}
func (m *GetApplicationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetApplicationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetApplicationRequest proto.InternalMessageInfo

func (m *GetApplicationRequest) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

type ApplicationMetric struct {
	Time                 string   `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
	Count                int32    `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ApplicationMetric) Reset()         { *m = ApplicationMetric{} }
func (m *ApplicationMetric) String() string { return proto.CompactTextString(m) }
func (*ApplicationMetric) ProtoMessage()    {}
func (*ApplicationMetric) Descriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{4}
}
func (m *ApplicationMetric) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApplicationMetric.Unmarshal(m, b)
}
func (m *ApplicationMetric) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApplicationMetric.Marshal(b, m, deterministic)
}
func (dst *ApplicationMetric) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApplicationMetric.Merge(dst, src)
}
func (m *ApplicationMetric) XXX_Size() int {
	return xxx_messageInfo_ApplicationMetric.Size(m)
}
func (m *ApplicationMetric) XXX_DiscardUnknown() {
	xxx_messageInfo_ApplicationMetric.DiscardUnknown(m)
}

var xxx_messageInfo_ApplicationMetric proto.InternalMessageInfo

func (m *ApplicationMetric) GetTime() string {
	if m != nil {
		return m.Time
	}
	return ""
}

func (m *ApplicationMetric) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

type ApplicationMetricsResponse struct {
	Metrics              []*ApplicationMetric `protobuf:"bytes,1,rep,name=metrics,proto3" json:"metrics,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *ApplicationMetricsResponse) Reset()         { *m = ApplicationMetricsResponse{} }
func (m *ApplicationMetricsResponse) String() string { return proto.CompactTextString(m) }
func (*ApplicationMetricsResponse) ProtoMessage()    {}
func (*ApplicationMetricsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{5}
}
func (m *ApplicationMetricsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ApplicationMetricsResponse.Unmarshal(m, b)
}
func (m *ApplicationMetricsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ApplicationMetricsResponse.Marshal(b, m, deterministic)
}
func (dst *ApplicationMetricsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ApplicationMetricsResponse.Merge(dst, src)
}
func (m *ApplicationMetricsResponse) XXX_Size() int {
	return xxx_messageInfo_ApplicationMetricsResponse.Size(m)
}
func (m *ApplicationMetricsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ApplicationMetricsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ApplicationMetricsResponse proto.InternalMessageInfo

func (m *ApplicationMetricsResponse) GetMetrics() []*ApplicationMetric {
	if m != nil {
		return m.Metrics
	}
	return nil
}

type GetApplicationMetricsRequest struct {
	Id                   int32      `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	MetricType           MetricType `protobuf:"varint,2,opt,name=metric_type,json=metricType,proto3,enum=soapbox.MetricType" json:"metric_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *GetApplicationMetricsRequest) Reset()         { *m = GetApplicationMetricsRequest{} }
func (m *GetApplicationMetricsRequest) String() string { return proto.CompactTextString(m) }
func (*GetApplicationMetricsRequest) ProtoMessage()    {}
func (*GetApplicationMetricsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_application_323a0ad033092da1, []int{6}
}
func (m *GetApplicationMetricsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetApplicationMetricsRequest.Unmarshal(m, b)
}
func (m *GetApplicationMetricsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetApplicationMetricsRequest.Marshal(b, m, deterministic)
}
func (dst *GetApplicationMetricsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetApplicationMetricsRequest.Merge(dst, src)
}
func (m *GetApplicationMetricsRequest) XXX_Size() int {
	return xxx_messageInfo_GetApplicationMetricsRequest.Size(m)
}
func (m *GetApplicationMetricsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetApplicationMetricsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetApplicationMetricsRequest proto.InternalMessageInfo

func (m *GetApplicationMetricsRequest) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *GetApplicationMetricsRequest) GetMetricType() MetricType {
	if m != nil {
		return m.MetricType
	}
	return MetricType_REQUEST_COUNT
}

func init() {
	proto.RegisterType((*Application)(nil), "soapbox.Application")
	proto.RegisterType((*ListApplicationRequest)(nil), "soapbox.ListApplicationRequest")
	proto.RegisterType((*ListApplicationResponse)(nil), "soapbox.ListApplicationResponse")
	proto.RegisterType((*GetApplicationRequest)(nil), "soapbox.GetApplicationRequest")
	proto.RegisterType((*ApplicationMetric)(nil), "soapbox.ApplicationMetric")
	proto.RegisterType((*ApplicationMetricsResponse)(nil), "soapbox.ApplicationMetricsResponse")
	proto.RegisterType((*GetApplicationMetricsRequest)(nil), "soapbox.GetApplicationMetricsRequest")
	proto.RegisterEnum("soapbox.ApplicationType", ApplicationType_name, ApplicationType_value)
	proto.RegisterEnum("soapbox.MetricType", MetricType_name, MetricType_value)
	proto.RegisterEnum("soapbox.CreationState", CreationState_name, CreationState_value)
	proto.RegisterEnum("soapbox.DeletionState", DeletionState_name, DeletionState_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ApplicationsClient is the client API for Applications service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ApplicationsClient interface {
	ListApplications(ctx context.Context, in *ListApplicationRequest, opts ...grpc.CallOption) (*ListApplicationResponse, error)
	CreateApplication(ctx context.Context, in *Application, opts ...grpc.CallOption) (*Application, error)
	GetApplication(ctx context.Context, in *GetApplicationRequest, opts ...grpc.CallOption) (*Application, error)
	DeleteApplication(ctx context.Context, in *Application, opts ...grpc.CallOption) (*Empty, error)
	GetApplicationMetrics(ctx context.Context, in *GetApplicationMetricsRequest, opts ...grpc.CallOption) (*ApplicationMetricsResponse, error)
}

type applicationsClient struct {
	cc *grpc.ClientConn
}

func NewApplicationsClient(cc *grpc.ClientConn) ApplicationsClient {
	return &applicationsClient{cc}
}

func (c *applicationsClient) ListApplications(ctx context.Context, in *ListApplicationRequest, opts ...grpc.CallOption) (*ListApplicationResponse, error) {
	out := new(ListApplicationResponse)
	err := c.cc.Invoke(ctx, "/soapbox.Applications/ListApplications", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationsClient) CreateApplication(ctx context.Context, in *Application, opts ...grpc.CallOption) (*Application, error) {
	out := new(Application)
	err := c.cc.Invoke(ctx, "/soapbox.Applications/CreateApplication", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationsClient) GetApplication(ctx context.Context, in *GetApplicationRequest, opts ...grpc.CallOption) (*Application, error) {
	out := new(Application)
	err := c.cc.Invoke(ctx, "/soapbox.Applications/GetApplication", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationsClient) DeleteApplication(ctx context.Context, in *Application, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/soapbox.Applications/DeleteApplication", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationsClient) GetApplicationMetrics(ctx context.Context, in *GetApplicationMetricsRequest, opts ...grpc.CallOption) (*ApplicationMetricsResponse, error) {
	out := new(ApplicationMetricsResponse)
	err := c.cc.Invoke(ctx, "/soapbox.Applications/GetApplicationMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ApplicationsServer is the server API for Applications service.
type ApplicationsServer interface {
	ListApplications(context.Context, *ListApplicationRequest) (*ListApplicationResponse, error)
	CreateApplication(context.Context, *Application) (*Application, error)
	GetApplication(context.Context, *GetApplicationRequest) (*Application, error)
	DeleteApplication(context.Context, *Application) (*Empty, error)
	GetApplicationMetrics(context.Context, *GetApplicationMetricsRequest) (*ApplicationMetricsResponse, error)
}

func RegisterApplicationsServer(s *grpc.Server, srv ApplicationsServer) {
	s.RegisterService(&_Applications_serviceDesc, srv)
}

func _Applications_ListApplications_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListApplicationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationsServer).ListApplications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Applications/ListApplications",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationsServer).ListApplications(ctx, req.(*ListApplicationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Applications_CreateApplication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Application)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationsServer).CreateApplication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Applications/CreateApplication",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationsServer).CreateApplication(ctx, req.(*Application))
	}
	return interceptor(ctx, in, info, handler)
}

func _Applications_GetApplication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetApplicationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationsServer).GetApplication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Applications/GetApplication",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationsServer).GetApplication(ctx, req.(*GetApplicationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Applications_DeleteApplication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Application)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationsServer).DeleteApplication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Applications/DeleteApplication",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationsServer).DeleteApplication(ctx, req.(*Application))
	}
	return interceptor(ctx, in, info, handler)
}

func _Applications_GetApplicationMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetApplicationMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationsServer).GetApplicationMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/soapbox.Applications/GetApplicationMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationsServer).GetApplicationMetrics(ctx, req.(*GetApplicationMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Applications_serviceDesc = grpc.ServiceDesc{
	ServiceName: "soapbox.Applications",
	HandlerType: (*ApplicationsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListApplications",
			Handler:    _Applications_ListApplications_Handler,
		},
		{
			MethodName: "CreateApplication",
			Handler:    _Applications_CreateApplication_Handler,
		},
		{
			MethodName: "GetApplication",
			Handler:    _Applications_GetApplication_Handler,
		},
		{
			MethodName: "DeleteApplication",
			Handler:    _Applications_DeleteApplication_Handler,
		},
		{
			MethodName: "GetApplicationMetrics",
			Handler:    _Applications_GetApplicationMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "application.proto",
}

func init() { proto.RegisterFile("application.proto", fileDescriptor_application_323a0ad033092da1) }

var fileDescriptor_application_323a0ad033092da1 = []byte{
	// 864 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x55, 0xed, 0x6e, 0xdb, 0x36,
	0x14, 0xb5, 0x6c, 0x27, 0x5e, 0xae, 0xe3, 0x2f, 0xa6, 0x4b, 0x05, 0xa1, 0x68, 0x3c, 0x15, 0x5b,
	0x0d, 0x63, 0x70, 0x30, 0x37, 0x03, 0x36, 0x0c, 0xfd, 0xa1, 0xda, 0xec, 0x9a, 0x2d, 0xb3, 0x3b,
	0x5a, 0x5e, 0xbb, 0xfd, 0x21, 0x14, 0x8b, 0x4d, 0x84, 0xca, 0x92, 0x26, 0xd2, 0x6d, 0xfd, 0x04,
	0x7b, 0x85, 0x3d, 0xe0, 0x1e, 0x64, 0x10, 0x29, 0xcb, 0xb2, 0xa3, 0x00, 0xfd, 0x15, 0xf2, 0xdc,
	0x73, 0xa9, 0x93, 0x73, 0xef, 0x49, 0xa0, 0xe3, 0x44, 0x91, 0xef, 0x2d, 0x1c, 0xe1, 0x85, 0xc1,
	0x20, 0x8a, 0x43, 0x11, 0xa2, 0x1a, 0x0f, 0x9d, 0xe8, 0x3a, 0xfc, 0x64, 0x34, 0xd2, 0x83, 0xc2,
	0x8d, 0xb3, 0x9b, 0x30, 0xbc, 0xf1, 0xd9, 0xb9, 0xbc, 0x5d, 0xaf, 0xde, 0x9d, 0x0b, 0x6f, 0xc9,
	0xb8, 0x70, 0x96, 0x91, 0x22, 0x98, 0xff, 0x55, 0xa1, 0x6e, 0x6d, 0x9f, 0x43, 0x4d, 0x28, 0x7b,
	0xae, 0xae, 0x75, 0xb5, 0xde, 0x01, 0x29, 0x7b, 0x2e, 0x7a, 0x08, 0xb5, 0x15, 0x67, 0x31, 0xf5,
	0x5c, 0xbd, 0x2c, 0xc1, 0xc3, 0xe4, 0x7a, 0xe9, 0x22, 0x04, 0xd5, 0xc0, 0x59, 0x32, 0xbd, 0xd2,
	0xd5, 0x7a, 0x47, 0x44, 0x9e, 0x51, 0x17, 0xea, 0x2e, 0xe3, 0x8b, 0xd8, 0x8b, 0x92, 0xb7, 0xf4,
	0xaa, 0x2c, 0xe5, 0x21, 0xf4, 0x15, 0x1c, 0xb3, 0x4f, 0x82, 0xc5, 0x81, 0xe3, 0x53, 0x37, 0xe0,
	0xfa, 0x81, 0xa2, 0x6c, 0xb0, 0x71, 0xc0, 0xd1, 0x37, 0xd0, 0xba, 0xf1, 0xc4, 0xed, 0xea, 0x9a,
	0xc6, 0x2c, 0x0a, 0xe9, 0x2a, 0xf6, 0xf5, 0x43, 0xc9, 0x6a, 0x28, 0x98, 0xb0, 0x28, 0x9c, 0xc7,
	0x3e, 0x7a, 0x0a, 0x2d, 0x37, 0x5c, 0xbc, 0x67, 0xf1, 0x3b, 0xcf, 0x67, 0x34, 0x72, 0xc4, 0xad,
	0x5e, 0x93, 0xbc, 0xe6, 0x16, 0x7e, 0xed, 0x88, 0x5b, 0x74, 0x0e, 0x27, 0x2c, 0x10, 0xf1, 0x3a,
	0x0a, 0xbd, 0x40, 0xd0, 0xf0, 0x03, 0x8b, 0x63, 0xcf, 0x65, 0xfa, 0x17, 0x92, 0x8c, 0xb6, 0xa5,
	0x69, 0x5a, 0x41, 0xdf, 0x42, 0x55, 0xac, 0x23, 0xa6, 0x1f, 0x75, 0xb5, 0x5e, 0x73, 0xa8, 0x0f,
	0x36, 0x96, 0xe6, 0x7c, 0xb2, 0xd7, 0x11, 0x23, 0x92, 0x85, 0x7e, 0x04, 0x58, 0xc4, 0xcc, 0x11,
	0xcc, 0xa5, 0x8e, 0xd0, 0xa1, 0xab, 0xf5, 0xea, 0x43, 0x63, 0xa0, 0x7c, 0x1f, 0x6c, 0x7c, 0x1f,
	0xd8, 0x1b, 0xdf, 0xc9, 0x51, 0xca, 0xb6, 0x44, 0xe2, 0x21, 0xf7, 0x57, 0x37, 0x7a, 0x5d, 0x79,
	0x98, 0x9c, 0x13, 0x87, 0xbc, 0x20, 0xe7, 0xd0, 0xb1, 0x72, 0x68, 0x83, 0x25, 0x0e, 0x3d, 0x87,
	0xa6, 0x7c, 0xc3, 0x0b, 0x03, 0xca, 0x85, 0x23, 0x98, 0xde, 0x90, 0x4a, 0x4f, 0x33, 0xa5, 0xa3,
	0xb4, 0x3c, 0x4b, 0xaa, 0xa4, 0xb1, 0xc8, 0x5f, 0x93, 0x76, 0x97, 0xf9, 0x2c, 0xd7, 0xde, 0xdc,
	0x6b, 0x1f, 0xa7, 0xe5, 0xb4, 0xdd, 0xcd, 0x5f, 0xd1, 0x33, 0x38, 0x75, 0x3e, 0x72, 0xca, 0x82,
	0x45, 0xbc, 0x96, 0x43, 0xa5, 0xef, 0xd9, 0x9a, 0x3a, 0x71, 0xa0, 0xb7, 0xa4, 0xd4, 0x13, 0xe7,
	0x23, 0xc7, 0x59, 0xf1, 0x57, 0xb6, 0xb6, 0xe2, 0xc0, 0xfc, 0x0e, 0x4e, 0xaf, 0x3c, 0x2e, 0x72,
	0x0e, 0x12, 0xf6, 0xf7, 0x8a, 0x71, 0x91, 0x5f, 0x30, 0x2d, 0xbf, 0x60, 0xe6, 0x0c, 0x1e, 0xde,
	0x69, 0xe1, 0x51, 0x18, 0x70, 0x86, 0x7e, 0x80, 0xe3, 0x5c, 0x04, 0xb8, 0xae, 0x75, 0x2b, 0xbd,
	0xfa, 0xf0, 0x41, 0xd1, 0xa0, 0xc8, 0x0e, 0xd3, 0x7c, 0x0a, 0x5f, 0xfe, 0xcc, 0x8a, 0x64, 0xec,
	0xed, 0xbd, 0xf9, 0x1c, 0x3a, 0x39, 0xd6, 0x6f, 0x4c, 0xc4, 0xde, 0x22, 0x99, 0x57, 0x92, 0x1f,
	0x49, 0x3b, 0x22, 0xf2, 0x8c, 0x1e, 0xc0, 0xc1, 0x22, 0x5c, 0x05, 0x22, 0x8d, 0x87, 0xba, 0x98,
	0x04, 0x8c, 0x3b, 0xed, 0x3c, 0xd3, 0x7f, 0x01, 0xb5, 0xa5, 0x82, 0x52, 0xe9, 0x46, 0x91, 0x74,
	0xd5, 0x45, 0x36, 0x54, 0xd3, 0x85, 0x47, 0xbb, 0xda, 0xb3, 0x67, 0x0b, 0x7f, 0x05, 0x74, 0x01,
	0x75, 0xd5, 0x4a, 0xe5, 0x36, 0x97, 0xe5, 0x90, 0x4f, 0xb2, 0x2f, 0xa9, 0x6e, 0xb9, 0xc8, 0xb0,
	0xcc, 0xce, 0xfd, 0x3e, 0xb4, 0xf6, 0xf6, 0x1c, 0x01, 0x1c, 0xce, 0x30, 0xf9, 0x03, 0x93, 0x76,
	0x09, 0xd5, 0xa1, 0x36, 0x22, 0xd3, 0xc9, 0x2f, 0xd3, 0x17, 0x6d, 0xad, 0x7f, 0x0b, 0xb0, 0x7d,
	0x05, 0x75, 0xa0, 0x41, 0xf0, 0xef, 0x73, 0x3c, 0xb3, 0xe9, 0x68, 0x3a, 0x9f, 0xd8, 0x8a, 0x7d,
	0x65, 0xd9, 0x78, 0x32, 0xfa, 0xb3, 0xad, 0x21, 0x04, 0xcd, 0x57, 0xb6, 0xfd, 0x9a, 0x7e, 0xff,
	0xf6, 0x6d, 0x4a, 0x28, 0x67, 0xd8, 0x45, 0x86, 0x55, 0x32, 0x6c, 0x98, 0x61, 0xd5, 0xfe, 0x07,
	0x68, 0xec, 0xec, 0x34, 0x7a, 0x0c, 0xc6, 0x88, 0x60, 0xcb, 0xc6, 0xf4, 0x72, 0xf2, 0x92, 0x58,
	0x33, 0x9b, 0xcc, 0x47, 0xf6, 0x9c, 0x60, 0xfa, 0xc6, 0xba, 0x4c, 0xbe, 0xfc, 0x04, 0xce, 0x8a,
	0xeb, 0xb3, 0xf9, 0x68, 0x84, 0xf1, 0x18, 0x8f, 0xdb, 0x1a, 0xea, 0xc2, 0xa3, 0x62, 0xd2, 0x4b,
	0xeb, 0xf2, 0x0a, 0x8f, 0xdb, 0xe5, 0xfe, 0x3f, 0x1a, 0x34, 0x76, 0xd2, 0x80, 0x5a, 0x50, 0x9f,
	0x4c, 0x6d, 0x3a, 0xc6, 0x57, 0xd8, 0xc6, 0xe3, 0x76, 0x29, 0x51, 0xa2, 0x2e, 0x85, 0x4a, 0xb4,
	0x44, 0x49, 0x71, 0x7d, 0xab, 0xa4, 0x9c, 0x28, 0x29, 0x26, 0xa5, 0x4a, 0x2a, 0xc3, 0x7f, 0x2b,
	0x70, 0x9c, 0x1b, 0x0c, 0x47, 0x6f, 0xa0, 0xbd, 0x97, 0x0f, 0x8e, 0xce, 0xb2, 0xe9, 0x16, 0xa7,
	0xcd, 0xe8, 0xde, 0x4f, 0x50, 0xbb, 0x69, 0x96, 0x90, 0x05, 0x1d, 0xe9, 0x35, 0xcb, 0xff, 0x5f,
	0x28, 0x0c, 0x97, 0x51, 0x88, 0x9a, 0x25, 0xf4, 0x0a, 0x9a, 0xbb, 0xab, 0x8a, 0x1e, 0x67, 0xcc,
	0xc2, 0xfc, 0xdd, 0xfb, 0xd2, 0x4f, 0xd0, 0x91, 0xfe, 0x7f, 0x86, 0x98, 0x66, 0x86, 0xe2, 0x65,
	0x24, 0xd6, 0x66, 0x09, 0xb1, 0xfd, 0xb4, 0xa7, 0x89, 0x41, 0x5f, 0xdf, 0xa3, 0x66, 0x37, 0x51,
	0xc6, 0x93, 0xfb, 0x63, 0xc9, 0xb7, 0x86, 0xbd, 0xa8, 0xfd, 0x75, 0xa0, 0xfe, 0xce, 0x1f, 0xca,
	0x1f, 0xcf, 0xfe, 0x0f, 0x00, 0x00, 0xff, 0xff, 0x1c, 0x61, 0x62, 0x22, 0xa1, 0x07, 0x00, 0x00,
}
