package extractor

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// 预编译：原实现每处理一篇文章都对着全量 consts 列表 MustCompile 一遍（33 条正则 ×
// 每次请求），编译开销与正文长度无关、纯属重复劳动。改成包级一次编译。
var (
	datetimeSubmatchRegexps = compileAllRegexps(consts.DATETIME_SUBMATCH_PATTERN)
	datetimeRegexps         = compileAllRegexps(consts.DATETIME_PATTERN)
)

func compileAllRegexps(patterns []string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		out = append(out, regexp.MustCompile(p))
	}
	return out
}

type PublishTimeExtractor struct {
	Doc    *doc.Document
	result string
	ctx    context.Context
}

func NewPublishTimeExtractor(ctx context.Context, doc *doc.Document) *PublishTimeExtractor {
	return &PublishTimeExtractor{
		Doc: doc,
		ctx: ctx,
	}
}

func (t *PublishTimeExtractor) tryParseAndNorm(dateTime string) (string, error) {
	for _, pattern := range consts.DATETIME_FORMATS {
		parse, err := time.Parse(pattern, dateTime)
		if err == nil {
			formatted := parse.Format(consts.NormalizedDateTime)
			formatted = strings.TrimSuffix(formatted, ":00")
			formatted = strings.TrimSuffix(formatted, "00:00")
			formatted = strings.TrimSpace(formatted)
			return formatted, nil
		}
	}
	return "", errors.New("failed to parse and norm datetime")
}

func (t *PublishTimeExtractor) Norm(dateTime string) string {
	if formatted, err := t.tryParseAndNorm(dateTime); err == nil {
		return formatted
	}

	hlog.CtxInfof(t.ctx, "failed to parse and norm datetime, try to search regex: %s", dateTime)
	// 无法解析，尝试搜索正则时间
	for i, regex := range datetimeRegexps {
		pattern := consts.DATETIME_PATTERN[i]
		if match := regex.FindString(dateTime); match != "" {
			hlog.CtxInfof(t.ctx, "found regex datetime, pattern: %v, match: %v", pattern, match)
			if formatted, err := t.tryParseAndNorm(match); err == nil {
				return formatted
			}
		}
	}

	return dateTime
}

func (t *PublishTimeExtractor) Extract() (string, error) {
	if rule := t.Doc.Rule; rule != nil {
		if rule.PubTime == consts.EmptyExtractXpath {
			return "", nil
		}
	}

	var result string
	var err error

	nodes := []func() (string, error){
		t.bySiteRule, t.byMeta, t.byRegex,
	}
	for _, node := range nodes {
		result, err = node()
		if err != nil {
			return "", err
		}
		result = strings.TrimSpace(result)
		if result != "" && t.isTimestamp(result) {
			result = t.timestampToString(result)
		}
		if result != "" {
			return t.Norm(result), nil
		}
	}
	return t.Norm(result), nil
}

func (t *PublishTimeExtractor) isTimestamp(text string) bool {
	// 检查文本是否为数字且长度大于等于8
	if len(text) >= 8 {
		for _, char := range text {
			if char < '0' || char > '9' {
				return false
			}
		}
		return true
	}
	return false
}

func (t *PublishTimeExtractor) timestampToString(text string) string {
	timestamp, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return text
	}
	switch len(text) {
	case 10:
		// 秒级
	case 13:
		// 毫秒级
		timestamp = timestamp / 1000
	default:

	}
	// 转换为时间字符串
	//loc, _ := time.LoadLocation("Asia/Shanghai")
	loc := time.Local
	dt := time.Unix(timestamp, 0).In(loc)
	return dt.Format(time.DateTime)
}

func (t *PublishTimeExtractor) bySiteRule() (string, error) {
	xpath := ""
	if t.Doc.Rule != nil {
		xpath = t.Doc.Rule.PubTime
	}
	if xpath == "" {
		return "", nil
	}
	result, err := ExtractByXpath(xpath, t.Doc)
	if err != nil {
		hlog.CtxErrorf(t.ctx, "extract publish time by site rule failed: %v", err)
		return "", err
	}
	return result, err

}
func (t *PublishTimeExtractor) byMeta() (string, error) {
	for _, xpath := range consts.PUBLISH_TIME_META {
		elems := t.Doc.Xpath(xpath)
		for _, elem := range elems {
			if value := elem.SelectAttrValue("content", ""); value != "" {
				if value != "" {
					// 如果是时间戳，则转换为时间字符串
					if t.isTimestamp(value) {
						value = t.timestampToString(value)
					}
					return value, nil
				}
			}
		}
	}
	return "", nil
}

func (t *PublishTimeExtractor) byRegex() (string, error) {
	rawHtml := t.Doc.GetRawHtmlStr()
	// 时间类正则扫全量 raw HTML 是本函数的主要开销（33 条正则 × 整篇 HTML）。
	// 先扫前 128KB（实测绝大多数页面的发布时间都在这个窗口内），只有窗口内找不到
	// 能解析成时间的候选时才回退扫全文——命中率不变，长页少扫 2/3 以上。
	limit := 128 * 1024
	if match := t.searchDatetime(rawHtml, limit); match != "" {
		return match, nil
	}
	if limit > 0 && len(rawHtml) > limit {
		if match := t.searchDatetime(rawHtml, 0); match != "" {
			return match, nil
		}
	}
	return "", nil
}

// plausibleDatetime 是廉价预筛：发布日期一定带 19xx/20xx 的年份，
// 版本号（91.0.1）、区间数字（7125-52.40）这类命中直接跳过，省掉 47 种格式的 time.Parse。
func plausibleDatetime(s string) bool {
	for i := 0; i+3 < len(s); i++ {
		if (s[i] == '1' && s[i+1] == '9') || (s[i] == '2' && s[i+1] == '0') {
			if s[i+2] >= '0' && s[i+2] <= '9' && s[i+3] >= '0' && s[i+3] <= '9' {
				return true
			}
		}
	}
	return false
}

// searchDatetime 在 rawHtml 的前 limit 字节（0=全文）里按既有优先级找发布时间。
// 每个候选先用 tryParseAndNorm 校验，避免把版本号（"91.0.1"）、区间数字（"7125-52.40"）
// 这类文本当成时间返回。
func (t *PublishTimeExtractor) searchDatetime(rawHtml string, limit int) string {
	text := rawHtml
	if limit > 0 && len(text) > limit {
		text = text[:limit]
	}
	for _, regex := range datetimeSubmatchRegexps {
		if matches := regex.FindStringSubmatch(text); len(matches) >= 2 {
			if _, err := t.tryParseAndNorm(matches[1]); err == nil {
				return matches[1]
			}
		}
	}
	for _, regex := range datetimeRegexps {
		match := regex.FindString(text)
		if match == "" || !plausibleDatetime(match) {
			continue
		}
		if _, err := t.tryParseAndNorm(match); err == nil {
			return match
		}
	}
	return ""
}
