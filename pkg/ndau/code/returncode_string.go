// Code generated by "stringer -type=ReturnCode"; DO NOT EDIT.

package code

import "strconv"

const _ReturnCode_name = "OKUnknownTransactionInvalidTransactionErrorApplyingTransaction"

var _ReturnCode_index = [...]uint8{0, 2, 20, 38, 62}

func (i ReturnCode) String() string {
	if i >= ReturnCode(len(_ReturnCode_index)-1) {
		return "ReturnCode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ReturnCode_name[_ReturnCode_index[i]:_ReturnCode_index[i+1]]
}
