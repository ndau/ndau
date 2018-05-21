package config

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *B64Data) DecodeMsg(dc *msgp.Reader) (err error) {
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
func (z B64Data) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z B64Data) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 0
	o = append(o, 0x80)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *B64Data) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
func (z B64Data) Msgsize() (s int) {
	s = 1
	return
}

// DecodeMsg implements msgp.Decodable
func (z *NamespacedKey) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	var field []byte
	_ = field
	var zb0002 uint32
	zb0002, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0002 > 0 {
		zb0002--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	var zb0003 uint32
	zb0003, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0003 > 0 {
		zb0003--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
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
func (z *NamespacedKey) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 2
	// map header, size 0
	err = en.Append(0x92, 0x80)
	if err != nil {
		return
	}
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *NamespacedKey) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 2
	// map header, size 0
	o = append(o, 0x92, 0x80)
	// map header, size 0
	o = append(o, 0x80)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *NamespacedKey) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	var field []byte
	_ = field
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
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	var zb0003 uint32
	zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0003 > 0 {
		zb0003--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
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
func (z *NamespacedKey) Msgsize() (s int) {
	s = 1 + 1 + 1
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SVIDeferredChange) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Current":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if zb0002 != 2 {
				err = msgp.ArrayError{Wanted: 2, Got: zb0002}
				return
			}
			var zb0003 uint32
			zb0003, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
			var zb0004 uint32
			zb0004, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zb0004 > 0 {
				zb0004--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "Future":
			var zb0005 uint32
			zb0005, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if zb0005 != 2 {
				err = msgp.ArrayError{Wanted: 2, Got: zb0005}
				return
			}
			var zb0006 uint32
			zb0006, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zb0006 > 0 {
				zb0006--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
			var zb0007 uint32
			zb0007, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			for zb0007 > 0 {
				zb0007--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				default:
					err = dc.Skip()
					if err != nil {
						return
					}
				}
			}
		case "ChangeOn":
			z.ChangeOn, err = dc.ReadUint64()
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
func (z *SVIDeferredChange) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Current"
	// array header, size 2
	// map header, size 0
	err = en.Append(0x83, 0xa7, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x92, 0x80)
	if err != nil {
		return
	}
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return
	}
	// write "Future"
	// array header, size 2
	// map header, size 0
	err = en.Append(0xa6, 0x46, 0x75, 0x74, 0x75, 0x72, 0x65, 0x92, 0x80)
	if err != nil {
		return
	}
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return
	}
	// write "ChangeOn"
	err = en.Append(0xa8, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4f, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteUint64(z.ChangeOn)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SVIDeferredChange) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Current"
	// array header, size 2
	// map header, size 0
	o = append(o, 0x83, 0xa7, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x92, 0x80)
	// map header, size 0
	o = append(o, 0x80)
	// string "Future"
	// array header, size 2
	// map header, size 0
	o = append(o, 0xa6, 0x46, 0x75, 0x74, 0x75, 0x72, 0x65, 0x92, 0x80)
	// map header, size 0
	o = append(o, 0x80)
	// string "ChangeOn"
	o = append(o, 0xa8, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4f, 0x6e)
	o = msgp.AppendUint64(o, z.ChangeOn)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SVIDeferredChange) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Current":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if zb0002 != 2 {
				err = msgp.ArrayError{Wanted: 2, Got: zb0002}
				return
			}
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
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
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "Future":
			var zb0005 uint32
			zb0005, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if zb0005 != 2 {
				err = msgp.ArrayError{Wanted: 2, Got: zb0005}
				return
			}
			var zb0006 uint32
			zb0006, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zb0006 > 0 {
				zb0006--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
			var zb0007 uint32
			zb0007, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			for zb0007 > 0 {
				zb0007--
				field, bts, err = msgp.ReadMapKeyZC(bts)
				if err != nil {
					return
				}
				switch msgp.UnsafeString(field) {
				default:
					bts, err = msgp.Skip(bts)
					if err != nil {
						return
					}
				}
			}
		case "ChangeOn":
			z.ChangeOn, bts, err = msgp.ReadUint64Bytes(bts)
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
func (z *SVIDeferredChange) Msgsize() (s int) {
	s = 1 + 8 + 1 + 1 + 1 + 7 + 1 + 1 + 1 + 9 + msgp.Uint64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SVIMap) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0003 uint32
	zb0003, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	if (*z) == nil && zb0003 > 0 {
		(*z) = make(SVIMap, zb0003)
	} else if len((*z)) > 0 {
		for key := range *z {
			delete((*z), key)
		}
	}
	for zb0003 > 0 {
		zb0003--
		var zb0001 string
		var zb0002 SVIDeferredChange
		zb0001, err = dc.ReadString()
		if err != nil {
			return
		}
		err = zb0002.DecodeMsg(dc)
		if err != nil {
			return
		}
		(*z)[zb0001] = zb0002
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z SVIMap) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zb0004, zb0005 := range z {
		err = en.WriteString(zb0004)
		if err != nil {
			return
		}
		err = zb0005.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z SVIMap) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for zb0004, zb0005 := range z {
		o = msgp.AppendString(o, zb0004)
		o, err = zb0005.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SVIMap) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0003 uint32
	zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	if (*z) == nil && zb0003 > 0 {
		(*z) = make(SVIMap, zb0003)
	} else if len((*z)) > 0 {
		for key := range *z {
			delete((*z), key)
		}
	}
	for zb0003 > 0 {
		var zb0001 string
		var zb0002 SVIDeferredChange
		zb0003--
		zb0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		bts, err = zb0002.UnmarshalMsg(bts)
		if err != nil {
			return
		}
		(*z)[zb0001] = zb0002
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SVIMap) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zb0004, zb0005 := range z {
			_ = zb0005
			s += msgp.StringPrefixSize + len(zb0004) + zb0005.Msgsize()
		}
	}
	return
}
