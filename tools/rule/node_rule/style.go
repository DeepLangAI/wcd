package node_rule

import (
	"strings"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
)

type StyleNodeRule struct {
	Xpath string
}

func NewStyleNodeRule() *StyleNodeRule {
	return &StyleNodeRule{
		//Xpath: "//*[re:match(@style, \".+\")]",
		Xpath: "//*[string-length(@style) > 0]",
	}
}

func (r *StyleNodeRule) removeListStyle(tagName string, styleDict map[string]string) map[string]string {
	if tagName != "ol" {
		delete(styleDict, "list-style")
	}
	return styleDict
}

func (r *StyleNodeRule) removeExtremeBg(styleDict map[string]string) map[string]string {
	bg := styleDict["background"]
	if bg == "" {
		bg = styleDict["background-color"]
	}

	if bg != "" {
		// 如果背景色是白色，则删除背景色
		// 如果背景色是深色，则删除背景色
		if utils.IsWhiteColor(bg) || utils.IsDarkColor(bg) {
			delete(styleDict, "background")
			delete(styleDict, "background-color")
		}
	}
	return styleDict
}

func (r *StyleNodeRule) removeWhiteColor(styleDict map[string]string) map[string]string {
	color := styleDict["color"]

	if color != "" {
		// 如果文字是白色，则删除
		if utils.IsWhiteColor(color) {
			delete(styleDict, "color")
		}
	}
	return styleDict
}

func (r *StyleNodeRule) Act(elem *etree.Element) {
	style := elem.SelectAttrValue("style", "")
	if style == "" {
		return
	}
	styleDict := utils.GetStyleMap(style)
	reservedAttrs := []string{
		"color", "font-style", "background", "background-color",
		"list-style",
		"font-weight", // bold加粗
		"font-style",  // italic斜体
		"text-indent", // 首行缩进
		"border-left",
		"border-radius",
		//"padding-top",
		//"padding-bottom",
		"padding-right",
		"padding-left",
	}
	imgStyleAttrs := []string{
		"width", // 宽高，用于图片等
		"height",
		"vertical-align", // svg图片垂直方向位置
	}
	if _, ok := consts.IMG_TAGS[elem.Tag]; ok {
		reservedAttrs = append(reservedAttrs, imgStyleAttrs...)
	}
	newStyleDict := utils.DictFromItems(
		utils.Filter(
			utils.Map(
				reservedAttrs,
				func(key string) utils.DictItem[string, string] {
					return utils.DictItem[string, string]{
						Key:   key,
						Value: strings.TrimSpace(strings.Replace(styleDict[key], "!important", "", -1)),
					}
				}),
			func(item utils.DictItem[string, string]) bool {
				return item.Value != ""
			},
		),
	)
	// 如果宽度是100%，则删除宽度
	if newStyleDict["width"] == "100%" {
		delete(newStyleDict, "width")
	}
	newStyleDict = r.removeExtremeBg(newStyleDict)
	newStyleDict = r.removeWhiteColor(newStyleDict)
	newStyleDict = r.removeListStyle(elem.Tag, newStyleDict)

	style = utils.StyleMapToString(newStyleDict)
	cleanElemAttr(elem)
	elem.CreateAttr("style", style)
}
func (r *StyleNodeRule) GetXpath() string {
	return r.Xpath
}
