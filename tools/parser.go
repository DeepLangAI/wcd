package tools

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/tools/extractor"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type Parser struct {
	ctx    context.Context
	Doc    *doc.Document
	Labels []*wcd.TextParseLabelSentence
	text   string
}

func NewParser(ctx context.Context, doc *doc.Document, labels []*wcd.TextParseLabelSentence) *Parser {
	text := doc.GetRawDocText(doc.Doc.Root())
	return &Parser{
		ctx:    ctx,
		Doc:    doc,
		Labels: labels,
		text:   text,
	}
}

func (p *Parser) Title() string {
	extractor := extractor.NewTitleExtractor(p.ctx, p.Doc)
	result, err := extractor.Extract()
	if err != nil {
		hlog.CtxErrorf(p.ctx, "extractor.Extract err:%v", err)
		return ""
	}
	result = strings.TrimSpace(result)
	return result
}
func (p *Parser) Author() string {
	extractor := extractor.NewAuthorExtractor(p.ctx, p.Doc)
	result, err := extractor.Extract()
	if err != nil {
		hlog.CtxErrorf(p.ctx, "extractor.Extract err:%v", err)
		return ""
	}
	result = strings.TrimSpace(result)
	return result
}

func (p *Parser) AuthorMeta() *wcd.ArticleAuthorMeta {
	extractor := extractor.NewAuthorExtractor(p.ctx, p.Doc)
	result, err := extractor.ExtractMeta()
	if err != nil {
		hlog.CtxErrorf(p.ctx, "extractor.Extract err:%v", err)
		return nil
	}
	return result
}

func (p *Parser) ContentSource() string {
	// 根据模型标签来提取
	return ""
}

func (p *Parser) PublishTime() string {
	timeExtractor := extractor.NewPublishTimeExtractor(p.ctx, p.Doc)
	result, err := timeExtractor.Extract()
	if err != nil {
		hlog.CtxErrorf(p.ctx, "extractor.Extract err:%v", err)
		return ""
	}
	result = strings.TrimSpace(result)
	return result
}

func (p *Parser) checkTrueWorthless(articleMeta *wcd.ArticleMeta) bool {
	if utf8.RuneCountInString(p.text) >= consts.WORTHLESS_TXT_LEN {
		return false
	}
	if articleMeta.Title != "" {
		titleCompacted := utils.Clean(articleMeta.Title)
		findString := consts.WORTHLESS_PAGE_TITLE_REGEX.FindString(titleCompacted)
		if findString != "" {
			return true
		}
	}
	for _, keywords := range consts.WORTHLESS_PAGE_KEYWORDS {
		if len(keywords) < 2 {
			continue
		}
		matchCount := 0
		for _, keyword := range keywords {
			if strings.Contains(p.text, keyword) {
				matchCount += 1
			}
		}
		if float32(matchCount)/float32(len(keywords)) >= consts.MULTIPLE_WORDS_MATCH_RATIO || matchCount+1 == len(keywords) {
			return true
		}
	}
	return false
}

func (p *Parser) hasValueableImage() bool {
	return utils.Any(p.Labels, func(info *wcd.TextParseLabelSentence) bool {
		return info.Label == consts.LABEL_FIGURE
	})
}

func (p *Parser) checkContentWorthless(articleMeta *wcd.ArticleMeta) bool {
	text := strings.Replace(p.text, articleMeta.Title, "", -1)
	if utf8.RuneCountInString(text) <= consts.WORTHLESS_TXT_LEN && !p.hasValueableImage() {
		return true
	}
	return false
}

func (p *Parser) CheckWorthless(articleMeta *wcd.ArticleMeta) int {
	if p.checkTrueWorthless(articleMeta) {
		return consts.WorthType_404
	}
	if p.checkContentWorthless(articleMeta) {
		return consts.WorthType_NoContent
	}
	return consts.WorthType_Valueable
}

func (p *Parser) SiteIcon() (iconLink string) {
	x := extractor.NewSiteIconExtractor(p.ctx, p.Doc)
	siteIcon, err := x.Extract()
	if err != nil {
		hlog.CtxErrorf(p.ctx, "extractor.Extract err:%v", err)
		return
	}
	return siteIcon
}

func (p *Parser) SiteDescription() string {
	x := extractor.NewDescriptionExtractor(p.ctx, p.Doc)
	description, err := x.Extract()
	if err != nil {
		hlog.CtxErrorf(p.ctx, "description extractor.Extract err:%v", err)
		return ""
	}
	return strings.TrimSpace(description)
}

func (p *Parser) SurfaceImage() string {
	x := extractor.NewSurfaceImgExtractor(p.ctx, p.Doc)
	val, err := x.Extract()
	if err != nil {
		hlog.CtxErrorf(p.ctx, "surface image extractor.Extract err:%v", err)
		return ""
	}
	return strings.TrimSpace(val)
}
