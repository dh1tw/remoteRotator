// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: rotator.proto

package sb_rotator

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

type None struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *None) Reset() {
	*x = None{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rotator_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *None) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*None) ProtoMessage() {}

func (x *None) ProtoReflect() protoreflect.Message {
	mi := &file_rotator_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use None.ProtoReflect.Descriptor instead.
func (*None) Descriptor() ([]byte, []int) {
	return file_rotator_proto_rawDescGZIP(), []int{0}
}

type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error       string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Code        int32  `protobuf:"varint,2,opt,name=code,proto3" json:"code,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rotator_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_rotator_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_rotator_proto_rawDescGZIP(), []int{1}
}

func (x *Error) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *Error) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *Error) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

type HeadingReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Heading int32 `protobuf:"varint,1,opt,name=heading,proto3" json:"heading,omitempty"`
}

func (x *HeadingReq) Reset() {
	*x = HeadingReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rotator_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeadingReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeadingReq) ProtoMessage() {}

func (x *HeadingReq) ProtoReflect() protoreflect.Message {
	mi := &file_rotator_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeadingReq.ProtoReflect.Descriptor instead.
func (*HeadingReq) Descriptor() ([]byte, []int) {
	return file_rotator_proto_rawDescGZIP(), []int{2}
}

func (x *HeadingReq) GetHeading() int32 {
	if x != nil {
		return x.Heading
	}
	return 0
}

type HeadingResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Heading int32 `protobuf:"varint,1,opt,name=heading,proto3" json:"heading,omitempty"`
	Preset  int32 `protobuf:"varint,2,opt,name=preset,proto3" json:"preset,omitempty"`
}

func (x *HeadingResp) Reset() {
	*x = HeadingResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rotator_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HeadingResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeadingResp) ProtoMessage() {}

func (x *HeadingResp) ProtoReflect() protoreflect.Message {
	mi := &file_rotator_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeadingResp.ProtoReflect.Descriptor instead.
func (*HeadingResp) Descriptor() ([]byte, []int) {
	return file_rotator_proto_rawDescGZIP(), []int{3}
}

func (x *HeadingResp) GetHeading() int32 {
	if x != nil {
		return x.Heading
	}
	return 0
}

func (x *HeadingResp) GetPreset() int32 {
	if x != nil {
		return x.Preset
	}
	return 0
}

type State struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Azimuth         int32 `protobuf:"varint,1,opt,name=azimuth,proto3" json:"azimuth,omitempty"`
	AzimuthPreset   int32 `protobuf:"varint,2,opt,name=azimuth_preset,json=azimuthPreset,proto3" json:"azimuth_preset,omitempty"`
	Elevation       int32 `protobuf:"varint,3,opt,name=elevation,proto3" json:"elevation,omitempty"`
	ElevationPreset int32 `protobuf:"varint,4,opt,name=elevation_preset,json=elevationPreset,proto3" json:"elevation_preset,omitempty"`
}

func (x *State) Reset() {
	*x = State{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rotator_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *State) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*State) ProtoMessage() {}

func (x *State) ProtoReflect() protoreflect.Message {
	mi := &file_rotator_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use State.ProtoReflect.Descriptor instead.
func (*State) Descriptor() ([]byte, []int) {
	return file_rotator_proto_rawDescGZIP(), []int{4}
}

func (x *State) GetAzimuth() int32 {
	if x != nil {
		return x.Azimuth
	}
	return 0
}

func (x *State) GetAzimuthPreset() int32 {
	if x != nil {
		return x.AzimuthPreset
	}
	return 0
}

func (x *State) GetElevation() int32 {
	if x != nil {
		return x.Elevation
	}
	return 0
}

func (x *State) GetElevationPreset() int32 {
	if x != nil {
		return x.ElevationPreset
	}
	return 0
}

type Metadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AzimuthStop  int32 `protobuf:"varint,1,opt,name=azimuth_stop,json=azimuthStop,proto3" json:"azimuth_stop,omitempty"`
	AzimuthMin   int32 `protobuf:"varint,2,opt,name=azimuth_min,json=azimuthMin,proto3" json:"azimuth_min,omitempty"`
	AzimuthMax   int32 `protobuf:"varint,3,opt,name=azimuth_max,json=azimuthMax,proto3" json:"azimuth_max,omitempty"`
	ElevationMin int32 `protobuf:"varint,4,opt,name=elevation_min,json=elevationMin,proto3" json:"elevation_min,omitempty"`
	ElevationMax int32 `protobuf:"varint,5,opt,name=elevation_max,json=elevationMax,proto3" json:"elevation_max,omitempty"`
	HasAzimuth   bool  `protobuf:"varint,6,opt,name=has_azimuth,json=hasAzimuth,proto3" json:"has_azimuth,omitempty"`
	HasElevation bool  `protobuf:"varint,7,opt,name=has_elevation,json=hasElevation,proto3" json:"has_elevation,omitempty"`
}

func (x *Metadata) Reset() {
	*x = Metadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rotator_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata) ProtoMessage() {}

func (x *Metadata) ProtoReflect() protoreflect.Message {
	mi := &file_rotator_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metadata.ProtoReflect.Descriptor instead.
func (*Metadata) Descriptor() ([]byte, []int) {
	return file_rotator_proto_rawDescGZIP(), []int{5}
}

func (x *Metadata) GetAzimuthStop() int32 {
	if x != nil {
		return x.AzimuthStop
	}
	return 0
}

func (x *Metadata) GetAzimuthMin() int32 {
	if x != nil {
		return x.AzimuthMin
	}
	return 0
}

func (x *Metadata) GetAzimuthMax() int32 {
	if x != nil {
		return x.AzimuthMax
	}
	return 0
}

func (x *Metadata) GetElevationMin() int32 {
	if x != nil {
		return x.ElevationMin
	}
	return 0
}

func (x *Metadata) GetElevationMax() int32 {
	if x != nil {
		return x.ElevationMax
	}
	return 0
}

func (x *Metadata) GetHasAzimuth() bool {
	if x != nil {
		return x.HasAzimuth
	}
	return false
}

func (x *Metadata) GetHasElevation() bool {
	if x != nil {
		return x.HasElevation
	}
	return false
}

var File_rotator_proto protoreflect.FileDescriptor

var file_rotator_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x10, 0x73, 0x68, 0x61, 0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f,
	0x72, 0x22, 0x06, 0x0a, 0x04, 0x4e, 0x6f, 0x6e, 0x65, 0x22, 0x53, 0x0a, 0x05, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x20, 0x0a, 0x0b,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x26,
	0x0a, 0x0a, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x12, 0x18, 0x0a, 0x07,
	0x68, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x68,
	0x65, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x22, 0x3f, 0x0a, 0x0b, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6e,
	0x67, 0x52, 0x65, 0x73, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x67,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x12,
	0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x06, 0x70, 0x72, 0x65, 0x73, 0x65, 0x74, 0x22, 0x91, 0x01, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x07, 0x61, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x12, 0x25, 0x0a, 0x0e, 0x61,
	0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x5f, 0x70, 0x72, 0x65, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0d, 0x61, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x50, 0x72, 0x65, 0x73,
	0x65, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x29, 0x0a, 0x10, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x72,
	0x65, 0x73, 0x65, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x65, 0x6c, 0x65, 0x76,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x65, 0x73, 0x65, 0x74, 0x22, 0xff, 0x01, 0x0a, 0x08,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x21, 0x0a, 0x0c, 0x61, 0x7a, 0x69, 0x6d,
	0x75, 0x74, 0x68, 0x5f, 0x73, 0x74, 0x6f, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b,
	0x61, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x53, 0x74, 0x6f, 0x70, 0x12, 0x1f, 0x0a, 0x0b, 0x61,
	0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x5f, 0x6d, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0a, 0x61, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x4d, 0x69, 0x6e, 0x12, 0x1f, 0x0a, 0x0b,
	0x61, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x5f, 0x6d, 0x61, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0a, 0x61, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x4d, 0x61, 0x78, 0x12, 0x23, 0x0a,
	0x0d, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6d, 0x69, 0x6e, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d,
	0x69, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x6d, 0x61, 0x78, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x65, 0x6c, 0x65, 0x76, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x61, 0x78, 0x12, 0x1f, 0x0a, 0x0b, 0x68, 0x61, 0x73, 0x5f, 0x61,
	0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x68, 0x61,
	0x73, 0x41, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x12, 0x23, 0x0a, 0x0d, 0x68, 0x61, 0x73, 0x5f,
	0x65, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0c, 0x68, 0x61, 0x73, 0x45, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x32, 0x93, 0x03,
	0x0a, 0x07, 0x52, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x42, 0x0a, 0x0a, 0x53, 0x65, 0x74,
	0x41, 0x7a, 0x69, 0x6d, 0x75, 0x74, 0x68, 0x12, 0x1c, 0x2e, 0x73, 0x68, 0x61, 0x63, 0x6b, 0x62,
	0x75, 0x73, 0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x71, 0x1a, 0x16, 0x2e, 0x73, 0x68, 0x61, 0x63, 0x6b, 0x62, 0x75, 0x73,
	0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e, 0x6f, 0x6e, 0x65, 0x12, 0x44, 0x0a,
	0x0c, 0x53, 0x65, 0x74, 0x45, 0x6c, 0x65, 0x76, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x2e,
	0x73, 0x68, 0x61, 0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72,
	0x2e, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x1a, 0x16, 0x2e, 0x73, 0x68,
	0x61, 0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e,
	0x6f, 0x6e, 0x65, 0x12, 0x3d, 0x0a, 0x0b, 0x53, 0x74, 0x6f, 0x70, 0x41, 0x7a, 0x69, 0x6d, 0x75,
	0x74, 0x68, 0x12, 0x16, 0x2e, 0x73, 0x68, 0x61, 0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f,
	0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e, 0x6f, 0x6e, 0x65, 0x1a, 0x16, 0x2e, 0x73, 0x68, 0x61,
	0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e, 0x6f,
	0x6e, 0x65, 0x12, 0x3f, 0x0a, 0x0d, 0x53, 0x74, 0x6f, 0x70, 0x45, 0x6c, 0x65, 0x76, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x16, 0x2e, 0x73, 0x68, 0x61, 0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72,
	0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e, 0x6f, 0x6e, 0x65, 0x1a, 0x16, 0x2e, 0x73, 0x68,
	0x61, 0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e,
	0x6f, 0x6e, 0x65, 0x12, 0x41, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x12, 0x16, 0x2e, 0x73, 0x68, 0x61, 0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f,
	0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e, 0x6f, 0x6e, 0x65, 0x1a, 0x1a, 0x2e, 0x73, 0x68, 0x61,
	0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x3b, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x16, 0x2e, 0x73, 0x68, 0x61, 0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f,
	0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x4e, 0x6f, 0x6e, 0x65, 0x1a, 0x17, 0x2e, 0x73, 0x68, 0x61,
	0x63, 0x6b, 0x62, 0x75, 0x73, 0x2e, 0x72, 0x6f, 0x74, 0x61, 0x74, 0x6f, 0x72, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x42, 0x0e, 0x5a, 0x0c, 0x2e, 0x2f, 0x73, 0x62, 0x5f, 0x72, 0x6f, 0x74, 0x61,
	0x74, 0x6f, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rotator_proto_rawDescOnce sync.Once
	file_rotator_proto_rawDescData = file_rotator_proto_rawDesc
)

func file_rotator_proto_rawDescGZIP() []byte {
	file_rotator_proto_rawDescOnce.Do(func() {
		file_rotator_proto_rawDescData = protoimpl.X.CompressGZIP(file_rotator_proto_rawDescData)
	})
	return file_rotator_proto_rawDescData
}

var file_rotator_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_rotator_proto_goTypes = []interface{}{
	(*None)(nil),        // 0: shackbus.rotator.None
	(*Error)(nil),       // 1: shackbus.rotator.Error
	(*HeadingReq)(nil),  // 2: shackbus.rotator.HeadingReq
	(*HeadingResp)(nil), // 3: shackbus.rotator.HeadingResp
	(*State)(nil),       // 4: shackbus.rotator.State
	(*Metadata)(nil),    // 5: shackbus.rotator.Metadata
}
var file_rotator_proto_depIdxs = []int32{
	2, // 0: shackbus.rotator.Rotator.SetAzimuth:input_type -> shackbus.rotator.HeadingReq
	2, // 1: shackbus.rotator.Rotator.SetElevation:input_type -> shackbus.rotator.HeadingReq
	0, // 2: shackbus.rotator.Rotator.StopAzimuth:input_type -> shackbus.rotator.None
	0, // 3: shackbus.rotator.Rotator.StopElevation:input_type -> shackbus.rotator.None
	0, // 4: shackbus.rotator.Rotator.GetMetadata:input_type -> shackbus.rotator.None
	0, // 5: shackbus.rotator.Rotator.GetState:input_type -> shackbus.rotator.None
	0, // 6: shackbus.rotator.Rotator.SetAzimuth:output_type -> shackbus.rotator.None
	0, // 7: shackbus.rotator.Rotator.SetElevation:output_type -> shackbus.rotator.None
	0, // 8: shackbus.rotator.Rotator.StopAzimuth:output_type -> shackbus.rotator.None
	0, // 9: shackbus.rotator.Rotator.StopElevation:output_type -> shackbus.rotator.None
	5, // 10: shackbus.rotator.Rotator.GetMetadata:output_type -> shackbus.rotator.Metadata
	4, // 11: shackbus.rotator.Rotator.GetState:output_type -> shackbus.rotator.State
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_rotator_proto_init() }
func file_rotator_proto_init() {
	if File_rotator_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rotator_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*None); i {
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
		file_rotator_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
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
		file_rotator_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeadingReq); i {
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
		file_rotator_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HeadingResp); i {
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
		file_rotator_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*State); i {
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
		file_rotator_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metadata); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rotator_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rotator_proto_goTypes,
		DependencyIndexes: file_rotator_proto_depIdxs,
		MessageInfos:      file_rotator_proto_msgTypes,
	}.Build()
	File_rotator_proto = out.File
	file_rotator_proto_rawDesc = nil
	file_rotator_proto_goTypes = nil
	file_rotator_proto_depIdxs = nil
}
