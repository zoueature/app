package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ConvertToApiContext(c *gin.Context) *ApiContext {
	return &ApiContext{c}
}

type ApiContext struct {
	*gin.Context
}

const ApiStatusOK = 0

func (c *ApiContext) ResponseJson(code int, msg string, data any) {

	res := &ApiResponse{
		Code:    code,
		Message: msg,
		Data:    gin.H{},
	}
	if data != nil {
		res.Data = data
	}
	c.JSON(http.StatusOK, res)
}

func (c *ApiContext) SuccessData(data ...any) {
	var resp interface{} = gin.H{}
	if len(data) > 0 && data[0] != nil {
		resp = data[0]
	}
	c.ResponseJson(ApiStatusOK, "OK", resp)
}

func (c *ApiContext) Success() {
	c.ResponseJson(ApiStatusOK, "OK", gin.H{})

}
