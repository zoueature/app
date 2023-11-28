package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
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

func (c *ApiContext) ResponseErrorCode(code Errcode) {
	c.ResponseJson(code.Code(), code.Error(), nil)
}

// ------------------------------------
// http rpc 客户端
// ------------------------------------

type HttpRpcClient struct {
	Host   string
	Client http.Client
}

func NewHttpRpcClient(baseURL string) *HttpRpcClient {
	return &HttpRpcClient{
		Host:   baseURL,
		Client: http.Client{},
	}
}

type RpcRequest struct {
	URL    string
	Method string
	Param  Encoder
}

type Encoder interface {
	Encode() string
}

type JsonReq struct {
	data interface{}
}

func (j JsonReq) Encode() string {
	str, _ := json.Marshal(j.data)
	return string(str)
}

func NewHttpRpcRequest(method, uri string, params Encoder) *RpcRequest {
	return &RpcRequest{
		URL:    uri,
		Method: method,
		Param:  params,
	}
}

func (cli *HttpRpcClient) HttpRemoteCall(ctx context.Context, req *RpcRequest) (*ApiResponse, error) {
	content, err := cli.call(ctx, req)
	if err != nil {
		return nil, err
	}
	data := new(ApiResponse)
	err = json.Unmarshal(content, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (cli *HttpRpcClient) call(ctx context.Context, req *RpcRequest) ([]byte, error) {
	url := req.URL
	param := req.Param.Encode()

	var body io.Reader
	if req.Method == http.MethodGet {
		url += "?" + param
	} else {
		body = strings.NewReader(param)
	}
	request, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Client.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (cli HttpRpcClient) HttpCallWithResp(ctx context.Context, req *RpcRequest, resp interface{}) error {
	content, err := cli.call(ctx, req)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, resp)
	if err != nil {
		return err
	}
	return nil
}
