package attendance

import "errors"

var (
	ErrAdminAttendanceNotFound = errors.New("admin attendance record not found")
	ErrAdminAttendanceRange    = errors.New("admin attendance date range is invalid")
)
