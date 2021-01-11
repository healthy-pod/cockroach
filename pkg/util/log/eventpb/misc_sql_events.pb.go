// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: util/log/eventpb/misc_sql_events.proto

package eventpb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// SetClusterSetting is recorded when a cluster setting is changed.
type SetClusterSetting struct {
	CommonEventDetails    `protobuf:"bytes,1,opt,name=common,proto3,embedded=common" json:""`
	CommonSQLEventDetails `protobuf:"bytes,2,opt,name=sql,proto3,embedded=sql" json:""`
	// The name of the affected cluster setting.
	SettingName string `protobuf:"bytes,3,opt,name=setting_name,json=settingName,proto3" json:",omitempty" redact:"nonsensitive"`
	// The new value of the cluster setting.
	Value string `protobuf:"bytes,4,opt,name=value,proto3" json:",omitempty"`
}

func (m *SetClusterSetting) Reset()         { *m = SetClusterSetting{} }
func (m *SetClusterSetting) String() string { return proto.CompactTextString(m) }
func (*SetClusterSetting) ProtoMessage()    {}
func (*SetClusterSetting) Descriptor() ([]byte, []int) {
	return fileDescriptor_misc_sql_events_a7ef825afe3b6935, []int{0}
}
func (m *SetClusterSetting) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SetClusterSetting) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *SetClusterSetting) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetClusterSetting.Merge(dst, src)
}
func (m *SetClusterSetting) XXX_Size() int {
	return m.Size()
}
func (m *SetClusterSetting) XXX_DiscardUnknown() {
	xxx_messageInfo_SetClusterSetting.DiscardUnknown(m)
}

var xxx_messageInfo_SetClusterSetting proto.InternalMessageInfo

func init() {
	proto.RegisterType((*SetClusterSetting)(nil), "cockroach.util.log.eventpb.SetClusterSetting")
}
func (m *SetClusterSetting) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SetClusterSetting) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintMiscSqlEvents(dAtA, i, uint64(m.CommonEventDetails.Size()))
	n1, err := m.CommonEventDetails.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	dAtA[i] = 0x12
	i++
	i = encodeVarintMiscSqlEvents(dAtA, i, uint64(m.CommonSQLEventDetails.Size()))
	n2, err := m.CommonSQLEventDetails.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	if len(m.SettingName) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintMiscSqlEvents(dAtA, i, uint64(len(m.SettingName)))
		i += copy(dAtA[i:], m.SettingName)
	}
	if len(m.Value) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintMiscSqlEvents(dAtA, i, uint64(len(m.Value)))
		i += copy(dAtA[i:], m.Value)
	}
	return i, nil
}

func encodeVarintMiscSqlEvents(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *SetClusterSetting) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.CommonEventDetails.Size()
	n += 1 + l + sovMiscSqlEvents(uint64(l))
	l = m.CommonSQLEventDetails.Size()
	n += 1 + l + sovMiscSqlEvents(uint64(l))
	l = len(m.SettingName)
	if l > 0 {
		n += 1 + l + sovMiscSqlEvents(uint64(l))
	}
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovMiscSqlEvents(uint64(l))
	}
	return n
}

func sovMiscSqlEvents(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMiscSqlEvents(x uint64) (n int) {
	return sovMiscSqlEvents(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SetClusterSetting) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMiscSqlEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SetClusterSetting: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SetClusterSetting: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommonEventDetails", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMiscSqlEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMiscSqlEvents
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.CommonEventDetails.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommonSQLEventDetails", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMiscSqlEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMiscSqlEvents
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.CommonSQLEventDetails.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SettingName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMiscSqlEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMiscSqlEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SettingName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMiscSqlEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMiscSqlEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMiscSqlEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMiscSqlEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMiscSqlEvents(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMiscSqlEvents
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMiscSqlEvents
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMiscSqlEvents
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthMiscSqlEvents
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMiscSqlEvents
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipMiscSqlEvents(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthMiscSqlEvents = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMiscSqlEvents   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("util/log/eventpb/misc_sql_events.proto", fileDescriptor_misc_sql_events_a7ef825afe3b6935)
}

var fileDescriptor_misc_sql_events_a7ef825afe3b6935 = []byte{
	// 327 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0xd0, 0x4d, 0x4b, 0x3a, 0x41,
	0x1c, 0x07, 0xf0, 0x1d, 0xfd, 0xff, 0x8d, 0x46, 0x09, 0x5a, 0x0a, 0x16, 0xa1, 0x59, 0x91, 0x28,
	0x83, 0x98, 0xa5, 0xba, 0x75, 0xd4, 0xba, 0x44, 0x04, 0x6a, 0xa7, 0x2e, 0xb2, 0x6e, 0x3f, 0xb6,
	0xa1, 0x79, 0xd0, 0x9d, 0x51, 0xe8, 0x5d, 0xf4, 0x4a, 0x7a, 0x1d, 0x1e, 0x3d, 0x7a, 0x5a, 0x6a,
	0xbd, 0x79, 0xec, 0x15, 0xc4, 0x8e, 0x7b, 0x28, 0xa2, 0xba, 0xcd, 0xc3, 0xf7, 0xfb, 0x99, 0x07,
	0x7c, 0x30, 0x31, 0x8c, 0x07, 0x5c, 0xc5, 0x01, 0x4c, 0x41, 0x9a, 0xd1, 0x30, 0x10, 0x4c, 0x47,
	0x03, 0x3d, 0xe6, 0x03, 0xbb, 0xa0, 0xe9, 0x28, 0x51, 0x46, 0xb9, 0xf5, 0x48, 0x45, 0x8f, 0x89,
	0x0a, 0xa3, 0x07, 0x9a, 0x37, 0x28, 0x57, 0x31, 0x2d, 0x1a, 0xf5, 0x9d, 0x58, 0xc5, 0xca, 0xc6,
	0x82, 0x7c, 0xb4, 0x6e, 0xd4, 0xf7, 0xbe, 0xc9, 0x9f, 0xc1, 0xe6, 0x4b, 0x09, 0x6f, 0xf7, 0xc1,
	0x74, 0xf8, 0x44, 0x1b, 0x48, 0xfa, 0x60, 0x0c, 0x93, 0xb1, 0x7b, 0x8b, 0x2b, 0x91, 0x12, 0x42,
	0x49, 0x0f, 0x35, 0x50, 0xab, 0x7a, 0x4a, 0xe9, 0xcf, 0xe7, 0xd2, 0x8e, 0x4d, 0x5e, 0xe6, 0xb3,
	0x0b, 0x30, 0x21, 0xe3, 0xba, 0x5d, 0x9b, 0xa5, 0xbe, 0x33, 0x4f, 0x7d, 0xb4, 0x4a, 0x7d, 0xa7,
	0x57, 0x58, 0x6e, 0x17, 0x97, 0xf5, 0x98, 0x7b, 0x25, 0x4b, 0x9e, 0xfc, 0x4d, 0xf6, 0xbb, 0xd7,
	0xbf, 0xa8, 0xb9, 0xe5, 0x5e, 0xe1, 0x9a, 0x5e, 0xdf, 0x79, 0x20, 0x43, 0x01, 0x5e, 0xb9, 0x81,
	0x5a, 0x9b, 0xed, 0xc3, 0x55, 0xea, 0xe3, 0x63, 0x25, 0x98, 0x01, 0x31, 0x32, 0x4f, 0xef, 0xa9,
	0xbf, 0x9b, 0xc0, 0x7d, 0x18, 0x99, 0xf3, 0xa6, 0x54, 0x52, 0x83, 0xd4, 0xcc, 0xb0, 0x29, 0x34,
	0x7b, 0xd5, 0xa2, 0x7c, 0x13, 0x0a, 0x70, 0xf7, 0xf1, 0xff, 0x69, 0xc8, 0x27, 0xe0, 0xfd, 0xb3,
	0xc8, 0xd6, 0x57, 0xa4, 0xb7, 0xde, 0x6c, 0x1f, 0xcd, 0xde, 0x88, 0x33, 0xcb, 0x08, 0x9a, 0x67,
	0x04, 0x2d, 0x32, 0x82, 0x5e, 0x33, 0x82, 0x9e, 0x97, 0xc4, 0x99, 0x2f, 0x89, 0xb3, 0x58, 0x12,
	0xe7, 0x6e, 0xa3, 0x78, 0xc5, 0xb0, 0x62, 0xbf, 0xf8, 0xec, 0x23, 0x00, 0x00, 0xff, 0xff, 0xf1,
	0xba, 0x8c, 0xd6, 0xdd, 0x01, 0x00, 0x00,
}
