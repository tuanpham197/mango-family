package common

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type successRes struct {
	Data   any `json:"data"`
	Paging any `json:"paging,omitempty"`
}

func Response(data any) successRes { return successRes{Data: data} }

func PagedResponse(data any, paging Paging) successRes {
	return successRes{Data: data, Paging: paging}
}

type errorBody struct {
	Error map[string]any `json:"error"`
}

// WriteError chuẩn hóa mọi lỗi về format {error:{code,message,field?,...extra}}.
func WriteError(c *gin.Context, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = NewInternal(err)
	}
	body := map[string]any{"code": appErr.Code, "message": appErr.Message}
	if appErr.Field != "" {
		body["field"] = appErr.Field
	}
	for k, v := range appErr.Extra {
		body[k] = v
	}
	c.AbortWithStatusJSON(appErr.StatusCode, errorBody{Error: body})
}

func WriteData(c *gin.Context, status int, data any) {
	c.JSON(status, Response(data))
}

func WriteOK(c *gin.Context, data any) { WriteData(c, http.StatusOK, data) }
