package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/tools/node"
	"github.com/DeepLangAI/wcd/tools/sentence"
	"github.com/DeepLangAI/wcd/utils"

	"github.com/beevik/etree"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type Splitter struct {
	Doc *doc.Document
	So  *sentence.SentenceOp
	Sik *sentence.SegmentIdKeeper
	ctx context.Context
}

func NewSplitter(ctx context.Context, doc *doc.Document) *Splitter {
	splitter := &Splitter{
		Doc: doc,
		So:  sentence.NewSentenceOp(),
		Sik: sentence.NewSegmentIdKeeper(),
		ctx: ctx,
	}
	return splitter
}

func (s *Splitter) HandleSpecial(elem *etree.Element) ([]*wcd.AtomicText, error) {
	imageNode, err := node.NewImageNode(elem, s.Doc, s.Sik, s.ctx)
	if err != nil {
		return nil, err
	}
	tableNode, err := node.NewTableNode(elem, s.Doc, s.Sik, s.ctx)
	if err != nil {
		return nil, err
	}
	videoNode, err := node.NewVideoNode(elem, s.Doc, s.Sik, s.ctx)
	if err != nil {
		return nil, err
	}
	var specialNodes = []node.BaseNode{imageNode, tableNode, videoNode}
	for _, node := range specialNodes {
		if node.Match() {
			atoms, err := node.Handle()
			if err != nil {
				hlog.CtxErrorf(s.ctx, "handle special node error: %v", err)
				return nil, err
			}
			return atoms, nil
		}
	}
	return nil, nil
}

func (s *Splitter) addArticleMeta(atoms []*wcd.AtomicText, articleMeta *wcd.ArticleMeta) []*wcd.AtomicText {
	if articleMeta.Title != "" {
		titleXpath := "//title"
		if rule := s.Doc.Rule; rule != nil && rule.Title != "" {
			titleXpath = s.Doc.Rule.Title
		}
		atoms = append(atoms, &wcd.AtomicText{
			Text:       articleMeta.Title,
			PositionID: 0, // 虚拟节点，没有实际位置
			Xpath:      titleXpath,
			Tags:       utils.XpathToTags(titleXpath),
			SegmentID:  0, // 虚拟节点，没有实际位置
		})
	}
	return atoms
}

func (s *Splitter) SplitWithSpecial(
	articleMeta *wcd.ArticleMeta,
	specialHandler func(element *etree.Element) ([]*wcd.AtomicText, error),
	handlerAfterSplit func([]*wcd.AtomicText) ([]*wcd.AtomicSentence, error),
) ([]*wcd.AtomicSentence, error) {
	atoms := []*wcd.AtomicText{}
	atoms = s.addArticleMeta(atoms, articleMeta)

	tailStack := utils.Stack[*wcd.AtomicText]{}
	prefixTree := utils.NewPrefixTree()
	s.Doc.Traverse(s.Doc.Doc.Root(), &doc.TraverseParams{
		ElementFunc: func(elem *etree.Element) {
			xpath := s.Doc.GetElemXpath(elem)
			for !tailStack.IsEmpty() {
				peek, _ := tailStack.Peek()
				if !strings.HasPrefix(xpath, peek.Xpath) {
					tailStack.Pop()
					atoms = append(atoms, peek)
				} else {
					break
				}
			}
			if prefixTree.HasPrefixOf(xpath) {
				return
			}
			if specialAtoms, err := specialHandler(elem); err == nil && specialAtoms != nil {
				atoms = append(atoms, specialAtoms...)
				prefixTree.Insert(xpath)
			} else if elem.Text() != "" {
				if atom := s.ElemToAtom(elem, false); atom != nil {
					atoms = append(atoms, atom)
				}
			}
			if elem.Tail() != "" {
				if atom := s.ElemToAtom(elem, true); atom != nil {
					tailStack.Push(atom)
				}
			}
		},
	})

	for !tailStack.IsEmpty() {
		tail, _ := tailStack.Pop()
		atoms = append(atoms, tail)
	}
	return handlerAfterSplit(atoms)
}

func (s *Splitter) Split(articleMeta *wcd.ArticleMeta) ([]*wcd.AtomicSentence, error) {
	return s.SplitWithSpecial(articleMeta, s.HandleSpecial, func(atoms []*wcd.AtomicText) ([]*wcd.AtomicSentence, error) {
		sentences := s.CutSentences(atoms)
		if len(sentences) == 0 {
			return nil, fmt.Errorf("no sentences found")
		}
		sentences = s.cleanDuplicatedImage(sentences)
		err := s.SplitDomWithVnode(sentences)
		if err != nil {
			return nil, err
		}
		hlog.CtxInfof(s.ctx, "[splitter] split done, num atoms: %v, num sentences: %v", len(atoms), len(sentences))
		return sentences, nil
	})
}

func (s *Splitter) sentenceIsImage(sentence *wcd.AtomicSentence) bool {
	return sentence.Meta.URL != ""
}

func (s *Splitter) cleanDuplicatedImage(sents []*wcd.AtomicSentence) []*wcd.AtomicSentence {
	// 去除连续且URL相同的图片
	newSents := make([]*wcd.AtomicSentence, 0, len(sents))
	for _, sent := range sents {
		if len(newSents) == 0 {
			newSents = append(newSents, sent)
		} else {
			if !s.sentenceIsImage(sent) || !s.sentenceIsImage(newSents[len(newSents)-1]) {
				newSents = append(newSents, sent)
			} else {
				lastSent := newSents[len(newSents)-1]
				if lastSent.Meta.URL == sent.Meta.URL {
					pid := sent.Atoms[0].PositionID
					xpath := utils.GetPositionIdXpath(pid)
					s.Doc.RemoveByXpath(xpath)
				} else {
					newSents = append(newSents, sent)
				}
			}
		}
	}
	return newSents
}
func (s *Splitter) getAtomWindows(sentences []*wcd.AtomicSentence) ([][]*wcd.AtomicText, error) {
	windows := [][]*wcd.AtomicText{}
	atoms := []*wcd.AtomicText{}
	for _, sentence := range sentences {
		for _, atom := range sentence.Atoms {
			atoms = append(atoms, atom)
		}
	}
	i := 0
	j := 0
	for i < len(atoms) {
		j = i
		for j < len(atoms) && atoms[j].Xpath == atoms[i].Xpath && atoms[j].Tail == atoms[i].Tail {
			j += 1
		}
		if j-i >= 1 { // 所有文本都挂到vnode下
			windows = append(windows, atoms[i:j])
		}
		if j == i {
			i += 1
		} else {
			i = j
		}
	}
	return windows, nil
}

func (s *Splitter) addTextVnode(window []*wcd.AtomicText) {
	xpath := fmt.Sprintf("//*[@%v='%v']", consts.KeyPositionId, window[0].PositionID)
	elems := s.Doc.Xpath(xpath)
	if len(elems) == 0 {
		return
	}
	elem := elems[0]
	elem.SetText("")
	vnodeGroup := etree.Element{
		Tag: consts.VNODE_TAG_PREFIX + "pos-group",
		Attr: []etree.Attr{
			{
				Key:   consts.KeyPositionId,
				Value: fmt.Sprintf("%v", s.Doc.IncreasePositionId()),
			},
		},
	}
	for _, atom := range window {
		vnode := etree.Element{
			Tag: consts.VNODE_TAG_PREFIX + "pos",
			Attr: []etree.Attr{
				{
					Key:   consts.KeyPositionId,
					Value: fmt.Sprintf("%v", s.Doc.IncreasePositionId()),
				},
			},
			Child: nil,
		}
		// 要先addChild，再SetTail，否则tail内容会被清空。可能是etree的bug
		vnodeGroup.AddChild(&vnode)
		vnode.SetText(atom.Text)

		atom.Tail = false
		atom.PositionID = int32(s.Doc.MaxPositionId)
	}
	// 添加到父节点的第一个子节点之前
	elem.InsertChildAt(0, &vnodeGroup)
}

func (s *Splitter) addTailVnode(window []*wcd.AtomicText) {
	xpath := fmt.Sprintf("//*[@%v='%v']", consts.KeyPositionId, window[0].PositionID)
	elems := s.Doc.Xpath(xpath)
	if len(elems) == 0 {
		return
	}
	elem := elems[0]
	elem.SetTail("")
	vnodeGroup := etree.Element{
		Tag: consts.VNODE_TAG_PREFIX + "pos-group",
		Attr: []etree.Attr{
			{
				Key:   consts.KeyPositionId,
				Value: fmt.Sprintf("%v", s.Doc.IncreasePositionId()),
			},
		},
	}
	for _, atom := range window {
		vnode := etree.Element{
			Tag: consts.VNODE_TAG_PREFIX + "pos",
			Attr: []etree.Attr{
				{
					Key:   consts.KeyPositionId,
					Value: fmt.Sprintf("%v", s.Doc.IncreasePositionId()),
				},
			},
			Child: nil,
		}
		// 要先addChild，再SetTail，否则tail内容会被清空。可能是etree的bug
		vnodeGroup.AddChild(&vnode)
		vnode.SetText(atom.Text)

		atom.Tail = false
		atom.PositionID = int32(s.Doc.MaxPositionId)
	}
	// 添加到elem之后，作为邻居节点
	idx := elem.Index()
	elem.Parent().InsertChildAt(idx+1, &vnodeGroup)
}

func (s *Splitter) addTextTailVnode(atoms []*wcd.AtomicText) {
	// 给既有文本节点，又有尾节点的情况，添加尾节点的虚拟定位节点
	pidCounter := map[int32]int{}
	for _, atom := range atoms {
		pidCounter[atom.PositionID] = pidCounter[atom.PositionID] + 1
	}
	for _, atom := range atoms {
		if pidCounter[atom.PositionID] > 1 && atom.Tail {
			s.addTailVnode([]*wcd.AtomicText{atom})
		}
	}
}

func (s *Splitter) isSpecialNode(atom *wcd.AtomicText) bool {
	if strings.HasPrefix(atom.Text, consts.IMG_TAG_PREFIX) || strings.HasPrefix(atom.Text, consts.TABLE_TAG_PREFIX) {
		return true
	}

	parents := utils.SplitXpath(atom.Xpath)
	if utils.Contains(parents, "math") {
		return true
	}

	return false
}
func (s *Splitter) SplitDomWithVnode(sentences []*wcd.AtomicSentence) error {
	atomWindows, err := s.getAtomWindows(sentences)
	if err != nil {
		return err
	}
	for _, window := range atomWindows {
		window = utils.Filter(window, func(text *wcd.AtomicText) bool {
			return !s.isSpecialNode(text)
		})
		if len(window) == 0 {
			continue
		}
		if window[0].Tail {
			s.addTailVnode(window)
		} else {
			s.addTextVnode(window)
		}
	}
	//atoms := []*wcd.AtomicText{}
	//for _, sentence := range sentences {
	//	atoms = append(atoms, sentence.Atoms...)
	//}
	//s.addTextTailVnode(atoms)
	return nil
}

func (s *Splitter) CutSentences(atoms []*wcd.AtomicText) []*wcd.AtomicSentence {
	result := make([]*wcd.AtomicSentence, 0)
	paragraph := []*wcd.AtomicText{}
	atomOperator := sentence.AtomOperator{}
	for _, atom := range atoms {
		if utils.Any([]string{consts.IMG_TAG_PREFIX, consts.TABLE_TAG_PREFIX}, func(prefix string) bool {
			return strings.HasPrefix(atom.Text, prefix)
		}) {
			result = append(result, s.cutParagraph(paragraph)...)
			paragraph = []*wcd.AtomicText{}
			result = append(result, atomOperator.AtomToSentence(atom))
			continue
		}
		if len(paragraph) == 0 {
			paragraph = append(paragraph, atom)
		} else {
			if atomOperator.IsSameParagraph(paragraph[len(paragraph)-1], atom) {
				paragraph = append(paragraph, atom)
			} else {
				result = append(result, s.cutParagraph(paragraph)...)
				paragraph = []*wcd.AtomicText{}
				paragraph = append(paragraph, atom)
			}
		}
	}
	if len(paragraph) > 0 {
		result = append(result, s.cutParagraph(paragraph)...)
	}
	return result
}

func (s *Splitter) cutParagraph(atoms []*wcd.AtomicText) []*wcd.AtomicSentence {
	atomOperator := sentence.AtomOperator{}

	// 拼接所有原子文本
	joinedStr := strings.Join(utils.Map(atoms, func(atom *wcd.AtomicText) string {
		return atom.Text
	}), "")

	// 计算每个原子文本在拼接字符串中的位置
	intervals := make([]sentence.Interval, len(atoms))
	start := 0
	for i, atom := range atoms {
		end := start + len(atom.Text)
		intervals[i] = sentence.Interval{
			Start: start,
			End:   end - 1,
			ID:    i,
		}
		start = end // 更新起始位置
	}

	// 使用分句方法分割文本
	sents := s.So.Cut(joinedStr)
	atomSentences := []*wcd.AtomicSentence{}

	for _, sent := range sents {
		index, success := sentence.FindInterval(sent.Start, intervals)
		if !success {
			hlog.CtxErrorf(s.ctx, "cannot find interval for sent:%v", sent)
			return nil
		}

		// 创建新的句子
		lastSent := atomOperator.NewEmptyAtomSentence()
		for lastAtomIndex := index; len(lastSent.Text) < len(sent.Text); lastAtomIndex++ {
			if len(lastSent.Text)+len(atoms[lastAtomIndex].Text) > len(sent.Text) {
				atomA, atomB := atomOperator.AtomSplit(atoms[lastAtomIndex], len(sent.Text)-len(lastSent.Text))
				lastSent = atomOperator.SentenceAdd(lastSent, atomA)
				if atomB != nil {
					atoms[lastAtomIndex] = atomB
				}
			} else {
				lastSent = atomOperator.SentenceAdd(lastSent, atoms[lastAtomIndex])
			}
		}
		atomSentences = append(atomSentences, lastSent)
	}

	atomSentences = utils.Filter(atomSentences, func(atomicSentence *wcd.AtomicSentence) bool {
		return strings.TrimSpace(atomicSentence.Text) != ""
	})
	return atomSentences
}

func (s *Splitter) ElemToAtom(elem *etree.Element, tail bool) *wcd.AtomicText {
	var err error

	xpath := s.Doc.GetElemXpath(elem)
	segmentId := s.Sik.GetSegmentdId(xpath)
	text := elem.Text()
	if tail {
		text = elem.Tail()
	}
	if strings.TrimSpace(text) == "" {
		return nil
	}

	atom, err := s.Doc.ElemToAtom(elem, text, tail, segmentId, false)
	if err != nil {
		hlog.CtxErrorf(s.ctx, "ElemToAtom error:%v", err)
		return nil
	}
	return atom
}
