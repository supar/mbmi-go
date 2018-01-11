package main

import (
	"encoding/json"
)

type ResponseIface interface {
	Get() ([]byte, error)
	Ok() bool
	Status() int
}

// Default response structure
type Response struct {
	Success bool        `json:"success"`
	Count   uint64      `json:"count"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Common error structure in response
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Title   string `json:"title"`
}

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

//func NewResponseOk(v interface{}) *Response {
//	return &Response{
//		Data:    v,
//		Success: true,
//	}
//}
//
//func NewResponseError(code int, err interface{}) *Response {
//	var msg string
//
//	if code == 0 {
//		code = 500
//	}
//
//	if c := http.StatusText(code); c != "" {
//		msg = c
//	}
//
//	if err != nil {
//		switch err.(type) {
//		case error:
//			msg = err.(error).Error()
//
//		case string:
//			msg = err.(string)
//		}
//	}
//
//	return &Response{
//		Success: false,
//		Error: &Error{
//			Code:    code,
//			Message: msg,
//		},
//	}
//}

func (s *Response) Get() (data []byte, err error) {
	return json.Marshal(s)
}

func (s *Response) Ok() bool {
	return s.Success
}

func (s *Response) Status() int {
	if s.Error != nil {
		return s.Error.Code
	}

	return 200
}
