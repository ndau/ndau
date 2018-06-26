package ndau

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// MarshalMsg implements msgp.Marshaler
func (z *ChangeEscrowPeriod) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Target"
	o = append(o, 0x83, 0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Period"
	o = append(o, 0xa6, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64)
	o, err = z.Period.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Signature"
	o = append(o, 0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	o, err = z.Signature.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ChangeEscrowPeriod) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Target":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Period":
			bts, err = z.Period.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Signature":
			bts, err = z.Signature.UnmarshalMsg(bts)
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
func (z *ChangeEscrowPeriod) Msgsize() (s int) {
	s = 1 + 7 + z.Target.Msgsize() + 7 + z.Period.Msgsize() + 10 + z.Signature.Msgsize()
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ChangeTransferKey) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "Target"
	o = append(o, 0x85, 0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "NewKey"
	o = append(o, 0xa6, 0x4e, 0x65, 0x77, 0x4b, 0x65, 0x79)
	o, err = z.NewKey.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "SigningKey"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x4b, 0x65, 0x79)
	o, err = z.SigningKey.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "KeyKind"
	o = append(o, 0xa7, 0x4b, 0x65, 0x79, 0x4b, 0x69, 0x6e, 0x64)
	o = msgp.AppendByte(o, byte(z.KeyKind))
	// string "Signature"
	o = append(o, 0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	o, err = z.Signature.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ChangeTransferKey) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Target":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "NewKey":
			bts, err = z.NewKey.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "SigningKey":
			bts, err = z.SigningKey.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "KeyKind":
			{
				var zb0002 byte
				zb0002, bts, err = msgp.ReadByteBytes(bts)
				if err != nil {
					return
				}
				z.KeyKind = SigningKeyKind(zb0002)
			}
		case "Signature":
			bts, err = z.Signature.UnmarshalMsg(bts)
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
func (z *ChangeTransferKey) Msgsize() (s int) {
	s = 1 + 7 + z.Target.Msgsize() + 7 + z.NewKey.Msgsize() + 11 + z.SigningKey.Msgsize() + 8 + msgp.ByteSize + 10 + z.Signature.Msgsize()
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

// MarshalMsg implements msgp.Marshaler
func (z *ReleaseFromEndowment) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Destination"
	o = append(o, 0x83, 0xab, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e)
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
	// string "Signature"
	o = append(o, 0xa9, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	o, err = z.Signature.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ReleaseFromEndowment) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Signature":
			bts, err = z.Signature.UnmarshalMsg(bts)
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
func (z *ReleaseFromEndowment) Msgsize() (s int) {
	s = 1 + 12 + z.Destination.Msgsize() + 4 + z.Qty.Msgsize() + 10 + z.Signature.Msgsize()
	return
}

// MarshalMsg implements msgp.Marshaler
func (z SigningKeyKind) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendByte(o, byte(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SigningKeyKind) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 byte
		zb0001, bts, err = msgp.ReadByteBytes(bts)
		if err != nil {
			return
		}
		(*z) = SigningKeyKind(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SigningKeyKind) Msgsize() (s int) {
	s = msgp.ByteSize
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
