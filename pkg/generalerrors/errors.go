package generalerrors

import "errors"

var (
	ErrNilPointerInInterface = errors.New("nil pointer in interface")

	ErrWrongPassword = errors.New("wrong password")

	ErrUserNotFound = errors.New("user not found")

	ErrRefreshInBlackList = errors.New("token found in blacklist")
)
