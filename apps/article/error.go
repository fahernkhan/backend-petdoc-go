package article

import "errors"

var (
	ErrArticleNotFound    = errors.New("article not found")
	ErrInvalidArticleData = errors.New("invalid article data")
	ErrUnauthorizedAccess = errors.New("unauthorized access to article")
	ErrImageUploadFailed  = errors.New("failed to upload image")
	ErrInvalidImageFormat = errors.New("invalid image format")
	ErrPaginationInvalid  = errors.New("invalid pagination parameters")
)
