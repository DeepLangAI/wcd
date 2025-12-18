package node_rule

import "github.com/beevik/etree"

type DivNodeRule struct {
	Xpath string
}

func NewDivNodeRule() *DivNodeRule {
	return &DivNodeRule{
		Xpath: "//div | //section",
	}
}

func (r *DivNodeRule) Act(elem *etree.Element) {
	//	删除style属性
	cleanElemAttr(elem)
}

func (r *DivNodeRule) GetXpath() string {
	return r.Xpath
}
