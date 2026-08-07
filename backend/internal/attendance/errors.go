package attendance

import "errors"

var (
	ErrInvalidInput      = errors.New("attendance input is invalid")
	ErrInactiveAccount   = errors.New("attendance user is not active")
	ErrScheduleNotFound  = errors.New("attendance schedule assignment not found")
	ErrInactiveSchedule  = errors.New("attendance schedule is inactive")
	ErrAlreadyCheckedIn  = errors.New("attendance already checked in")
	ErrNotCheckedIn      = errors.New("attendance not checked in")
	ErrAlreadyCheckedOut = errors.New("attendance already checked out")
	ErrLocationNotFound  = errors.New("attendance location assignment not found")
	ErrInactiveLocation  = errors.New("attendance office location is inactive")
	ErrOutsideGeofence   = errors.New("attendance location is outside geofence")
	ErrPoorAccuracy      = errors.New("attendance location accuracy is too low")
	ErrInternal          = errors.New("attendance operation failed")
)
