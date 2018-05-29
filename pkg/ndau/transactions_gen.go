package ndau

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *GTValidatorChange) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "PublicKey":
			z.PublicKey, err = dc.ReadBytes(z.PublicKey)
			if err != nil {
				return
			}
		case "Power":
			z.Power, err = dc.ReadInt64()
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
func (z *GTValidatorChange) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "PublicKey"
	err = en.Append(0x82, 0xa9, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.PublicKey)
	if err != nil {
		return
	}
	// write "Power"
	err = en.Append(0xa5, 0x50, 0x6f, 0x77, 0x65, 0x72)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Power)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *GTValidatorChange) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "PublicKey"
	o = append(o, 0x82, 0xa9, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79)
	o = msgp.AppendBytes(o, z.PublicKey)
	// string "Power"
	o = append(o, 0xa5, 0x50, 0x6f, 0x77, 0x65, 0x72)
	o = msgp.AppendInt64(o, z.Power)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *GTValidatorChange) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "PublicKey":
			z.PublicKey, bts, err = msgp.ReadBytesBytes(bts, z.PublicKey)
			if err != nil {
				return
			}
		case "Power":
			z.Power, bts, err = msgp.ReadInt64Bytes(bts)
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
func (z *GTValidatorChange) Msgsize() (s int) {
	s = 1 + 10 + msgp.BytesPrefixSize + len(z.PublicKey) + 6 + msgp.Int64Size
	return
}
