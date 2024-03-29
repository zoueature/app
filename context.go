package app

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"net/http"
	"time"
)

type BaseApiResponse struct {
	Code      int    `json:"code"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

type ApiResponse struct {
	BaseApiResponse
	Data interface{} `json:"data"`
}

// ConvertToApiContext 将gin上下文转化为api上下文
func ConvertToApiContext(c *gin.Context) *ApiContext {
	return &ApiContext{c}
}

type ApiContext struct {
	*gin.Context
}

const userIDKeyInContext = "user_id"

func (c *ApiContext) SetUserID(userId int64) {
	c.Set(userIDKeyInContext, userId)
}

func (c *ApiContext) AuthUserID() int64 {
	v, ok := c.Get(userIDKeyInContext)
	if !ok {
		return 0
	}
	return cast.ToInt64(v)
}

func (c *ApiContext) RequestURI() string {
	return c.Request.RequestURI
}

func (c *ApiContext) MustGetAuthUserID() int64 {
	v, ok := c.Get(userIDKeyInContext)
	if !ok {
		panic("auth user id is empty")
	}
	return cast.ToInt64(v)
}

// ResponseJson 响应json数据
func (c *ApiContext) ResponseJson(code int, msg string, data any) {

	res := &ApiResponse{
		BaseApiResponse: BaseApiResponse{
			Code:      code,
			Message:   msg,
			Timestamp: time.Now().Unix(),
		},
		Data: gin.H{},
	}
	if data != nil {
		res.Data = data
	}
	c.JSON(http.StatusOK, res)
}

// SuccessData 成功返回并响应对应数据
func (c *ApiContext) SuccessData(data ...any) {
	var resp interface{} = gin.H{}
	if len(data) > 0 && data[0] != nil {
		resp = data[0]
	}
	c.ResponseJson(ApiStatusOK.Code(), ApiStatusOK.Error(), resp)
}

// Success 返回成功
func (c *ApiContext) Success() {
	c.ResponseError(ApiStatusOK)

}

// ResponseErrorCode 响应错误码
func (c *ApiContext) ResponseErrorCode(code Errcode, msg ...string) {
	respMsg := code.Error()
	if len(msg) > 0 && msg[0] != "" {
		respMsg = msg[0]
	}
	c.ResponseJson(code.Code(), respMsg, nil)
}

// ResponseError 返回err
func (c *ApiContext) ResponseError(err error) {
	ec, ok := err.(Errcode)
	code := ErrServer.Code()
	if ok {
		code = ec.Code()
	}
	c.ResponseJson(code, err.Error(), nil)
}
