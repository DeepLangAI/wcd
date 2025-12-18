package middleware

import (
	"context"
	"fmt"
	constslib "github.com/DeepLangAI/go_lib/consts"
	"github.com/DeepLangAI/go_lib/utillib"
	hertz_client "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
)

// TraceClientMiddleware 链路信息透传中间件
func TraceClientMiddleware(endpoint hertz_client.Endpoint) hertz_client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		if traceId, ok := ctx.Value(constslib.TraceIdKey).(string); ok && traceId != "" {
			req.Header.Set(constslib.HttpHeaderTraceId, traceId)
		}

		return endpoint(ctx, req, resp)
	}
}

// ResponseCheckClientMiddleware 下游响应status_code校验
func ResponseCheckClientMiddleware(endpoint hertz_client.Endpoint) hertz_client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {
		err = endpoint(ctx, req, resp)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			respBody := ""
			if !resp.IsBodyStream() {
				respBody = string(resp.Body())
			}
			return utillib.NewErr(constslib.ErrCode_REMOTE_ERROR, fmt.Sprintf("status_code: %v msg %v", resp.StatusCode(), respBody))
		}
		return nil
	}
}
