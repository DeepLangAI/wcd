package formatter

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/tools/rule"
	"github.com/DeepLangAI/wcd/utils"

	"github.com/beevik/etree"
)

type SubtreeFormatter struct {
	ctx         context.Context
	doc         *doc.Document
	labels      []*wcd.TextParseLabelSentence
	labelToRule map[string]consts.SubtreeLabelRule
}

func NewSubtreeFormatter(ctx context.Context, doc *doc.Document, labels []*wcd.TextParseLabelSentence) *SubtreeFormatter {
	labelToRule := map[string]consts.SubtreeLabelRule{}
	for _, rule := range consts.SUBTREE_RULES {
		labelToRule[rule.Label] = rule
	}
	return &SubtreeFormatter{
		ctx:         ctx,
		doc:         doc,
		labels:      labels,
		labelToRule: labelToRule,
	}
}
func (f *SubtreeFormatter) getElemOrigXpath(elem *etree.Element) string {
	origXpath := f.doc.GetElemXpath(elem)
	noiseRe := regexp.MustCompile(fmt.Sprintf("%v.*", consts.VNODE_TAG_PREFIX))
	return strings.TrimRight(noiseRe.ReplaceAllString(origXpath, ""), "/")
}

func (f *SubtreeFormatter) mergeParagraph(paragraph []*etree.Element) []*etree.Element {
	// 尝试直接找到当前段落的公共父节点
	paragraphContent := utils.RemoveSpace(strings.Join(utils.Map(paragraph, func(elem *etree.Element) string {
		return f.doc.GetRawDocText(elem)
	}), ""))

	for parent := paragraph[0].Parent(); parent != nil; parent = parent.Parent() {
		parentContent := utils.RemoveSpace(f.doc.GetRawDocText(parent))
		rate := float32(len(paragraphContent)) / float32(len(parentContent))
		if 0.95 <= rate && rate <= 1.05 {
			return []*etree.Element{parent}
		}
		if rate < 0.9 {
			return paragraph
		}
	}
	return paragraph
}
func (f *SubtreeFormatter) tryFindTopParent(joinedText string, initialElem *etree.Element) *etree.Element {
	for parent := initialElem.Parent(); parent != nil; parent = parent.Parent() {
		parentContent := utils.RemoveSpace(f.doc.GetRawDocText(parent))
		rate := float32(len(joinedText)) / float32(len(parentContent))
		if 0.97 <= rate && rate <= 1.03 {
			return parent
		}
		if rate < 0.9 {
			return nil
		}
	}
	return nil
}

func (f *SubtreeFormatter) getFormatWindows() [][]*wcd.TextParseLabelSentence {
	window := []*wcd.TextParseLabelSentence{}
	windows := [][]*wcd.TextParseLabelSentence{}
	for _, sentence := range f.labels {
		if rule, ok := f.labelToRule[sentence.Label]; ok {
			if len(window) != 0 {
				if window[0].Label != sentence.Label {
					windows = append(windows, window)
					window = []*wcd.TextParseLabelSentence{}
				} else if window[0].SegmentID != sentence.SegmentID {
					// 如果要求只能合并同一段落的，则遇到不同段落的，则窗口结束
					if rule.SameParagraph {
						windows = append(windows, window)
						window = []*wcd.TextParseLabelSentence{}
					}
				}
			}
			//if len(window) != 0 && (window[0].Label != sentence.Label || window[0].SegmentID != sentence.SegmentID) {
			//	windows = append(windows, window)
			//	window = []*wcd.TextParseLabelSentence{}
			//}
			window = append(window, sentence)
		} else {
			if len(window) > 0 {
				windows = append(windows, window)
				window = []*wcd.TextParseLabelSentence{}
			}
		}
	}
	if len(window) > 0 {
		windows = append(windows, window)
	}
	return windows
}

func (f *SubtreeFormatter) formatWindow(window []*wcd.TextParseLabelSentence) error {
	formatRule := f.labelToRule[window[0].Label]

	pids := []int32{}
	pidToSegmentId := map[string]int32{}
	for _, sentence := range window {
		for _, atom := range sentence.Atoms {
			pids = append(pids, atom.PositionID)
			pidToSegmentId[fmt.Sprintf("%v", atom.PositionID)] = atom.SegmentID
		}
	}
	xpath := strings.Join(
		utils.Map(pids, func(pid int32) string {
			return utils.GetPositionIdXpath(pid)
		}), "|",
	)
	elems := f.doc.Xpath(xpath)
	if len(elems) == 0 {
		// 添加的全文标题可能是虚拟的，没有对应的元素
		return nil
	}
	// 如果能找到一个父节点，**正好包含**当前窗口若干段落的所有内容，直接修改这个父节点的属性即可
	if topElem := f.tryFindTopParent(utils.RemoveSpace(strings.Join(utils.Map(window, func(s *wcd.TextParseLabelSentence) string {
		return s.Text
	}), "")), elems[0]); topElem != nil {
		mostTop := f.getMostTopElem(topElem)
		mostTop.CreateAttr(formatRule.NewAttr, "")
		return nil
	}
	sort.Slice(elems, func(i, j int) bool {
		pidI, err := strconv.ParseInt(elems[i].SelectAttrValue(consts.KeyPositionId, "0"), 10, 32)
		if err != nil {
			return false
		}
		pidJ, err := strconv.ParseInt(elems[j].SelectAttrValue(consts.KeyPositionId, "0"), 10, 32)
		if err != nil {
			return true
		}
		return pidI < pidJ
	})

	paragraphs := [][]*etree.Element{}
	paragraph := []*etree.Element{}

	lastSegmentId := int32(0)
	for _, elem := range elems {
		if len(paragraph) == 0 {
			topElem := f.getMostTopElem(elem)
			//topElem := elem
			paragraph = append(paragraph, topElem)
			if pid := elem.SelectAttrValue(consts.KeyPositionId, ""); pid != "" {
				lastSegmentId = pidToSegmentId[pid]
			}
		} else {
			currentSegmentId := int32(0)
			if pid := elem.SelectAttrValue(consts.KeyPositionId, ""); pid != "" {
				currentSegmentId = pidToSegmentId[pid]
			}
			if lastSegmentId != currentSegmentId {
				paragraphs = append(paragraphs, paragraph)
				paragraph = []*etree.Element{}
				lastSegmentId = currentSegmentId
			}
			topElem := f.getMostTopElem(elem)
			//topElem := elem
			paragraph = append(paragraph, topElem)
		}
	}
	if len(paragraph) > 0 {
		paragraphs = append(paragraphs, paragraph)
	}
	if len(paragraphs) <= 0 {
		return nil
	}
	paragraphs = utils.Map(paragraphs, f.mergeParagraph)

	firstTopElem := paragraphs[0][0]
	//firstParent := firstTopElem.Parent()
	//firstIndex := firstTopElem.Index()

	firstParents := []*etree.Element{}
	firstIndexs := []int{}
	for p := firstTopElem; p != nil; p = p.Parent() {
		firstParents = append(firstParents, p)
		firstIndexs = append(firstIndexs, p.Index())
	}

	subtreeTag := "span"
	// 用自定义的节点会渲染错位，如p标签内不能有div节点。此处统一用span
	//if formatRule.NewTag != "" {
	//	subtreeTag = formatRule.NewTag
	//}
	subtreeVnode := etree.Element{
		Tag:   subtreeTag,
		Attr:  make([]etree.Attr, 0),
		Child: make([]etree.Token, 0),
	}
	subtreeVnode.CreateAttr(formatRule.NewAttr, "")
	subtreeVnode.CreateAttr(consts.KeyPositionId, fmt.Sprintf("%v", f.doc.IncreasePositionId()))
	subtreeVnode.CreateAttr(consts.KeySubtree, "")

	for paragraphIndex, paragraph := range paragraphs {
		uf := utils.NewUnionFind[int64]()
		pidToElem := map[int64]*etree.Element{}
		pidIndex := map[int64]int{}
		var mostTopElem *etree.Element
		var mostTopXpathDepth = 0
		for i, topElem := range paragraph {
			topElemPid := f.doc.GetElemPositionId(topElem)
			pidToElem[topElemPid] = topElem
			pidIndex[topElemPid] = i
			f.doc.Traverse(topElem, &doc.TraverseParams{Traversefunc: func(elem *etree.Element) {
				elemPid := f.doc.GetElemPositionId(elem)
				uf.Union(topElemPid, elemPid)
			}})

			origXpath := f.getElemOrigXpath(topElem)
			if mostTopXpathDepth == 0 {
				mostTopXpathDepth = len(strings.Split(origXpath, "/"))
				mostTopElem = topElem
			} else {
				if depth := len(strings.Split(origXpath, "/")); depth < mostTopXpathDepth {
					mostTopXpathDepth = depth
					mostTopElem = topElem
				}
			}
		}

		tagName := "span"
		vnode := etree.Element{
			Tag:   tagName,
			Attr:  make([]etree.Attr, 0),
			Child: make([]etree.Token, 0),
		}
		vnode.CreateAttr(consts.KeyPositionId, fmt.Sprintf("%v", f.doc.IncreasePositionId()))
		// 用display block来当作段落
		vnode.CreateAttr("style", "display: block; margin: 0;")

		roots := uf.GetRoots()
		// 按先后顺序排序，而不是按pid字面大小排序
		sort.Slice(roots, func(i, j int) bool {
			return pidIndex[roots[i]] < pidIndex[roots[j]]
		})
		// 如果同一个段落有多个句子，且都在同一个父节点下，则不进行合并，直接修改父节点的属性
		if len(roots) >= 2 {
			paragraphText := utils.RemoveSpace(strings.Join(utils.Map(roots, func(root int64) string {
				return f.doc.GetRawDocText(pidToElem[root])
			}), ""))
			parentText := utils.RemoveSpace(f.doc.GetRawDocText(mostTopElem.Parent()))
			if float32(len(paragraphText))/float32(len(parentText)) >= 0.95 {
				parent := f.getMostTopElem(mostTopElem.Parent())
				if len(paragraphs) == 1 {
					// 都在同一个父节点下，且只有一个段落，则直接修改父节点的属性，而不是创建新的节点
					parent.CreateAttr(formatRule.NewAttr, "")
					return nil
				}
				if paragraphIndex == 0 {
					firstParents = make([]*etree.Element, 0)
					firstIndexs = make([]int, 0)
					for p := parent; p != nil; p = p.Parent() {
						firstParents = append(firstParents, p)
						firstIndexs = append(firstIndexs, p.Index())
					}
				}

				parent.Parent().RemoveChildAt(parent.Index())
				vnode.Child = append(vnode.Child, parent)
				subtreeVnode.Child = append(subtreeVnode.Child, &vnode)
				continue
			}
		}
		sort.Slice(roots, func(i, j int) bool {
			return pidIndex[roots[i]] < pidIndex[roots[j]]
		})
		for _, pid := range roots {
			topElem := pidToElem[pid]
			//topElem.Parent().RemoveChild(topElem)
			topElem.Parent().RemoveChildAt(topElem.Index())
			vnode.Child = append(vnode.Child, topElem)
		}
		subtreeVnode.Child = append(subtreeVnode.Child, &vnode)
	}

	for i := 0; i < len(firstParents); i++ {
		firstParent := firstParents[i]
		//firstIndex := firstIndexs[i]
		if parentPid := firstParent.SelectAttrValue(consts.KeyPositionId, ""); parentPid != "" {
			if len(f.doc.Xpath(utils.GetPositionIdXpath(parentPid))) > 0 {
				childIndex := firstIndexs[i-1]
				firstParent.InsertChildAt(childIndex, &subtreeVnode)
				break
			}
		}
	}
	//firstParent.InsertChildAt(firstIndex, &subtreeVnode)
	// 避免p标签内不能有div节点，此处统一用span
	modifyTags := map[string]int{
		"p":   1,
		"div": 1,
		"ul":  1,
	}
	f.doc.Traverse(&subtreeVnode, &doc.TraverseParams{
		Traversefunc: func(elem *etree.Element) {
			if modifyTags[elem.Tag] == 1 {
				elem.Tag = "span"
			}
		},
	})

	// 避免原网页的样式被覆盖
	subtreeTop := f.getMostTopElem(&subtreeVnode)
	if f.doc.GetElemPositionId(subtreeTop) != f.doc.GetElemPositionId(&subtreeVnode) {
		//subtreeTop.CreateAttr(consts.KeySubtree, "")
		//subtreeVnode.RemoveAttr(consts.KeySubtree)

		subtreeTop.CreateAttr(formatRule.NewAttr, "")
		subtreeVnode.RemoveAttr(formatRule.NewAttr)
	}
	return nil
}

func (f *SubtreeFormatter) getMostTopElem(elem *etree.Element) *etree.Element {
	return f.doc.GetMostTopElem(elem)
}

func (f *SubtreeFormatter) formatAtom(sentence *wcd.TextParseLabelSentence) error {
	labelRuleExecutor := rule.NewLabelRuleExecutor()
	formatRule := f.labelToRule[sentence.Label]
	xpath := utils.GetPositionIdXpath(sentence.Atoms[0].PositionID)
	if elems := f.doc.Xpath(xpath); len(elems) > 0 {
		elem := f.getMostTopElem(elems[0])
		elem = f.getMostTopElem(elem)
		labelRuleExecutor.Execute(elem, formatRule.LabelRule, sentence.Atoms[0].Tail, sentence.Text)
	}
	return nil
}

func (f *SubtreeFormatter) Format() {
	windows := f.getFormatWindows()
	for _, window := range windows {
		numAtoms := 0
		for _, sentence := range window {
			numAtoms += len(sentence.Atoms)
		}
		if numAtoms >= 2 {
			f.formatWindow(window)
		} else if numAtoms == 1 {
			f.formatAtom(window[0])
		}
	}
}
