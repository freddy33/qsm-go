// Code generated by protoc-gen-go. DO NOT EDIT.
// source: m3path.proto

package m3api

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type PathContextRequestMsg struct {
	GrowthType           int32    `protobuf:"varint,1,opt,name=growth_type,json=growthType,proto3" json:"growth_type"`
	GrowthIndex          int32    `protobuf:"varint,2,opt,name=growth_index,json=growthIndex,proto3" json:"growth_index"`
	GrowthOffset         int32    `protobuf:"varint,3,opt,name=growth_offset,json=growthOffset,proto3" json:"growth_offset"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PathContextRequestMsg) Reset()         { *m = PathContextRequestMsg{} }
func (m *PathContextRequestMsg) String() string { return proto.CompactTextString(m) }
func (*PathContextRequestMsg) ProtoMessage()    {}
func (*PathContextRequestMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8f5151eb02dd926, []int{0}
}

func (m *PathContextRequestMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PathContextRequestMsg.Unmarshal(m, b)
}
func (m *PathContextRequestMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PathContextRequestMsg.Marshal(b, m, deterministic)
}
func (m *PathContextRequestMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PathContextRequestMsg.Merge(m, src)
}
func (m *PathContextRequestMsg) XXX_Size() int {
	return xxx_messageInfo_PathContextRequestMsg.Size(m)
}
func (m *PathContextRequestMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_PathContextRequestMsg.DiscardUnknown(m)
}

var xxx_messageInfo_PathContextRequestMsg proto.InternalMessageInfo

func (m *PathContextRequestMsg) GetGrowthType() int32 {
	if m != nil {
		return m.GrowthType
	}
	return 0
}

func (m *PathContextRequestMsg) GetGrowthIndex() int32 {
	if m != nil {
		return m.GrowthIndex
	}
	return 0
}

func (m *PathContextRequestMsg) GetGrowthOffset() int32 {
	if m != nil {
		return m.GrowthOffset
	}
	return 0
}

type PathNodeMsg struct {
	PathNodeId           int64     `protobuf:"varint,1,opt,name=path_node_id,json=pathNodeId,proto3" json:"path_node_id"`
	Point                *PointMsg `protobuf:"bytes,2,opt,name=point,proto3" json:"point,omitempty"`
	D                    int32     `protobuf:"varint,3,opt,name=d,proto3" json:"d"`
	TrioId               int32     `protobuf:"varint,4,opt,name=trio_id,json=trioId,proto3" json:"trio_id"`
	ConnectionMask       uint32    `protobuf:"varint,5,opt,name=connection_mask,json=connectionMask,proto3" json:"connection_mask,omitempty"`
	LinkedPathNodeIds    []int64   `protobuf:"varint,6,rep,packed,name=linked_path_node_ids,json=linkedPathNodeIds,proto3" json:"linked_path_node_ids,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *PathNodeMsg) Reset()         { *m = PathNodeMsg{} }
func (m *PathNodeMsg) String() string { return proto.CompactTextString(m) }
func (*PathNodeMsg) ProtoMessage()    {}
func (*PathNodeMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8f5151eb02dd926, []int{1}
}

func (m *PathNodeMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PathNodeMsg.Unmarshal(m, b)
}
func (m *PathNodeMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PathNodeMsg.Marshal(b, m, deterministic)
}
func (m *PathNodeMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PathNodeMsg.Merge(m, src)
}
func (m *PathNodeMsg) XXX_Size() int {
	return xxx_messageInfo_PathNodeMsg.Size(m)
}
func (m *PathNodeMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_PathNodeMsg.DiscardUnknown(m)
}

var xxx_messageInfo_PathNodeMsg proto.InternalMessageInfo

func (m *PathNodeMsg) GetPathNodeId() int64 {
	if m != nil {
		return m.PathNodeId
	}
	return 0
}

func (m *PathNodeMsg) GetPoint() *PointMsg {
	if m != nil {
		return m.Point
	}
	return nil
}

func (m *PathNodeMsg) GetD() int32 {
	if m != nil {
		return m.D
	}
	return 0
}

func (m *PathNodeMsg) GetTrioId() int32 {
	if m != nil {
		return m.TrioId
	}
	return 0
}

func (m *PathNodeMsg) GetConnectionMask() uint32 {
	if m != nil {
		return m.ConnectionMask
	}
	return 0
}

func (m *PathNodeMsg) GetLinkedPathNodeIds() []int64 {
	if m != nil {
		return m.LinkedPathNodeIds
	}
	return nil
}

type PathContextResponseMsg struct {
	PathCtxId            int32        `protobuf:"varint,1,opt,name=path_ctx_id,json=pathCtxId,proto3" json:"path_ctx_id"`
	GrowthContextId      int32        `protobuf:"varint,2,opt,name=growth_context_id,json=growthContextId,proto3" json:"growth_context_id"`
	GrowthOffset         int32        `protobuf:"varint,3,opt,name=growth_offset,json=growthOffset,proto3" json:"growth_offset"`
	RootPathNode         *PathNodeMsg `protobuf:"bytes,4,opt,name=root_path_node,json=rootPathNode,proto3" json:"root_path_node,omitempty"`
	MaxDist              int32        `protobuf:"varint,5,opt,name=max_dist,json=maxDist,proto3" json:"max_dist"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *PathContextResponseMsg) Reset()         { *m = PathContextResponseMsg{} }
func (m *PathContextResponseMsg) String() string { return proto.CompactTextString(m) }
func (*PathContextResponseMsg) ProtoMessage()    {}
func (*PathContextResponseMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8f5151eb02dd926, []int{2}
}

func (m *PathContextResponseMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PathContextResponseMsg.Unmarshal(m, b)
}
func (m *PathContextResponseMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PathContextResponseMsg.Marshal(b, m, deterministic)
}
func (m *PathContextResponseMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PathContextResponseMsg.Merge(m, src)
}
func (m *PathContextResponseMsg) XXX_Size() int {
	return xxx_messageInfo_PathContextResponseMsg.Size(m)
}
func (m *PathContextResponseMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_PathContextResponseMsg.DiscardUnknown(m)
}

var xxx_messageInfo_PathContextResponseMsg proto.InternalMessageInfo

func (m *PathContextResponseMsg) GetPathCtxId() int32 {
	if m != nil {
		return m.PathCtxId
	}
	return 0
}

func (m *PathContextResponseMsg) GetGrowthContextId() int32 {
	if m != nil {
		return m.GrowthContextId
	}
	return 0
}

func (m *PathContextResponseMsg) GetGrowthOffset() int32 {
	if m != nil {
		return m.GrowthOffset
	}
	return 0
}

func (m *PathContextResponseMsg) GetRootPathNode() *PathNodeMsg {
	if m != nil {
		return m.RootPathNode
	}
	return nil
}

func (m *PathContextResponseMsg) GetMaxDist() int32 {
	if m != nil {
		return m.MaxDist
	}
	return 0
}

type PathNodesRequestMsg struct {
	PathCtxId            int32    `protobuf:"varint,1,opt,name=path_ctx_id,json=pathCtxId,proto3" json:"path_ctx_id" query:"path_ctx_id"`
	Dist                 int32    `protobuf:"varint,2,opt,name=dist,proto3" json:"dist" query:"dist"`
	ToDist               int32    `protobuf:"varint,3,opt,name=to_dist,json=toDist,proto3" json:"to_dist" query:"to_dist"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PathNodesRequestMsg) Reset()         { *m = PathNodesRequestMsg{} }
func (m *PathNodesRequestMsg) String() string { return proto.CompactTextString(m) }
func (*PathNodesRequestMsg) ProtoMessage()    {}
func (*PathNodesRequestMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8f5151eb02dd926, []int{3}
}

func (m *PathNodesRequestMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PathNodesRequestMsg.Unmarshal(m, b)
}
func (m *PathNodesRequestMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PathNodesRequestMsg.Marshal(b, m, deterministic)
}
func (m *PathNodesRequestMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PathNodesRequestMsg.Merge(m, src)
}
func (m *PathNodesRequestMsg) XXX_Size() int {
	return xxx_messageInfo_PathNodesRequestMsg.Size(m)
}
func (m *PathNodesRequestMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_PathNodesRequestMsg.DiscardUnknown(m)
}

var xxx_messageInfo_PathNodesRequestMsg proto.InternalMessageInfo

func (m *PathNodesRequestMsg) GetPathCtxId() int32 {
	if m != nil {
		return m.PathCtxId
	}
	return 0
}

func (m *PathNodesRequestMsg) GetDist() int32 {
	if m != nil {
		return m.Dist
	}
	return 0
}

func (m *PathNodesRequestMsg) GetToDist() int32 {
	if m != nil {
		return m.ToDist
	}
	return 0
}

type PathNodesResponseMsg struct {
	PathCtxId            int32          `protobuf:"varint,1,opt,name=path_ctx_id,json=pathCtxId,proto3" json:"path_ctx_id"`
	Dist                 int32          `protobuf:"varint,2,opt,name=dist,proto3" json:"dist"`
	ToDist               int32          `protobuf:"varint,4,opt,name=to_dist,json=toDist,proto3" json:"to_dist"`
	MaxDist              int32          `protobuf:"varint,5,opt,name=max_dist,json=maxDist,proto3" json:"max_dist"`
	NbPathNodes          int32          `protobuf:"varint,6,opt,name=nb_path_nodes,json=nbPathNodes,proto3" json:"nb_path_nodes,omitempty"`
	PathNodes            []*PathNodeMsg `protobuf:"bytes,3,rep,name=path_nodes,json=pathNodes,proto3" json:"path_nodes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *PathNodesResponseMsg) Reset()         { *m = PathNodesResponseMsg{} }
func (m *PathNodesResponseMsg) String() string { return proto.CompactTextString(m) }
func (*PathNodesResponseMsg) ProtoMessage()    {}
func (*PathNodesResponseMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8f5151eb02dd926, []int{4}
}

func (m *PathNodesResponseMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PathNodesResponseMsg.Unmarshal(m, b)
}
func (m *PathNodesResponseMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PathNodesResponseMsg.Marshal(b, m, deterministic)
}
func (m *PathNodesResponseMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PathNodesResponseMsg.Merge(m, src)
}
func (m *PathNodesResponseMsg) XXX_Size() int {
	return xxx_messageInfo_PathNodesResponseMsg.Size(m)
}
func (m *PathNodesResponseMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_PathNodesResponseMsg.DiscardUnknown(m)
}

var xxx_messageInfo_PathNodesResponseMsg proto.InternalMessageInfo

func (m *PathNodesResponseMsg) GetPathCtxId() int32 {
	if m != nil {
		return m.PathCtxId
	}
	return 0
}

func (m *PathNodesResponseMsg) GetDist() int32 {
	if m != nil {
		return m.Dist
	}
	return 0
}

func (m *PathNodesResponseMsg) GetToDist() int32 {
	if m != nil {
		return m.ToDist
	}
	return 0
}

func (m *PathNodesResponseMsg) GetMaxDist() int32 {
	if m != nil {
		return m.MaxDist
	}
	return 0
}

func (m *PathNodesResponseMsg) GetNbPathNodes() int32 {
	if m != nil {
		return m.NbPathNodes
	}
	return 0
}

func (m *PathNodesResponseMsg) GetPathNodes() []*PathNodeMsg {
	if m != nil {
		return m.PathNodes
	}
	return nil
}

func init() {
	proto.RegisterType((*PathContextRequestMsg)(nil), "m3api.PathContextRequestMsg")
	proto.RegisterType((*PathNodeMsg)(nil), "m3api.PathNodeMsg")
	proto.RegisterType((*PathContextResponseMsg)(nil), "m3api.PathContextResponseMsg")
	proto.RegisterType((*PathNodesRequestMsg)(nil), "m3api.PathNodesRequestMsg")
	proto.RegisterType((*PathNodesResponseMsg)(nil), "m3api.PathNodesResponseMsg")
}

func init() {
	proto.RegisterFile("m3path.proto", fileDescriptor_e8f5151eb02dd926)
}

var fileDescriptor_e8f5151eb02dd926 = []byte{
	// 456 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xcf, 0x6f, 0xd3, 0x30,
	0x14, 0x96, 0xc9, 0xd2, 0xb1, 0x97, 0x74, 0xd5, 0xcc, 0x80, 0xc0, 0x01, 0x42, 0x10, 0xa2, 0xe2,
	0x50, 0xc4, 0x7a, 0xe1, 0x3e, 0x2e, 0x39, 0x14, 0x2a, 0x8b, 0x7b, 0x94, 0xd6, 0x5e, 0x6b, 0x95,
	0xd8, 0xa1, 0x36, 0x22, 0xbb, 0xf1, 0x4f, 0x22, 0x6e, 0xfc, 0x2d, 0xc8, 0xcf, 0xce, 0x1a, 0x0e,
	0xb0, 0xdd, 0xda, 0xcf, 0x9f, 0xdf, 0xf7, 0xe3, 0x39, 0x90, 0x36, 0xf3, 0xb6, 0xb6, 0xdb, 0x59,
	0xbb, 0xd7, 0x56, 0xd3, 0xb8, 0x99, 0xd7, 0xad, 0x7c, 0x3a, 0x6e, 0xe6, 0xad, 0x96, 0xca, 0x7a,
	0xb4, 0xf8, 0x41, 0xe0, 0xe1, 0xb2, 0xb6, 0xdb, 0x4b, 0xad, 0xac, 0xe8, 0x2c, 0x13, 0x5f, 0xbf,
	0x09, 0x63, 0x17, 0x66, 0x43, 0x9f, 0x43, 0xb2, 0xd9, 0xeb, 0xef, 0x76, 0x5b, 0xd9, 0xeb, 0x56,
	0x64, 0x24, 0x27, 0xd3, 0x98, 0x81, 0x87, 0x3e, 0x5f, 0xb7, 0x82, 0xbe, 0x80, 0x34, 0x10, 0xa4,
	0xe2, 0xa2, 0xcb, 0xee, 0x21, 0x23, 0x5c, 0x2a, 0x1d, 0x44, 0x5f, 0xc2, 0x38, 0x50, 0xf4, 0xd5,
	0x95, 0x11, 0x36, 0x8b, 0x90, 0x13, 0xee, 0x7d, 0x42, 0xac, 0xf8, 0x45, 0x20, 0x71, 0x16, 0x3e,
	0x6a, 0x2e, 0x9c, 0x70, 0x0e, 0xa9, 0xb3, 0x5d, 0x29, 0xcd, 0x45, 0x25, 0x39, 0x2a, 0x47, 0x0c,
	0xda, 0x40, 0x29, 0x39, 0x7d, 0x05, 0x31, 0x66, 0x40, 0xc9, 0xe4, 0x62, 0x32, 0xc3, 0x68, 0xb3,
	0xa5, 0xc3, 0x16, 0x66, 0xc3, 0xfc, 0x29, 0x4d, 0x81, 0xf0, 0xa0, 0x48, 0x38, 0x7d, 0x0c, 0xc7,
	0x76, 0x2f, 0xb5, 0x9b, 0x78, 0x84, 0xd8, 0xc8, 0xfd, 0x2d, 0x39, 0x7d, 0x0d, 0x93, 0xb5, 0x56,
	0x4a, 0xac, 0xad, 0xd4, 0xaa, 0x6a, 0x6a, 0xb3, 0xcb, 0xe2, 0x9c, 0x4c, 0xc7, 0xec, 0xf4, 0x00,
	0x2f, 0x6a, 0xb3, 0xa3, 0x6f, 0xe1, 0xfc, 0x8b, 0x54, 0x3b, 0xc1, 0xab, 0xa1, 0x3f, 0x93, 0x8d,
	0xf2, 0x68, 0x1a, 0xb1, 0x33, 0x7f, 0xb6, 0xbc, 0xb1, 0x69, 0x8a, 0xdf, 0x04, 0x1e, 0xfd, 0x55,
	0xae, 0x69, 0xb5, 0x32, 0x18, 0xf2, 0x19, 0x24, 0x38, 0x64, 0x6d, 0xbb, 0x3e, 0x63, 0xcc, 0x4e,
	0x1c, 0x74, 0x69, 0xbb, 0x92, 0xd3, 0x37, 0x70, 0x16, 0x9a, 0x5b, 0xfb, 0xcb, 0x8e, 0xe5, 0x1b,
	0x9e, 0xf8, 0x83, 0x30, 0xb4, 0xe4, 0x77, 0x6a, 0x99, 0xbe, 0x87, 0xd3, 0xbd, 0xd6, 0xf6, 0x60,
	0x1d, 0x5b, 0x48, 0x2e, 0x68, 0x5f, 0xde, 0x61, 0x03, 0x2c, 0x75, 0xcc, 0x1e, 0xa0, 0x4f, 0xe0,
	0x7e, 0x53, 0x77, 0x15, 0x97, 0xc6, 0x62, 0x31, 0x31, 0x3b, 0x6e, 0xea, 0xee, 0x83, 0x34, 0xb6,
	0x58, 0xc1, 0x83, 0x9e, 0x66, 0x06, 0x4f, 0xe7, 0xb6, 0x70, 0x14, 0x8e, 0x70, 0x9a, 0xcf, 0x83,
	0xbf, 0x71, 0x3d, 0xda, 0x8b, 0x44, 0x61, 0x3d, 0x1a, 0x35, 0x7e, 0x12, 0x38, 0x1f, 0x88, 0xdc,
	0xbd, 0xc2, 0x5b, 0x54, 0x8e, 0x86, 0x2a, 0xff, 0x09, 0x49, 0x0b, 0x18, 0xab, 0xd5, 0xa1, 0x37,
	0xb7, 0x6f, 0x7c, 0xe8, 0x6a, 0x75, 0x63, 0x8b, 0xbe, 0x03, 0x18, 0x10, 0xa2, 0x3c, 0xfa, 0x47,
	0xb3, 0x27, 0xfd, 0x2b, 0x36, 0xab, 0x11, 0x7e, 0x80, 0xf3, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff,
	0xf1, 0x8e, 0xef, 0xe5, 0xa6, 0x03, 0x00, 0x00,
}
