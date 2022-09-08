// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: ingester/v1/ingester.proto

package ingesterv1

import (
	v1 "github.com/grafana/fire/pkg/gen/common/v1"
	v11 "github.com/grafana/fire/pkg/gen/push/v1"
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

type LabelValuesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *LabelValuesRequest) Reset() {
	*x = LabelValuesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LabelValuesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LabelValuesRequest) ProtoMessage() {}

func (x *LabelValuesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LabelValuesRequest.ProtoReflect.Descriptor instead.
func (*LabelValuesRequest) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{0}
}

func (x *LabelValuesRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type LabelValuesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Names []string `protobuf:"bytes,1,rep,name=names,proto3" json:"names,omitempty"`
}

func (x *LabelValuesResponse) Reset() {
	*x = LabelValuesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LabelValuesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LabelValuesResponse) ProtoMessage() {}

func (x *LabelValuesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LabelValuesResponse.ProtoReflect.Descriptor instead.
func (*LabelValuesResponse) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{1}
}

func (x *LabelValuesResponse) GetNames() []string {
	if x != nil {
		return x.Names
	}
	return nil
}

type LabelNamesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LabelNamesRequest) Reset() {
	*x = LabelNamesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LabelNamesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LabelNamesRequest) ProtoMessage() {}

func (x *LabelNamesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LabelNamesRequest.ProtoReflect.Descriptor instead.
func (*LabelNamesRequest) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{2}
}

type LabelNamesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Names []string `protobuf:"bytes,1,rep,name=names,proto3" json:"names,omitempty"`
}

func (x *LabelNamesResponse) Reset() {
	*x = LabelNamesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LabelNamesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LabelNamesResponse) ProtoMessage() {}

func (x *LabelNamesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LabelNamesResponse.ProtoReflect.Descriptor instead.
func (*LabelNamesResponse) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{3}
}

func (x *LabelNamesResponse) GetNames() []string {
	if x != nil {
		return x.Names
	}
	return nil
}

type ProfileTypesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ProfileTypesRequest) Reset() {
	*x = ProfileTypesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProfileTypesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProfileTypesRequest) ProtoMessage() {}

func (x *ProfileTypesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProfileTypesRequest.ProtoReflect.Descriptor instead.
func (*ProfileTypesRequest) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{4}
}

type ProfileTypesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProfileTypes []*v1.ProfileType `protobuf:"bytes,1,rep,name=profile_types,json=profileTypes,proto3" json:"profile_types,omitempty"`
}

func (x *ProfileTypesResponse) Reset() {
	*x = ProfileTypesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProfileTypesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProfileTypesResponse) ProtoMessage() {}

func (x *ProfileTypesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProfileTypesResponse.ProtoReflect.Descriptor instead.
func (*ProfileTypesResponse) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{5}
}

func (x *ProfileTypesResponse) GetProfileTypes() []*v1.ProfileType {
	if x != nil {
		return x.ProfileTypes
	}
	return nil
}

type SeriesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Matchers []string `protobuf:"bytes,1,rep,name=matchers,proto3" json:"matchers,omitempty"`
}

func (x *SeriesRequest) Reset() {
	*x = SeriesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SeriesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SeriesRequest) ProtoMessage() {}

func (x *SeriesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SeriesRequest.ProtoReflect.Descriptor instead.
func (*SeriesRequest) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{6}
}

func (x *SeriesRequest) GetMatchers() []string {
	if x != nil {
		return x.Matchers
	}
	return nil
}

type SeriesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LabelsSet []*v1.Labels `protobuf:"bytes,2,rep,name=labels_set,json=labelsSet,proto3" json:"labels_set,omitempty"`
}

func (x *SeriesResponse) Reset() {
	*x = SeriesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SeriesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SeriesResponse) ProtoMessage() {}

func (x *SeriesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SeriesResponse.ProtoReflect.Descriptor instead.
func (*SeriesResponse) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{7}
}

func (x *SeriesResponse) GetLabelsSet() []*v1.Labels {
	if x != nil {
		return x.LabelsSet
	}
	return nil
}

type FlushRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *FlushRequest) Reset() {
	*x = FlushRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FlushRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FlushRequest) ProtoMessage() {}

func (x *FlushRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FlushRequest.ProtoReflect.Descriptor instead.
func (*FlushRequest) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{8}
}

type FlushResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *FlushResponse) Reset() {
	*x = FlushResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FlushResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FlushResponse) ProtoMessage() {}

func (x *FlushResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FlushResponse.ProtoReflect.Descriptor instead.
func (*FlushResponse) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{9}
}

type SelectProfilesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LabelSelector string          `protobuf:"bytes,1,opt,name=label_selector,json=labelSelector,proto3" json:"label_selector,omitempty"`
	Type          *v1.ProfileType `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Start         int64           `protobuf:"varint,3,opt,name=start,proto3" json:"start,omitempty"`
	End           int64           `protobuf:"varint,4,opt,name=end,proto3" json:"end,omitempty"`
}

func (x *SelectProfilesRequest) Reset() {
	*x = SelectProfilesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SelectProfilesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SelectProfilesRequest) ProtoMessage() {}

func (x *SelectProfilesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SelectProfilesRequest.ProtoReflect.Descriptor instead.
func (*SelectProfilesRequest) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{10}
}

func (x *SelectProfilesRequest) GetLabelSelector() string {
	if x != nil {
		return x.LabelSelector
	}
	return ""
}

func (x *SelectProfilesRequest) GetType() *v1.ProfileType {
	if x != nil {
		return x.Type
	}
	return nil
}

func (x *SelectProfilesRequest) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *SelectProfilesRequest) GetEnd() int64 {
	if x != nil {
		return x.End
	}
	return 0
}

type SelectProfilesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Profiles      []*Profile `protobuf:"bytes,1,rep,name=profiles,proto3" json:"profiles,omitempty"`
	FunctionNames []string   `protobuf:"bytes,2,rep,name=function_names,json=functionNames,proto3" json:"function_names,omitempty"`
}

func (x *SelectProfilesResponse) Reset() {
	*x = SelectProfilesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SelectProfilesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SelectProfilesResponse) ProtoMessage() {}

func (x *SelectProfilesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SelectProfilesResponse.ProtoReflect.Descriptor instead.
func (*SelectProfilesResponse) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{11}
}

func (x *SelectProfilesResponse) GetProfiles() []*Profile {
	if x != nil {
		return x.Profiles
	}
	return nil
}

func (x *SelectProfilesResponse) GetFunctionNames() []string {
	if x != nil {
		return x.FunctionNames
	}
	return nil
}

// Profile represents a point in time profile.
type Profile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The ID of the profile.
	ID string `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	// The name and type of the profile.
	Type *v1.ProfileType `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	// LabelPair is the key value pairs to identify the corresponding profile
	Labels []*v1.LabelPair `protobuf:"bytes,3,rep,name=labels,proto3" json:"labels,omitempty"`
	// Timestamp is when that profile was created
	Timestamp int64 `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// The list of stracktraces for the profile with their respective value
	Stacktraces []*StacktraceSample `protobuf:"bytes,5,rep,name=stacktraces,proto3" json:"stacktraces,omitempty"`
}

func (x *Profile) Reset() {
	*x = Profile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Profile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Profile) ProtoMessage() {}

func (x *Profile) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Profile.ProtoReflect.Descriptor instead.
func (*Profile) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{12}
}

func (x *Profile) GetID() string {
	if x != nil {
		return x.ID
	}
	return ""
}

func (x *Profile) GetType() *v1.ProfileType {
	if x != nil {
		return x.Type
	}
	return nil
}

func (x *Profile) GetLabels() []*v1.LabelPair {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *Profile) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *Profile) GetStacktraces() []*StacktraceSample {
	if x != nil {
		return x.Stacktraces
	}
	return nil
}

type StacktraceSample struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FunctionIds []int32 `protobuf:"varint,1,rep,packed,name=function_ids,json=functionIds,proto3" json:"function_ids,omitempty"`
	Value       int64   `protobuf:"varint,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *StacktraceSample) Reset() {
	*x = StacktraceSample{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ingester_v1_ingester_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StacktraceSample) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StacktraceSample) ProtoMessage() {}

func (x *StacktraceSample) ProtoReflect() protoreflect.Message {
	mi := &file_ingester_v1_ingester_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StacktraceSample.ProtoReflect.Descriptor instead.
func (*StacktraceSample) Descriptor() ([]byte, []int) {
	return file_ingester_v1_ingester_proto_rawDescGZIP(), []int{13}
}

func (x *StacktraceSample) GetFunctionIds() []int32 {
	if x != nil {
		return x.FunctionIds
	}
	return nil
}

func (x *StacktraceSample) GetValue() int64 {
	if x != nil {
		return x.Value
	}
	return 0
}

var File_ingester_v1_ingester_proto protoreflect.FileDescriptor

var file_ingester_v1_ingester_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6e,
	0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x69, 0x6e,
	0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x16, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x12, 0x70, 0x75, 0x73, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x75, 0x73, 0x68, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x28, 0x0a, 0x12, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22,
	0x2b, 0x0a, 0x13, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x13, 0x0a, 0x11,
	0x4c, 0x61, 0x62, 0x65, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x22, 0x2a, 0x0a, 0x12, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x15, 0x0a,
	0x13, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x53, 0x0a, 0x14, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54,
	0x79, 0x70, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3b, 0x0a, 0x0d,
	0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0c, 0x70, 0x72, 0x6f,
	0x66, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x22, 0x2b, 0x0a, 0x0d, 0x53, 0x65, 0x72,
	0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x61,
	0x74, 0x63, 0x68, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x6d, 0x61,
	0x74, 0x63, 0x68, 0x65, 0x72, 0x73, 0x22, 0x42, 0x0a, 0x0e, 0x53, 0x65, 0x72, 0x69, 0x65, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x0a, 0x6c, 0x61, 0x62, 0x65,
	0x6c, 0x73, 0x5f, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x52,
	0x09, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x53, 0x65, 0x74, 0x22, 0x0e, 0x0a, 0x0c, 0x46, 0x6c,
	0x75, 0x73, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x0f, 0x0a, 0x0d, 0x46, 0x6c,
	0x75, 0x73, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x92, 0x01, 0x0a, 0x15,
	0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x0e, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x5f, 0x73,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6c,
	0x61, 0x62, 0x65, 0x6c, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x2a, 0x0a, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x65, 0x6e, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x65, 0x6e, 0x64,
	0x22, 0x71, 0x0a, 0x16, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x08, 0x70, 0x72,
	0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x69,
	0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69,
	0x6c, 0x65, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x25, 0x0a, 0x0e,
	0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x0d, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x22, 0xd2, 0x01, 0x0a, 0x07, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x12,
	0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x44, 0x12,
	0x2a, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x6c,
	0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x50, 0x61, 0x69,
	0x72, 0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x3f, 0x0a, 0x0b, 0x73, 0x74, 0x61, 0x63, 0x6b,
	0x74, 0x72, 0x61, 0x63, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x69,
	0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x63, 0x6b,
	0x74, 0x72, 0x61, 0x63, 0x65, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x0b, 0x73, 0x74, 0x61,
	0x63, 0x6b, 0x74, 0x72, 0x61, 0x63, 0x65, 0x73, 0x22, 0x4b, 0x0a, 0x10, 0x53, 0x74, 0x61, 0x63,
	0x6b, 0x74, 0x72, 0x61, 0x63, 0x65, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x21, 0x0a, 0x0c,
	0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x05, 0x52, 0x0b, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x73, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x32, 0xa8, 0x04, 0x0a, 0x0f, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74,
	0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x35, 0x0a, 0x04, 0x50, 0x75, 0x73,
	0x68, 0x12, 0x14, 0x2e, 0x70, 0x75, 0x73, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x73, 0x68,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x70, 0x75, 0x73, 0x68, 0x2e, 0x76,
	0x31, 0x2e, 0x50, 0x75, 0x73, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x52, 0x0a, 0x0b, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x12,
	0x1f, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x61,
	0x62, 0x65, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x20, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c,
	0x61, 0x62, 0x65, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x4f, 0x0a, 0x0a, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x4e, 0x61, 0x6d,
	0x65, 0x73, 0x12, 0x1e, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x55, 0x0a, 0x0c, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x73, 0x12, 0x20, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x54, 0x79, 0x70,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x06,
	0x53, 0x65, 0x72, 0x69, 0x65, 0x73, 0x12, 0x1a, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x65, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x40, 0x0a, 0x05, 0x46, 0x6c, 0x75, 0x73, 0x68, 0x12, 0x19, 0x2e, 0x69, 0x6e, 0x67,
	0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6c, 0x75, 0x73, 0x68, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6c, 0x75, 0x73, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x5b, 0x0a, 0x0e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x50, 0x72, 0x6f,
	0x66, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x22, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c,
	0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x69, 0x6e, 0x67, 0x65,
	0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x50, 0x72,
	0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x42, 0xa5, 0x01, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x42, 0x0d, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x67, 0x72, 0x61, 0x66, 0x61, 0x6e, 0x61, 0x2f, 0x66, 0x69, 0x72, 0x65, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2f,
	0x76, 0x31, 0x3b, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x76, 0x31, 0xa2, 0x02, 0x03,
	0x49, 0x58, 0x58, 0xaa, 0x02, 0x0b, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x56,
	0x31, 0xca, 0x02, 0x0b, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x5c, 0x56, 0x31, 0xe2,
	0x02, 0x17, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50,
	0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0c, 0x49, 0x6e, 0x67, 0x65,
	0x73, 0x74, 0x65, 0x72, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ingester_v1_ingester_proto_rawDescOnce sync.Once
	file_ingester_v1_ingester_proto_rawDescData = file_ingester_v1_ingester_proto_rawDesc
)

func file_ingester_v1_ingester_proto_rawDescGZIP() []byte {
	file_ingester_v1_ingester_proto_rawDescOnce.Do(func() {
		file_ingester_v1_ingester_proto_rawDescData = protoimpl.X.CompressGZIP(file_ingester_v1_ingester_proto_rawDescData)
	})
	return file_ingester_v1_ingester_proto_rawDescData
}

var file_ingester_v1_ingester_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_ingester_v1_ingester_proto_goTypes = []interface{}{
	(*LabelValuesRequest)(nil),     // 0: ingester.v1.LabelValuesRequest
	(*LabelValuesResponse)(nil),    // 1: ingester.v1.LabelValuesResponse
	(*LabelNamesRequest)(nil),      // 2: ingester.v1.LabelNamesRequest
	(*LabelNamesResponse)(nil),     // 3: ingester.v1.LabelNamesResponse
	(*ProfileTypesRequest)(nil),    // 4: ingester.v1.ProfileTypesRequest
	(*ProfileTypesResponse)(nil),   // 5: ingester.v1.ProfileTypesResponse
	(*SeriesRequest)(nil),          // 6: ingester.v1.SeriesRequest
	(*SeriesResponse)(nil),         // 7: ingester.v1.SeriesResponse
	(*FlushRequest)(nil),           // 8: ingester.v1.FlushRequest
	(*FlushResponse)(nil),          // 9: ingester.v1.FlushResponse
	(*SelectProfilesRequest)(nil),  // 10: ingester.v1.SelectProfilesRequest
	(*SelectProfilesResponse)(nil), // 11: ingester.v1.SelectProfilesResponse
	(*Profile)(nil),                // 12: ingester.v1.Profile
	(*StacktraceSample)(nil),       // 13: ingester.v1.StacktraceSample
	(*v1.ProfileType)(nil),         // 14: common.v1.ProfileType
	(*v1.Labels)(nil),              // 15: common.v1.Labels
	(*v1.LabelPair)(nil),           // 16: common.v1.LabelPair
	(*v11.PushRequest)(nil),        // 17: push.v1.PushRequest
	(*v11.PushResponse)(nil),       // 18: push.v1.PushResponse
}
var file_ingester_v1_ingester_proto_depIdxs = []int32{
	14, // 0: ingester.v1.ProfileTypesResponse.profile_types:type_name -> common.v1.ProfileType
	15, // 1: ingester.v1.SeriesResponse.labels_set:type_name -> common.v1.Labels
	14, // 2: ingester.v1.SelectProfilesRequest.type:type_name -> common.v1.ProfileType
	12, // 3: ingester.v1.SelectProfilesResponse.profiles:type_name -> ingester.v1.Profile
	14, // 4: ingester.v1.Profile.type:type_name -> common.v1.ProfileType
	16, // 5: ingester.v1.Profile.labels:type_name -> common.v1.LabelPair
	13, // 6: ingester.v1.Profile.stacktraces:type_name -> ingester.v1.StacktraceSample
	17, // 7: ingester.v1.IngesterService.Push:input_type -> push.v1.PushRequest
	0,  // 8: ingester.v1.IngesterService.LabelValues:input_type -> ingester.v1.LabelValuesRequest
	2,  // 9: ingester.v1.IngesterService.LabelNames:input_type -> ingester.v1.LabelNamesRequest
	4,  // 10: ingester.v1.IngesterService.ProfileTypes:input_type -> ingester.v1.ProfileTypesRequest
	6,  // 11: ingester.v1.IngesterService.Series:input_type -> ingester.v1.SeriesRequest
	8,  // 12: ingester.v1.IngesterService.Flush:input_type -> ingester.v1.FlushRequest
	10, // 13: ingester.v1.IngesterService.SelectProfiles:input_type -> ingester.v1.SelectProfilesRequest
	18, // 14: ingester.v1.IngesterService.Push:output_type -> push.v1.PushResponse
	1,  // 15: ingester.v1.IngesterService.LabelValues:output_type -> ingester.v1.LabelValuesResponse
	3,  // 16: ingester.v1.IngesterService.LabelNames:output_type -> ingester.v1.LabelNamesResponse
	5,  // 17: ingester.v1.IngesterService.ProfileTypes:output_type -> ingester.v1.ProfileTypesResponse
	7,  // 18: ingester.v1.IngesterService.Series:output_type -> ingester.v1.SeriesResponse
	9,  // 19: ingester.v1.IngesterService.Flush:output_type -> ingester.v1.FlushResponse
	11, // 20: ingester.v1.IngesterService.SelectProfiles:output_type -> ingester.v1.SelectProfilesResponse
	14, // [14:21] is the sub-list for method output_type
	7,  // [7:14] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_ingester_v1_ingester_proto_init() }
func file_ingester_v1_ingester_proto_init() {
	if File_ingester_v1_ingester_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ingester_v1_ingester_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LabelValuesRequest); i {
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
		file_ingester_v1_ingester_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LabelValuesResponse); i {
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
		file_ingester_v1_ingester_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LabelNamesRequest); i {
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
		file_ingester_v1_ingester_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LabelNamesResponse); i {
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
		file_ingester_v1_ingester_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProfileTypesRequest); i {
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
		file_ingester_v1_ingester_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProfileTypesResponse); i {
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
		file_ingester_v1_ingester_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SeriesRequest); i {
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
		file_ingester_v1_ingester_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SeriesResponse); i {
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
		file_ingester_v1_ingester_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FlushRequest); i {
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
		file_ingester_v1_ingester_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FlushResponse); i {
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
		file_ingester_v1_ingester_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SelectProfilesRequest); i {
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
		file_ingester_v1_ingester_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SelectProfilesResponse); i {
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
		file_ingester_v1_ingester_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Profile); i {
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
		file_ingester_v1_ingester_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StacktraceSample); i {
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
			RawDescriptor: file_ingester_v1_ingester_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ingester_v1_ingester_proto_goTypes,
		DependencyIndexes: file_ingester_v1_ingester_proto_depIdxs,
		MessageInfos:      file_ingester_v1_ingester_proto_msgTypes,
	}.Build()
	File_ingester_v1_ingester_proto = out.File
	file_ingester_v1_ingester_proto_rawDesc = nil
	file_ingester_v1_ingester_proto_goTypes = nil
	file_ingester_v1_ingester_proto_depIdxs = nil
}
