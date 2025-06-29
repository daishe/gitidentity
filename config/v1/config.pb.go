// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: config/v1/config.proto

package configv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type VersionEntity struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Version       string                 `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VersionEntity) Reset() {
	*x = VersionEntity{}
	mi := &file_config_v1_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VersionEntity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VersionEntity) ProtoMessage() {}

func (x *VersionEntity) ProtoReflect() protoreflect.Message {
	mi := &file_config_v1_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VersionEntity.ProtoReflect.Descriptor instead.
func (*VersionEntity) Descriptor() ([]byte, []int) {
	return file_config_v1_config_proto_rawDescGZIP(), []int{0}
}

func (x *VersionEntity) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Version       string                 `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"` // for this object must equal to "v1"
	List          []*Identity            `protobuf:"bytes,2,rep,name=list,proto3" json:"list,omitempty"`       // list of targets
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_config_v1_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_config_v1_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_config_v1_config_proto_rawDescGZIP(), []int{1}
}

func (x *Config) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *Config) GetList() []*Identity {
	if x != nil {
		return x.List
	}
	return nil
}

type Identity struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`   // git user.name property
	Email         string                 `protobuf:"bytes,2,opt,name=email,proto3" json:"email,omitempty"` // git user.email property
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Identity) Reset() {
	*x = Identity{}
	mi := &file_config_v1_config_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Identity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Identity) ProtoMessage() {}

func (x *Identity) ProtoReflect() protoreflect.Message {
	mi := &file_config_v1_config_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Identity.ProtoReflect.Descriptor instead.
func (*Identity) Descriptor() ([]byte, []int) {
	return file_config_v1_config_proto_rawDescGZIP(), []int{2}
}

func (x *Identity) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Identity) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

var File_config_v1_config_proto protoreflect.FileDescriptor

const file_config_v1_config_proto_rawDesc = "" +
	"\n" +
	"\x16config/v1/config.proto\x12\x15gitidentity.config.v1\")\n" +
	"\rVersionEntity\x12\x18\n" +
	"\aversion\x18\x01 \x01(\tR\aversion\"W\n" +
	"\x06Config\x12\x18\n" +
	"\aversion\x18\x01 \x01(\tR\aversion\x123\n" +
	"\x04list\x18\x02 \x03(\v2\x1f.gitidentity.config.v1.IdentityR\x04list\"4\n" +
	"\bIdentity\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x14\n" +
	"\x05email\x18\x02 \x01(\tR\x05emailB\xd0\x01\n" +
	"\x19com.gitidentity.config.v1B\vConfigProtoP\x01Z0github.com/daishe/gitidentity/config/v1;configv1\xa2\x02\x03GCX\xaa\x02\x15Gitidentity.Config.V1\xca\x02\x15Gitidentity\\Config\\V1\xe2\x02!Gitidentity\\Config\\V1\\GPBMetadata\xea\x02\x17Gitidentity::Config::V1b\x06proto3"

var (
	file_config_v1_config_proto_rawDescOnce sync.Once
	file_config_v1_config_proto_rawDescData []byte
)

func file_config_v1_config_proto_rawDescGZIP() []byte {
	file_config_v1_config_proto_rawDescOnce.Do(func() {
		file_config_v1_config_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_config_v1_config_proto_rawDesc), len(file_config_v1_config_proto_rawDesc)))
	})
	return file_config_v1_config_proto_rawDescData
}

var file_config_v1_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_config_v1_config_proto_goTypes = []any{
	(*VersionEntity)(nil), // 0: gitidentity.config.v1.VersionEntity
	(*Config)(nil),        // 1: gitidentity.config.v1.Config
	(*Identity)(nil),      // 2: gitidentity.config.v1.Identity
}
var file_config_v1_config_proto_depIdxs = []int32{
	2, // 0: gitidentity.config.v1.Config.list:type_name -> gitidentity.config.v1.Identity
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_config_v1_config_proto_init() }
func file_config_v1_config_proto_init() {
	if File_config_v1_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_config_v1_config_proto_rawDesc), len(file_config_v1_config_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_v1_config_proto_goTypes,
		DependencyIndexes: file_config_v1_config_proto_depIdxs,
		MessageInfos:      file_config_v1_config_proto_msgTypes,
	}.Build()
	File_config_v1_config_proto = out.File
	file_config_v1_config_proto_goTypes = nil
	file_config_v1_config_proto_depIdxs = nil
}
