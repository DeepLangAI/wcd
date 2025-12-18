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
	for _, pattern := range consts.DATETIME_PATTERN {
		regex := regexp.MustCompile(pattern)
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
	//text := t.Doc.GetRawDocText(t.Doc.Doc.Root())
	text := t.Doc.GetRawHtmlStr()
	for _, pattern := range consts.DATETIME_SUBMATCH_PATTERN {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindStringSubmatch(text); len(matches) >= 1 {
			return matches[1], nil
		}
	}
	for _, pattern := range consts.DATETIME_PATTERN {
		regex := regexp.MustCompile(pattern)
		if match := regex.FindString(text); match != "" {
			return match, nil
		}
	}
	return "", nil
}
