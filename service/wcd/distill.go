package wcd

import (
	"context"
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools"
	wcdDoc "github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/utils"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type DistillService struct {
	doc *wcdDoc.Document
}

func NewDistillService(ctx context.Context, htmlStr string, htmlUrl string, ruleStageGroup wcd.RuleStageGroupEnum) *DistillService {
	doc := wcdDoc.LoadDocumentFromSegmentResult(ctx, htmlStr, htmlUrl, ruleStageGroup)
	return &DistillService{
		doc: doc,
	}
}

func (d *DistillService) CheckWorthless(ctx context.Context, req wcd.DistillReq, doc *wcdDoc.Document) int {
	if req.ArticleMeta == nil {
		req.ArticleMeta = &wcd.ArticleMeta{}
	}
	parser := tools.NewParser(ctx, doc, req.Sentences)
	worthType := parser.CheckWorthless(req.ArticleMeta)
	return worthType
}

func (d *DistillService) Distill(ctx context.Context, req wcd.DistillReq) (*wcd.DistillResp, *consts.BizCode) {
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeBegin)
	//doc := wcdDoc.LoadDocumentFromSegmentResult(ctx, req.GetHTML(), req.GetURL(), wcd.RuleStageGroupEnum_ProdOnly)
	doc := d.doc
	if doc == nil {
		hlog.CtxErrorf(ctx, "load document failed")
		return nil, &consts.ParseWorthless
	}
	formatter := tools.NewFormatter(ctx, doc, req.Sentences)
	formatter.PostFormat()
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeNamePostDistillFormat)

	cleaner := tools.NewCleaner(ctx, doc)
	cleaner.SetSentences(req.Sentences)
	cleaner.PostPurify()
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeNamePostDistillDistill)

	doc.InsertMeta(req.ArticleMeta)
	readerHtml, err := doc.ToString()
	if err != nil {
		hlog.CtxErrorf(ctx, "readerHtml err: %v", err)
		return nil, &consts.ParseWorthless
	}
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeNamePostDistillRenderHtml)

	worthType := d.CheckWorthless(ctx, req, doc)
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeNamePostDistillCheckWorthless)

	// todo 这个节点很耗时，需要优化
	//sentenceIds := d.getExistingSentences(ctx, doc, req.Sentences)
	sentenceIds := make([]int32, 0)
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeNamePostDistillGetExistSentence)

	text := doc.GetRawDocText(doc.Doc.Root())
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeNamePostDistillGetText)

	images := doc.GetImages()
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeNamePostDistillGetImg)
	result := &wcd.DistillResp{
		SentenceIds: sentenceIds,
		HTML:        readerHtml,
		Text:        text,
		Images:      images,
		Worthless:   worthType != consts.WorthType_Valueable,
		WorthType:   int32(worthType),
	}
	req.GetArticleMeta()
	utils.CoreLog(ctx, utils.CoreNamePostDistill, utils.NodeDone)
	return result, nil
}

func (d *DistillService) getExistingImages(ctx context.Context, images map[string]string) []string {
	hlog.CtxInfof(ctx, "before denoise, num images: %v", len(images))
	pids := []string{}
	pidToImages := map[string]string{}
	for imgUrl, pid := range images {
		pidToImages[pid] = imgUrl
		pids = append(pids, pid)
	}

	xpath := strings.Join(utils.Map(pids, func(pid string) string {
		return utils.GetPositionIdXpath(pid)
	}), " | ")

	imgUrls := []string{}
	if elems := d.doc.Xpath(xpath); len(elems) > 0 {
		for _, elem := range elems {
			if pid := elem.SelectAttrValue(consts.KeyPositionId, ""); pid != "" {
				if imgUrl, ok := pidToImages[pid]; ok {
					imgUrls = append(imgUrls, imgUrl)
				}
			}
		}
	}
	hlog.CtxInfof(ctx, "after denoise, num images: %v", len(imgUrls))
	return imgUrls
}
