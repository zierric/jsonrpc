package jsonrpc

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type (
	Client struct {
		endpoint string
		timeout  uint64
		resp     *Response
	}
)

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		timeout:  10,
	}
}

func (rpc *Client) SetTimeout(sec uint64) {
	rpc.timeout = sec
}

func (rpc *Client) Call(method string, params ...interface{}) (*Response, error) {
	if strings.TrimSpace(method) == "" {
		return nil, errors.WithStack(errors.New("Empty 'method' parameter"))
	}

	r := &Request{
		JsonRPC: "2.0",
		Method:  method,
	}

	if len(params) == 1 {
		r.Params = params[0]
	}

	b, err := json.Marshal(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req, err := http.NewRequest("POST", rpc.endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	rsp, err := (&http.Client{
		Timeout: time.Duration(rpc.timeout) * time.Second,
	}).Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := rsp.Body.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	rpc.resp = &Response{}
	if err := json.Unmarshal(body, &rpc.resp); err != nil {
		return nil, errors.WithStack(err)
	}

	return rpc.resp, nil
}

func (rpc *Client) ToObject(obj interface{}) {

}
