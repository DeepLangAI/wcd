package middleware

import (
	"context"
	constslib "github.com/DeepLangAI/go_lib/consts"
	"github.com/DeepLangAI/go_lib/utillib"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/logger/zap"
)

// TraceServerMiddleware 链路信息透传中间件，配合TraceClientMiddleware使用
func TraceServerMiddleware() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		traceId := string(ctx.GetHeader(constslib.HttpHeaderTraceId))
		if traceId == "" {
			traceId = utillib.GenerateTraceId()
		}
		c = context.WithValue(c, zap.ExtraKey(constslib.TraceIdKey), traceId)
		c = context.WithValue(c, constslib.TraceIdKey, traceId)

		c = context.WithValue(c, zap.ExtraKey(constslib.OperationIdKey), primitive.NewObjectID().Hex())
		ctx.Next(c)
		ctx.Response.Header.Set(constslib.HttpHeaderTraceId, traceId)
	}
}

func ResponseVersionServerMiddleware(version string) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.Next(c)
		ctx.Response.Header.Set(constslib.HttpHeaderVersion, version)
	}
}

type BaseResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}
type RequestInfo struct {
	Route  string `json:"route"`
	Params string `json:"params"`
}

func ReqRespLogMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 记录请求时间
		beginAt := time.Now()
		method := string(c.Method())
		req := &RequestInfo{
			Route: c.FullPath(),
		}
		if method == "GET" {
			params := c.Request.URI().QueryString()
			req.Params = string(params)
		}
		if method == "POST" {
			body := c.Request.Body()
			req.Params = utillib.TranslateJsonIO(ctx, string(body))
		}
		// 打印请求信息
		hlog.CtxInfof(ctx, "Request route: %s, Method: %s, RequestBody: %+v", req.Route, method, req.Params)
		// 执行请求处理程序和其他中间件函数
		c.Next(ctx)
		hlog.CtxInfof(ctx, "Response route: %v, http_code: %v, cost: %.4f s", req.Route, c.Response.StatusCode(), time.Since(beginAt).Seconds())
	}
}
