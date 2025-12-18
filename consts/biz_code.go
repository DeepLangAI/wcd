package consts

type BizCode struct {
	// 错误码
	Code int32
	// 错误描述
	Msg string
}

var (
	ResSuccess      = BizCode{0, "success"}
	ParseWorthless  = BizCode{0, "解析失败，文章内容无意义"}
	CrawlFailed     = BizCode{0, "抓取失败"}
	TextParseFailed = BizCode{10204, "text-parse解析失败"}

	ReqParamError = BizCode{10400, "参数错误"}

	SystemErr      = BizCode{10500, "服务繁忙，请稍后重试"}
	QueryDataError = BizCode{10501, "数据查询异常"}
	WriteDbError   = BizCode{10502, "数据写入异常"}
)
