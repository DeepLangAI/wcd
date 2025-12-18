package extractor

import (
	"context"
	"strings"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type SurfaceImgExtractor struct {
	Doc    *doc.Document
	result string
	ctx    context.Context
}

func NewSurfaceImgExtractor(ctx context.Context, doc *doc.Document) *SurfaceImgExtractor {
	return &SurfaceImgExtractor{
		Doc: doc,
		ctx: ctx,
	}
}

func (t *SurfaceImgExtractor) Extract() (string, error) {
	if strings.Contains(t.Doc.Url, "tmtpost.com") {
		// 钛媒体的封面图都是 logo，所以屏蔽掉
		return "", nil
	}

	if t.result != "" {
		return t.result, nil
	}
	var description string
	var err error

	nodes := []func() (string, error){
		t.byMeta,
	}
	for _, node := range nodes {
		description, err = node()
		if err != nil {
			hlog.CtxErrorf(t.ctx, "extract surface img failed, err: %v", err)
			return "", err
		}
		if description != "" {
			return description, nil
		}
	}
	return description, nil
}

func (t *SurfaceImgExtractor) byMeta() (string, error) {
	me := NewMetaExtractor(t.ctx, t.Doc)
	metaContent := me.Extract()

	for _, key := range consts.SURFACE_IMG_META_KEYS {
		if val, ok := metaContent[key]; ok {
			return utils.EnsureLinkAbsolute(val, t.Doc.Url), nil
		}
	}
	return "", nil
}
