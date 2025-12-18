package node

import (
	"github.com/DeepLangAI/wcd/biz/model/wcd"
	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
)

type ParagraphNode struct {
	Elem *etree.Element
	Doc  *doc.Document
}

func (n *ParagraphNode) Match() bool {
	if !utils.Contains(consts.PARAGRAPH_TAGS, n.Elem.Tag) {
		return false
	}
	if ContainsNestedParagraph(n.Elem) {
		return false
	}
	return true
}

func (n *ParagraphNode) Handle() ([]*wcd.AtomicText, error) {
	return nil, nil
}
