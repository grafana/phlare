// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0-devel
// 	protoc        (unknown)
// source: parca/metastore/v1alpha1/metastore.proto

package metastorev1alpha1

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

// Sample is a stack trace with optional labels.
type Sample struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// location_ids are locations that define the stack trace.
	LocationIds [][]byte `protobuf:"bytes,1,rep,name=location_ids,json=locationIds,proto3" json:"location_ids,omitempty"`
	// labels are extra labels for a stack trace.
	Labels map[string]*SampleLabel `protobuf:"bytes,2,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// num_labels are the num of labels.
	NumLabels map[string]*SampleNumLabel `protobuf:"bytes,3,rep,name=num_labels,json=numLabels,proto3" json:"num_labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// num_units are the units for the labels.
	NumUnits map[string]*SampleNumUnit `protobuf:"bytes,4,rep,name=num_units,json=numUnits,proto3" json:"num_units,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Sample) Reset() {
	*x = Sample{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Sample) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Sample) ProtoMessage() {}

func (x *Sample) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Sample.ProtoReflect.Descriptor instead.
func (*Sample) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{0}
}

func (x *Sample) GetLocationIds() [][]byte {
	if x != nil {
		return x.LocationIds
	}
	return nil
}

func (x *Sample) GetLabels() map[string]*SampleLabel {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *Sample) GetNumLabels() map[string]*SampleNumLabel {
	if x != nil {
		return x.NumLabels
	}
	return nil
}

func (x *Sample) GetNumUnits() map[string]*SampleNumUnit {
	if x != nil {
		return x.NumUnits
	}
	return nil
}

// SampleLabel are the labels added to a Sample.
type SampleLabel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// labels for a label in a Sample.
	Labels []string `protobuf:"bytes,1,rep,name=labels,proto3" json:"labels,omitempty"`
}

func (x *SampleLabel) Reset() {
	*x = SampleLabel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SampleLabel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SampleLabel) ProtoMessage() {}

func (x *SampleLabel) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SampleLabel.ProtoReflect.Descriptor instead.
func (*SampleLabel) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{1}
}

func (x *SampleLabel) GetLabels() []string {
	if x != nil {
		return x.Labels
	}
	return nil
}

// SampleNumLabel are the num of labels of a Sample.
type SampleNumLabel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// num_labels are the num_label of a Sample.
	NumLabels []int64 `protobuf:"varint,1,rep,packed,name=num_labels,json=numLabels,proto3" json:"num_labels,omitempty"`
}

func (x *SampleNumLabel) Reset() {
	*x = SampleNumLabel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SampleNumLabel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SampleNumLabel) ProtoMessage() {}

func (x *SampleNumLabel) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SampleNumLabel.ProtoReflect.Descriptor instead.
func (*SampleNumLabel) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{2}
}

func (x *SampleNumLabel) GetNumLabels() []int64 {
	if x != nil {
		return x.NumLabels
	}
	return nil
}

// SampleNumUnit are the num units of a Sample.
type SampleNumUnit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// units of a labels of a Sample.
	Units []string `protobuf:"bytes,1,rep,name=units,proto3" json:"units,omitempty"`
}

func (x *SampleNumUnit) Reset() {
	*x = SampleNumUnit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SampleNumUnit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SampleNumUnit) ProtoMessage() {}

func (x *SampleNumUnit) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SampleNumUnit.ProtoReflect.Descriptor instead.
func (*SampleNumUnit) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{3}
}

func (x *SampleNumUnit) GetUnits() []string {
	if x != nil {
		return x.Units
	}
	return nil
}

// Location describes a single location of a stack traces.
type Location struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id is the unique identifier for the location.
	Id []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// address is the memory address of the location if present.
	Address uint64 `protobuf:"varint,2,opt,name=address,proto3" json:"address,omitempty"`
	// mapping_id is the unique identifier for the mapping associated with the location.
	MappingId []byte `protobuf:"bytes,4,opt,name=mapping_id,json=mappingId,proto3" json:"mapping_id,omitempty"`
	// is_folded indicates whether the location is folded into the previous location.
	IsFolded bool `protobuf:"varint,5,opt,name=is_folded,json=isFolded,proto3" json:"is_folded,omitempty"`
}

func (x *Location) Reset() {
	*x = Location{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Location) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Location) ProtoMessage() {}

func (x *Location) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Location.ProtoReflect.Descriptor instead.
func (*Location) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{4}
}

func (x *Location) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Location) GetAddress() uint64 {
	if x != nil {
		return x.Address
	}
	return 0
}

func (x *Location) GetMappingId() []byte {
	if x != nil {
		return x.MappingId
	}
	return nil
}

func (x *Location) GetIsFolded() bool {
	if x != nil {
		return x.IsFolded
	}
	return false
}

// LocationLines describes a set of lines of a location.
type LocationLines struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id is the unique identifier for the location.
	Id []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Lines is the list of lines associated with the location.
	Lines []*Line `protobuf:"bytes,2,rep,name=lines,proto3" json:"lines,omitempty"`
}

func (x *LocationLines) Reset() {
	*x = LocationLines{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LocationLines) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LocationLines) ProtoMessage() {}

func (x *LocationLines) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LocationLines.ProtoReflect.Descriptor instead.
func (*LocationLines) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{5}
}

func (x *LocationLines) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *LocationLines) GetLines() []*Line {
	if x != nil {
		return x.Lines
	}
	return nil
}

// Line describes a source code function and its line number.
type Line struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// function_id is the ID of the function.
	FunctionId []byte `protobuf:"bytes,1,opt,name=function_id,json=functionId,proto3" json:"function_id,omitempty"`
	// line is the line number in the source file of the referenced function.
	Line int64 `protobuf:"varint,2,opt,name=line,proto3" json:"line,omitempty"`
}

func (x *Line) Reset() {
	*x = Line{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Line) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Line) ProtoMessage() {}

func (x *Line) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Line.ProtoReflect.Descriptor instead.
func (*Line) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{6}
}

func (x *Line) GetFunctionId() []byte {
	if x != nil {
		return x.FunctionId
	}
	return nil
}

func (x *Line) GetLine() int64 {
	if x != nil {
		return x.Line
	}
	return 0
}

// Function describes metadata of a source code function.
type Function struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id is the unique identifier for the function.
	Id []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// start_line is the line number in the source file of the first line of the function.
	StartLine int64 `protobuf:"varint,2,opt,name=start_line,json=startLine,proto3" json:"start_line,omitempty"`
	// name is the name of the function.
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// system_name describes the name of the function, as identified by the
	// system. For instance, it can be a C++ mangled name.
	SystemName string `protobuf:"bytes,4,opt,name=system_name,json=systemName,proto3" json:"system_name,omitempty"`
	// filename is the name of the source file of the function.
	Filename string `protobuf:"bytes,5,opt,name=filename,proto3" json:"filename,omitempty"`
}

func (x *Function) Reset() {
	*x = Function{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Function) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Function) ProtoMessage() {}

func (x *Function) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Function.ProtoReflect.Descriptor instead.
func (*Function) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{7}
}

func (x *Function) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Function) GetStartLine() int64 {
	if x != nil {
		return x.StartLine
	}
	return 0
}

func (x *Function) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Function) GetSystemName() string {
	if x != nil {
		return x.SystemName
	}
	return ""
}

func (x *Function) GetFilename() string {
	if x != nil {
		return x.Filename
	}
	return ""
}

// Mapping describes a memory mapping.
type Mapping struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id is the unique identifier for the mapping.
	Id []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// start is the start address of the mapping.
	Start uint64 `protobuf:"varint,2,opt,name=start,proto3" json:"start,omitempty"`
	// limit is the length of the address space of the mapping.
	Limit uint64 `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	// offset is the offset of the mapping.
	Offset uint64 `protobuf:"varint,4,opt,name=offset,proto3" json:"offset,omitempty"`
	// file is the name of the file associated with the mapping.
	File string `protobuf:"bytes,5,opt,name=file,proto3" json:"file,omitempty"`
	// build_id is the build ID of the mapping.
	BuildId string `protobuf:"bytes,6,opt,name=build_id,json=buildId,proto3" json:"build_id,omitempty"`
	// has_functions indicates whether the mapping has associated functions.
	HasFunctions bool `protobuf:"varint,7,opt,name=has_functions,json=hasFunctions,proto3" json:"has_functions,omitempty"`
	// has_filenames indicates whether the mapping has associated filenames.
	HasFilenames bool `protobuf:"varint,8,opt,name=has_filenames,json=hasFilenames,proto3" json:"has_filenames,omitempty"`
	// has_line_numbers indicates whether the mapping has associated line numbers.
	HasLineNumbers bool `protobuf:"varint,9,opt,name=has_line_numbers,json=hasLineNumbers,proto3" json:"has_line_numbers,omitempty"`
	// has_inline_frames indicates whether the mapping has associated inline frames.
	HasInlineFrames bool `protobuf:"varint,10,opt,name=has_inline_frames,json=hasInlineFrames,proto3" json:"has_inline_frames,omitempty"`
}

func (x *Mapping) Reset() {
	*x = Mapping{}
	if protoimpl.UnsafeEnabled {
		mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Mapping) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Mapping) ProtoMessage() {}

func (x *Mapping) ProtoReflect() protoreflect.Message {
	mi := &file_parca_metastore_v1alpha1_metastore_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Mapping.ProtoReflect.Descriptor instead.
func (*Mapping) Descriptor() ([]byte, []int) {
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP(), []int{8}
}

func (x *Mapping) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Mapping) GetStart() uint64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *Mapping) GetLimit() uint64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *Mapping) GetOffset() uint64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *Mapping) GetFile() string {
	if x != nil {
		return x.File
	}
	return ""
}

func (x *Mapping) GetBuildId() string {
	if x != nil {
		return x.BuildId
	}
	return ""
}

func (x *Mapping) GetHasFunctions() bool {
	if x != nil {
		return x.HasFunctions
	}
	return false
}

func (x *Mapping) GetHasFilenames() bool {
	if x != nil {
		return x.HasFilenames
	}
	return false
}

func (x *Mapping) GetHasLineNumbers() bool {
	if x != nil {
		return x.HasLineNumbers
	}
	return false
}

func (x *Mapping) GetHasInlineFrames() bool {
	if x != nil {
		return x.HasInlineFrames
	}
	return false
}

var File_parca_metastore_v1alpha1_metastore_proto protoreflect.FileDescriptor

var file_parca_metastore_v1alpha1_metastore_proto_rawDesc = []byte{
	0x0a, 0x28, 0x70, 0x61, 0x72, 0x63, 0x61, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72,
	0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x70, 0x61, 0x72, 0x63,
	0x61, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x22, 0xbe, 0x04, 0x0a, 0x06, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12,
	0x21, 0x0a, 0x0c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0b, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x64, 0x73, 0x12, 0x44, 0x0a, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x70, 0x61, 0x72, 0x63, 0x61, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x4e, 0x0a, 0x0a, 0x6e, 0x75, 0x6d, 0x5f,
	0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x70,
	0x61, 0x72, 0x63, 0x61, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x4e,
	0x75, 0x6d, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x6e,
	0x75, 0x6d, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x4b, 0x0a, 0x09, 0x6e, 0x75, 0x6d, 0x5f,
	0x75, 0x6e, 0x69, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x70, 0x61,
	0x72, 0x63, 0x61, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x4e, 0x75,
	0x6d, 0x55, 0x6e, 0x69, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6e, 0x75, 0x6d,
	0x55, 0x6e, 0x69, 0x74, 0x73, 0x1a, 0x60, 0x0a, 0x0b, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x3b, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x70, 0x61, 0x72, 0x63, 0x61, 0x2e, 0x6d, 0x65,
	0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x66, 0x0a, 0x0e, 0x4e, 0x75, 0x6d, 0x4c, 0x61,
	0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x3e, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x70, 0x61, 0x72,
	0x63, 0x61, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x4e, 0x75, 0x6d, 0x4c,
	0x61, 0x62, 0x65, 0x6c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a,
	0x64, 0x0a, 0x0d, 0x4e, 0x75, 0x6d, 0x55, 0x6e, 0x69, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x3d, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x27, 0x2e, 0x70, 0x61, 0x72, 0x63, 0x61, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x73, 0x74,
	0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x4e, 0x75, 0x6d, 0x55, 0x6e, 0x69, 0x74, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x25, 0x0a, 0x0b, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x4c,
	0x61, 0x62, 0x65, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x22, 0x2f, 0x0a, 0x0e,
	0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x4e, 0x75, 0x6d, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x1d,
	0x0a, 0x0a, 0x6e, 0x75, 0x6d, 0x5f, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x03, 0x52, 0x09, 0x6e, 0x75, 0x6d, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x22, 0x25, 0x0a,
	0x0d, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x4e, 0x75, 0x6d, 0x55, 0x6e, 0x69, 0x74, 0x12, 0x14,
	0x0a, 0x05, 0x75, 0x6e, 0x69, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x75,
	0x6e, 0x69, 0x74, 0x73, 0x22, 0x70, 0x0a, 0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x6d, 0x61,
	0x70, 0x70, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09,
	0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x69, 0x73, 0x5f,
	0x66, 0x6f, 0x6c, 0x64, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x69, 0x73,
	0x46, 0x6f, 0x6c, 0x64, 0x65, 0x64, 0x22, 0x55, 0x0a, 0x0d, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x4c, 0x69, 0x6e, 0x65, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x34, 0x0a, 0x05, 0x6c, 0x69, 0x6e, 0x65, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x70, 0x61, 0x72, 0x63, 0x61, 0x2e, 0x6d,
	0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x4c, 0x69, 0x6e, 0x65, 0x52, 0x05, 0x6c, 0x69, 0x6e, 0x65, 0x73, 0x22, 0x3b, 0x0a,
	0x04, 0x4c, 0x69, 0x6e, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x6c, 0x69, 0x6e, 0x65, 0x22, 0x8a, 0x01, 0x0a, 0x08, 0x46,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x5f, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x4c, 0x69, 0x6e, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x66,
	0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66,
	0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0xac, 0x02, 0x0a, 0x07, 0x4d, 0x61, 0x70, 0x70,
	0x69, 0x6e, 0x67, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d,
	0x69, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x62,
	0x75, 0x69, 0x6c, 0x64, 0x49, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x68, 0x61, 0x73, 0x5f, 0x66, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x68,
	0x61, 0x73, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x68,
	0x61, 0x73, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0c, 0x68, 0x61, 0x73, 0x46, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x12, 0x28, 0x0a, 0x10, 0x68, 0x61, 0x73, 0x5f, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x6e, 0x75, 0x6d,
	0x62, 0x65, 0x72, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x68, 0x61, 0x73, 0x4c,
	0x69, 0x6e, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x12, 0x2a, 0x0a, 0x11, 0x68, 0x61,
	0x73, 0x5f, 0x69, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x73, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x68, 0x61, 0x73, 0x49, 0x6e, 0x6c, 0x69, 0x6e, 0x65,
	0x46, 0x72, 0x61, 0x6d, 0x65, 0x73, 0x42, 0x84, 0x02, 0x0a, 0x1c, 0x63, 0x6f, 0x6d, 0x2e, 0x70,
	0x61, 0x72, 0x63, 0x61, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x0e, 0x4d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f,
	0x72, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x52, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x61, 0x72, 0x63, 0x61, 0x2d, 0x64, 0x65, 0x76, 0x2f,
	0x70, 0x61, 0x72, 0x63, 0x61, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x67, 0x6f, 0x2f, 0x70, 0x61, 0x72, 0x63, 0x61, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f,
	0x72, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x3b, 0x6d, 0x65, 0x74, 0x61,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xa2, 0x02, 0x03,
	0x50, 0x4d, 0x58, 0xaa, 0x02, 0x18, 0x50, 0x61, 0x72, 0x63, 0x61, 0x2e, 0x4d, 0x65, 0x74, 0x61,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xca, 0x02,
	0x18, 0x50, 0x61, 0x72, 0x63, 0x61, 0x5c, 0x4d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65,
	0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xe2, 0x02, 0x24, 0x50, 0x61, 0x72, 0x63,
	0x61, 0x5c, 0x4d, 0x65, 0x74, 0x61, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x5c, 0x56, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x1a, 0x50, 0x61, 0x72, 0x63, 0x61, 0x3a, 0x3a, 0x4d, 0x65, 0x74, 0x61, 0x73, 0x74,
	0x6f, 0x72, 0x65, 0x3a, 0x3a, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_parca_metastore_v1alpha1_metastore_proto_rawDescOnce sync.Once
	file_parca_metastore_v1alpha1_metastore_proto_rawDescData = file_parca_metastore_v1alpha1_metastore_proto_rawDesc
)

func file_parca_metastore_v1alpha1_metastore_proto_rawDescGZIP() []byte {
	file_parca_metastore_v1alpha1_metastore_proto_rawDescOnce.Do(func() {
		file_parca_metastore_v1alpha1_metastore_proto_rawDescData = protoimpl.X.CompressGZIP(file_parca_metastore_v1alpha1_metastore_proto_rawDescData)
	})
	return file_parca_metastore_v1alpha1_metastore_proto_rawDescData
}

var file_parca_metastore_v1alpha1_metastore_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_parca_metastore_v1alpha1_metastore_proto_goTypes = []interface{}{
	(*Sample)(nil),         // 0: parca.metastore.v1alpha1.Sample
	(*SampleLabel)(nil),    // 1: parca.metastore.v1alpha1.SampleLabel
	(*SampleNumLabel)(nil), // 2: parca.metastore.v1alpha1.SampleNumLabel
	(*SampleNumUnit)(nil),  // 3: parca.metastore.v1alpha1.SampleNumUnit
	(*Location)(nil),       // 4: parca.metastore.v1alpha1.Location
	(*LocationLines)(nil),  // 5: parca.metastore.v1alpha1.LocationLines
	(*Line)(nil),           // 6: parca.metastore.v1alpha1.Line
	(*Function)(nil),       // 7: parca.metastore.v1alpha1.Function
	(*Mapping)(nil),        // 8: parca.metastore.v1alpha1.Mapping
	nil,                    // 9: parca.metastore.v1alpha1.Sample.LabelsEntry
	nil,                    // 10: parca.metastore.v1alpha1.Sample.NumLabelsEntry
	nil,                    // 11: parca.metastore.v1alpha1.Sample.NumUnitsEntry
}
var file_parca_metastore_v1alpha1_metastore_proto_depIdxs = []int32{
	9,  // 0: parca.metastore.v1alpha1.Sample.labels:type_name -> parca.metastore.v1alpha1.Sample.LabelsEntry
	10, // 1: parca.metastore.v1alpha1.Sample.num_labels:type_name -> parca.metastore.v1alpha1.Sample.NumLabelsEntry
	11, // 2: parca.metastore.v1alpha1.Sample.num_units:type_name -> parca.metastore.v1alpha1.Sample.NumUnitsEntry
	6,  // 3: parca.metastore.v1alpha1.LocationLines.lines:type_name -> parca.metastore.v1alpha1.Line
	1,  // 4: parca.metastore.v1alpha1.Sample.LabelsEntry.value:type_name -> parca.metastore.v1alpha1.SampleLabel
	2,  // 5: parca.metastore.v1alpha1.Sample.NumLabelsEntry.value:type_name -> parca.metastore.v1alpha1.SampleNumLabel
	3,  // 6: parca.metastore.v1alpha1.Sample.NumUnitsEntry.value:type_name -> parca.metastore.v1alpha1.SampleNumUnit
	7,  // [7:7] is the sub-list for method output_type
	7,  // [7:7] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_parca_metastore_v1alpha1_metastore_proto_init() }
func file_parca_metastore_v1alpha1_metastore_proto_init() {
	if File_parca_metastore_v1alpha1_metastore_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Sample); i {
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
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SampleLabel); i {
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
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SampleNumLabel); i {
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
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SampleNumUnit); i {
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
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Location); i {
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
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LocationLines); i {
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
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Line); i {
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
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Function); i {
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
		file_parca_metastore_v1alpha1_metastore_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Mapping); i {
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
			RawDescriptor: file_parca_metastore_v1alpha1_metastore_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_parca_metastore_v1alpha1_metastore_proto_goTypes,
		DependencyIndexes: file_parca_metastore_v1alpha1_metastore_proto_depIdxs,
		MessageInfos:      file_parca_metastore_v1alpha1_metastore_proto_msgTypes,
	}.Build()
	File_parca_metastore_v1alpha1_metastore_proto = out.File
	file_parca_metastore_v1alpha1_metastore_proto_rawDesc = nil
	file_parca_metastore_v1alpha1_metastore_proto_goTypes = nil
	file_parca_metastore_v1alpha1_metastore_proto_depIdxs = nil
}
