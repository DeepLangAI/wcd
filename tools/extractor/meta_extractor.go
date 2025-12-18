package extractor

import (
	"context"

	"github.com/DeepLangAI/wcd/tools/doc"
)

type MetaExtractor struct {
	Doc *doc.Document
	ctx context.Context
}

func NewMetaExtractor(ctx context.Context, doc *doc.Document) *MetaExtractor {
	return &MetaExtractor{
		ctx: ctx,
		Doc: doc,
	}
}

func (m *MetaExtractor) Extract() map[string]string {
	elements := m.Doc.Xpath("//meta")
	metaContent := map[string]string{}
	for _, elem := range elements {
		key := ""
		for _, tmpKey := range []string{"name", "property"} {
			if attr := elem.SelectAttrValue(tmpKey, ""); attr != "" {
				key = attr
				break
			}
		}
		value := elem.SelectAttrValue("content", "")

		if key != "" && value != "" {
			metaContent[key] = value
		}
	}
	return metaContent
}
