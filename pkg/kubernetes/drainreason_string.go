// Code generated by "stringer -type=DrainReason"; DO NOT EDIT.

package kubernetes

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Updating-0]
	_ = x[Rebooting-1]
	_ = x[Deleting-2]
}

const _DrainReason_name = "UpdatingRebootingDeleting"

var _DrainReason_index = [...]uint8{0, 8, 17, 25}

func (i DrainReason) String() string {
	if i < 0 || i >= DrainReason(len(_DrainReason_index)-1) {
		return "DrainReason(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DrainReason_name[_DrainReason_index[i]:_DrainReason_index[i+1]]
}
