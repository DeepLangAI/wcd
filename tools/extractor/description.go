package extractor

import (
	"context"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type DescriptionExtractor struct {
	Doc    *doc.Document
	result string
	ctx    context.Context
}

func NewDescriptionExtractor(ctx context.Context, doc *doc.Document) *DescriptionExtractor {
	return &DescriptionExtractor{
		Doc: doc,
		ctx: ctx,
	}
}

func (t *DescriptionExtractor) Extract() (string, error) {
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
			hlog.CtxErrorf(t.ctx, "extract description failed, err: %v", err)
			return "", err
		}
		if description != "" {
			return description, nil
		}
	}
	return description, nil
}

func (t *DescriptionExtractor) byMeta() (string, error) {
	me := NewMetaExtractor(t.ctx, t.Doc)
	metaContent := me.Extract()

	for _, key := range consts.DESCRIPTION_META_KEYS {
		if val, ok := metaContent[key]; ok {
			return val, nil
		}
	}
	return "", nil
}
