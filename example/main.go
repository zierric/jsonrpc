package main

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/zierric/jsonrpc"
)

func main() {
	// server
	server := jsonrpc.NewServer([]string{
		"127.0.0.1",
	})

	server.AddHandler("test", func(ctx *fasthttp.RequestCtx, params interface{}) (interface{}, error) {
		return map[string]interface{}{
			"test":         []string{"ok", "slice", "string"},
			"input_params": params,
		}, nil
	})

	server.AddHandler("test.error", func(ctx *fasthttp.RequestCtx, params interface{}) (interface{}, error) {
		return nil, errors.New("test error message")
	})

	server.Listen(8080)

	// clients
	client := jsonrpc.NewClient("http://127.0.0.1:8080/")
	resp, err := client.Call("test", []uint64{1, 2, 3, 8, 9})
	if err != nil {
		logrus.Fatal(err)
	}

	if resp.Error == nil {
		fmt.Println("Result:", resp.Result)
	} else {
		fmt.Println("Error:", resp.Error)
	}
}
