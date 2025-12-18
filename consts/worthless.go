package consts

import (
	"regexp"
	"strings"
)

const (
	MUST_WORTHLESS_TXT_LEN     = 15
	WORTHLESS_TXT_LEN          = 100 // if too large, page will be treated as nonce by mistake
	MULTIPLE_WORDS_MATCH_RATIO = float32(0.75)
)

var WORTHLESS_PAGE_TITLES = []string{
	"账号已迁移", "page not found", "Captcha Interception", "安全验证", "404页面", "404 Not Found", "403 Forbidden",
	"页面没有找到", "找不到我要找的页面",
}

var WORTHLESS_PAGE_TITLE_REGEX = regexp.MustCompile(
	strings.Join(WORTHLESS_PAGE_TITLES, "|"),
)

var WORTHLESS_PAGE_KEYWORDS = [][]string{
	{"已删除", "不存在", "页面", "跳回"},
	{"节点", "域名", "存在"},
	{"端口", "域名", "绑定"},
	{"nginx", "404", "403"},
	{"请求", "非法"},
	{"网站", "无法", "访问"},
	{"content", "deleted", "author"},
	{"No", "page", "found", "not"},
	{"内容", "发布者", "删除", "作者"},
	{"抱歉", "页面", "没", "找到"},
	{"抱歉", "网页", "出错"},
	{"自动", "跳转", "首页"},
	{"内容", "违规", "无法", "查看"},
	{"链接", "不", "访问"},
	{"公众号", "已迁移"},
	{"网络", "稍后", "重试"},
	{"文章", "找不到"},
	{"检查", "网址", "是否正确"},
	{"输入", "错误", "重新"},
	// {"暂", "无"},
}
