package node_rule

import (
	"strings"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
)

func cleanElemAttr(elem *etree.Element) {
	reserveKeys := []string{
		consts.KeyXpath, consts.KeyPositionId,
		"viewBox", // svg图片的属性，移除了会无法显示
		"style",   // 保留图片style，不然宽高等会异常
		"height",  // 保留svg图片高度
	}
	reserveKeys = append(reserveKeys, consts.IMG_ATTRS...)
	reserveKeys = append(reserveKeys, consts.A_ATTRS...)

	elem.Attr = utils.Filter(elem.Attr, func(attr etree.Attr) bool {
		return utils.Contains(reserveKeys, attr.Key) || strings.Contains(strings.ToLower(attr.Key), "deeplang")
	})
}

type NodeRule interface {
	Act(elem *etree.Element)
	GetXpath() string
}
