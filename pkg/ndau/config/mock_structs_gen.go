package config

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *MockChaosChain) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Namespaces":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Namespaces) >= int(zb0002) {
				z.Namespaces = (z.Namespaces)[:zb0002]
			} else {
				z.Namespaces = make([]MockNamespace, zb0002)
			}
			for za0001 := range z.Namespaces {
				err = z.Namespaces[za0001].DecodeMsg(dc)
				if err != nil {
					return
				}
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
func (z *MockChaosChain) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Namespaces"
	err = en.Append(0x81, 0xaa, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Namespaces)))
	if err != nil {
		return
	}
	for za0001 := range z.Namespaces {
		err = z.Namespaces[za0001].EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *MockChaosChain) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Namespaces"
	o = append(o, 0x81, 0xaa, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Namespaces)))
	for za0001 := range z.Namespaces {
		o, err = z.Namespaces[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MockChaosChain) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Namespaces":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Namespaces) >= int(zb0002) {
				z.Namespaces = (z.Namespaces)[:zb0002]
			} else {
				z.Namespaces = make([]MockNamespace, zb0002)
			}
			for za0001 := range z.Namespaces {
				bts, err = z.Namespaces[za0001].UnmarshalMsg(bts)
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
func (z *MockChaosChain) Msgsize() (s int) {
	s = 1 + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.Namespaces {
		s += z.Namespaces[za0001].Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MockKeyValue) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	err = z.Key.DecodeMsg(dc)
	if err != nil {
		return
	}
	err = z.Value.DecodeMsg(dc)
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *MockKeyValue) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 2
	err = en.Append(0x92)
	if err != nil {
		return
	}
	err = z.Key.EncodeMsg(en)
	if err != nil {
		return
	}
	err = z.Value.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *MockKeyValue) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 2
	o = append(o, 0x92)
	o, err = z.Key.MarshalMsg(o)
	if err != nil {
		return
	}
	o, err = z.Value.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MockKeyValue) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	bts, err = z.Key.UnmarshalMsg(bts)
	if err != nil {
		return
	}
	bts, err = z.Value.UnmarshalMsg(bts)
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *MockKeyValue) Msgsize() (s int) {
	s = 1 + z.Key.Msgsize() + z.Value.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MockNamespace) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Namespace":
			err = z.Namespace.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Data":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Data) >= int(zb0002) {
				z.Data = (z.Data)[:zb0002]
			} else {
				z.Data = make([]MockKeyValue, zb0002)
			}
			for za0001 := range z.Data {
				var zb0003 uint32
				zb0003, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if zb0003 != 2 {
					err = msgp.ArrayError{Wanted: 2, Got: zb0003}
					return
				}
				err = z.Data[za0001].Key.DecodeMsg(dc)
				if err != nil {
					return
				}
				err = z.Data[za0001].Value.DecodeMsg(dc)
				if err != nil {
					return
				}
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
func (z *MockNamespace) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Namespace"
	err = en.Append(0x82, 0xa9, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65)
	if err != nil {
		return
	}
	err = z.Namespace.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Data"
	err = en.Append(0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Data)))
	if err != nil {
		return
	}
	for za0001 := range z.Data {
		// array header, size 2
		err = en.Append(0x92)
		if err != nil {
			return
		}
		err = z.Data[za0001].Key.EncodeMsg(en)
		if err != nil {
			return
		}
		err = z.Data[za0001].Value.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *MockNamespace) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Namespace"
	o = append(o, 0x82, 0xa9, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65)
	o, err = z.Namespace.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Data"
	o = append(o, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Data)))
	for za0001 := range z.Data {
		// array header, size 2
		o = append(o, 0x92)
		o, err = z.Data[za0001].Key.MarshalMsg(o)
		if err != nil {
			return
		}
		o, err = z.Data[za0001].Value.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MockNamespace) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "Namespace":
			bts, err = z.Namespace.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Data":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Data) >= int(zb0002) {
				z.Data = (z.Data)[:zb0002]
			} else {
				z.Data = make([]MockKeyValue, zb0002)
			}
			for za0001 := range z.Data {
				var zb0003 uint32
				zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if zb0003 != 2 {
					err = msgp.ArrayError{Wanted: 2, Got: zb0003}
					return
				}
				bts, err = z.Data[za0001].Key.UnmarshalMsg(bts)
				if err != nil {
					return
				}
				bts, err = z.Data[za0001].Value.UnmarshalMsg(bts)
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
func (z *MockNamespace) Msgsize() (s int) {
	s = 1 + 10 + z.Namespace.Msgsize() + 5 + msgp.ArrayHeaderSize
	for za0001 := range z.Data {
		s += 1 + z.Data[za0001].Key.Msgsize() + z.Data[za0001].Value.Msgsize()
	}
	return
}
