package attendance

import "errors"

var (
	ErrAttendanceCorrectionInvalid   = errors.New("attendance correction invalid")
	ErrAttendanceCorrectionReason    = errors.New("attendance correction reason invalid")
	ErrAttendanceCorrectionForbidden = errors.New("attendance correction forbidden")
)
