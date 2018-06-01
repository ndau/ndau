package address

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *DataStore) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "key":
			err = z.Key.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "drv":
			z.Derivation, err = dc.ReadBytes(z.Derivation)
			if err != nil {
				return
			}
		case "crc":
			z.Crc, err = dc.ReadUint32()
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
func (z *DataStore) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "key"
	err = en.Append(0x83, 0xa3, 0x6b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = z.Key.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "drv"
	err = en.Append(0xa3, 0x64, 0x72, 0x76)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Derivation)
	if err != nil {
		return
	}
	// write "crc"
	err = en.Append(0xa3, 0x63, 0x72, 0x63)
	if err != nil {
		return
	}
	err = en.WriteUint32(z.Crc)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *DataStore) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "key"
	o = append(o, 0x83, 0xa3, 0x6b, 0x65, 0x79)
	o, err = z.Key.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "drv"
	o = append(o, 0xa3, 0x64, 0x72, 0x76)
	o = msgp.AppendBytes(o, z.Derivation)
	// string "crc"
	o = append(o, 0xa3, 0x63, 0x72, 0x63)
	o = msgp.AppendUint32(o, z.Crc)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DataStore) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "key":
			bts, err = z.Key.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "drv":
			z.Derivation, bts, err = msgp.ReadBytesBytes(bts, z.Derivation)
			if err != nil {
				return
			}
		case "crc":
			z.Crc, bts, err = msgp.ReadUint32Bytes(bts)
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
func (z *DataStore) Msgsize() (s int) {
	s = 1 + 4 + z.Key.Msgsize() + 4 + msgp.BytesPrefixSize + len(z.Derivation) + 4 + msgp.Uint32Size
	return
}
