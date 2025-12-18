package extractor

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type AuthorExtractor struct {
	Doc    *doc.Document
	result string
	ctx    context.Context
}

func NewAuthorExtractor(ctx context.Context, doc *doc.Document) *AuthorExtractor {
	return &AuthorExtractor{
		Doc: doc,
		ctx: ctx,
	}
}

func (t *AuthorExtractor) Extract() (string, error) {
	if rule := t.Doc.Rule; rule != nil {
		if rule.Author == consts.EmptyExtractXpath {
			return "", nil
		}
	}
	var result string
	var err error

	nodes := []func() (string, error){
		t.bySiteRule, t.byMeta, t.byNode, t.byRegex,
	}
	for _, node := range nodes {
		result, err = node()
		if err != nil {
			return "", err
		}
		if result != "" {
			return result, nil
		}
	}
	return result, nil
}
func (t *AuthorExtractor) bySiteRule() (string, error) {
	xpath := ""
	if t.Doc.Rule != nil {
		xpath = t.Doc.Rule.Author
	}
	if xpath == "" {
		return "", nil
	}
	title, err := ExtractByXpath(xpath, t.Doc)
	if err != nil {
		hlog.CtxErrorf(t.ctx, "extract author by site rule failed: %v", err)
		return "", err
	}
	return title, err
}

func (t *AuthorExtractor) byMeta() (string, error) {
	me := NewMetaExtractor(t.ctx, t.Doc)
	metaContent := me.Extract()
	for _, key := range consts.AUTHOR_META_KEYS {
		if value, ok := metaContent[key]; ok {
			return value, nil
		}
	}
	return "", nil
}

func (t *AuthorExtractor) byNode() (string, error) {
	// TODO: 测试是否能支持这样的复杂xpath查询
	xpath := `
//*[
	(@*="author" or contains(@*,"author") or contains(@*,"Author") or contains(@*,"作者"))
	and not(contains(@*,"footer")) and not(contains(@*,"statement")) and not(contains(@*,"authorize"))
	and not(contains(@*,"comment"))
] |
//*[
	(@*="作者信息" or @*="author" or contains(@*,"author") or contains(@class,"author"))
	and not(contains(@*,"comment"))
]//*[@*="name" or contains(@*,"name")]
`
	elems := t.Doc.Xpath(xpath)
	if len(elems) != 0 {
		for _, elem := range elems {
			if elem.Text() != "" {
				return elem.Text(), nil
			}
		}
	}
	return "", nil
}
func (t *AuthorExtractor) byRegex() (string, error) {
	AUTHOR_PATTERN := "(%s)\\s*[：|:| |丨|/]\\s*([\u4E00-\u9FA5a-zA-Z]{2,20})[^\u4E00-\u9FA5|:|：]*"

	var keywords []string
	for _, keyword := range consts.AUTHOR_KEYWORDS {
		chars := []rune(keyword)
		if len(chars) == 2 {
			keywords = append(keywords, fmt.Sprintf("%v\\s{0,1}%v", string(chars[0]), string(chars[1])))
		} else {
			keywords = append(keywords, keyword)
		}
	}

	authorPattern := fmt.Sprintf(AUTHOR_PATTERN, strings.Join(keywords, "|"))

	re := regexp.MustCompile(authorPattern)
	text := t.Doc.GetRawDocText(t.Doc.Doc.Root())

	match := re.FindStringSubmatch(text)
	if match != nil && len(match) > 2 {
		return match[2], nil
	}
	if match = consts.RegexRule_AuthroName.FindStringSubmatch(text); match != nil && len(match) > 1 {
		return match[1], nil
	}
	return "", nil
}

func (t *AuthorExtractor) ExtractMeta() (*wcd.ArticleAuthorMeta, error) {
	authorMeta := &wcd.ArticleAuthorMeta{
		Name:        "",
		ProfileURL:  "",
		Description: "",
		UID:         "",
	}
	name, err := t.Extract()
	if err != nil {
		hlog.CtxErrorf(t.ctx, "extract author meta failed: %v", err)
		return nil, err
	}
	authorMeta.Name = name
	authorMeta.UID = t.extractMetaUid()
	authorMeta.Description = t.extractMetaDescription()
	authorMeta.ProfileURL = t.extractMetaProfileUrl()

	return authorMeta, nil
}

func (t *AuthorExtractor) extractMetaUid() string {
	me := NewMetaExtractor(t.ctx, t.Doc)
	metaContent := me.Extract()
	urlStr := metaContent["og:url"]
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		hlog.CtxErrorf(t.ctx, "extract author meta uid failed: %v", err)
		return ""
	}
	biz := parsedUrl.Query().Get("__biz")
	if biz == "" {
		if matches := consts.RegexRule_AuthorId.FindStringSubmatch(t.Doc.GetRawHtmlStr()); matches != nil && len(matches) > 1 {
			return matches[1]
		}
	}
	return biz
}

func (t *AuthorExtractor) extractMetaDescription() string {
	if matches := consts.RegexRule_AuthroDescription.FindStringSubmatch(t.Doc.GetRawHtmlStr()); matches != nil && len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (t *AuthorExtractor) extractMetaProfileUrl() string {
	if matches := consts.RegexRule_ProfilImg.FindStringSubmatch(t.Doc.GetRawHtmlStr()); matches != nil && len(matches) > 1 {
		return matches[1]
	}
	return ""
}
