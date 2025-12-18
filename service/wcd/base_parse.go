package wcd

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/conf"
	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/dal/mongo"
	"github.com/DeepLangAI/wcd/http"
	"github.com/DeepLangAI/wcd/model/http_model"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/utils"

	"github.com/beevik/etree"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type BaseParseService struct {
}

func (b *BaseParseService) BaseParse(ctx context.Context, req wcd.BaseParseReq) (
	fnResp *wcd.WcdParseResp,
	fnBizErr *consts.BizCode,
) {
	defer func() {
		if fnResp != nil {
			modelInput := http_model.LabelModelReq{}
			sonic.UnmarshalString(fnResp.ModelInputStr, &modelInput)

			hlog.CtxInfof(ctx, "BaseParse result: text_length: %v, tp_text_length: %v, num_sentences: %v, url: %v",
				len(fnResp.GetText()),
				len(strings.Join(utils.Map(modelInput.Infos, func(t *http_model.LabelInfo) string {
					return t.Txt
				}), "")),
				len(modelInput.Infos),
				req.GetURL(),
			)
		}
	}()
	return b.webBaseParse(ctx, req)
}

const (
	CrawlerName_Unk     = "unk"
	CrawlerName_Cache   = "cache"
	CrawlerName_Crawler = "crawler"
)

type CrawlResult struct {
	Html        string
	NeedCache   bool
	CrawlerName string
}

func (b *BaseParseService) checkUrlNeedBrowserCrawl(ctx context.Context, htmlUrl string, ruleStageGeoup wcd.RuleStageGroupEnum) bool {
	doc := doc.Document{Url: htmlUrl, RuleStageGroup: ruleStageGeoup}
	rule, err := doc.MatchRule()
	if err != nil || rule == nil {
		return false
	}
	return rule.NeedBrowserCrawl
}

func (b *BaseParseService) checkNeedBrowserCrawl(ctx context.Context, htmlUrl string, html string, ruleStageGeoup wcd.RuleStageGroupEnum) bool {
	if strings.Contains(htmlUrl, "mp.weixin.qq.com") {
		document, err := doc.NewDocument(ctx, html, htmlUrl, ruleStageGeoup)
		if err != nil {
			return false
		}
		if nodes := document.Xpath("//p[@id='js_text_desc']"); len(nodes) > 0 {
			content := strings.Join(
				utils.Map(nodes, func(n *etree.Element) string {
					return document.GetRawDocText(n)
				}), "",
			)
			// 如果内容太短，则使用浏览器抓取
			if len(utils.RemoveSpace(content)) < 10 {
				hlog.CtxInfof(ctx, "checkNeedBrowserCrawl content too short")
				return true
			}
		}
	}
	return false
}

func (b *BaseParseService) crawlHtmlWithCache(
	ctx context.Context,
	htmlUrl string,
	ruleStageGroup wcd.RuleStageGroupEnum,
	needBrowserCrawl bool,
) (*CrawlResult, error) {
	expireDuration := time.Hour * time.Duration(conf.GetConfig().Parse.Crawl.HtmlCacheHours)
	model, err := mongo.CrawlHtmlModelDal.FindByUrl(ctx, htmlUrl, expireDuration)
	if err == nil && model.Html != "" {
		hlog.CtxInfof(ctx, "crawlHtmlWithCache hit cache")
		return &CrawlResult{
			Html:        model.Html,
			NeedCache:   false,
			CrawlerName: CrawlerName_Cache,
		}, nil
	} else {
		hlog.CtxInfof(ctx, "crawlHtmlWithCache not hit cache")
	}
	// 用crawler
	if !needBrowserCrawl {
		crawlResult, err := http.CrawlHtml(ctx, htmlUrl, nil)
		data := crawlResult.Data
		if err == nil && data.Html != "" {
			if !b.checkNeedBrowserCrawl(ctx, htmlUrl, data.Html, ruleStageGroup) {
				return &CrawlResult{
					Html:        data.Html,
					NeedCache:   true,
					CrawlerName: CrawlerName_Crawler,
				}, nil
			} else {
				needBrowserCrawl = true
			}
		}
	}

	// 用crawler，用浏览器抓取
	if needBrowserCrawl {
		crawlResult, err := http.CrawlHtml(ctx, htmlUrl, &http.CrawlExtraParams{ForceBroswer: true})
		data := crawlResult.Data
		if err == nil && data.Html != "" {
			return &CrawlResult{
				Html:        data.Html,
				NeedCache:   true,
				CrawlerName: CrawlerName_Crawler,
			}, nil
		}
	}

	return nil, errors.New("all crawler failed")
}

func (b *BaseParseService) webBaseParse(ctx context.Context, req wcd.BaseParseReq) (*wcd.WcdParseResp, *consts.BizCode) {

	// 如果传入了html，但检测到需要浏览器渲染后的抓取结果，则忽略传入的html
	needBrowserCrawl := false

	// 1. 抓取网页
	var err error
	htmlStr := ""
	needCacheHtml := false
	crawlerName := ""

	if req.GetHTML() == "" {
		if b.checkUrlNeedBrowserCrawl(ctx, req.URL, req.GetRuleStageGroup()) {
			hlog.CtxInfof(ctx, "checkUrlNeedBrowserCrawl before crawl, check need browser to crawl")
			needBrowserCrawl = true
		}
		crawlResult, err := b.crawlHtmlWithCache(ctx, req.URL, req.GetRuleStageGroup(), needBrowserCrawl)
		if err != nil {
			hlog.CtxErrorf(ctx, "webBaseParse crawlHtmlWithCache err:%v", err)
			return nil, &consts.CrawlFailed
		}
		htmlStr = crawlResult.Html
		needCacheHtml = crawlResult.NeedCache
		crawlerName = crawlResult.CrawlerName
	} else {
		htmlStr = req.GetHTML()
		needCacheHtml = true
		crawlerName = CrawlerName_Unk
	}
	htmlStr = utils.UnescapeHtml(htmlStr)
	hlog.CtxInfof(ctx, "crawler_name: %v, crawl html success, url:%v", crawlerName, req.URL)

	// 2. 解析
	service := WcdParseService{}
	parseResult, bizErr := service.WcdParse(ctx, wcd.WcdParseReq{
		HTML:           htmlStr,
		URL:            req.URL,
		RuleStageGroup: req.RuleStageGroup,
	})

	if req.GetWithRawHTML() == true {
		parseResult.RawHTML = &htmlStr
	}

	if bizErr != nil {
		hlog.CtxErrorf(ctx, "webBaseParse WcdParse err:%v", err)
		return parseResult, bizErr
	}

	// 如果本次重新抓取，且解析结果有效，则缓存html。之后可以根据缓存来判断是否需要重新抓取
	if needCacheHtml && parseResult.Worthless == false {
		hlog.CtxInfof(ctx, "crawler_name: %v, parse result is valid", crawlerName)
		if req.GetSaveCrawlHTML() {
			err = mongo.CrawlHtmlModelDal.SaveOne(ctx, mongo.CrawlHtmlModel{
				Url:        req.URL,
				Html:       htmlStr,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
				Status:     consts.StatusValid,
			})
			if err != nil {
				hlog.CtxErrorf(ctx, "crawlHtmlWithCache SaveOne err:%v", err)
			}
		}
	}
	return parseResult, nil
}
