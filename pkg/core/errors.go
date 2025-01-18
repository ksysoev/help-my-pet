package core

import "errors"

// ErrRateLimit is returned when the API rate limit is exceeded
var ErrRateLimit = errors.New("rate limit exceeded")
