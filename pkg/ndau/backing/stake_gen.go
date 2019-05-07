package backing

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *StakeData) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Point":
			err = z.Point.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "Point")
				return
			}
		case "RulesAcct":
			err = z.RulesAcct.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "RulesAcct")
				return
			}
		case "StakeTo":
			err = z.StakeTo.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "StakeTo")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *StakeData) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Point"
	err = en.Append(0x83, 0xa5, 0x50, 0x6f, 0x69, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = z.Point.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "Point")
		return
	}
	// write "RulesAcct"
	err = en.Append(0xa9, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x41, 0x63, 0x63, 0x74)
	if err != nil {
		return
	}
	err = z.RulesAcct.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "RulesAcct")
		return
	}
	// write "StakeTo"
	err = en.Append(0xa7, 0x53, 0x74, 0x61, 0x6b, 0x65, 0x54, 0x6f)
	if err != nil {
		return
	}
	err = z.StakeTo.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "StakeTo")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *StakeData) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Point"
	o = append(o, 0x83, 0xa5, 0x50, 0x6f, 0x69, 0x6e, 0x74)
	o, err = z.Point.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "Point")
		return
	}
	// string "RulesAcct"
	o = append(o, 0xa9, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x41, 0x63, 0x63, 0x74)
	o, err = z.RulesAcct.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "RulesAcct")
		return
	}
	// string "StakeTo"
	o = append(o, 0xa7, 0x53, 0x74, 0x61, 0x6b, 0x65, 0x54, 0x6f)
	o, err = z.StakeTo.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "StakeTo")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *StakeData) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Point":
			bts, err = z.Point.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "Point")
				return
			}
		case "RulesAcct":
			bts, err = z.RulesAcct.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "RulesAcct")
				return
			}
		case "StakeTo":
			bts, err = z.StakeTo.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "StakeTo")
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
func (z *StakeData) Msgsize() (s int) {
	s = 1 + 6 + z.Point.Msgsize() + 10 + z.RulesAcct.Msgsize() + 8 + z.StakeTo.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *StakeRules) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Script":
			z.Script, err = dc.ReadBytes(z.Script)
			if err != nil {
				err = msgp.WrapError(err, "Script")
				return
			}
		case "Inbound":
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, "Inbound")
				return
			}
			if z.Inbound == nil {
				z.Inbound = make(map[string]uint64, zb0002)
			} else if len(z.Inbound) > 0 {
				for key := range z.Inbound {
					delete(z.Inbound, key)
				}
			}
			for zb0002 > 0 {
				zb0002--
				var za0001 string
				var za0002 uint64
				za0001, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Inbound")
					return
				}
				za0002, err = dc.ReadUint64()
				if err != nil {
					err = msgp.WrapError(err, "Inbound", za0001)
					return
				}
				z.Inbound[za0001] = za0002
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *StakeRules) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Script"
	err = en.Append(0x82, 0xa6, 0x53, 0x63, 0x72, 0x69, 0x70, 0x74)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Script)
	if err != nil {
		err = msgp.WrapError(err, "Script")
		return
	}
	// write "Inbound"
	err = en.Append(0xa7, 0x49, 0x6e, 0x62, 0x6f, 0x75, 0x6e, 0x64)
	if err != nil {
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Inbound)))
	if err != nil {
		err = msgp.WrapError(err, "Inbound")
		return
	}
	for za0001, za0002 := range z.Inbound {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Inbound")
			return
		}
		err = en.WriteUint64(za0002)
		if err != nil {
			err = msgp.WrapError(err, "Inbound", za0001)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *StakeRules) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Script"
	o = append(o, 0x82, 0xa6, 0x53, 0x63, 0x72, 0x69, 0x70, 0x74)
	o = msgp.AppendBytes(o, z.Script)
	// string "Inbound"
	o = append(o, 0xa7, 0x49, 0x6e, 0x62, 0x6f, 0x75, 0x6e, 0x64)
	o = msgp.AppendMapHeader(o, uint32(len(z.Inbound)))
	for za0001, za0002 := range z.Inbound {
		o = msgp.AppendString(o, za0001)
		o = msgp.AppendUint64(o, za0002)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *StakeRules) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Script":
			z.Script, bts, err = msgp.ReadBytesBytes(bts, z.Script)
			if err != nil {
				err = msgp.WrapError(err, "Script")
				return
			}
		case "Inbound":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Inbound")
				return
			}
			if z.Inbound == nil {
				z.Inbound = make(map[string]uint64, zb0002)
			} else if len(z.Inbound) > 0 {
				for key := range z.Inbound {
					delete(z.Inbound, key)
				}
			}
			for zb0002 > 0 {
				var za0001 string
				var za0002 uint64
				zb0002--
				za0001, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Inbound")
					return
				}
				za0002, bts, err = msgp.ReadUint64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Inbound", za0001)
					return
				}
				z.Inbound[za0001] = za0002
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
func (z *StakeRules) Msgsize() (s int) {
	s = 1 + 7 + msgp.BytesPrefixSize + len(z.Script) + 8 + msgp.MapHeaderSize
	if z.Inbound != nil {
		for za0001, za0002 := range z.Inbound {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.Uint64Size
		}
	}
	return
}
