package main

import (
	"encoding/json"
)

// ResponseIface represents interface to work with
// response data
type ResponseIface interface {
	Get() ([]byte, error)
	Ok() bool
	Status() int
}

// Response is structured response
type Response struct {
	Success bool        `json:"success"`
	Count   uint64      `json:"count"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error is a trivial implementation of error with
// code, message and title to use in the response
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Title   string `json:"title"`
}

// NewResponse returns Response given a data
func NewResponse(send interface{}) (response *Response) {
	response = &Response{
		Success: true,
	}

	if send == nil {
		return
	}

	switch send.(type) {
	case error:
		response.Success = false
		response.Error = &Error{
			Message: send.(error).Error(),
			Title:   "Server error",
		}

	case *Error:
		response.Success = false
		response.Error = send.(*Error)

	case string:
		response.Data = send.(string)

	default:
		response.Data = send
	}

	if response.Error != nil && response.Error.Code == 0 {
		response.Error.Code = 500
	}

	return
}

// Get returns encoded data ready to send to client
func (s *Response) Get() (data []byte, err error) {
	return json.Marshal(s)
}

// Ok returns response state, if Success field is false
func (s *Response) Ok() bool {
	return s.Success
}

// Status returns response code, default 200
func (s *Response) Status() int {
	if s.Error != nil {
		return s.Error.Code
	}

	return 200
}
