package domain

import "errors"

type LogError struct {
	Message string `json:"message"`
	Err     error  `json:"err"`
	Code    int    `json:"code"`
}

func (l *LogError) Error() string {
	return l.Message
}

var ErrorMetaNotFound = errors.New("meta info not found")
