package utillib

import (
	"context"
	constslib "github.com/DeepLangAI/go_lib/consts"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func TranslateJsonIO(ctx context.Context, data string) string {
	if data == "" {
		return ""
	}
	var jsonData any
	if err := sonic.Unmarshal([]byte(data), &jsonData); err != nil {
		hlog.CtxErrorf(ctx, "Error parsing JSON: %v", err)
		if len(data) > constslib.DataTooLongUpper {
			return constslib.DataTooLongUpperErrMsg
		}
		return data
	}
	// 处理 JSON
	processedData := processJSON(jsonData)

	// 转换回 JSON 字符串
	result, err := sonic.Marshal(processedData)
	if err != nil {
		hlog.CtxErrorf(ctx, "Error converting JSON back to string: %v", err)
		if len(data) > constslib.DataTooLongUpper {
			return constslib.DataTooLongUpperErrMsg
		}
		return data
	}
	return string(result)
}

// 递归处理 JSON 数据
func processJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			v[key] = processJSON(value)
		}
	case []interface{}:
		for i, value := range v {
			v[i] = processJSON(value)
		}
	case string:
		if len(v) > constslib.DataTooLongUpper {
			return constslib.DataTooLongUpperErrMsg
		}
	}
	return data
}
