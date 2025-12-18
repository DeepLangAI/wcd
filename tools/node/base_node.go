package node

import (
	"fmt"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/beevik/etree"
)

type BaseNode interface {
	Match() bool
	Handle() ([]*wcd.AtomicText, error)
}

func containsNested(elem *etree.Element, tags []string) bool {
	for _, tag := range tags {
		xpath := fmt.Sprintf(".//%s", tag)
		if len(elem.FindElements(xpath)) != 0 {
			return true
		}
	}
	return false
}
func ContainsNestedTable(elem *etree.Element) bool {
	return containsNested(elem, consts.TABLE_TAGS)
}

func ContainsNestedParagraph(elem *etree.Element) bool {
	return containsNested(elem, consts.PARAGRAPH_TAGS)
}
