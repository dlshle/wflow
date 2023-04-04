// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: proto/wflow.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Status int32

const (
	Status_OK       Status = 0
	Status_INTERNAL Status = 1
	Status_INVALID  Status = 2
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "OK",
		1: "INTERNAL",
		2: "INVALID",
	}
	Status_value = map[string]int32{
		"OK":       0,
		"INTERNAL": 1,
		"INVALID":  2,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_wflow_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_proto_wflow_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_proto_wflow_proto_rawDescGZIP(), []int{0}
}

type Type int32

const (
	Type_PING                Type = 0
	Type_PONG                Type = 1
	Type_RESPONSE            Type = 2
	Type_DISPATCH_JOB        Type = 3   // from server only
	Type_QUERY_JOB           Type = 4   // from server only
	Type_CANCEL_JOB          Type = 5   // from server only
	Type_QUERY_WORKER_STATUS Type = 6   // from server only
	Type_DISCONNECT_WORKER   Type = 7   // from server only
	Type_CUSTOM              Type = 100 // custom message for both server and worker
)

// Enum value maps for Type.
var (
	Type_name = map[int32]string{
		0:   "PING",
		1:   "PONG",
		2:   "RESPONSE",
		3:   "DISPATCH_JOB",
		4:   "QUERY_JOB",
		5:   "CANCEL_JOB",
		6:   "QUERY_WORKER_STATUS",
		7:   "DISCONNECT_WORKER",
		100: "CUSTOM",
	}
	Type_value = map[string]int32{
		"PING":                0,
		"PONG":                1,
		"RESPONSE":            2,
		"DISPATCH_JOB":        3,
		"QUERY_JOB":           4,
		"CANCEL_JOB":          5,
		"QUERY_WORKER_STATUS": 6,
		"DISCONNECT_WORKER":   7,
		"CUSTOM":              100,
	}
)

func (x Type) Enum() *Type {
	p := new(Type)
	*p = x
	return p
}

func (x Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Type) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_wflow_proto_enumTypes[1].Descriptor()
}

func (Type) Type() protoreflect.EnumType {
	return &file_proto_wflow_proto_enumTypes[1]
}

func (x Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Type.Descriptor instead.
func (Type) EnumDescriptor() ([]byte, []int) {
	return file_proto_wflow_proto_rawDescGZIP(), []int{1}
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"` // uuid
	// string worker_id = 2;
	// string server_id = 3;
	Header  map[string]string `protobuf:"bytes,4,rep,name=header,proto3" json:"header,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Payload []byte            `protobuf:"bytes,5,opt,name=payload,proto3" json:"payload,omitempty"`
	Type    Type              `protobuf:"varint,6,opt,name=type,proto3,enum=com.github.dlshle.wflow.Type" json:"type,omitempty"`
	Status  Status            `protobuf:"varint,7,opt,name=status,proto3,enum=com.github.dlshle.wflow.Status" json:"status,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_wflow_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_proto_wflow_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_proto_wflow_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Message) GetHeader() map[string]string {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Message) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Message) GetType() Type {
	if x != nil {
		return x.Type
	}
	return Type_PING
}

func (x *Message) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_OK
}

type Job struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                    string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Description           *string `protobuf:"bytes,2,opt,name=description,proto3,oneof" json:"description,omitempty"`
	Param                 []byte  `protobuf:"bytes,3,opt,name=param,proto3" json:"param,omitempty"`
	DispatchTimeInSeconds int32   `protobuf:"varint,4,opt,name=dispatch_time_in_seconds,json=dispatchTimeInSeconds,proto3" json:"dispatch_time_in_seconds,omitempty"`
}

func (x *Job) Reset() {
	*x = Job{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_wflow_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Job) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Job) ProtoMessage() {}

func (x *Job) ProtoReflect() protoreflect.Message {
	mi := &file_proto_wflow_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Job.ProtoReflect.Descriptor instead.
func (*Job) Descriptor() ([]byte, []int) {
	return file_proto_wflow_proto_rawDescGZIP(), []int{1}
}

func (x *Job) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Job) GetDescription() string {
	if x != nil && x.Description != nil {
		return *x.Description
	}
	return ""
}

func (x *Job) GetParam() []byte {
	if x != nil {
		return x.Param
	}
	return nil
}

func (x *Job) GetDispatchTimeInSeconds() int32 {
	if x != nil {
		return x.DispatchTimeInSeconds
	}
	return 0
}

type Worker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id              string      `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	SystemStat      *SystemStat `protobuf:"bytes,2,opt,name=system_stat,json=systemStat,proto3" json:"system_stat,omitempty"`
	ActiveJobs      []string    `protobuf:"bytes,3,rep,name=active_jobs,json=activeJobs,proto3" json:"active_jobs,omitempty"`
	ConnectedServer *string     `protobuf:"bytes,4,opt,name=connected_server,json=connectedServer,proto3,oneof" json:"connected_server,omitempty"`
}

func (x *Worker) Reset() {
	*x = Worker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_wflow_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Worker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Worker) ProtoMessage() {}

func (x *Worker) ProtoReflect() protoreflect.Message {
	mi := &file_proto_wflow_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Worker.ProtoReflect.Descriptor instead.
func (*Worker) Descriptor() ([]byte, []int) {
	return file_proto_wflow_proto_rawDescGZIP(), []int{2}
}

func (x *Worker) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Worker) GetSystemStat() *SystemStat {
	if x != nil {
		return x.SystemStat
	}
	return nil
}

func (x *Worker) GetActiveJobs() []string {
	if x != nil {
		return x.ActiveJobs
	}
	return nil
}

func (x *Worker) GetConnectedServer() string {
	if x != nil && x.ConnectedServer != nil {
		return *x.ConnectedServer
	}
	return ""
}

type Server struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id               string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ConnectedWorkers []string `protobuf:"bytes,2,rep,name=connected_workers,json=connectedWorkers,proto3" json:"connected_workers,omitempty"`
	UptimeInSeconds  int32    `protobuf:"varint,3,opt,name=uptime_in_seconds,json=uptimeInSeconds,proto3" json:"uptime_in_seconds,omitempty"`
}

func (x *Server) Reset() {
	*x = Server{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_wflow_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Server) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Server) ProtoMessage() {}

func (x *Server) ProtoReflect() protoreflect.Message {
	mi := &file_proto_wflow_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Server.ProtoReflect.Descriptor instead.
func (*Server) Descriptor() ([]byte, []int) {
	return file_proto_wflow_proto_rawDescGZIP(), []int{3}
}

func (x *Server) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Server) GetConnectedWorkers() []string {
	if x != nil {
		return x.ConnectedWorkers
	}
	return nil
}

func (x *Server) GetUptimeInSeconds() int32 {
	if x != nil {
		return x.UptimeInSeconds
	}
	return 0
}

type SystemStat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CpuCount               int32 `protobuf:"varint,2,opt,name=cpu_count,json=cpuCount,proto3" json:"cpu_count,omitempty"`
	AvailableMemoryInBytes int32 `protobuf:"varint,3,opt,name=available_memory_in_bytes,json=availableMemoryInBytes,proto3" json:"available_memory_in_bytes,omitempty"`
	TotalMemoryInBytes     int32 `protobuf:"varint,4,opt,name=total_memory_in_bytes,json=totalMemoryInBytes,proto3" json:"total_memory_in_bytes,omitempty"`
}

func (x *SystemStat) Reset() {
	*x = SystemStat{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_wflow_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemStat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemStat) ProtoMessage() {}

func (x *SystemStat) ProtoReflect() protoreflect.Message {
	mi := &file_proto_wflow_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemStat.ProtoReflect.Descriptor instead.
func (*SystemStat) Descriptor() ([]byte, []int) {
	return file_proto_wflow_proto_rawDescGZIP(), []int{4}
}

func (x *SystemStat) GetCpuCount() int32 {
	if x != nil {
		return x.CpuCount
	}
	return 0
}

func (x *SystemStat) GetAvailableMemoryInBytes() int32 {
	if x != nil {
		return x.AvailableMemoryInBytes
	}
	return 0
}

func (x *SystemStat) GetTotalMemoryInBytes() int32 {
	if x != nil {
		return x.TotalMemoryInBytes
	}
	return 0
}

var File_proto_wflow_proto protoreflect.FileDescriptor

var file_proto_wflow_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x17, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e, 0x77, 0x66, 0x6c, 0x6f, 0x77, 0x22, 0xa0, 0x02, 0x0a,
	0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x44, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e, 0x77, 0x66, 0x6c,
	0x6f, 0x77, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x18,
	0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x31, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1d, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e, 0x77, 0x66, 0x6c, 0x6f, 0x77,
	0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x37, 0x0a, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1f, 0x2e, 0x63, 0x6f,
	0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2e,
	0x77, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x1a, 0x39, 0x0a, 0x0b, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x9b, 0x01, 0x0a, 0x03, 0x4a, 0x6f, 0x62, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x25, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0b,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x12, 0x14,
	0x0a, 0x05, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x70,
	0x61, 0x72, 0x61, 0x6d, 0x12, 0x37, 0x0a, 0x18, 0x64, 0x69, 0x73, 0x70, 0x61, 0x74, 0x63, 0x68,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x15, 0x64, 0x69, 0x73, 0x70, 0x61, 0x74, 0x63, 0x68,
	0x54, 0x69, 0x6d, 0x65, 0x49, 0x6e, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x42, 0x0e, 0x0a,
	0x0c, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xc4, 0x01,
	0x0a, 0x06, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x44, 0x0a, 0x0b, 0x73, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e,
	0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x64, 0x6c, 0x73, 0x68, 0x6c,
	0x65, 0x2e, 0x77, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74,
	0x61, 0x74, 0x52, 0x0a, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x12, 0x1f,
	0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x6a, 0x6f, 0x62, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x4a, 0x6f, 0x62, 0x73, 0x12,
	0x2e, 0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x65, 0x64, 0x5f, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0f, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x65, 0x64, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x88, 0x01, 0x01, 0x42,
	0x13, 0x0a, 0x11, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x65, 0x64, 0x5f, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x22, 0x71, 0x0a, 0x06, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x2b,
	0x0a, 0x11, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x65, 0x64, 0x5f, 0x77, 0x6f, 0x72, 0x6b,
	0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x10, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x65, 0x64, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x73, 0x12, 0x2a, 0x0a, 0x11, 0x75,
	0x70, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x69, 0x6e, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x75, 0x70, 0x74, 0x69, 0x6d, 0x65, 0x49, 0x6e,
	0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x22, 0x97, 0x01, 0x0a, 0x0a, 0x53, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x70, 0x75, 0x5f, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x63, 0x70, 0x75, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x12, 0x39, 0x0a, 0x19, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65,
	0x5f, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x5f, 0x69, 0x6e, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x16, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c,
	0x65, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x49, 0x6e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x31,
	0x0a, 0x15, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x5f, 0x69,
	0x6e, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x12, 0x74,
	0x6f, 0x74, 0x61, 0x6c, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x49, 0x6e, 0x42, 0x79, 0x74, 0x65,
	0x73, 0x2a, 0x2b, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x06, 0x0a, 0x02, 0x4f,
	0x4b, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c, 0x10,
	0x01, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10, 0x02, 0x2a, 0x95,
	0x01, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x49, 0x4e, 0x47, 0x10,
	0x00, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x4f, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x52,
	0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x10, 0x02, 0x12, 0x10, 0x0a, 0x0c, 0x44, 0x49, 0x53,
	0x50, 0x41, 0x54, 0x43, 0x48, 0x5f, 0x4a, 0x4f, 0x42, 0x10, 0x03, 0x12, 0x0d, 0x0a, 0x09, 0x51,
	0x55, 0x45, 0x52, 0x59, 0x5f, 0x4a, 0x4f, 0x42, 0x10, 0x04, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x41,
	0x4e, 0x43, 0x45, 0x4c, 0x5f, 0x4a, 0x4f, 0x42, 0x10, 0x05, 0x12, 0x17, 0x0a, 0x13, 0x51, 0x55,
	0x45, 0x52, 0x59, 0x5f, 0x57, 0x4f, 0x52, 0x4b, 0x45, 0x52, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55,
	0x53, 0x10, 0x06, 0x12, 0x15, 0x0a, 0x11, 0x44, 0x49, 0x53, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43,
	0x54, 0x5f, 0x57, 0x4f, 0x52, 0x4b, 0x45, 0x52, 0x10, 0x07, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x55,
	0x53, 0x54, 0x4f, 0x4d, 0x10, 0x64, 0x42, 0x1f, 0x5a, 0x1d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6c, 0x73, 0x68, 0x6c, 0x65, 0x2f, 0x77, 0x66, 0x6c, 0x6f,
	0x77, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_wflow_proto_rawDescOnce sync.Once
	file_proto_wflow_proto_rawDescData = file_proto_wflow_proto_rawDesc
)

func file_proto_wflow_proto_rawDescGZIP() []byte {
	file_proto_wflow_proto_rawDescOnce.Do(func() {
		file_proto_wflow_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_wflow_proto_rawDescData)
	})
	return file_proto_wflow_proto_rawDescData
}

var file_proto_wflow_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_proto_wflow_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_proto_wflow_proto_goTypes = []interface{}{
	(Status)(0),        // 0: com.github.dlshle.wflow.Status
	(Type)(0),          // 1: com.github.dlshle.wflow.Type
	(*Message)(nil),    // 2: com.github.dlshle.wflow.Message
	(*Job)(nil),        // 3: com.github.dlshle.wflow.Job
	(*Worker)(nil),     // 4: com.github.dlshle.wflow.Worker
	(*Server)(nil),     // 5: com.github.dlshle.wflow.Server
	(*SystemStat)(nil), // 6: com.github.dlshle.wflow.SystemStat
	nil,                // 7: com.github.dlshle.wflow.Message.HeaderEntry
}
var file_proto_wflow_proto_depIdxs = []int32{
	7, // 0: com.github.dlshle.wflow.Message.header:type_name -> com.github.dlshle.wflow.Message.HeaderEntry
	1, // 1: com.github.dlshle.wflow.Message.type:type_name -> com.github.dlshle.wflow.Type
	0, // 2: com.github.dlshle.wflow.Message.status:type_name -> com.github.dlshle.wflow.Status
	6, // 3: com.github.dlshle.wflow.Worker.system_stat:type_name -> com.github.dlshle.wflow.SystemStat
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_wflow_proto_init() }
func file_proto_wflow_proto_init() {
	if File_proto_wflow_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_wflow_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_wflow_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Job); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_wflow_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Worker); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_wflow_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Server); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_wflow_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemStat); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_proto_wflow_proto_msgTypes[1].OneofWrappers = []interface{}{}
	file_proto_wflow_proto_msgTypes[2].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_wflow_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_wflow_proto_goTypes,
		DependencyIndexes: file_proto_wflow_proto_depIdxs,
		EnumInfos:         file_proto_wflow_proto_enumTypes,
		MessageInfos:      file_proto_wflow_proto_msgTypes,
	}.Build()
	File_proto_wflow_proto = out.File
	file_proto_wflow_proto_rawDesc = nil
	file_proto_wflow_proto_goTypes = nil
	file_proto_wflow_proto_depIdxs = nil
}
