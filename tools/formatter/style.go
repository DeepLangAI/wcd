package formatter

import (
	"context"
	"fmt"
	"strings"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/beevik/etree"
)

type StyleFormatter struct {
	ctx context.Context
	doc *doc.Document
}

func NewStyleFormatter(ctx context.Context, doc *doc.Document) *StyleFormatter {
	return &StyleFormatter{
		ctx: ctx,
		doc: doc,
	}
}

func (f *StyleFormatter) Format() {
	f.formatLi()
	f.formatGlobal()
}

func (f *StyleFormatter) formatGlobal() {
	styleTag := &etree.Element{
		Tag: "style",
	}
	styleStr := ""
	if strings.Contains(f.doc.Url, "web.okjike.com") {
		styleStr += `
br {
  line-height: normal;
}`
	}

	styleTag.SetText(styleStr)
	if styleStr != "" {
		f.doc.Doc.Root().AddChild(styleTag)
	}
}
func (f *StyleFormatter) formatLi() {
	xpath := fmt.Sprintf("//%v//li/*", consts.VNODE_TAG_PREFIX+"intro")
	for _, elem := range f.doc.Xpath(xpath) {
		style := elem.SelectAttrValue(consts.StyleAttr, "")
		elem.CreateAttr(consts.StyleAttr, style+"; display: contents;")
	}
}
