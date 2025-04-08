package clinic

import "errors"

var (
	ErrClinicNotFound     = errors.New("clinic not found")
	ErrInvalidClinicData  = errors.New("invalid clinic data")
	ErrUnauthorizedAccess = errors.New("unauthorized access to clinic")
	ErrImageUploadFailed  = errors.New("failed to upload image")
	ErrInvalidImageFormat = errors.New("invalid image format")
	ErrPaginationInvalid  = errors.New("invalid pagination parameters")
)
