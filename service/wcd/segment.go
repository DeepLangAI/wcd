package wcd

import (
	"context"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools"
	wcdDoc "github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type SegmentService struct {
}

func (s *SegmentService) HtmlSegment(ctx context.Context, req wcd.SegmentReq) (*wcd.SegmentResp, *consts.BizCode) {
	utils.CoreLog(ctx, utils.CoreNameSegment, utils.NodeBegin)
	doc, err := wcdDoc.NewDocument(ctx, req.HTML, req.URL, req.GetRuleStageGroup())

	if err != nil {
		hlog.CtxErrorf(ctx, "html解析失败%v", err)
		return nil, &consts.ParseWorthless
	}

	parser := tools.NewParser(ctx, doc, nil)
	authorMeta := parser.AuthorMeta()
	parsedData := &wcd.ArticleMeta{
		Title:         parser.Title(),
		PublishTime:   parser.PublishTime(),
		Author:        authorMeta.Name,
		ContentSource: parser.ContentSource(),
		AuthorMeta:    authorMeta,
		SiteIcon:      thrift.StringPtr(parser.SiteIcon()),
		Description:   thrift.StringPtr(parser.SiteDescription()),
		SurfaceImage:  thrift.StringPtr(parser.SurfaceImage()),
	}

	cleaner := tools.NewCleaner(ctx, doc)
	err = cleaner.Purify()
	utils.CoreLog(ctx, utils.CoreNameSegment, utils.NodeNameSegmentPreDistill)
	if err != nil {
		hlog.CtxErrorf(ctx, "purify failed, err: %v", err)
		return nil, &consts.ParseWorthless
	}
	if doc.Doc.Root() == nil {
		hlog.CtxErrorf(ctx, "after cleaner purify, html is empty")
		return nil, &consts.ParseWorthless
	}

	spliter := tools.NewSplitter(ctx, doc)
	sents, err := spliter.Split(parsedData)
	utils.CoreLog(ctx, utils.CoreNameSegment, utils.NodeNameSegmentSegment)
	if err != nil {
		hlog.CtxErrorf(ctx, "html切分失败%v", err)
		return nil, &consts.ParseWorthless
	}

	formatter := tools.NewFormatter(ctx, doc, nil)
	formatter.PreFormat()
	utils.CoreLog(ctx, utils.CoreNameSegment, utils.NodeNameSegmentPreFormat)

	htmlStr, err := doc.ToString()
	if err != nil {
		hlog.CtxErrorf(ctx, "html转字符串失败%v", err)
		return nil, &consts.ParseWorthless
	}
	resp := &wcd.SegmentResp{
		Sentences:            sents,
		HTML:                 htmlStr,
		OperationID:          utils.GetCtxOperationId(ctx),
		ArticleMeta:          parsedData,
		ImagesWithPositionID: doc.GetImagesWithPositionId(),
	}

	utils.CoreLog(ctx, utils.CoreNameSegment, utils.NodeDone)
	return resp, nil
}
