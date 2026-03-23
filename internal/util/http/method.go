package http

import (
	"net/http"
	"strings"
)

func IsValid(method string) bool {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodGet:
	case http.MethodHead:
	case http.MethodPut:
	case http.MethodPost:
	case http.MethodDelete:
	case http.MethodConnect:
	case http.MethodTrace:
	case http.MethodOptions:
	case http.MethodPatch:
	default:
		return false
	}
	return true
}
