package ndau

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/tinylib/msgp/msgp"
)

// MarshalMsg implements msgp.Marshaler
func (z *ChangeSettlementPeriod) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "tgt"
	o = append(o, 0x84, 0xa3, 0x74, 0x67, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "per"
	o = append(o, 0xa3, 0x70, 0x65, 0x72)
	o, err = z.Period.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0001 := range z.Signatures {
		o, err = z.Signatures[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ChangeSettlementPeriod) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "tgt":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "per":
			bts, err = z.Period.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0002) {
				z.Signatures = (z.Signatures)[:zb0002]
			} else {
				z.Signatures = make([]signature.Signature, zb0002)
			}
			for za0001 := range z.Signatures {
				bts, err = z.Signatures[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
func (z *ChangeSettlementPeriod) Msgsize() (s int) {
	s = 1 + 4 + z.Target.Msgsize() + 4 + z.Period.Msgsize() + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ChangeValidation) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "tgt"
	o = append(o, 0x85, 0xa3, 0x74, 0x67, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "key"
	o = append(o, 0xa3, 0x6b, 0x65, 0x79)
	o = msgp.AppendArrayHeader(o, uint32(len(z.NewKeys)))
	for za0001 := range z.NewKeys {
		o, err = z.NewKeys[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "val"
	o = append(o, 0xa3, 0x76, 0x61, 0x6c)
	o = msgp.AppendBytes(o, z.ValidationScript)
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0002 := range z.Signatures {
		o, err = z.Signatures[za0002].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ChangeValidation) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "tgt":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "key":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.NewKeys) >= int(zb0002) {
				z.NewKeys = (z.NewKeys)[:zb0002]
			} else {
				z.NewKeys = make([]signature.PublicKey, zb0002)
			}
			for za0001 := range z.NewKeys {
				bts, err = z.NewKeys[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "val":
			z.ValidationScript, bts, err = msgp.ReadBytesBytes(bts, z.ValidationScript)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0003) {
				z.Signatures = (z.Signatures)[:zb0003]
			} else {
				z.Signatures = make([]signature.Signature, zb0003)
			}
			for za0002 := range z.Signatures {
				bts, err = z.Signatures[za0002].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
func (z *ChangeValidation) Msgsize() (s int) {
	s = 1 + 4 + z.Target.Msgsize() + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.NewKeys {
		s += z.NewKeys[za0001].Msgsize()
	}
	s += 4 + msgp.BytesPrefixSize + len(z.ValidationScript) + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0002 := range z.Signatures {
		s += z.Signatures[za0002].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ClaimAccount) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "tgt"
	o = append(o, 0x86, 0xa3, 0x74, 0x67, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "own"
	o = append(o, 0xa3, 0x6f, 0x77, 0x6e)
	o, err = z.Ownership.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "key"
	o = append(o, 0xa3, 0x6b, 0x65, 0x79)
	o = msgp.AppendArrayHeader(o, uint32(len(z.TransferKeys)))
	for za0001 := range z.TransferKeys {
		o, err = z.TransferKeys[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "val"
	o = append(o, 0xa3, 0x76, 0x61, 0x6c)
	o = msgp.AppendBytes(o, z.ValidationScript)
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o, err = z.Signature.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ClaimAccount) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "tgt":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "own":
			bts, err = z.Ownership.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "key":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.TransferKeys) >= int(zb0002) {
				z.TransferKeys = (z.TransferKeys)[:zb0002]
			} else {
				z.TransferKeys = make([]signature.PublicKey, zb0002)
			}
			for za0001 := range z.TransferKeys {
				bts, err = z.TransferKeys[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "val":
			z.ValidationScript, bts, err = msgp.ReadBytesBytes(bts, z.ValidationScript)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
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
func (z *ClaimAccount) Msgsize() (s int) {
	s = 1 + 4 + z.Target.Msgsize() + 4 + z.Ownership.Msgsize() + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.TransferKeys {
		s += z.TransferKeys[za0001].Msgsize()
	}
	s += 4 + msgp.BytesPrefixSize + len(z.ValidationScript) + 4 + msgp.Uint64Size + 4 + z.Signature.Msgsize()
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CreditEAI) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "nod"
	o = append(o, 0x83, 0xa3, 0x6e, 0x6f, 0x64)
	o, err = z.Node.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0001 := range z.Signatures {
		o, err = z.Signatures[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CreditEAI) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "nod":
			bts, err = z.Node.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0002) {
				z.Signatures = (z.Signatures)[:zb0002]
			} else {
				z.Signatures = make([]signature.Signature, zb0002)
			}
			for za0001 := range z.Signatures {
				bts, err = z.Signatures[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
func (z *CreditEAI) Msgsize() (s int) {
	s = 1 + 4 + z.Node.Msgsize() + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Delegate) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "tgt"
	o = append(o, 0x84, 0xa3, 0x74, 0x67, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "nod"
	o = append(o, 0xa3, 0x6e, 0x6f, 0x64)
	o, err = z.Node.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0001 := range z.Signatures {
		o, err = z.Signatures[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Delegate) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "tgt":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "nod":
			bts, err = z.Node.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0002) {
				z.Signatures = (z.Signatures)[:zb0002]
			} else {
				z.Signatures = make([]signature.Signature, zb0002)
			}
			for za0001 := range z.Signatures {
				bts, err = z.Signatures[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
func (z *Delegate) Msgsize() (s int) {
	s = 1 + 4 + z.Target.Msgsize() + 4 + z.Node.Msgsize() + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
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

// MarshalMsg implements msgp.Marshaler
func (z *Lock) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "tgt"
	o = append(o, 0x84, 0xa3, 0x74, 0x67, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "per"
	o = append(o, 0xa3, 0x70, 0x65, 0x72)
	o, err = z.Period.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0001 := range z.Signatures {
		o, err = z.Signatures[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Lock) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "tgt":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "per":
			bts, err = z.Period.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0002) {
				z.Signatures = (z.Signatures)[:zb0002]
			} else {
				z.Signatures = make([]signature.Signature, zb0002)
			}
			for za0001 := range z.Signatures {
				bts, err = z.Signatures[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
func (z *Lock) Msgsize() (s int) {
	s = 1 + 4 + z.Target.Msgsize() + 4 + z.Period.Msgsize() + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Notify) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "tgt"
	o = append(o, 0x83, 0xa3, 0x74, 0x67, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0001 := range z.Signatures {
		o, err = z.Signatures[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Notify) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "tgt":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0002) {
				z.Signatures = (z.Signatures)[:zb0002]
			} else {
				z.Signatures = make([]signature.Signature, zb0002)
			}
			for za0001 := range z.Signatures {
				bts, err = z.Signatures[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
func (z *Notify) Msgsize() (s int) {
	s = 1 + 4 + z.Target.Msgsize() + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ReleaseFromEndowment) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "dst"
	o = append(o, 0x85, 0xa3, 0x64, 0x73, 0x74)
	o, err = z.Destination.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "qty"
	o = append(o, 0xa3, 0x71, 0x74, 0x79)
	o, err = z.Qty.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "fee"
	o = append(o, 0xa3, 0x66, 0x65, 0x65)
	o, err = z.TxFeeAcct.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0001 := range z.Signatures {
		o, err = z.Signatures[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
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
		case "dst":
			bts, err = z.Destination.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "qty":
			bts, err = z.Qty.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "fee":
			bts, err = z.TxFeeAcct.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0002) {
				z.Signatures = (z.Signatures)[:zb0002]
			} else {
				z.Signatures = make([]signature.Signature, zb0002)
			}
			for za0001 := range z.Signatures {
				bts, err = z.Signatures[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
	s = 1 + 4 + z.Destination.Msgsize() + 4 + z.Qty.Msgsize() + 4 + z.TxFeeAcct.Msgsize() + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SetRewardsDestination) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "src"
	o = append(o, 0x84, 0xa3, 0x73, 0x72, 0x63)
	o, err = z.Source.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "dst"
	o = append(o, 0xa3, 0x64, 0x73, 0x74)
	o, err = z.Destination.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0001 := range z.Signatures {
		o, err = z.Signatures[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SetRewardsDestination) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "src":
			bts, err = z.Source.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "dst":
			bts, err = z.Destination.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0002) {
				z.Signatures = (z.Signatures)[:zb0002]
			} else {
				z.Signatures = make([]signature.Signature, zb0002)
			}
			for za0001 := range z.Signatures {
				bts, err = z.Signatures[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
func (z *SetRewardsDestination) Msgsize() (s int) {
	s = 1 + 4 + z.Source.Msgsize() + 4 + z.Destination.Msgsize() + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Transfer) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "src"
	o = append(o, 0x85, 0xa3, 0x73, 0x72, 0x63)
	o, err = z.Source.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "dst"
	o = append(o, 0xa3, 0x64, 0x73, 0x74)
	o, err = z.Destination.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "qty"
	o = append(o, 0xa3, 0x71, 0x74, 0x79)
	o, err = z.Qty.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "seq"
	o = append(o, 0xa3, 0x73, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "sig"
	o = append(o, 0xa3, 0x73, 0x69, 0x67)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Signatures)))
	for za0001 := range z.Signatures {
		o, err = z.Signatures[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
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
		case "src":
			bts, err = z.Source.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "dst":
			bts, err = z.Destination.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "qty":
			bts, err = z.Qty.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "seq":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "sig":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Signatures) >= int(zb0002) {
				z.Signatures = (z.Signatures)[:zb0002]
			} else {
				z.Signatures = make([]signature.Signature, zb0002)
			}
			for za0001 := range z.Signatures {
				bts, err = z.Signatures[za0001].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
	s = 1 + 4 + z.Source.Msgsize() + 4 + z.Destination.Msgsize() + 4 + z.Qty.Msgsize() + 4 + msgp.Uint64Size + 4 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}
