package response

import (
	"smart-slowquery/internal/util/errors"

	"net/http"

	"github.com/gin-gonic/gin"
)

func ToResponse(ctx *gin.Context, data interface{}, err error) {
	if err != nil {
		resp := gin.H{"error": NewResponseError(errors.ErrApp.Code, err.Error()), "data": nil}
		ctx.JSON(http.StatusOK, resp)
	} else {
		resp := gin.H{"data": data}
		ctx.JSON(http.StatusOK, resp)
	}
}

func ToAbortWithErrorCodeResponse(ctx *gin.Context, err errors.Error) {
	resp := gin.H{"error": NewResponseError(err.Code, err.Error())}
	ctx.AbortWithStatusJSON(err.Code, resp)
}

func ToAbortErrorResponse(ctx *gin.Context, err errors.Error) {
	resp := gin.H{"error": NewResponseError(err.Code, err.Error())}
	ctx.AbortWithStatusJSON(http.StatusOK, resp)
}

type ResError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewResponseError(code int, message string) *ResError {
	return &ResError{
		Code:    code,
		Message: message,
	}
}

//func statusCode(code int) int {
//	switch c := code; {
//	case c == 0:
//		return http.StatusOK
//	case c == errors.ErrApp.Code:
//		return http.StatusInternalServerError
//	case c == errors.ErrParam.Code:
//		return http.StatusBadRequest
//	}
//	return http.StatusInternalServerError
//}
