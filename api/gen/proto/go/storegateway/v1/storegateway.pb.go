// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: storegateway/v1/storegateway.proto

package storegatewayv1

import (
	_ "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	v1 "github.com/grafana/phlare/api/gen/proto/go/ingester/v1"
	_ "github.com/grafana/phlare/api/gen/proto/go/push/v1"
	_ "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_storegateway_v1_storegateway_proto protoreflect.FileDescriptor

var file_storegateway_v1_storegateway_proto_rawDesc = []byte{
	0x0a, 0x22, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2f, 0x76,
	0x31, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x2e, 0x76, 0x31, 0x1a, 0x17, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x76, 0x31,
	0x2f, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a,
	0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6e, 0x67, 0x65,
	0x73, 0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x70, 0x75, 0x73, 0x68,
	0x2f, 0x76, 0x31, 0x2f, 0x70, 0x75, 0x73, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x14,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x32, 0xf1, 0x02, 0x0a, 0x13, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x47, 0x61,
	0x74, 0x65, 0x77, 0x61, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7d, 0x0a, 0x18,
	0x4d, 0x65, 0x72, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x53, 0x74, 0x61,
	0x63, 0x6b, 0x74, 0x72, 0x61, 0x63, 0x65, 0x73, 0x12, 0x2c, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x72, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x66,
	0x69, 0x6c, 0x65, 0x73, 0x53, 0x74, 0x61, 0x63, 0x6b, 0x74, 0x72, 0x61, 0x63, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x72, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c,
	0x65, 0x73, 0x53, 0x74, 0x61, 0x63, 0x6b, 0x74, 0x72, 0x61, 0x63, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x12, 0x6e, 0x0a, 0x13, 0x4d,
	0x65, 0x72, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x4c, 0x61, 0x62, 0x65,
	0x6c, 0x73, 0x12, 0x27, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x4d, 0x65, 0x72, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x4c, 0x61,
	0x62, 0x65, 0x6c, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x69, 0x6e,
	0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x72, 0x67, 0x65, 0x50,
	0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x12, 0x6b, 0x0a, 0x12, 0x4d,
	0x65, 0x72, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x50, 0x70, 0x72, 0x6f,
	0x66, 0x12, 0x26, 0x2e, 0x69, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x4d, 0x65, 0x72, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x50, 0x70, 0x72,
	0x6f, 0x66, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x69, 0x6e, 0x67, 0x65,
	0x73, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x72, 0x67, 0x65, 0x50, 0x72, 0x6f,
	0x66, 0x69, 0x6c, 0x65, 0x73, 0x50, 0x70, 0x72, 0x6f, 0x66, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0xd0, 0x01, 0x0a, 0x13, 0x63, 0x6f, 0x6d,
	0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31,
	0x42, 0x11, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x49, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x67, 0x72, 0x61, 0x66, 0x61, 0x6e, 0x61, 0x2f, 0x70, 0x68, 0x6c, 0x61, 0x72, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67,
	0x6f, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2f, 0x76,
	0x31, 0x3b, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x76, 0x31,
	0xa2, 0x02, 0x03, 0x53, 0x58, 0x58, 0xaa, 0x02, 0x0f, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x67, 0x61,
	0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0f, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x1b, 0x53, 0x74, 0x6f,
	0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x10, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var file_storegateway_v1_storegateway_proto_goTypes = []interface{}{
	(*v1.MergeProfilesStacktracesRequest)(nil),  // 0: ingester.v1.MergeProfilesStacktracesRequest
	(*v1.MergeProfilesLabelsRequest)(nil),       // 1: ingester.v1.MergeProfilesLabelsRequest
	(*v1.MergeProfilesPprofRequest)(nil),        // 2: ingester.v1.MergeProfilesPprofRequest
	(*v1.MergeProfilesStacktracesResponse)(nil), // 3: ingester.v1.MergeProfilesStacktracesResponse
	(*v1.MergeProfilesLabelsResponse)(nil),      // 4: ingester.v1.MergeProfilesLabelsResponse
	(*v1.MergeProfilesPprofResponse)(nil),       // 5: ingester.v1.MergeProfilesPprofResponse
}
var file_storegateway_v1_storegateway_proto_depIdxs = []int32{
	0, // 0: storegateway.v1.StoreGatewayService.MergeProfilesStacktraces:input_type -> ingester.v1.MergeProfilesStacktracesRequest
	1, // 1: storegateway.v1.StoreGatewayService.MergeProfilesLabels:input_type -> ingester.v1.MergeProfilesLabelsRequest
	2, // 2: storegateway.v1.StoreGatewayService.MergeProfilesPprof:input_type -> ingester.v1.MergeProfilesPprofRequest
	3, // 3: storegateway.v1.StoreGatewayService.MergeProfilesStacktraces:output_type -> ingester.v1.MergeProfilesStacktracesResponse
	4, // 4: storegateway.v1.StoreGatewayService.MergeProfilesLabels:output_type -> ingester.v1.MergeProfilesLabelsResponse
	5, // 5: storegateway.v1.StoreGatewayService.MergeProfilesPprof:output_type -> ingester.v1.MergeProfilesPprofResponse
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_storegateway_v1_storegateway_proto_init() }
func file_storegateway_v1_storegateway_proto_init() {
	if File_storegateway_v1_storegateway_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_storegateway_v1_storegateway_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_storegateway_v1_storegateway_proto_goTypes,
		DependencyIndexes: file_storegateway_v1_storegateway_proto_depIdxs,
	}.Build()
	File_storegateway_v1_storegateway_proto = out.File
	file_storegateway_v1_storegateway_proto_rawDesc = nil
	file_storegateway_v1_storegateway_proto_goTypes = nil
	file_storegateway_v1_storegateway_proto_depIdxs = nil
}
