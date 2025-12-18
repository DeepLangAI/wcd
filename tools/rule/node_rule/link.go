package node_rule

import (
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
)

type LinkNodeRule struct {
	Xpath string
	Url   string
}

func NewLinkNodeRule(url string) *LinkNodeRule {
	return &LinkNodeRule{
		Xpath: "//a",
		Url:   url,
	}
}

func (r *LinkNodeRule) Act(elem *etree.Element) {
	//	只保留a标签的href属性
	oldHref := elem.SelectAttrValue("href", "")
	cleanElemAttr(elem)
	link := utils.EnsureLinkAbsolute(oldHref, r.Url)
	elem.CreateAttr("href", link)
}
func (r *LinkNodeRule) GetXpath() string {
	return r.Xpath
}
