package http

import (
	"context"
	"fmt"
	"github.com/DeepLangAI/go_lib/middleware"
	"github.com/DeepLangAI/go_lib/utillib"

	"github.com/DeepLangAI/wcd/conf"
	consts2 "github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/model/http_model"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var labelClient *client.Client

func InitLabelClient() {
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use([]client.Middleware{
		middleware.TraceClientMiddleware,
		middleware.ResponseCheckClientMiddleware,
	}...)

	labelClient = cli
}

type ModelRawReqResp struct {
	Req  string
	Resp string
	Code int
}

func ParseLabel(ctx context.Context, reqBody *http_model.LabelModelReq) (*http_model.LabelModelResp, *ModelRawReqResp, error) {
	io := &ModelRawReqResp{}
	io.Code = http_model.LabelModelCode_Unk
	req := protocol.AcquireRequest()
	resp := protocol.AcquireResponse()
	defer func() {
		protocol.ReleaseRequest(req)
		protocol.ReleaseResponse(resp)
	}()
	bodyStr := utillib.JsonMarshal(ctx, reqBody)
	utils.PrintRequestLog(ctx, consts2.ActionTextParse, "", consts2.ActionType_Request, reqBody)
	req.SetBody([]byte(bodyStr))
	io.Req = string(req.Body())
	req.SetHeader(consts.HeaderContentType, consts.MIMEApplicationJSON)
	req.SetMethod(consts.MethodPost)
	req.SetRequestURI("parser-v2/text")
	req.SetHost(conf.GetConfig().ApiDomain.AiApi)

	err := labelClient.DoTimeout(ctx, req, resp, consts2.TextParseReadTimeOut)
	if err != nil {
		hlog.CtxErrorf(ctx, "ParseLabel do req error %v", err)
		return nil, io, err
	}

	res := &http_model.LabelModelResp{}
	io.Resp = string(resp.Body())
	if err = sonic.UnmarshalString(io.Resp, res); err != nil {
		hlog.CtxErrorf(ctx, "ParseLabel unmarshal %v error %v", string(resp.Body()), err)
		return nil, io, err
	}
	if res.Code != http_model.LabelModelCode_Success {
		io.Code = res.Code
		err = fmt.Errorf("ParseLabel code: %v, msg: %v", res.Code, res.Msg)
		hlog.CtxErrorf(ctx, "%v", err)
		return nil, io, err
	}
	utils.PrintRequestLog(ctx, consts2.ActionTextParse, "", consts2.ActionType_Response, res)
	io.Code = http_model.LabelModelCode_Success
	return res, io, nil
}
