// Code generated by protoc-gen-go. DO NOT EDIT.
// source: m3space.proto

package m3api

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

type SpaceMsg struct {
	SpaceId              int32    `protobuf:"varint,1,opt,name=space_id,json=spaceId,proto3" json:"space_id" query:"space_id"`
	SpaceName            string   `protobuf:"bytes,2,opt,name=space_name,json=spaceName,proto3" json:"space_name" query:"space_name"`
	ActiveThreshold      int32    `protobuf:"varint,3,opt,name=active_threshold,json=activeThreshold,proto3" json:"active_threshold" query:"active_threshold"`
	MaxTriosPerPoint     int32    `protobuf:"varint,4,opt,name=max_trios_per_point,json=maxTriosPerPoint,proto3" json:"max_trios_per_point" query:"max_trios_per_point"`
	MaxNodesPerPoint     int32    `protobuf:"varint,5,opt,name=max_nodes_per_point,json=maxNodesPerPoint,proto3" json:"max_nodes_per_point" query:"max_nodes_per_point"`
	MaxTime              int32    `protobuf:"varint,6,opt,name=max_time,json=maxTime,proto3" json:"max_time" query:"max_time"`
	MaxCoord             int32    `protobuf:"varint,8,opt,name=max_coord,json=maxCoord,proto3" json:"max_coord" query:"max_coord"`
	EventIds             []int32  `protobuf:"varint,9,rep,packed,name=event_ids,json=eventIds,proto3" json:"event_ids,omitempty" query:"event_ids"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" query:"-"`
	XXX_unrecognized     []byte   `json:"-" query:"-"`
	XXX_sizecache        int32    `json:"-" query:"-"`
}

func (m *SpaceMsg) Reset()         { *m = SpaceMsg{} }
func (m *SpaceMsg) String() string { return proto.CompactTextString(m) }
func (*SpaceMsg) ProtoMessage()    {}
func (*SpaceMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{0}
}

func (m *SpaceMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SpaceMsg.Unmarshal(m, b)
}
func (m *SpaceMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SpaceMsg.Marshal(b, m, deterministic)
}
func (m *SpaceMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpaceMsg.Merge(m, src)
}
func (m *SpaceMsg) XXX_Size() int {
	return xxx_messageInfo_SpaceMsg.Size(m)
}
func (m *SpaceMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SpaceMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SpaceMsg proto.InternalMessageInfo

func (m *SpaceMsg) GetSpaceId() int32 {
	if m != nil {
		return m.SpaceId
	}
	return 0
}

func (m *SpaceMsg) GetSpaceName() string {
	if m != nil {
		return m.SpaceName
	}
	return ""
}

func (m *SpaceMsg) GetActiveThreshold() int32 {
	if m != nil {
		return m.ActiveThreshold
	}
	return 0
}

func (m *SpaceMsg) GetMaxTriosPerPoint() int32 {
	if m != nil {
		return m.MaxTriosPerPoint
	}
	return 0
}

func (m *SpaceMsg) GetMaxNodesPerPoint() int32 {
	if m != nil {
		return m.MaxNodesPerPoint
	}
	return 0
}

func (m *SpaceMsg) GetMaxTime() int32 {
	if m != nil {
		return m.MaxTime
	}
	return 0
}

func (m *SpaceMsg) GetMaxCoord() int32 {
	if m != nil {
		return m.MaxCoord
	}
	return 0
}

func (m *SpaceMsg) GetEventIds() []int32 {
	if m != nil {
		return m.EventIds
	}
	return nil
}

type SpaceListMsg struct {
	Spaces               []*SpaceMsg `protobuf:"bytes,1,rep,name=spaces,proto3" json:"spaces,omitempty" query:"-"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-" query:"-"`
	XXX_unrecognized     []byte      `json:"-" query:"-"`
	XXX_sizecache        int32       `json:"-" query:"-"`
}

func (m *SpaceListMsg) Reset()         { *m = SpaceListMsg{} }
func (m *SpaceListMsg) String() string { return proto.CompactTextString(m) }
func (*SpaceListMsg) ProtoMessage()    {}
func (*SpaceListMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{1}
}

func (m *SpaceListMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SpaceListMsg.Unmarshal(m, b)
}
func (m *SpaceListMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SpaceListMsg.Marshal(b, m, deterministic)
}
func (m *SpaceListMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpaceListMsg.Merge(m, src)
}
func (m *SpaceListMsg) XXX_Size() int {
	return xxx_messageInfo_SpaceListMsg.Size(m)
}
func (m *SpaceListMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SpaceListMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SpaceListMsg proto.InternalMessageInfo

func (m *SpaceListMsg) GetSpaces() []*SpaceMsg {
	if m != nil {
		return m.Spaces
	}
	return nil
}

type CreateEventRequestMsg struct {
	SpaceId              int32     `protobuf:"varint,2,opt,name=space_id,json=spaceId,proto3" json:"space_id" query:"space_id"`
	GrowthType           int32     `protobuf:"varint,3,opt,name=growth_type,json=growthType,proto3" json:"growth_type" query:"growth_type"`
	GrowthIndex          int32     `protobuf:"varint,4,opt,name=growth_index,json=growthIndex,proto3" json:"growth_index" query:"growth_index"`
	GrowthOffset         int32     `protobuf:"varint,5,opt,name=growth_offset,json=growthOffset,proto3" json:"growth_offset" query:"growth_offset"`
	CreationTime         int32     `protobuf:"varint,6,opt,name=creation_time,json=creationTime,proto3" json:"creation_time" query:"creation_time"`
	Center               *PointMsg `protobuf:"bytes,7,opt,name=center,proto3" json:"center,omitempty" query:"center"`
	Color                uint32    `protobuf:"varint,8,opt,name=color,proto3" json:"color" query:"color"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-" query:"-"`
	XXX_unrecognized     []byte    `json:"-" query:"-"`
	XXX_sizecache        int32     `json:"-" query:"-"`
}

func (m *CreateEventRequestMsg) Reset()         { *m = CreateEventRequestMsg{} }
func (m *CreateEventRequestMsg) String() string { return proto.CompactTextString(m) }
func (*CreateEventRequestMsg) ProtoMessage()    {}
func (*CreateEventRequestMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{2}
}

func (m *CreateEventRequestMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateEventRequestMsg.Unmarshal(m, b)
}
func (m *CreateEventRequestMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateEventRequestMsg.Marshal(b, m, deterministic)
}
func (m *CreateEventRequestMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateEventRequestMsg.Merge(m, src)
}
func (m *CreateEventRequestMsg) XXX_Size() int {
	return xxx_messageInfo_CreateEventRequestMsg.Size(m)
}
func (m *CreateEventRequestMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateEventRequestMsg.DiscardUnknown(m)
}

var xxx_messageInfo_CreateEventRequestMsg proto.InternalMessageInfo

func (m *CreateEventRequestMsg) GetSpaceId() int32 {
	if m != nil {
		return m.SpaceId
	}
	return 0
}

func (m *CreateEventRequestMsg) GetGrowthType() int32 {
	if m != nil {
		return m.GrowthType
	}
	return 0
}

func (m *CreateEventRequestMsg) GetGrowthIndex() int32 {
	if m != nil {
		return m.GrowthIndex
	}
	return 0
}

func (m *CreateEventRequestMsg) GetGrowthOffset() int32 {
	if m != nil {
		return m.GrowthOffset
	}
	return 0
}

func (m *CreateEventRequestMsg) GetCreationTime() int32 {
	if m != nil {
		return m.CreationTime
	}
	return 0
}

func (m *CreateEventRequestMsg) GetCenter() *PointMsg {
	if m != nil {
		return m.Center
	}
	return nil
}

func (m *CreateEventRequestMsg) GetColor() uint32 {
	if m != nil {
		return m.Color
	}
	return 0
}

type NodeEventMsg struct {
	EventNodeId          int64     `protobuf:"varint,1,opt,name=event_node_id,json=eventNodeId,proto3" json:"event_node_id" query:"event_node_id"`
	EventId              int32     `protobuf:"varint,2,opt,name=event_id,json=eventId,proto3" json:"event_id" query:"event_id"`
	Point                *PointMsg `protobuf:"bytes,3,opt,name=point,proto3" json:"point,omitempty" query:"point"`
	CreationTime         int32     `protobuf:"varint,4,opt,name=creation_time,json=creationTime,proto3" json:"creation_time" query:"creation_time"`
	D                    int32     `protobuf:"varint,5,opt,name=d,proto3" json:"d" query:"d"`
	TrioId               int32     `protobuf:"varint,6,opt,name=trio_id,json=trioId,proto3" json:"trio_id" query:"trio_id"`
	ConnectionMask       uint32    `protobuf:"varint,7,opt,name=connection_mask,json=connectionMask,proto3" json:"connection_mask" query:"connection_mask"`
	PathNodeId           int64     `protobuf:"varint,8,opt,name=path_node_id,json=pathNodeId,proto3" json:"path_node_id" query:"path_node_id"`
	LinkedNodeIds        []int64   `protobuf:"varint,9,rep,packed,name=linked_node_ids,json=linkedNodeIds,proto3" json:"linked_node_ids,omitempty" query:"linked_node_ids"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-" query:"-"`
	XXX_unrecognized     []byte    `json:"-" query:"-"`
	XXX_sizecache        int32     `json:"-" query:"-"`
}

func (m *NodeEventMsg) Reset()         { *m = NodeEventMsg{} }
func (m *NodeEventMsg) String() string { return proto.CompactTextString(m) }
func (*NodeEventMsg) ProtoMessage()    {}
func (*NodeEventMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{3}
}

func (m *NodeEventMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeEventMsg.Unmarshal(m, b)
}
func (m *NodeEventMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeEventMsg.Marshal(b, m, deterministic)
}
func (m *NodeEventMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeEventMsg.Merge(m, src)
}
func (m *NodeEventMsg) XXX_Size() int {
	return xxx_messageInfo_NodeEventMsg.Size(m)
}
func (m *NodeEventMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeEventMsg.DiscardUnknown(m)
}

var xxx_messageInfo_NodeEventMsg proto.InternalMessageInfo

func (m *NodeEventMsg) GetEventNodeId() int64 {
	if m != nil {
		return m.EventNodeId
	}
	return 0
}

func (m *NodeEventMsg) GetEventId() int32 {
	if m != nil {
		return m.EventId
	}
	return 0
}

func (m *NodeEventMsg) GetPoint() *PointMsg {
	if m != nil {
		return m.Point
	}
	return nil
}

func (m *NodeEventMsg) GetCreationTime() int32 {
	if m != nil {
		return m.CreationTime
	}
	return 0
}

func (m *NodeEventMsg) GetD() int32 {
	if m != nil {
		return m.D
	}
	return 0
}

func (m *NodeEventMsg) GetTrioId() int32 {
	if m != nil {
		return m.TrioId
	}
	return 0
}

func (m *NodeEventMsg) GetConnectionMask() uint32 {
	if m != nil {
		return m.ConnectionMask
	}
	return 0
}

func (m *NodeEventMsg) GetPathNodeId() int64 {
	if m != nil {
		return m.PathNodeId
	}
	return 0
}

func (m *NodeEventMsg) GetLinkedNodeIds() []int64 {
	if m != nil {
		return m.LinkedNodeIds
	}
	return nil
}

type FindEventsMsg struct {
	EventId              int32    `protobuf:"varint,1,opt,name=event_id,json=eventId,proto3" json:"event_id" query:"event_id"`
	SpaceId              int32    `protobuf:"varint,2,opt,name=space_id,json=spaceId,proto3" json:"space_id" query:"space_id"`
	AtTime               int32    `protobuf:"varint,3,opt,name=at_time,json=atTime,proto3" json:"at_time" query:"at_time"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" query:"-"`
	XXX_unrecognized     []byte   `json:"-" query:"-"`
	XXX_sizecache        int32    `json:"-" query:"-"`
}

func (m *FindEventsMsg) Reset()         { *m = FindEventsMsg{} }
func (m *FindEventsMsg) String() string { return proto.CompactTextString(m) }
func (*FindEventsMsg) ProtoMessage()    {}
func (*FindEventsMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{4}
}

func (m *FindEventsMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FindEventsMsg.Unmarshal(m, b)
}
func (m *FindEventsMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FindEventsMsg.Marshal(b, m, deterministic)
}
func (m *FindEventsMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FindEventsMsg.Merge(m, src)
}
func (m *FindEventsMsg) XXX_Size() int {
	return xxx_messageInfo_FindEventsMsg.Size(m)
}
func (m *FindEventsMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_FindEventsMsg.DiscardUnknown(m)
}

var xxx_messageInfo_FindEventsMsg proto.InternalMessageInfo

func (m *FindEventsMsg) GetEventId() int32 {
	if m != nil {
		return m.EventId
	}
	return 0
}

func (m *FindEventsMsg) GetSpaceId() int32 {
	if m != nil {
		return m.SpaceId
	}
	return 0
}

func (m *FindEventsMsg) GetAtTime() int32 {
	if m != nil {
		return m.AtTime
	}
	return 0
}

type EventMsg struct {
	EventId              int32         `protobuf:"varint,1,opt,name=event_id,json=eventId,proto3" json:"event_id" query:"event_id"`
	SpaceId              int32         `protobuf:"varint,2,opt,name=space_id,json=spaceId,proto3" json:"space_id" query:"space_id"`
	GrowthType           int32         `protobuf:"varint,3,opt,name=growth_type,json=growthType,proto3" json:"growth_type" query:"growth_type"`
	GrowthIndex          int32         `protobuf:"varint,4,opt,name=growth_index,json=growthIndex,proto3" json:"growth_index" query:"growth_index"`
	GrowthOffset         int32         `protobuf:"varint,5,opt,name=growth_offset,json=growthOffset,proto3" json:"growth_offset" query:"growth_offset"`
	CreationTime         int32         `protobuf:"varint,6,opt,name=creation_time,json=creationTime,proto3" json:"creation_time" query:"creation_time"`
	PathCtxId            int32         `protobuf:"varint,7,opt,name=path_ctx_id,json=pathCtxId,proto3" json:"path_ctx_id" query:"path_ctx_id"`
	Color                uint32        `protobuf:"varint,8,opt,name=color,proto3" json:"color" query:"color"`
	RootNode             *NodeEventMsg `protobuf:"bytes,9,opt,name=root_node,json=rootNode,proto3" json:"root_node,omitempty" query:"-"`
	MaxNodeTime          int32         `protobuf:"varint,10,opt,name=max_node_time,json=maxNodeTime,proto3" json:"max_node_time" query:"max_node_time"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-" query:"-"`
	XXX_unrecognized     []byte        `json:"-" query:"-"`
	XXX_sizecache        int32         `json:"-" query:"-"`
}

func (m *EventMsg) Reset()         { *m = EventMsg{} }
func (m *EventMsg) String() string { return proto.CompactTextString(m) }
func (*EventMsg) ProtoMessage()    {}
func (*EventMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{5}
}

func (m *EventMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EventMsg.Unmarshal(m, b)
}
func (m *EventMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EventMsg.Marshal(b, m, deterministic)
}
func (m *EventMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventMsg.Merge(m, src)
}
func (m *EventMsg) XXX_Size() int {
	return xxx_messageInfo_EventMsg.Size(m)
}
func (m *EventMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_EventMsg.DiscardUnknown(m)
}

var xxx_messageInfo_EventMsg proto.InternalMessageInfo

func (m *EventMsg) GetEventId() int32 {
	if m != nil {
		return m.EventId
	}
	return 0
}

func (m *EventMsg) GetSpaceId() int32 {
	if m != nil {
		return m.SpaceId
	}
	return 0
}

func (m *EventMsg) GetGrowthType() int32 {
	if m != nil {
		return m.GrowthType
	}
	return 0
}

func (m *EventMsg) GetGrowthIndex() int32 {
	if m != nil {
		return m.GrowthIndex
	}
	return 0
}

func (m *EventMsg) GetGrowthOffset() int32 {
	if m != nil {
		return m.GrowthOffset
	}
	return 0
}

func (m *EventMsg) GetCreationTime() int32 {
	if m != nil {
		return m.CreationTime
	}
	return 0
}

func (m *EventMsg) GetPathCtxId() int32 {
	if m != nil {
		return m.PathCtxId
	}
	return 0
}

func (m *EventMsg) GetColor() uint32 {
	if m != nil {
		return m.Color
	}
	return 0
}

func (m *EventMsg) GetRootNode() *NodeEventMsg {
	if m != nil {
		return m.RootNode
	}
	return nil
}

func (m *EventMsg) GetMaxNodeTime() int32 {
	if m != nil {
		return m.MaxNodeTime
	}
	return 0
}

type EventListMsg struct {
	Events               []*EventMsg `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty" query:"-"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-" query:"-"`
	XXX_unrecognized     []byte      `json:"-" query:"-"`
	XXX_sizecache        int32       `json:"-" query:"-"`
}

func (m *EventListMsg) Reset()         { *m = EventListMsg{} }
func (m *EventListMsg) String() string { return proto.CompactTextString(m) }
func (*EventListMsg) ProtoMessage()    {}
func (*EventListMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{6}
}

func (m *EventListMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EventListMsg.Unmarshal(m, b)
}
func (m *EventListMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EventListMsg.Marshal(b, m, deterministic)
}
func (m *EventListMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventListMsg.Merge(m, src)
}
func (m *EventListMsg) XXX_Size() int {
	return xxx_messageInfo_EventListMsg.Size(m)
}
func (m *EventListMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_EventListMsg.DiscardUnknown(m)
}

var xxx_messageInfo_EventListMsg proto.InternalMessageInfo

func (m *EventListMsg) GetEvents() []*EventMsg {
	if m != nil {
		return m.Events
	}
	return nil
}

type FindNodeEventsMsg struct {
	SpaceId              int32    `protobuf:"varint,1,opt,name=space_id,json=spaceId,proto3" json:"space_id" query:"space_id"`
	EventId              int32    `protobuf:"varint,2,opt,name=event_id,json=eventId,proto3" json:"event_id" query:"event_id"`
	AtTime               int32    `protobuf:"varint,3,opt,name=at_time,json=atTime,proto3" json:"at_time" query:"at_time"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" query:"-"`
	XXX_unrecognized     []byte   `json:"-" query:"-"`
	XXX_sizecache        int32    `json:"-" query:"-"`
}

func (m *FindNodeEventsMsg) Reset()         { *m = FindNodeEventsMsg{} }
func (m *FindNodeEventsMsg) String() string { return proto.CompactTextString(m) }
func (*FindNodeEventsMsg) ProtoMessage()    {}
func (*FindNodeEventsMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{7}
}

func (m *FindNodeEventsMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FindNodeEventsMsg.Unmarshal(m, b)
}
func (m *FindNodeEventsMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FindNodeEventsMsg.Marshal(b, m, deterministic)
}
func (m *FindNodeEventsMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FindNodeEventsMsg.Merge(m, src)
}
func (m *FindNodeEventsMsg) XXX_Size() int {
	return xxx_messageInfo_FindNodeEventsMsg.Size(m)
}
func (m *FindNodeEventsMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_FindNodeEventsMsg.DiscardUnknown(m)
}

var xxx_messageInfo_FindNodeEventsMsg proto.InternalMessageInfo

func (m *FindNodeEventsMsg) GetSpaceId() int32 {
	if m != nil {
		return m.SpaceId
	}
	return 0
}

func (m *FindNodeEventsMsg) GetEventId() int32 {
	if m != nil {
		return m.EventId
	}
	return 0
}

func (m *FindNodeEventsMsg) GetAtTime() int32 {
	if m != nil {
		return m.AtTime
	}
	return 0
}

type NodeEventListMsg struct {
	Nodes                []*NodeEventMsg `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty" query:"-"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-" query:"-"`
	XXX_unrecognized     []byte          `json:"-" query:"-"`
	XXX_sizecache        int32           `json:"-" query:"-"`
}

func (m *NodeEventListMsg) Reset()         { *m = NodeEventListMsg{} }
func (m *NodeEventListMsg) String() string { return proto.CompactTextString(m) }
func (*NodeEventListMsg) ProtoMessage()    {}
func (*NodeEventListMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{8}
}

func (m *NodeEventListMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeEventListMsg.Unmarshal(m, b)
}
func (m *NodeEventListMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeEventListMsg.Marshal(b, m, deterministic)
}
func (m *NodeEventListMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeEventListMsg.Merge(m, src)
}
func (m *NodeEventListMsg) XXX_Size() int {
	return xxx_messageInfo_NodeEventListMsg.Size(m)
}
func (m *NodeEventListMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeEventListMsg.DiscardUnknown(m)
}

var xxx_messageInfo_NodeEventListMsg proto.InternalMessageInfo

func (m *NodeEventListMsg) GetNodes() []*NodeEventMsg {
	if m != nil {
		return m.Nodes
	}
	return nil
}

type SpaceTimeRequestMsg struct {
	SpaceId              int32    `protobuf:"varint,1,opt,name=space_id,json=spaceId,proto3" json:"space_id" query:"space_id"`
	CurrentTime          int32    `protobuf:"varint,2,opt,name=current_time,json=currentTime,proto3" json:"current_time" query:"current_time"`
	MinNbEventsFilter    int32    `protobuf:"varint,3,opt,name=min_nb_events_filter,json=minNbEventsFilter,proto3" json:"min_nb_events_filter" query:"min_nb_events_filter"`
	ColorMaskFilter      uint32   `protobuf:"varint,4,opt,name=color_mask_filter,json=colorMaskFilter,proto3" json:"color_mask_filter" query:"color_mask_filter"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" query:"-"`
	XXX_unrecognized     []byte   `json:"-" query:"-"`
	XXX_sizecache        int32    `json:"-" query:"-"`
}

func (m *SpaceTimeRequestMsg) Reset()         { *m = SpaceTimeRequestMsg{} }
func (m *SpaceTimeRequestMsg) String() string { return proto.CompactTextString(m) }
func (*SpaceTimeRequestMsg) ProtoMessage()    {}
func (*SpaceTimeRequestMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{9}
}

func (m *SpaceTimeRequestMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SpaceTimeRequestMsg.Unmarshal(m, b)
}
func (m *SpaceTimeRequestMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SpaceTimeRequestMsg.Marshal(b, m, deterministic)
}
func (m *SpaceTimeRequestMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpaceTimeRequestMsg.Merge(m, src)
}
func (m *SpaceTimeRequestMsg) XXX_Size() int {
	return xxx_messageInfo_SpaceTimeRequestMsg.Size(m)
}
func (m *SpaceTimeRequestMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SpaceTimeRequestMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SpaceTimeRequestMsg proto.InternalMessageInfo

func (m *SpaceTimeRequestMsg) GetSpaceId() int32 {
	if m != nil {
		return m.SpaceId
	}
	return 0
}

func (m *SpaceTimeRequestMsg) GetCurrentTime() int32 {
	if m != nil {
		return m.CurrentTime
	}
	return 0
}

func (m *SpaceTimeRequestMsg) GetMinNbEventsFilter() int32 {
	if m != nil {
		return m.MinNbEventsFilter
	}
	return 0
}

func (m *SpaceTimeRequestMsg) GetColorMaskFilter() uint32 {
	if m != nil {
		return m.ColorMaskFilter
	}
	return 0
}

type SpaceTimeResponseMsg struct {
	SpaceId              int32               `protobuf:"varint,1,opt,name=space_id,json=spaceId,proto3" json:"space_id" query:"space_id"`
	CurrentTime          int32               `protobuf:"varint,2,opt,name=current_time,json=currentTime,proto3" json:"current_time" query:"current_time"`
	ActiveEvents         []*EventMsg         `protobuf:"bytes,3,rep,name=active_events,json=activeEvents,proto3" json:"active_events,omitempty" query:"-"`
	NbActiveNodes        int32               `protobuf:"varint,4,opt,name=nb_active_nodes,json=nbActiveNodes,proto3" json:"nb_active_nodes" query:"nb_active_nodes"`
	FilteredNodes        []*SpaceTimeNodeMsg `protobuf:"bytes,5,rep,name=filtered_nodes,json=filteredNodes,proto3" json:"filtered_nodes,omitempty" query:"-"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-" query:"-"`
	XXX_unrecognized     []byte              `json:"-" query:"-"`
	XXX_sizecache        int32               `json:"-" query:"-"`
}

func (m *SpaceTimeResponseMsg) Reset()         { *m = SpaceTimeResponseMsg{} }
func (m *SpaceTimeResponseMsg) String() string { return proto.CompactTextString(m) }
func (*SpaceTimeResponseMsg) ProtoMessage()    {}
func (*SpaceTimeResponseMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{10}
}

func (m *SpaceTimeResponseMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SpaceTimeResponseMsg.Unmarshal(m, b)
}
func (m *SpaceTimeResponseMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SpaceTimeResponseMsg.Marshal(b, m, deterministic)
}
func (m *SpaceTimeResponseMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpaceTimeResponseMsg.Merge(m, src)
}
func (m *SpaceTimeResponseMsg) XXX_Size() int {
	return xxx_messageInfo_SpaceTimeResponseMsg.Size(m)
}
func (m *SpaceTimeResponseMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SpaceTimeResponseMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SpaceTimeResponseMsg proto.InternalMessageInfo

func (m *SpaceTimeResponseMsg) GetSpaceId() int32 {
	if m != nil {
		return m.SpaceId
	}
	return 0
}

func (m *SpaceTimeResponseMsg) GetCurrentTime() int32 {
	if m != nil {
		return m.CurrentTime
	}
	return 0
}

func (m *SpaceTimeResponseMsg) GetActiveEvents() []*EventMsg {
	if m != nil {
		return m.ActiveEvents
	}
	return nil
}

func (m *SpaceTimeResponseMsg) GetNbActiveNodes() int32 {
	if m != nil {
		return m.NbActiveNodes
	}
	return 0
}

func (m *SpaceTimeResponseMsg) GetFilteredNodes() []*SpaceTimeNodeMsg {
	if m != nil {
		return m.FilteredNodes
	}
	return nil
}

type SpaceTimeNodeMsg struct {
	PointId              int64                    `protobuf:"varint,1,opt,name=point_id,json=pointId,proto3" json:"point_id" query:"point_id"`
	Point                *PointMsg                `protobuf:"bytes,2,opt,name=point,proto3" json:"point,omitempty" query:"point"`
	Nodes                []*SpaceTimeNodeEventMsg `protobuf:"bytes,3,rep,name=nodes,proto3" json:"nodes,omitempty" query:"-"`
	HasRoot              bool                     `protobuf:"varint,4,opt,name=has_root,json=hasRoot,proto3" json:"has_root,omitempty" query:"-"`
	ColorMask            uint32                   `protobuf:"varint,5,opt,name=color_mask,json=colorMask,proto3" json:"color_mask" query:"color_mask"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-" query:"-"`
	XXX_unrecognized     []byte                   `json:"-" query:"-"`
	XXX_sizecache        int32                    `json:"-" query:"-"`
}

func (m *SpaceTimeNodeMsg) Reset()         { *m = SpaceTimeNodeMsg{} }
func (m *SpaceTimeNodeMsg) String() string { return proto.CompactTextString(m) }
func (*SpaceTimeNodeMsg) ProtoMessage()    {}
func (*SpaceTimeNodeMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{11}
}

func (m *SpaceTimeNodeMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SpaceTimeNodeMsg.Unmarshal(m, b)
}
func (m *SpaceTimeNodeMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SpaceTimeNodeMsg.Marshal(b, m, deterministic)
}
func (m *SpaceTimeNodeMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpaceTimeNodeMsg.Merge(m, src)
}
func (m *SpaceTimeNodeMsg) XXX_Size() int {
	return xxx_messageInfo_SpaceTimeNodeMsg.Size(m)
}
func (m *SpaceTimeNodeMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SpaceTimeNodeMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SpaceTimeNodeMsg proto.InternalMessageInfo

func (m *SpaceTimeNodeMsg) GetPointId() int64 {
	if m != nil {
		return m.PointId
	}
	return 0
}

func (m *SpaceTimeNodeMsg) GetPoint() *PointMsg {
	if m != nil {
		return m.Point
	}
	return nil
}

func (m *SpaceTimeNodeMsg) GetNodes() []*SpaceTimeNodeEventMsg {
	if m != nil {
		return m.Nodes
	}
	return nil
}

func (m *SpaceTimeNodeMsg) GetHasRoot() bool {
	if m != nil {
		return m.HasRoot
	}
	return false
}

func (m *SpaceTimeNodeMsg) GetColorMask() uint32 {
	if m != nil {
		return m.ColorMask
	}
	return 0
}

type SpaceTimeNodeEventMsg struct {
	EventId              int32    `protobuf:"varint,2,opt,name=event_id,json=eventId,proto3" json:"event_id" query:"event_id"`
	CreationTime         int32    `protobuf:"varint,4,opt,name=creation_time,json=creationTime,proto3" json:"creation_time" query:"creation_time"`
	D                    int32    `protobuf:"varint,5,opt,name=d,proto3" json:"d" query:"d"`
	TrioId               int32    `protobuf:"varint,6,opt,name=trio_id,json=trioId,proto3" json:"trio_id" query:"trio_id"`
	ConnectionMask       uint32   `protobuf:"varint,7,opt,name=connection_mask,json=connectionMask,proto3" json:"connection_mask" query:"connection_mask"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" query:"-"`
	XXX_unrecognized     []byte   `json:"-" query:"-"`
	XXX_sizecache        int32    `json:"-" query:"-"`
}

func (m *SpaceTimeNodeEventMsg) Reset()         { *m = SpaceTimeNodeEventMsg{} }
func (m *SpaceTimeNodeEventMsg) String() string { return proto.CompactTextString(m) }
func (*SpaceTimeNodeEventMsg) ProtoMessage()    {}
func (*SpaceTimeNodeEventMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_c43524b64f4ebcab, []int{12}
}

func (m *SpaceTimeNodeEventMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SpaceTimeNodeEventMsg.Unmarshal(m, b)
}
func (m *SpaceTimeNodeEventMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SpaceTimeNodeEventMsg.Marshal(b, m, deterministic)
}
func (m *SpaceTimeNodeEventMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpaceTimeNodeEventMsg.Merge(m, src)
}
func (m *SpaceTimeNodeEventMsg) XXX_Size() int {
	return xxx_messageInfo_SpaceTimeNodeEventMsg.Size(m)
}
func (m *SpaceTimeNodeEventMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_SpaceTimeNodeEventMsg.DiscardUnknown(m)
}

var xxx_messageInfo_SpaceTimeNodeEventMsg proto.InternalMessageInfo

func (m *SpaceTimeNodeEventMsg) GetEventId() int32 {
	if m != nil {
		return m.EventId
	}
	return 0
}

func (m *SpaceTimeNodeEventMsg) GetCreationTime() int32 {
	if m != nil {
		return m.CreationTime
	}
	return 0
}

func (m *SpaceTimeNodeEventMsg) GetD() int32 {
	if m != nil {
		return m.D
	}
	return 0
}

func (m *SpaceTimeNodeEventMsg) GetTrioId() int32 {
	if m != nil {
		return m.TrioId
	}
	return 0
}

func (m *SpaceTimeNodeEventMsg) GetConnectionMask() uint32 {
	if m != nil {
		return m.ConnectionMask
	}
	return 0
}

func init() {
	proto.RegisterType((*SpaceMsg)(nil), "m3api.SpaceMsg")
	proto.RegisterType((*SpaceListMsg)(nil), "m3api.SpaceListMsg")
	proto.RegisterType((*CreateEventRequestMsg)(nil), "m3api.CreateEventRequestMsg")
	proto.RegisterType((*NodeEventMsg)(nil), "m3api.NodeEventMsg")
	proto.RegisterType((*FindEventsMsg)(nil), "m3api.FindEventsMsg")
	proto.RegisterType((*EventMsg)(nil), "m3api.EventMsg")
	proto.RegisterType((*EventListMsg)(nil), "m3api.EventListMsg")
	proto.RegisterType((*FindNodeEventsMsg)(nil), "m3api.FindNodeEventsMsg")
	proto.RegisterType((*NodeEventListMsg)(nil), "m3api.NodeEventListMsg")
	proto.RegisterType((*SpaceTimeRequestMsg)(nil), "m3api.SpaceTimeRequestMsg")
	proto.RegisterType((*SpaceTimeResponseMsg)(nil), "m3api.SpaceTimeResponseMsg")
	proto.RegisterType((*SpaceTimeNodeMsg)(nil), "m3api.SpaceTimeNodeMsg")
	proto.RegisterType((*SpaceTimeNodeEventMsg)(nil), "m3api.SpaceTimeNodeEventMsg")
}

func init() {
	proto.RegisterFile("m3space.proto", fileDescriptor_c43524b64f4ebcab)
}

var fileDescriptor_c43524b64f4ebcab = []byte{
	// 903 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x56, 0xdd, 0x6e, 0xdc, 0x44,
	0x14, 0x96, 0x77, 0xb3, 0x7f, 0x67, 0xed, 0x6e, 0xe2, 0xa4, 0x8a, 0xcb, 0xef, 0xd6, 0x08, 0x92,
	0x22, 0x11, 0x50, 0x83, 0xd4, 0x2b, 0x90, 0x50, 0x44, 0xa5, 0x95, 0x68, 0xa8, 0x4c, 0xae, 0xb1,
	0xbc, 0xf6, 0xa4, 0x3b, 0x4a, 0x3c, 0x63, 0x3c, 0xd3, 0xb2, 0x7d, 0x1d, 0x6e, 0xb8, 0x45, 0xe2,
	0x0d, 0x78, 0x0c, 0x1e, 0x83, 0x17, 0x00, 0x9d, 0x73, 0xc6, 0x8e, 0x1b, 0xb6, 0x8b, 0x04, 0x17,
	0xf4, 0x72, 0xbe, 0xf3, 0xcd, 0xf1, 0xf9, 0xbe, 0x73, 0xe6, 0xec, 0x42, 0x50, 0x9e, 0x9a, 0x2a,
	0xcb, 0xc5, 0x49, 0x55, 0x6b, 0xab, 0xc3, 0x41, 0x79, 0x9a, 0x55, 0xf2, 0xad, 0xa0, 0x3c, 0xad,
	0xb4, 0x54, 0x96, 0xd1, 0xf8, 0xa7, 0x1e, 0x8c, 0xbf, 0x43, 0xd6, 0x13, 0xf3, 0x2c, 0xbc, 0x07,
	0x63, 0xba, 0x91, 0xca, 0x22, 0xf2, 0xe6, 0xde, 0xf1, 0x20, 0x19, 0xd1, 0x79, 0x51, 0x84, 0xef,
	0x02, 0x70, 0x48, 0x65, 0xa5, 0x88, 0x7a, 0x73, 0xef, 0x78, 0x92, 0x4c, 0x08, 0x39, 0xcf, 0x4a,
	0x11, 0x3e, 0x80, 0xdd, 0x2c, 0xb7, 0xf2, 0x85, 0x48, 0xed, 0xaa, 0x16, 0x66, 0xa5, 0xaf, 0x8b,
	0xa8, 0x4f, 0x19, 0x66, 0x8c, 0x5f, 0x34, 0x70, 0xf8, 0x09, 0xec, 0x97, 0xd9, 0x3a, 0xb5, 0xb5,
	0xd4, 0x26, 0xad, 0x44, 0x9d, 0x52, 0x39, 0xd1, 0x0e, 0xb1, 0x77, 0xcb, 0x6c, 0x7d, 0x81, 0x91,
	0xa7, 0xa2, 0x7e, 0x8a, 0x78, 0x43, 0x57, 0xba, 0x10, 0x5d, 0xfa, 0xa0, 0xa5, 0x9f, 0x63, 0xa4,
	0xa5, 0xdf, 0x83, 0x31, 0x65, 0x97, 0xa5, 0x88, 0x86, 0x2c, 0x01, 0x53, 0xca, 0x52, 0x84, 0x6f,
	0xc3, 0x04, 0x43, 0xb9, 0xd6, 0x75, 0x11, 0x8d, 0x29, 0x86, 0xdc, 0x33, 0x3c, 0x63, 0x50, 0xbc,
	0x10, 0xca, 0xa6, 0xb2, 0x30, 0xd1, 0x64, 0xde, 0xc7, 0x20, 0x01, 0x8b, 0xc2, 0xc4, 0x8f, 0xc0,
	0x27, 0x8f, 0xbe, 0x91, 0xc6, 0xa2, 0x4f, 0x47, 0x30, 0x24, 0xe9, 0x26, 0xf2, 0xe6, 0xfd, 0xe3,
	0xe9, 0xc3, 0xd9, 0x09, 0x79, 0x7b, 0xd2, 0x18, 0x99, 0xb8, 0x70, 0xfc, 0xa7, 0x07, 0x77, 0xcf,
	0x6a, 0x91, 0x59, 0xf1, 0x35, 0xe6, 0x4a, 0xc4, 0x0f, 0xcf, 0x05, 0xa7, 0xe8, 0x5a, 0xdd, 0x7b,
	0xd5, 0xea, 0xf7, 0x61, 0xfa, 0xac, 0xd6, 0x3f, 0xda, 0x55, 0x6a, 0x5f, 0x56, 0xc2, 0xd9, 0x08,
	0x0c, 0x5d, 0xbc, 0xac, 0x44, 0x78, 0x1f, 0x7c, 0x47, 0x90, 0xaa, 0x10, 0x6b, 0x67, 0x9d, 0xbb,
	0xb4, 0x40, 0x28, 0xfc, 0x00, 0x02, 0x47, 0xd1, 0x97, 0x97, 0x46, 0x34, 0x7e, 0xb9, 0x7b, 0xdf,
	0x12, 0x86, 0xa4, 0x1c, 0x8b, 0x93, 0x5a, 0x75, 0x0d, 0xf3, 0x1b, 0x90, 0x5c, 0x3b, 0x82, 0x61,
	0x2e, 0x94, 0x15, 0x75, 0x34, 0x9a, 0x7b, 0x1d, 0xad, 0x64, 0x37, 0x69, 0xe5, 0x70, 0x78, 0x00,
	0x83, 0x5c, 0x5f, 0xeb, 0x9a, 0xac, 0x0d, 0x12, 0x3e, 0xc4, 0xbf, 0xf6, 0xc0, 0xc7, 0x0e, 0x91,
	0x7e, 0x14, 0x1e, 0x43, 0xc0, 0x46, 0x63, 0x47, 0x9b, 0x41, 0xeb, 0x27, 0x53, 0x02, 0x91, 0xb9,
	0x28, 0xd0, 0x9c, 0xa6, 0x19, 0x8d, 0x39, 0xae, 0x17, 0xe1, 0x87, 0x30, 0xe0, 0x01, 0xe8, 0x6f,
	0xae, 0x86, 0xa3, 0x7f, 0x97, 0xb6, 0xb3, 0x41, 0x9a, 0x0f, 0x5e, 0xe1, 0x8c, 0xf1, 0x8a, 0xf0,
	0x10, 0x46, 0x38, 0x93, 0xf8, 0x4d, 0xf6, 0x61, 0x88, 0xc7, 0x45, 0x11, 0x1e, 0xc1, 0x2c, 0xd7,
	0x4a, 0x89, 0x9c, 0xb2, 0x95, 0x99, 0xb9, 0x22, 0x2b, 0x82, 0xe4, 0xce, 0x0d, 0xfc, 0x24, 0x33,
	0x57, 0xe1, 0x1c, 0xfc, 0x2a, 0xb3, 0xab, 0x56, 0xd9, 0x98, 0x94, 0x01, 0x62, 0x4e, 0xd8, 0x47,
	0x30, 0xbb, 0x96, 0xea, 0x4a, 0x14, 0x0d, 0x87, 0x67, 0xad, 0x9f, 0x04, 0x0c, 0x33, 0xcd, 0xc4,
	0xdf, 0x43, 0xf0, 0x58, 0xaa, 0x82, 0x4c, 0x33, 0x6e, 0x5c, 0x5a, 0x47, 0xbc, 0x57, 0x1d, 0xd9,
	0x32, 0x49, 0x87, 0x30, 0xca, 0x2c, 0xeb, 0xe7, 0x29, 0x1a, 0x66, 0x16, 0x95, 0xc7, 0xbf, 0xf7,
	0x60, 0xdc, 0x76, 0xe4, 0xdf, 0xe5, 0x7e, 0xb3, 0xa6, 0xf4, 0x3d, 0x98, 0x92, 0xf5, 0xb9, 0x5d,
	0x63, 0xad, 0x23, 0xa2, 0x4c, 0x10, 0x3a, 0xb3, 0xeb, 0x45, 0xb1, 0x79, 0x38, 0xc3, 0xcf, 0x60,
	0x52, 0x6b, 0xcd, 0xa3, 0x18, 0x4d, 0x68, 0xa0, 0xf6, 0xdd, 0x40, 0x75, 0x67, 0x36, 0x19, 0x23,
	0x0b, 0x11, 0x9c, 0xde, 0x66, 0x1b, 0x71, 0x31, 0xc0, 0xaa, 0xdc, 0x1e, 0x22, 0x73, 0x1f, 0x81,
	0x4f, 0x37, 0x3b, 0xdb, 0x82, 0xfc, 0xbc, 0xbd, 0x2d, 0xda, 0xf4, 0x2e, 0x1c, 0x2f, 0x61, 0x0f,
	0xbb, 0xde, 0x7e, 0xda, 0xfc, 0xc3, 0x4e, 0xde, 0xf2, 0x4c, 0x5e, 0xdb, 0xf9, 0x2f, 0x60, 0xb7,
	0xcd, 0xdf, 0x14, 0xf8, 0x00, 0x06, 0xb4, 0x5e, 0x5d, 0x7d, 0x1b, 0x2d, 0x60, 0x46, 0xfc, 0x8b,
	0x07, 0xfb, 0xb4, 0xe5, 0x30, 0xd9, 0x6b, 0xd6, 0xd9, 0xad, 0x2a, 0xef, 0x83, 0x9f, 0x3f, 0xaf,
	0x6b, 0xac, 0x93, 0xea, 0xe1, 0x4a, 0xa7, 0x0e, 0xa3, 0xee, 0x7d, 0x0a, 0x07, 0xa5, 0x54, 0xa9,
	0x5a, 0xa6, 0xec, 0x44, 0x7a, 0x29, 0xaf, 0x71, 0xe3, 0x70, 0xe9, 0x7b, 0xa5, 0x54, 0xe7, 0x4b,
	0x76, 0xe4, 0x31, 0x05, 0xc2, 0x8f, 0x61, 0x8f, 0x3a, 0x48, 0xaf, 0xb1, 0x61, 0xef, 0x50, 0x6b,
	0x67, 0x14, 0xc0, 0xf7, 0xc8, 0xdc, 0xf8, 0x0f, 0x0f, 0x0e, 0x3a, 0x25, 0x9b, 0x4a, 0x2b, 0x23,
	0xfe, 0x7b, 0xcd, 0x9f, 0x43, 0xe0, 0x7e, 0xf1, 0x5c, 0x73, 0xfb, 0x9b, 0x9b, 0xeb, 0x33, 0x8b,
	0xcb, 0xc7, 0x05, 0xa0, 0x96, 0xa9, 0xbb, 0xc8, 0xa6, 0xf3, 0xbb, 0x08, 0xd4, 0xf2, 0x2b, 0x42,
	0xe9, 0xe7, 0x2c, 0xfc, 0x12, 0xee, 0xb0, 0x2a, 0xb7, 0x2a, 0x4c, 0x34, 0xa0, 0xf4, 0x87, 0xdd,
	0x5f, 0x1a, 0xac, 0x03, 0xe9, 0xf8, 0x99, 0xa0, 0xa1, 0xd3, 0xfd, 0xf8, 0x37, 0x0f, 0x76, 0x6f,
	0x73, 0x50, 0x30, 0x6d, 0xc7, 0x9b, 0xad, 0x3b, 0xa2, 0x73, 0x77, 0xad, 0xf6, 0xb6, 0xae, 0xd5,
	0x87, 0xcd, 0xa4, 0xb0, 0xd8, 0x77, 0x36, 0x55, 0x73, 0x6b, 0x64, 0xf0, 0xab, 0xab, 0xcc, 0xa4,
	0xf8, 0x84, 0x48, 0xeb, 0x38, 0x19, 0xad, 0x32, 0x93, 0x68, 0x6d, 0xf1, 0x4f, 0xc5, 0x4d, 0x1b,
	0xe9, 0xf1, 0x07, 0xc9, 0xa4, 0xed, 0x5f, 0xfc, 0xb3, 0x07, 0x77, 0x37, 0xa6, 0xde, 0x36, 0xf9,
	0xff, 0xc7, 0xe6, 0x5f, 0x0e, 0xe9, 0xcf, 0xd4, 0xe9, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x1c,
	0xc1, 0x1f, 0x56, 0x73, 0x09, 0x00, 0x00,
}
