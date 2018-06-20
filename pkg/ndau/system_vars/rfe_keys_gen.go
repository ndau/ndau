package system_vars

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// MarshalMsg implements msgp.Marshaler
func (z ReleaseFromEndowmentKeys) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for za0001 := range z {
		o, err = z[za0001].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ReleaseFromEndowmentKeys) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(ReleaseFromEndowmentKeys, zb0002)
	}
	for zb0001 := range *z {
		bts, err = (*z)[zb0001].UnmarshalMsg(bts)
		if err != nil {
			return
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ReleaseFromEndowmentKeys) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0003 := range z {
		s += z[zb0003].Msgsize()
	}
	return
}
