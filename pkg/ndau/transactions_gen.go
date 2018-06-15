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

// DecodeMsg implements msgp.Decodable
func (z *Transfer) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Timestamp":
			err = z.Timestamp.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Source":
			err = z.Source.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Destination":
			err = z.Destination.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Qty":
			err = z.Qty.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Sequence":
			z.Sequence, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "Signature":
			z.Signature, err = dc.ReadBytes(z.Signature)
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
func (z *Transfer) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "Timestamp"
	err = en.Append(0x86, 0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	if err != nil {
		return
	}
	err = z.Timestamp.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Source"
	err = en.Append(0xa6, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65)
	if err != nil {
		return
	}
	err = z.Source.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Destination"
	err = en.Append(0xab, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	if err != nil {
		return
	}
	err = z.Destination.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Qty"
	err = en.Append(0xa3, 0x51, 0x74, 0x79)
	if err != nil {
		return
	}
	err = z.Qty.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Sequence"
	err = en.Append(0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.Sequence)
	if err != nil {
		return
	}
	// write "Signature"
	err = en.Append(0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Signature)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Transfer) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "Timestamp"
	o = append(o, 0x86, 0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	o, err = z.Timestamp.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Source"
	o = append(o, 0xa6, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65)
	o, err = z.Source.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Destination"
	o = append(o, 0xab, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	o, err = z.Destination.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Qty"
	o = append(o, 0xa3, 0x51, 0x74, 0x79)
	o, err = z.Qty.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signature"
	o = append(o, 0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	o = msgp.AppendBytes(o, z.Signature)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Transfer) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Timestamp":
			bts, err = z.Timestamp.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Source":
			bts, err = z.Source.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Destination":
			bts, err = z.Destination.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Qty":
			bts, err = z.Qty.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signature":
			z.Signature, bts, err = msgp.ReadBytesBytes(bts, z.Signature)
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
func (z *Transfer) Msgsize() (s int) {
	s = 1 + 10 + z.Timestamp.Msgsize() + 7 + z.Source.Msgsize() + 12 + z.Destination.Msgsize() + 4 + z.Qty.Msgsize() + 9 + msgp.Uint64Size + 10 + msgp.BytesPrefixSize + len(z.Signature)
	return
}
