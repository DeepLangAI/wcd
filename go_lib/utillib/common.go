package utillib

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/redis/go-redis/v9"
	"runtime"
	"sort"
)

func JsonMarshal(ctx context.Context, in interface{}) string {
	res, err := sonic.Marshal(in)
	if err != nil {
		hlog.CtxErrorf(ctx, "marshal %v error %v", in, err)
		return ""
	}
	if string(res) == "null" {
		return ""
	}
	return string(res)
}

func DeduplicateString(strList []string) []string {
	sortMap := make(map[string]int)

	set := make(map[string]struct{})
	for idx, str := range strList {
		if _, ok := set[str]; !ok {
			sortMap[str] = idx
		}
		set[str] = struct{}{}
	}

	res := make([]string, 0, len(set))
	for str := range set {
		res = append(res, str)
	}

	sort.SliceStable(res, func(i, j int) bool {
		return sortMap[res[i]] < sortMap[res[j]]
	})
	return res
}

func GetCurrentFilePath() string {
	_, filePath, _, _ := runtime.Caller(1)
	return filePath
}

type GetRdbFunc func() *redis.ClusterClient
