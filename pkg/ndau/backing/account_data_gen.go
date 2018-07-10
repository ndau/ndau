package backing

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/tinylib/msgp/msgp"
)

// MarshalMsg implements msgp.Marshaler
func (z *AccountData) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 12
	// string "Balance"
	o = append(o, 0x8c, 0xa7, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	o, err = z.Balance.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "TransferKey"
	o = append(o, 0xab, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x4b, 0x65, 0x79)
	if z.TransferKey == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.TransferKey.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "RewardsTarget"
	o = append(o, 0xad, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x73, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74)
	if z.RewardsTarget == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.RewardsTarget.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "DelegationNode"
	o = append(o, 0xae, 0x44, 0x65, 0x6c, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x6f, 0x64, 0x65)
	if z.DelegationNode == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.DelegationNode.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Lock"
	o = append(o, 0xa4, 0x4c, 0x6f, 0x63, 0x6b)
	if z.Lock == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.Lock.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Stake"
	o = append(o, 0xa5, 0x53, 0x74, 0x61, 0x6b, 0x65)
	if z.Stake == nil {
		o = msgp.AppendNil(o)
	} else {
		// map header, size 2
		// string "Point"
		o = append(o, 0x82, 0xa5, 0x50, 0x6f, 0x69, 0x6e, 0x74)
		o, err = z.Stake.Point.MarshalMsg(o)
		if err != nil {
			return
		}
		// string "Address"
		o = append(o, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
		o, err = z.Stake.Address.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "LastEAIUpdate"
	o = append(o, 0xad, 0x4c, 0x61, 0x73, 0x74, 0x45, 0x41, 0x49, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65)
	o, err = z.LastEAIUpdate.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "LastWAAUpdate"
	o = append(o, 0xad, 0x4c, 0x61, 0x73, 0x74, 0x57, 0x41, 0x41, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65)
	o, err = z.LastWAAUpdate.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "WeightedAverageAge"
	o = append(o, 0xb2, 0x57, 0x65, 0x69, 0x67, 0x68, 0x74, 0x65, 0x64, 0x41, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x41, 0x67, 0x65)
	o, err = z.WeightedAverageAge.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Sequence"
	o = append(o, 0xa8, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65)
	o = msgp.AppendUint64(o, z.Sequence)
	// string "Escrows"
	o = append(o, 0xa7, 0x45, 0x73, 0x63, 0x72, 0x6f, 0x77, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Escrows)))
	for za0001 := range z.Escrows {
		// map header, size 2
		// string "Qty"
		o = append(o, 0x82, 0xa3, 0x51, 0x74, 0x79)
		o, err = z.Escrows[za0001].Qty.MarshalMsg(o)
		if err != nil {
			return
		}
		// string "Expiry"
		o = append(o, 0xa6, 0x45, 0x78, 0x70, 0x69, 0x72, 0x79)
		o, err = z.Escrows[za0001].Expiry.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "EscrowSettings"
	o = append(o, 0xae, 0x45, 0x73, 0x63, 0x72, 0x6f, 0x77, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73)
	o, err = z.EscrowSettings.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AccountData) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Balance":
			bts, err = z.Balance.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "TransferKey":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.TransferKey = nil
			} else {
				if z.TransferKey == nil {
					z.TransferKey = new(signature.PublicKey)
				}
				bts, err = z.TransferKey.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "RewardsTarget":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.RewardsTarget = nil
			} else {
				if z.RewardsTarget == nil {
					z.RewardsTarget = new(address.Address)
				}
				bts, err = z.RewardsTarget.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "DelegationNode":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.DelegationNode = nil
			} else {
				if z.DelegationNode == nil {
					z.DelegationNode = new(address.Address)
				}
				bts, err = z.DelegationNode.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "Lock":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Lock = nil
			} else {
				if z.Lock == nil {
					z.Lock = new(Lock)
				}
				bts, err = z.Lock.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "Stake":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Stake = nil
			} else {
				if z.Stake == nil {
					z.Stake = new(Stake)
				}
				var zb0002 uint32
				zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zb0002 > 0 {
					zb0002--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Point":
						bts, err = z.Stake.Point.UnmarshalMsg(bts)
						if err != nil {
							return
						}
					case "Address":
						bts, err = z.Stake.Address.UnmarshalMsg(bts)
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
			}
		case "LastEAIUpdate":
			bts, err = z.LastEAIUpdate.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "LastWAAUpdate":
			bts, err = z.LastWAAUpdate.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "WeightedAverageAge":
			bts, err = z.WeightedAverageAge.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Sequence":
			z.Sequence, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Escrows":
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Escrows) >= int(zb0003) {
				z.Escrows = (z.Escrows)[:zb0003]
			} else {
				z.Escrows = make([]Escrow, zb0003)
			}
			for za0001 := range z.Escrows {
				var zb0004 uint32
				zb0004, bts, err = msgp.ReadMapHeaderBytes(bts)
				if err != nil {
					return
				}
				for zb0004 > 0 {
					zb0004--
					field, bts, err = msgp.ReadMapKeyZC(bts)
					if err != nil {
						return
					}
					switch msgp.UnsafeString(field) {
					case "Qty":
						bts, err = z.Escrows[za0001].Qty.UnmarshalMsg(bts)
						if err != nil {
							return
						}
					case "Expiry":
						bts, err = z.Escrows[za0001].Expiry.UnmarshalMsg(bts)
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
			}
		case "EscrowSettings":
			bts, err = z.EscrowSettings.UnmarshalMsg(bts)
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
func (z *AccountData) Msgsize() (s int) {
	s = 1 + 8 + z.Balance.Msgsize() + 12
	if z.TransferKey == nil {
		s += msgp.NilSize
	} else {
		s += z.TransferKey.Msgsize()
	}
	s += 14
	if z.RewardsTarget == nil {
		s += msgp.NilSize
	} else {
		s += z.RewardsTarget.Msgsize()
	}
	s += 15
	if z.DelegationNode == nil {
		s += msgp.NilSize
	} else {
		s += z.DelegationNode.Msgsize()
	}
	s += 5
	if z.Lock == nil {
		s += msgp.NilSize
	} else {
		s += z.Lock.Msgsize()
	}
	s += 6
	if z.Stake == nil {
		s += msgp.NilSize
	} else {
		s += 1 + 6 + z.Stake.Point.Msgsize() + 8 + z.Stake.Address.Msgsize()
	}
	s += 14 + z.LastEAIUpdate.Msgsize() + 14 + z.LastWAAUpdate.Msgsize() + 19 + z.WeightedAverageAge.Msgsize() + 9 + msgp.Uint64Size + 8 + msgp.ArrayHeaderSize
	for za0001 := range z.Escrows {
		s += 1 + 4 + z.Escrows[za0001].Qty.Msgsize() + 7 + z.Escrows[za0001].Expiry.Msgsize()
	}
	s += 15 + z.EscrowSettings.Msgsize()
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Escrow) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Qty"
	o = append(o, 0x82, 0xa3, 0x51, 0x74, 0x79)
	o, err = z.Qty.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Expiry"
	o = append(o, 0xa6, 0x45, 0x78, 0x70, 0x69, 0x72, 0x79)
	o, err = z.Expiry.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Escrow) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Qty":
			bts, err = z.Qty.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Expiry":
			bts, err = z.Expiry.UnmarshalMsg(bts)
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
func (z *Escrow) Msgsize() (s int) {
	s = 1 + 4 + z.Qty.Msgsize() + 7 + z.Expiry.Msgsize()
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *EscrowSettings) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Duration"
	o = append(o, 0x83, 0xa8, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	o, err = z.Duration.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "ChangesAt"
	o = append(o, 0xa9, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x41, 0x74)
	if z.ChangesAt == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.ChangesAt.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Next"
	o = append(o, 0xa4, 0x4e, 0x65, 0x78, 0x74)
	if z.Next == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.Next.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *EscrowSettings) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Duration":
			bts, err = z.Duration.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "ChangesAt":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.ChangesAt = nil
			} else {
				if z.ChangesAt == nil {
					z.ChangesAt = new(math.Timestamp)
				}
				bts, err = z.ChangesAt.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "Next":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Next = nil
			} else {
				if z.Next == nil {
					z.Next = new(math.Duration)
				}
				bts, err = z.Next.UnmarshalMsg(bts)
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
func (z *EscrowSettings) Msgsize() (s int) {
	s = 1 + 9 + z.Duration.Msgsize() + 10
	if z.ChangesAt == nil {
		s += msgp.NilSize
	} else {
		s += z.ChangesAt.Msgsize()
	}
	s += 5
	if z.Next == nil {
		s += msgp.NilSize
	} else {
		s += z.Next.Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Stake) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Point"
	o = append(o, 0x82, 0xa5, 0x50, 0x6f, 0x69, 0x6e, 0x74)
	o, err = z.Point.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Address"
	o = append(o, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	o, err = z.Address.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Stake) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Point":
			bts, err = z.Point.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Address":
			bts, err = z.Address.UnmarshalMsg(bts)
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
func (z *Stake) Msgsize() (s int) {
	s = 1 + 6 + z.Point.Msgsize() + 8 + z.Address.Msgsize()
	return
}
