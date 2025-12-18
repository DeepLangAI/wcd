package utillib

import (
	"context"
	constslib "github.com/DeepLangAI/go_lib/consts"
	hertzzap "github.com/hertz-contrib/logger/zap"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewCtxWithTraceId(ctx context.Context) context.Context {
	traceId := GenerateTraceId()
	ctx = context.WithValue(ctx, constslib.TraceIdKey, traceId)
	ctx = context.WithValue(ctx, hertzzap.ExtraKey(constslib.TraceIdKey), traceId)
	return ctx
}

func GetCtxTraceId(ctx context.Context) string {
	if traceId, ok := ctx.Value(constslib.TraceIdKey).(string); ok && traceId != "" {
		return traceId
	}
	return ""
}

func GenerateTraceId() string {
	return primitive.NewObjectID().Hex()
}
