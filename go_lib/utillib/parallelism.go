package utillib

import (
	"context"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"runtime"
	"sync"
)

type AsyncFunc func() error

func AsyncExec(ctx context.Context, f AsyncFunc) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				hlog.CtxErrorf(ctx, "AsyncExec [panic] err: %v\nstack: %s\n", err, string(buf[:n]))
			}
		}()
		err := f()
		if err != nil {
			hlog.CtxErrorf(ctx, "AsyncExec error %v", err)
		}
	}()
}

func ParallelExec(ctx context.Context, funcList []AsyncFunc, workerNum int) []error {
	if workerNum > len(funcList) {
		workerNum = len(funcList)
	}
	wg := &sync.WaitGroup{}
	wg.Add(workerNum)

	funcCh := make(chan AsyncFunc, len(funcList))
	errCh := make(chan error, len(funcList))
	for i := 0; i < workerNum; i++ {
		AsyncExec(ctx, func() error {
			defer func() {
				wg.Done()
				if err := recover(); err != nil {
					var buf [4096]byte
					n := runtime.Stack(buf[:], false)
					hlog.CtxErrorf(context.Background(), "AsyncExec [panic] err: %v\nstack: %s\n", err, string(buf[:n]))
				}
			}()
			for f := range funcCh {
				if err := f(); err != nil {
					errCh <- err
				}
			}
			return nil
		})
	}

	for _, f := range funcList {
		funcCh <- f
	}
	close(funcCh)
	wg.Wait()

	close(errCh)
	errList := make([]error, 0)
	for err := range errCh {
		errList = append(errList, err)
	}
	return errList
}
