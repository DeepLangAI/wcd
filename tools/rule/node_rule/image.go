package node_rule

import (
	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
)

type ImgNodeRule struct {
	Xpath string
	Url   string
}

func NewImgNodeRule(url string) *ImgNodeRule {
	return &ImgNodeRule{
		Xpath: "//img",
		Url:   url,
	}
}

func (r *ImgNodeRule) Act(elem *etree.Element) {
	//只保留src、data-src属性
	type StrItem = utils.DictItem[string, string]

	attrs := utils.DictFromItems(
		utils.Filter(
			utils.Map(
				consts.IMG_ATTRS, // 链接属性
				func(key string) StrItem {
					return StrItem{Key: key, Value: elem.SelectAttrValue(key, "")}
				}),
			func(item StrItem) bool {
				return item.Value != ""
			},
		),
	)
	cleanElemAttr(elem)
	for key, val := range attrs {
		link := utils.EnsureLinkAbsolute(val, r.Url)
		elem.CreateAttr(key, link)
	}
}
func (r *ImgNodeRule) GetXpath() string {
	return r.Xpath
}
