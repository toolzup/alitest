// Code generated by "stringer -type ParameterLocation"; DO NOT EDIT.

package alitest

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Query-0]
	_ = x[Header-1]
	_ = x[Path-2]
	_ = x[Cookie-3]
}

const _ParameterLocation_name = "QueryHeaderPathCookie"

var _ParameterLocation_index = [...]uint8{0, 5, 11, 15, 21}

func (i ParameterLocation) String() string {
	if i < 0 || i >= ParameterLocation(len(_ParameterLocation_index)-1) {
		return "ParameterLocation(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ParameterLocation_name[_ParameterLocation_index[i]:_ParameterLocation_index[i+1]]
}
