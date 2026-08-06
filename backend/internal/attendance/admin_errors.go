package attendance

import "errors"

var (
	ErrScheduleDuplicate     = errors.New("work schedule name already exists")
	ErrScheduleInUse         = errors.New("work schedule has active or future assignment")
	ErrAssignmentNotFound    = errors.New("schedule assignment not found")
	ErrAssignmentOverlap     = errors.New("schedule assignment overlaps another assignment")
	ErrAssignmentInvalidUser = errors.New("schedule assignment user is invalid")
)
