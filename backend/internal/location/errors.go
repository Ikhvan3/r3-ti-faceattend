package location

import "errors"

var (
	ErrInvalidInput       = errors.New("location input is invalid")
	ErrOfficeNotFound     = errors.New("office location not found")
	ErrOfficeInUse        = errors.New("office location has active or future assignment")
	ErrInactiveOffice     = errors.New("office location is inactive")
	ErrAssignmentNotFound = errors.New("location assignment not found")
	ErrAssignmentOverlap  = errors.New("location assignment overlaps another assignment")
	ErrInvalidUser        = errors.New("location assignment user is invalid")
	ErrInternal           = errors.New("location operation failed")
)
