package app

import (
	"context"
	"encoding/json"
	"errors"
	"gitlab.jiebu.com/base/log"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

// ------------------------------------------------------
// http rpc 客户端
// ------------------------------------------------------

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
	url := cli.Host + req.URL
	param := req.Param.Encode()

	var body io.Reader
	if req.Method == http.MethodGet {
		url += "?" + param
	} else {
		body = strings.NewReader(param)
	}
	log.Debug(ctx, "call rpc ", zap.String("url", url), zap.Any("params", param))
	request, err := http.NewRequestWithContext(ctx, req.Method, url, body)
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

// ------------------------------------------------------
// grpc rpc 客户端
// ------------------------------------------------------
