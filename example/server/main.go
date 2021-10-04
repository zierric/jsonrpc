package main

import (
	"errors"
	"github.com/valyala/fasthttp"
	"github.com/zierric/jsonrpc"
	"time"
)

func main() {
	rpc := jsonrpc.NewServer([]string{
		"127.0.0.1",
	})

	rpc.AddHandler("test", func(ctx *fasthttp.RequestCtx, params interface{}) (interface{}, error) {
		return map[string]interface{}{
			"test":         []string{"ok", "slice", "string"},
			"input_params": params,
		}, nil
	})

	rpc.AddHandler("test.error", func(ctx *fasthttp.RequestCtx, params interface{}) (interface{}, error) {
		return nil, errors.New("test error message")
	})

	rpc.Listen(8080)

	time.Sleep(60 * time.Second)

	rpc.Shutdown()
}
