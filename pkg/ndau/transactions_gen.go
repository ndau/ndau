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
	// string "Target"
	o = append(o, 0x84, 0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
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
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signatures":
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
	s = 1 + 7 + z.Target.Msgsize() + 7 + z.Period.Msgsize() + 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ChangeValidation) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Target"
	o = append(o, 0x84, 0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "NewKeys"
	o = append(o, 0xa7, 0x4e, 0x65, 0x77, 0x4b, 0x65, 0x79, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.NewKeys)))
	for za0001 := range z.NewKeys {
		o, err = z.NewKeys[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
		case "Target":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "NewKeys":
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
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signatures":
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
	s = 1 + 7 + z.Target.Msgsize() + 8 + msgp.ArrayHeaderSize
	for za0001 := range z.NewKeys {
		s += z.NewKeys[za0001].Msgsize()
	}
	s += 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
	for za0002 := range z.Signatures {
		s += z.Signatures[za0002].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ClaimAccount) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Account"
	o = append(o, 0x84, 0xa7, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	o, err = z.Account.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Ownership"
	o = append(o, 0xa9, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70)
	o, err = z.Ownership.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "TransferKeys"
	o = append(o, 0xac, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x4b, 0x65, 0x79, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.TransferKeys)))
	for za0001 := range z.TransferKeys {
		o, err = z.TransferKeys[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
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
		case "Account":
			bts, err = z.Account.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Ownership":
			bts, err = z.Ownership.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "TransferKeys":
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
func (z *ClaimAccount) Msgsize() (s int) {
	s = 1 + 8 + z.Account.Msgsize() + 10 + z.Ownership.Msgsize() + 13 + msgp.ArrayHeaderSize
	for za0001 := range z.TransferKeys {
		s += z.TransferKeys[za0001].Msgsize()
	}
	s += 10 + z.Signature.Msgsize()
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ComputeEAI) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Node"
	o = append(o, 0x83, 0xa4, 0x4e, 0x6f, 0x64, 0x65)
	o, err = z.Node.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
func (z *ComputeEAI) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Node":
			bts, err = z.Node.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signatures":
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
func (z *ComputeEAI) Msgsize() (s int) {
	s = 1 + 5 + z.Node.Msgsize() + 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Delegate) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Target"
	o = append(o, 0x84, 0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Node"
	o = append(o, 0xa4, 0x4e, 0x6f, 0x64, 0x65)
	o, err = z.Node.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
		case "Target":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Node":
			bts, err = z.Node.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signatures":
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
	s = 1 + 7 + z.Target.Msgsize() + 5 + z.Node.Msgsize() + 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
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
	// string "Target"
	o = append(o, 0x84, 0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
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
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signatures":
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
	s = 1 + 7 + z.Target.Msgsize() + 7 + z.Period.Msgsize() + 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Notify) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Target"
	o = append(o, 0x83, 0xa6, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
	o, err = z.Target.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
		case "Target":
			bts, err = z.Target.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signatures":
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
	s = 1 + 7 + z.Target.Msgsize() + 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ReleaseFromEndowment) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "Destination"
	o = append(o, 0x85, 0xab, 0x44, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e)
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
	// string "TxFeeAcct"
	o = append(o, 0xa9, 0x54, 0x78, 0x46, 0x65, 0x65, 0x41, 0x63, 0x63, 0x74)
	o, err = z.TxFeeAcct.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
		case "TxFeeAcct":
			bts, err = z.TxFeeAcct.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signatures":
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
	s = 1 + 12 + z.Destination.Msgsize() + 4 + z.Qty.Msgsize() + 10 + z.TxFeeAcct.Msgsize() + 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SetRewardsDestination) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Source"
	o = append(o, 0x84, 0xa6, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65)
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
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Signatures":
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
	s = 1 + 7 + z.Source.Msgsize() + 12 + z.Destination.Msgsize() + 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
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
	// string "Signatures"
	o = append(o, 0xaa, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73)
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
		case "Signatures":
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
	s = 1 + 10 + z.Timestamp.Msgsize() + 7 + z.Source.Msgsize() + 12 + z.Destination.Msgsize() + 4 + z.Qty.Msgsize() + 9 + msgp.Uint64Size + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.Signatures {
		s += z.Signatures[za0001].Msgsize()
	}
	return
}
