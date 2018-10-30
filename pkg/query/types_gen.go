package query

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Summary) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "BlockHeight":
			z.BlockHeight, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "TotalNdau":
			err = z.TotalNdau.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "NumAccounts":
			z.NumAccounts, err = dc.ReadInt()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Summary) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "BlockHeight"
	err = en.Append(0x83, 0xab, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.BlockHeight)
	if err != nil {
		return
	}
	// write "TotalNdau"
	err = en.Append(0xa9, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x4e, 0x64, 0x61, 0x75)
	if err != nil {
		return
	}
	err = z.TotalNdau.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "NumAccounts"
	err = en.Append(0xab, 0x4e, 0x75, 0x6d, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x73)
	if err != nil {
		return
	}
	err = en.WriteInt(z.NumAccounts)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Summary) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "BlockHeight"
	o = append(o, 0x83, 0xab, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74)
	o = msgp.AppendUint64(o, z.BlockHeight)
	// string "TotalNdau"
	o = append(o, 0xa9, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x4e, 0x64, 0x61, 0x75)
	o, err = z.TotalNdau.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "NumAccounts"
	o = append(o, 0xab, 0x4e, 0x75, 0x6d, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x73)
	o = msgp.AppendInt(o, z.NumAccounts)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Summary) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "BlockHeight":
			z.BlockHeight, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "TotalNdau":
			bts, err = z.TotalNdau.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "NumAccounts":
			z.NumAccounts, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Summary) Msgsize() (s int) {
	s = 1 + 12 + msgp.Uint64Size + 10 + z.TotalNdau.Msgsize() + 12 + msgp.IntSize
	return
}
