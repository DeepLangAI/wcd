package utils

import (
	"context"
	"fmt"
	constslib "github.com/DeepLangAI/go_lib/consts"
	"github.com/DeepLangAI/go_lib/utillib"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/logger/zap"
)

func GetCtxOperationId(ctx context.Context) string {
	key := zap.ExtraKey(constslib.OperationIdKey)
	if value, ok := ctx.Value(key).(string); ok && value != "" {
		return value
	}
	return ""
}

func PrintRequestLog(ctx context.Context, action, key, actionType string, info interface{}) {
	var infoStr string
	if _, ok := info.(string); ok {
		infoStr = info.(string)
	} else {
		infoStr = utillib.JsonMarshal(ctx, info)
	}
	infoStr = utillib.TranslateJsonIO(ctx, infoStr)

	if len(infoStr) > constslib.DataTooLongUpper {
		infoStr, _ = utillib.StrGzip(infoStr)
	}

	if len(infoStr) > 300000 {
		infoStr = fmt.Sprintf("data too long. len: %d", len(infoStr))
	}
	hlog.CtxInfof(ctx, "OutRequest %s, %s, %s:%v", action, key, actionType, infoStr)
}
