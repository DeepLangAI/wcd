package utils

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func DumpAndLoad(ctx context.Context, from, to any) {
	marshal, err := sonic.Marshal(from)
	if err != nil {
		hlog.CtxErrorf(ctx, "sonic marshal failed: %v", err)
	}
	err = sonic.Unmarshal(marshal, to)
	if err != nil {
		hlog.CtxErrorf(ctx, "sonic unmarshal failed: %v", err)
	}
}
