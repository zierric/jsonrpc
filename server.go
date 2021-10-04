package jsonrpc

import (
	"encoding/json"
	"github.com/fasthttp/router"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type (
	ServerHandler func(*fasthttp.RequestCtx, interface{}) (interface{}, error)

	Server struct {
		whitelist []string
		handlers  map[string]ServerHandler

		mux  sync.Mutex
		http *fasthttp.Server
	}
)

func NewServer(whitelist []string) *Server {
	logrus.Debug("[jsonrpc] new server")

	return &Server{
		whitelist: whitelist,
		handlers:  map[string]ServerHandler{},
	}
}

func (rpc *Server) AddHandler(method string, callback ServerHandler) {
	logrus.Debug("[jsonrpc] add method ", method)

	rpc.handlers[method] = callback
}

func (rpc *Server) Listen(port uint64) {
	logrus.Debug("[jsonrpc] http listen preparing")

	route := router.New()
	route.NotFound = rpc.notFoundHandler
	route.MethodNotAllowed = rpc.notAllowedHandler
	route.PanicHandler = rpc.panicHandler
	route.Handle("POST", "/", rpc.rootHandler)

	handler := rpc.middleware(route.Handler)
	handler = fasthttp.TimeoutHandler(handler, 60*time.Second, http.StatusText(fasthttp.StatusRequestTimeout))
	handler = fasthttp.CompressHandlerLevel(handler, fasthttp.CompressDefaultCompression)

	rpc.mux.Lock()
	rpc.http = &fasthttp.Server{
		Handler:                            handler,
		Concurrency:                        1024,
		ReadTimeout:                        60 * time.Second,
		WriteTimeout:                       60 * time.Second,
		IdleTimeout:                        10 * time.Second,
		MaxRequestBodySize:                 8 * 1024 * 1024,
		LogAllErrors:                       true,
		SleepWhenConcurrencyLimitsExceeded: 30 * time.Second,
		NoDefaultServerHeader:              true,
		NoDefaultDate:                      true,
		CloseOnShutdown:                    true,
		Logger:                             logrus.StandardLogger(),
	}
	rpc.mux.Unlock()

	go func() {
		logrus.Debug("[jsonrpc] http listen on ", port)
		if err := rpc.http.ListenAndServe(":" + strconv.Itoa(int(port))); err != nil {
			logrus.Fatal(err)
		}
	}()
}

func (rpc *Server) Shutdown() {
	logrus.Debug("[jsonrpc] shutdown")

	rpc.mux.Lock()
	if err := rpc.http.Shutdown(); err != nil {
		logrus.Fatal(err)
	}
	rpc.mux.Unlock()
}

func (rpc *Server) rootHandler(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		rpc.sendError(ctx, errors.New(http.StatusText(http.StatusBadRequest)))
		return
	}

	req := &Request{}
	if err := json.Unmarshal(ctx.Request.Body(), &req); err != nil {
		rpc.sendError(ctx, err)
		return
	}

	if req.JsonRPC != "2.0" {
		rpc.sendError(ctx, errors.New("Invalid 'jsonrpc' parameter"))
		return
	}

	if req.Method == "" {
		rpc.sendError(ctx, errors.New("Empty 'method' parameter"))
		return
	}

	if _, ok := rpc.handlers[req.Method]; !ok {
		rpc.sendError(ctx, errors.New("Invalid 'method' parameter"))
		return
	}

	logrus.WithField("params", req.Params).Debug("[jsonrpc] call method " + req.Method)

	res, err := rpc.handlers[req.Method](ctx, req.Params)
	if err != nil {
		rpc.sendError(ctx, err)
		return
	}

	rpc.sendResponse(ctx, http.StatusOK, &Response{
		Result: res,
	})
}

func (rpc *Server) notFoundHandler(ctx *fasthttp.RequestCtx) {
	rpc.sendError(ctx, errors.New(http.StatusText(http.StatusNotFound)))
}

func (rpc *Server) notAllowedHandler(ctx *fasthttp.RequestCtx) {
	rpc.sendError(ctx, errors.New(http.StatusText(http.StatusMethodNotAllowed)))
}

func (rpc *Server) panicHandler(ctx *fasthttp.RequestCtx, e interface{}) {
	rpc.sendError(ctx, errors.New(http.StatusText(http.StatusInternalServerError)))

	logrus.WithField("error", e.(error).Error()).Error("[jsonrpc] panic error")
}

func (rpc *Server) sendResponse(ctx *fasthttp.RequestCtx, status uint64, req *Response) {
	logrus.Debug("[jsonrpc] send response")

	req.JsonRPC = "2.0"
	req.Id = 0

	b, err := json.Marshal(req)
	if err != nil {
		rpc.sendError(ctx, errors.WithStack(err))
		return
	}

	ctx.SetStatusCode(int(status))
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetBody(b)
}

func (rpc *Server) sendError(ctx *fasthttp.RequestCtx, err error) {
	logrus.Error(errors.WithStack(err))

	rpc.sendResponse(ctx, http.StatusInternalServerError, &Response{
		Error: err.Error(),
	})
}

func (rpc *Server) middleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	logrus.Debug("[jsonrpc] middleware")

	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json; charset=utf-8")

		if !contains(rpc.whitelist, ctx.RemoteIP().String()) {
			rpc.sendError(ctx, errors.New(http.StatusText(http.StatusForbidden)))
			return
		}

		next(ctx)
	}
}

func contains(s []string, v string) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}
