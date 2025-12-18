package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/tools/sentence"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type TableNode struct {
	Elem *etree.Element
	Doc  *doc.Document
	Sik  *sentence.SegmentIdKeeper
	ctx  context.Context
}

func NewTableNode(elem *etree.Element, doc *doc.Document, sik *sentence.SegmentIdKeeper, ctx context.Context) (*TableNode, error) {
	return &TableNode{
		Elem: elem,
		Doc:  doc,
		Sik:  sik,
		ctx:  ctx,
	}, nil
}

func (n *TableNode) tableTooLarge() bool {
	tableText := n.Doc.GetRawDocText(n.Elem)
	htmlText := n.Doc.GetRawDocText(n.Doc.Doc.Root())
	if float32(len(tableText))/float32(len(htmlText)) >= 0.8 {
		return true
	}
	return false
}
func (n *TableNode) Match() bool {
	if !utils.Contains([]string{"table"}, n.Elem.Tag) {
		return false
	}
	if ContainsNestedTable(n.Elem) {
		return false
	}
	if n.tableTooLarge() {
		return false
	}

	return true
}

func (n *TableNode) Handle() ([]*wcd.AtomicText, error) {
	writer := &strings.Builder{}
	n.Elem.WriteTo(writer, &n.Doc.Doc.WriteSettings)
	tableHtml := writer.String()
	xpath := n.Doc.GetElemXpath(n.Elem)
	segmentId := n.Sik.GetSegmentdId(xpath)

	nodeText := fmt.Sprintf("%v %v", consts.TABLE_TAG_PREFIX, tableHtml)
	atom, err := n.Doc.ElemToAtom(n.Elem, nodeText, false, segmentId, true)
	if err != nil {
		hlog.CtxErrorf(n.ctx, "ElemToAtom failed, err: %v", err)
		return nil, err
	}

	return []*wcd.AtomicText{atom}, nil
}
