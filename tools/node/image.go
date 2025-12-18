package node

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/tools/sentence"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
)

type ImageNode struct {
	Elem *etree.Element
	Doc  *doc.Document
	Sik  *sentence.SegmentIdKeeper
	ctx  context.Context
}

func NewImageNode(elem *etree.Element, doc *doc.Document, sik *sentence.SegmentIdKeeper, ctx context.Context) (*ImageNode, error) {
	n := &ImageNode{
		Elem: elem,
		Doc:  doc,
		Sik:  sik,
	}
	return n, nil
}

func (n *ImageNode) selectBgImgUrl() string {
	if style := n.Elem.SelectAttrValue(consts.StyleAttr, ""); style != "" {
		styleDict := utils.GetStyleMap(style)
		bg := styleDict["background-image"]
		re := regexp.MustCompile(`url\((.*)\)`)
		if match := re.FindStringSubmatch(bg); len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func (n *ImageNode) Match() bool {
	if consts.IMG_TAGS[n.Elem.Tag] == "" {
		if bgImgUrl := n.selectBgImgUrl(); bgImgUrl != "" {
			n.Elem.Tag = consts.TagNameImg
			n.Elem.CreateAttr("src", bgImgUrl)
			n.Elem.Child = nil
			return true
		} else {
			return false
		}
	}
	if utils.All(consts.IMG_ATTRS, func(key string) bool {
		value := n.Elem.SelectAttrValue(key, "")
		// 属性为空
		if value == "" {
			return true
		}
		// 属性不是链接
		if !n.attrIsLink(value) {
			return true
		}
		return false
	}) {
		return false
	}

	return true
}

func (n *ImageNode) attrIsLink(value string) bool {
	if strings.HasPrefix(value, "http") ||
		strings.HasPrefix(value, "//") ||
		strings.HasPrefix(value, "/") ||
		strings.HasPrefix(value, "://") ||
		strings.HasPrefix(value, "./") ||
		strings.HasPrefix(value, "../") {
		return true
	}
	return false
}

func (n *ImageNode) Handle() ([]*wcd.AtomicText, error) {
	url := n.Elem.Tag
	for _, key := range consts.IMG_ATTRS {
		if attr := n.Elem.SelectAttr(key); attr != nil {
			if n.attrIsLink(attr.Value) {
				url = attr.Value
				break
			}
		}
	}
	if url != n.Elem.Tag {
		// 解析过程中，将图片节点的图片链接转换为绝对链接，并仅保留src字段，防止解析过程中出现图片链接错误
		url = utils.EnsureLinkAbsolute(url, n.Doc.Url)
		for _, key := range consts.IMG_ATTRS {
			n.Elem.RemoveAttr(key)
		}
		attrs := utils.Filter(n.Elem.Attr, func(attr etree.Attr) bool {
			return !strings.Contains(attr.Value, "http")
		})
		n.Elem.Attr = attrs
		//for _, attr := range n.Elem.Attr {
		//	if strings.Contains(attr.Value, "http") {
		//		n.Elem.RemoveAttr(attr.Key)
		//	}
		//}
		n.Elem.CreateAttr("src", url)
	}

	nodeText := fmt.Sprintf("%v %v", consts.IMG_TAG_PREFIX, url)
	xpath := n.Doc.GetElemXpath(n.Elem)
	segmentId := n.Sik.GetSegmentdId(xpath)
	atom, err := n.Doc.ElemToAtom(n.Elem, nodeText, false, segmentId, true)
	if err != nil {
		return nil, err
	}
	return []*wcd.AtomicText{atom}, nil
}
