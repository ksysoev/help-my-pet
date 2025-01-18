package core

import "errors"

// ErrRateLimit is returned when the API rate limit is exceeded
var ErrRateLimit = errors.New("rate limit exceeded")

// ErrGlobalLimit is returned when the global daily request limit is exceeded
var ErrGlobalLimit = errors.New("global request limit exceeded for today, please try again tomorrow")
