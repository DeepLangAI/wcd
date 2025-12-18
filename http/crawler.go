package http

import (
	"context"
	"fmt"
	"github.com/DeepLangAI/go_lib/middleware"
	"net/http"
	"time"

	"github.com/DeepLangAI/wcd/conf"
	consts2 "github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

var crawlerClient *client.Client

func InitCrawlerClient() {
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use([]client.Middleware{
		middleware.TraceClientMiddleware,
		middleware.ResponseCheckClientMiddleware,
	}...)
	crawlerClient = cli
}

type crawlHtmlReq struct {
	Url          string `json:"url"`
	ForceBrowser *bool  `json:"force_browser,omitempty"`
}

type CrawlHtmlResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Html string `json:"html"`
	} `json:"data"`
}

type CrawlExtraParams struct {
	ForceBroswer bool `json:"force_browser"`
}

func CrawlHtml(ctx context.Context, url string, ext *CrawlExtraParams) (retData *CrawlHtmlResp, retErr error) {
	timeBegin := time.Now()
	defer func() {
		host := utils.ExtractUrlHost(url)
		if retErr != nil {
			hlog.CtxErrorf(ctx,
				"CrawlHtml failed, host: %v, url: %v, cost: %.2fs, err: %v",
				host,
				url,
				time.Since(timeBegin).Seconds(),
				retErr,
			)
		} else {
			hlog.CtxInfof(ctx,
				"CrawlHtml success, host: %v, url: %v, cost: %.2fs",
				host,
				url,
				time.Since(timeBegin).Seconds(),
			)
		}
	}()
	// 抓取网页
	reqBody := crawlHtmlReq{
		Url: url,
	}
	if ext != nil {
		if ext.ForceBroswer {
			reqBody.ForceBrowser = thrift.BoolPtr(ext.ForceBroswer)
		}
	}
	utils.PrintRequestLog(ctx, consts2.ActionCrawlHtml, url, consts2.ActionType_Request, reqBody)

	req := protocol.AcquireRequest()
	resp := protocol.AcquireResponse()
	defer func() {
		protocol.ReleaseRequest(req)
		protocol.ReleaseResponse(resp)
	}()

	bodyStr, err := sonic.MarshalString(reqBody)
	hlog.CtxDebugf(ctx, "CrawlHtml reqBody: %s", string(bodyStr))

	req.SetBody([]byte(bodyStr))
	req.SetMethod(consts.MethodPost)
	req.SetRequestURI("/crawl")
	req.SetHeader(consts.HeaderContentType, consts.MIMEApplicationJSON)
	req.SetHost(conf.GetConfig().ApiDomain.CrawlerApi)

	timeout := consts2.CrawlHtmlTimeout
	err = crawlerClient.DoTimeout(ctx, req, resp, timeout)
	if err != nil {
		hlog.CtxErrorf(ctx, "CrawlHtml crawlerClient.DoTimeout failed, err: %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("CrawlHtml crawlerClient.DoTimeout failed, status code: %d", resp.StatusCode())
		hlog.CtxErrorf(ctx, "%v", err)
		return nil, err
	}

	result := &CrawlHtmlResp{}
	err = sonic.Unmarshal(resp.Body(), result)
	if err != nil {
		hlog.CtxErrorf(ctx, "CrawlHtml sonic.Unmarshal failed, err: %v", err)
		return nil, err
	}

	if result.Code != 0 {
		err = fmt.Errorf("CrawlHtml failed, resp status not ok, msg: %s", result.Msg)
		hlog.CtxErrorf(ctx, "%v", err)
		return nil, err
	}
	utils.PrintRequestLog(ctx, consts2.ActionCrawlHtml, url, consts2.ActionType_Response, result)

	return result, nil
}
