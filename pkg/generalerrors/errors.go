package generalerrors

import "errors"

var (
	ErrNilPointerInInterface = errors.New("nil pointer in interface")

	ErrWrongPassword         = errors.New("wrong password")
	ErrUsernameAlreadyExists = errors.New("this username already exists")
	ErrUserNotFound          = errors.New("user not found")

	ErrRefreshInBlackList      = errors.New("token found in blacklist")
	ErrToManyUseOfRefreshToken = errors.New("to many uses of refresh token")

	ErrAliasAlreadyExists = errors.New("alias already exists")
	ErrAliasNotFound      = errors.New("alias not found")

	ErrCacheMiss = errors.New("not found in cache")
)
