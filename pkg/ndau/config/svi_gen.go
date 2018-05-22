package config

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *B64Data) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 []byte
		zb0001, err = dc.ReadBytes([]byte((*z)))
		if err != nil {
			return
		}
		(*z) = B64Data(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z B64Data) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteBytes([]byte(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z B64Data) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendBytes(o, []byte(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *B64Data) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 []byte
		zb0001, bts, err = msgp.ReadBytesBytes(bts, []byte((*z)))
		if err != nil {
			return
		}
		(*z) = B64Data(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z B64Data) Msgsize() (s int) {
	s = msgp.BytesPrefixSize + len([]byte(z))
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
	{
		var zb0002 []byte
		zb0002, err = dc.ReadBytes([]byte(z.Namespace))
		if err != nil {
			return
		}
		z.Namespace = B64Data(zb0002)
	}
	{
		var zb0003 []byte
		zb0003, err = dc.ReadBytes([]byte(z.Key))
		if err != nil {
			return
		}
		z.Key = B64Data(zb0003)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *NamespacedKey) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 2
	err = en.Append(0x92)
	if err != nil {
		return
	}
	err = en.WriteBytes([]byte(z.Namespace))
	if err != nil {
		return
	}
	err = en.WriteBytes([]byte(z.Key))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *NamespacedKey) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 2
	o = append(o, 0x92)
	o = msgp.AppendBytes(o, []byte(z.Namespace))
	o = msgp.AppendBytes(o, []byte(z.Key))
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
	{
		var zb0002 []byte
		zb0002, bts, err = msgp.ReadBytesBytes(bts, []byte(z.Namespace))
		if err != nil {
			return
		}
		z.Namespace = B64Data(zb0002)
	}
	{
		var zb0003 []byte
		zb0003, bts, err = msgp.ReadBytesBytes(bts, []byte(z.Key))
		if err != nil {
			return
		}
		z.Key = B64Data(zb0003)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *NamespacedKey) Msgsize() (s int) {
	s = 1 + msgp.BytesPrefixSize + len([]byte(z.Namespace)) + msgp.BytesPrefixSize + len([]byte(z.Key))
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
			err = z.Current.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Future":
			err = z.Future.DecodeMsg(dc)
			if err != nil {
				return
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
	err = en.Append(0x83, 0xa7, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = z.Current.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Future"
	err = en.Append(0xa6, 0x46, 0x75, 0x74, 0x75, 0x72, 0x65)
	if err != nil {
		return
	}
	err = z.Future.EncodeMsg(en)
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
	o = append(o, 0x83, 0xa7, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74)
	o, err = z.Current.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Future"
	o = append(o, 0xa6, 0x46, 0x75, 0x74, 0x75, 0x72, 0x65)
	o, err = z.Future.MarshalMsg(o)
	if err != nil {
		return
	}
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
			bts, err = z.Current.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Future":
			bts, err = z.Future.UnmarshalMsg(bts)
			if err != nil {
				return
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
	s = 1 + 8 + z.Current.Msgsize() + 7 + z.Future.Msgsize() + 9 + msgp.Uint64Size
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
		var field []byte
		_ = field
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
			case "Current":
				err = zb0002.Current.DecodeMsg(dc)
				if err != nil {
					return
				}
			case "Future":
				err = zb0002.Future.DecodeMsg(dc)
				if err != nil {
					return
				}
			case "ChangeOn":
				zb0002.ChangeOn, err = dc.ReadUint64()
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
	for zb0005, zb0006 := range z {
		err = en.WriteString(zb0005)
		if err != nil {
			return
		}
		// map header, size 3
		// write "Current"
		err = en.Append(0x83, 0xa7, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74)
		if err != nil {
			return
		}
		err = zb0006.Current.EncodeMsg(en)
		if err != nil {
			return
		}
		// write "Future"
		err = en.Append(0xa6, 0x46, 0x75, 0x74, 0x75, 0x72, 0x65)
		if err != nil {
			return
		}
		err = zb0006.Future.EncodeMsg(en)
		if err != nil {
			return
		}
		// write "ChangeOn"
		err = en.Append(0xa8, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4f, 0x6e)
		if err != nil {
			return
		}
		err = en.WriteUint64(zb0006.ChangeOn)
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
	for zb0005, zb0006 := range z {
		o = msgp.AppendString(o, zb0005)
		// map header, size 3
		// string "Current"
		o = append(o, 0x83, 0xa7, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74)
		o, err = zb0006.Current.MarshalMsg(o)
		if err != nil {
			return
		}
		// string "Future"
		o = append(o, 0xa6, 0x46, 0x75, 0x74, 0x75, 0x72, 0x65)
		o, err = zb0006.Future.MarshalMsg(o)
		if err != nil {
			return
		}
		// string "ChangeOn"
		o = append(o, 0xa8, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4f, 0x6e)
		o = msgp.AppendUint64(o, zb0006.ChangeOn)
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
		var field []byte
		_ = field
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
			case "Current":
				bts, err = zb0002.Current.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			case "Future":
				bts, err = zb0002.Future.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			case "ChangeOn":
				zb0002.ChangeOn, bts, err = msgp.ReadUint64Bytes(bts)
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
		(*z)[zb0001] = zb0002
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SVIMap) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zb0005, zb0006 := range z {
			_ = zb0006
			s += msgp.StringPrefixSize + len(zb0005) + 1 + 8 + zb0006.Current.Msgsize() + 7 + zb0006.Future.Msgsize() + 9 + msgp.Uint64Size
		}
	}
	return
}
