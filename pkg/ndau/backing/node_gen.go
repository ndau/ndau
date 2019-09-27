package backing

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// MarshalMsg implements msgp.Marshaler
func (z *Node) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Active"
	o = append(o, 0x84, 0xa6, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65)
	o = msgp.AppendBool(o, z.Active)
	// string "DistributionScript"
	o = append(o, 0xb2, 0x44, 0x69, 0x73, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x63, 0x72, 0x69, 0x70, 0x74)
	o = msgp.AppendBytes(o, z.DistributionScript)
	// string "TMAddress"
	o = append(o, 0xa9, 0x54, 0x4d, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	o = msgp.AppendString(o, z.TMAddress)
	// string "Key"
	o = append(o, 0xa3, 0x4b, 0x65, 0x79)
	o, err = z.Key.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "Key")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Node) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Active":
			z.Active, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Active")
				return
			}
		case "DistributionScript":
			z.DistributionScript, bts, err = msgp.ReadBytesBytes(bts, z.DistributionScript)
			if err != nil {
				err = msgp.WrapError(err, "DistributionScript")
				return
			}
		case "TMAddress":
			z.TMAddress, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "TMAddress")
				return
			}
		case "Key":
			bts, err = z.Key.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "Key")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Node) Msgsize() (s int) {
	s = 1 + 7 + msgp.BoolSize + 19 + msgp.BytesPrefixSize + len(z.DistributionScript) + 10 + msgp.StringPrefixSize + len(z.TMAddress) + 4 + z.Key.Msgsize()
	return
}