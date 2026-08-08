package face

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid face input")
	ErrForbidden            = errors.New("face access forbidden")
	ErrInactiveAccount      = errors.New("face account inactive")
	ErrProfileNotFound      = errors.New("face profile not found")
	ErrNotEnrolled          = errors.New("face profile not enrolled")
	ErrAlreadyEnrolled      = errors.New("face profile already enrolled")
	ErrDuplicateBiometric   = errors.New("face biometric already belongs to another user")
	ErrUnsupportedModel     = errors.New("face embedding model unsupported")
	ErrInvalidDimension     = errors.New("face embedding dimension invalid")
	ErrVerificationMismatch = errors.New("face verification mismatch")
	ErrRepositoryFailure    = errors.New("face repository operation failed")
)
