package wcd

import (
	"context"
	"fmt"
	"github.com/DeepLangAI/go_lib/utillib"
	"runtime"
	"time"

	"github.com/DeepLangAI/wcd/biz/model/text_parse"
	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/conf"
	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/http"
	"github.com/DeepLangAI/wcd/model/http_model"
	"github.com/DeepLangAI/wcd/utils"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type WcdParseService struct {
}

func (s *WcdParseService) textParseLabelize(ctx context.Context, labelModelReq *http_model.LabelModelReq) (*http_model.LabelModelResp, *http.ModelRawReqResp, error) {
	if conf.GetConfig().Parse.Label.UseMock {
		labelModelResp := &http_model.LabelModelResp{
			Code: 0,
			Msg:  "",
			Type: "web",
			Infos: utils.Map(labelModelReq.Infos, func(infoReq *http_model.LabelInfo) *http_model.TextParseInfo {
				return &http_model.TextParseInfo{
					LabelInfo: http_model.LabelInfo{
						Txt:          infoReq.Txt,
						Position:     infoReq.Position,
						Tags:         infoReq.Tags,
						Label:        consts.LABEL_CONTENT, // mock as content
						Meta:         infoReq.Meta,
						WebSegmentId: infoReq.WebSegmentId,
					},
				}
			}),
			ArticleMeta:  nil,
			LabelVersion: "",
		}

		rawReq, _ := sonic.MarshalString(labelModelReq)
		rawResp, _ := sonic.MarshalString(labelModelResp)

		io := &http.ModelRawReqResp{
			Req:  rawReq,
			Resp: rawResp,
			Code: 0,
		}
		return labelModelResp, io, nil
	} else {
		label, io, err := http.ParseLabel(ctx, labelModelReq)
		if err != nil {
			hlog.CtxErrorf(ctx, "http.ParseLabel failed, err: %v", err)
			return label, io, err
		}
		return label, io, nil
	}
}
func (s *WcdParseService) labelReqFromSegmentResult(ctx context.Context, url string, resp *wcd.SegmentResp) (*http_model.LabelModelReq, error) {
	var (
		labelInfos = make([]*http_model.LabelInfo, 0)
	)
	labelInfos = utils.Map(resp.Sentences, http_model.NewLabelInfoFromSentence)

	articleMeta := wcd.ArticleMeta{
		URL:           url,
		Title:         resp.ArticleMeta.Title,
		PublishTime:   resp.ArticleMeta.PublishTime,
		Author:        resp.ArticleMeta.Author,
		ContentSource: resp.ArticleMeta.ContentSource,
	}

	labelReq := &http_model.LabelModelReq{
		ArticleMeta: articleMeta,
		EntryId:     "",
		Type:        "web",
		Infos:       labelInfos,
	}
	return labelReq, nil
}

func (s *WcdParseService) recoverArticleMeta(ctx context.Context, articleMeta *wcd.ArticleMeta, modelResp *http_model.LabelModelResp) (*wcd.ArticleMeta, error) {
	if articleMeta == nil {
		return nil, nil
	}

	if modelRespArticleMeta := modelResp.ArticleMeta; modelRespArticleMeta != nil {
		if articleMeta.Author == "" {
			articleMeta.Author = modelRespArticleMeta.Author
		}
		if articleMeta.Title == "" {
			articleMeta.Title = modelRespArticleMeta.Title
		}
		if articleMeta.PublishTime == "" {
			articleMeta.PublishTime = modelRespArticleMeta.PublishTime
		}
		if articleMeta.ContentSource == "" {
			articleMeta.ContentSource = modelRespArticleMeta.ContentSource
		}
	}
	if articleMeta.ContentSource == "" {
		for _, info := range modelResp.Infos {
			if info.Label == consts.LABEL_SOURCE {
				articleMeta.ContentSource = info.Txt
				break
			}
		}
	}
	return articleMeta, nil
}

type WcdParseFunc func(ctx context.Context, req wcd.WcdParseReq) (*wcd.WcdParseResp, *consts.BizCode)

func (s *WcdParseService) WcdParse(ctx context.Context, req wcd.WcdParseReq) (wcdParseResp *wcd.WcdParseResp, fnErr *consts.BizCode) {
	beginTime := time.Now()
	labelDuration := 0.0
	crawlImagesDuration := 0.0
	numSentences := 0
	defer func() {
		// 收集panic
		if r := recover(); r != nil {
			// fnResp = nil
			buffer := make([]byte, 9024)
			n := runtime.Stack(buffer, true)

			err := fmt.Sprintf("%v", r)
			hlog.CtxErrorf(ctx, "WcdParse panic, err: %v\nstack:\n%v", err, string(buffer[:n]))

			fnErr = &consts.SystemErr
		}

		hlog.CtxInfof(ctx, "WcdParse end, cost: %.2fs, num_sentences: %v, url: %v",
			time.Since(beginTime).Seconds()-labelDuration-crawlImagesDuration,
			numSentences,
			req.GetURL(),
		)
		if wcdParseResp != nil {
			fnRespStr, err := sonic.MarshalString(wcdParseResp)
			if err == nil {
				hlog.CtxInfof(ctx, "WcdParse result, resp: %v", utillib.TranslateJsonIO(ctx, fnRespStr))
			}
		}
	}()

	wcdParseResp = &wcd.WcdParseResp{
		URL:          req.URL,
		WcdRequestID: utils.GetCtxOperationId(ctx),
		Worthless:    true,
		WorthType:    consts.WorthType_NoContent,
	}

	if req.URL == "" || req.HTML == "" {
		errMsg := "WcdParse failed, req.URL or req.HTML is empty"
		hlog.CtxErrorf(ctx, "WcdParse failed, err: %v", errMsg)
		return wcdParseResp, &consts.ReqParamError
	}

	segmentService := SegmentService{}
	segmentReq := wcd.SegmentReq{
		HTML:           req.HTML,
		URL:            req.URL,
		RuleStageGroup: req.RuleStageGroup,
	}
	// 1。切分
	segmentResp, bizErr := segmentService.HtmlSegment(ctx, segmentReq)
	if bizErr != nil {
		hlog.CtxErrorf(ctx, "segmentService.HtmlSegment failed, bizErr: %v", bizErr)

		return wcdParseResp, bizErr
	}
	numSentences = len(segmentResp.Sentences) // 切句数量
	wcdParseResp.Title = segmentResp.ArticleMeta.Title
	wcdParseResp.Author = segmentResp.ArticleMeta.Author
	wcdParseResp.PubTime = segmentResp.ArticleMeta.PublishTime
	wcdParseResp.AuthorMeta = segmentResp.ArticleMeta.AuthorMeta
	wcdParseResp.SiteIcon = segmentResp.ArticleMeta.SiteIcon
	wcdParseResp.Description = segmentResp.ArticleMeta.Description
	wcdParseResp.SurfaceImage = segmentResp.ArticleMeta.SurfaceImage

	// 2. 标注
	utils.CoreLog(ctx, utils.CoreNameLabel, utils.NodeBegin)
	result, err := s.labelReqFromSegmentResult(ctx, req.URL, segmentResp)
	if err != nil {
		hlog.CtxErrorf(ctx, "s.labelReqFromSegmentResult failed, err: %v", err)
		return wcdParseResp, &consts.SystemErr
	}

	timeBeginLabel := time.Now()
	labelResp, io, err := s.textParseLabelize(ctx, result)
	labelDuration = time.Since(timeBeginLabel).Seconds()

	if io != nil {
		wcdParseResp.ModelInputStr = io.Req
		wcdParseResp.ModelResultStr = io.Resp
	}
	if err != nil {
		hlog.CtxErrorf(ctx, "s.textParseLabelize failed, err: %v", err)
		if io != nil {
			switch io.Code {
			case http_model.LabelModelCode_AllEduO, http_model.LabelModelCode_InfosEmpty:
				return wcdParseResp, &consts.ParseWorthless
			case http_model.LabelModelCode_Unk:
				return wcdParseResp, &consts.TextParseFailed
			}
		}
		return wcdParseResp, &consts.TextParseFailed
	}
	utils.CoreLog(ctx, utils.CoreNameLabel, utils.NodeDone)

	// 3. 恢复文章元信息
	segmentResp.ArticleMeta, err = s.recoverArticleMeta(ctx, segmentResp.ArticleMeta, labelResp)
	if err != nil {
		hlog.CtxErrorf(ctx, "s.RecoverArticleMeta failed, err: %v", err)
		return wcdParseResp, &consts.SystemErr
	}
	articleMeta := segmentResp.ArticleMeta
	wcdParseResp.Title = articleMeta.Title
	wcdParseResp.Author = articleMeta.Author
	wcdParseResp.ContentSource = articleMeta.ContentSource
	wcdParseResp.PubTime = articleMeta.PublishTime

	// 4.去噪
	distillService := NewDistillService(ctx, segmentResp.HTML, req.URL, req.GetRuleStageGroup())
	distillReq := wcd.DistillReq{
		Sentences: utils.Map(labelResp.Infos, func(info *http_model.TextParseInfo) *wcd.TextParseLabelSentence {
			return &wcd.TextParseLabelSentence{
				Text: info.Txt,
				Atoms: utils.Map(info.Position.Atoms, func(atom *text_parse.AtomicTxt) *wcd.AtomicText {
					return &wcd.AtomicText{
						Text:       *atom.Txt,
						PositionID: atom.PositionID,
						Xpath:      atom.X,
						SegmentID:  info.WebSegmentId,
					}
				}),
				SegmentID: info.WebSegmentId,
				Label:     info.Label,
			}
		}),
		HTML:        segmentResp.HTML,
		URL:         req.URL,
		ArticleMeta: articleMeta,
	}

	distill, bizErr := distillService.Distill(ctx, distillReq)
	if bizErr != nil {
		hlog.CtxErrorf(ctx, "distillService.Distill failed, err: %v", err)
		return wcdParseResp, bizErr
	}

	wcdParseResp.URL = req.URL                                       // 网页链接
	wcdParseResp.Text = distill.Text                                 // 去噪文本
	wcdParseResp.Images = distill.Images                             // 图片
	wcdParseResp.ReadableHTML = distill.HTML                         // 阅读器网页
	wcdParseResp.Title = articleMeta.Title                           // 标题
	wcdParseResp.Author = articleMeta.Author                         // 作者名
	wcdParseResp.AuthorMeta = articleMeta.AuthorMeta                 // 作者详细信息
	wcdParseResp.SiteIcon = segmentResp.ArticleMeta.SiteIcon         // 网站图标
	wcdParseResp.Description = segmentResp.ArticleMeta.Description   // 网页描述
	wcdParseResp.SurfaceImage = segmentResp.ArticleMeta.SurfaceImage // 封面图
	wcdParseResp.ContentSource = articleMeta.ContentSource           // 来源
	wcdParseResp.PubTime = articleMeta.PublishTime                   // 发布时间
	wcdParseResp.ModelInputStr = io.Req                              // 请求模型原始数据
	wcdParseResp.ModelResultStr = io.Resp                            // 模型响应原始数据
	wcdParseResp.Worthless = distill.Worthless                       // 是否无意义
	wcdParseResp.WorthType = distill.WorthType                       // 意义类型
	wcdParseResp.WcdRequestID = utils.GetCtxOperationId(ctx)
	//wcdParseResp.OssInfo = nil

	return wcdParseResp, nil
}
