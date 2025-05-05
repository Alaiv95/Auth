package auth

import "errors"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserExists = errors.New("user already exists")
