package ethparser

import "errors"

var (
	ErrFailedToParseJSON          = errors.New("failed to parse json")
	ErrFailedToCreateRequest      = errors.New("failed to create request")
	ErrRequestFailed              = errors.New("request failed")
	ErrResultIsUnexpected         = errors.New("result is unexpected")
	ErrParseToIntFailed           = errors.New("failed to parse to int")
	ErrInvalidAddress             = errors.New("invalid address")
	ErrWriteMessageFailed         = errors.New("failed to write message")
	ErrBadHandshake               = errors.New("bad handshake")
	ErrGenerateChallengeKeyFailed = errors.New("failed to generate challenge key")
	ErrParseURIFailed             = errors.New("failed to parse uri")
	ErrDialFailed                 = errors.New("failed to dial")
)
