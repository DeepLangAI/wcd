package sentence

import (
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/utils"
)

type AtomOperator struct {
}

func (op *AtomOperator) NewAtom(
	text string,
	positionId int32,
	xpath string,
	immutable bool,
	tags []string,
	tail bool,
	segmentId int32,
	attrs map[string]string,
) *wcd.AtomicText {
	t := &wcd.AtomicText{
		Text:       text,
		PositionID: positionId,
		Xpath:      xpath,
		Immutable:  immutable,
		Tags:       tags,
		Tail:       tail,
		SegmentID:  segmentId,
		Attrs:      attrs,
	}
	return t
}

func (op *AtomOperator) NewEmptyAtomSentence() *wcd.AtomicSentence {
	return &wcd.AtomicSentence{
		Text:  "",
		Atoms: make([]*wcd.AtomicText, 0),
		Meta:  &wcd.SentenceMeta{},
	}
}

func (op *AtomOperator) IsSameParagraph(a, b *wcd.AtomicText) bool {
	return utils.Pdepth(a.Xpath) == utils.Pdepth(b.Xpath)
}

func (op *AtomOperator) SentenceAdd(a *wcd.AtomicSentence, b *wcd.AtomicText) *wcd.AtomicSentence {
	segmentId := a.SegmentID
	if segmentId == 0 {
		// 处理a是空句子的情况
		segmentId = b.SegmentID
	}
	return &wcd.AtomicSentence{
		Text:      utils.Join(a.Text, b.Text),
		Atoms:     append(a.Atoms, b),
		SegmentID: segmentId,
		Meta:      a.Meta,
		Tags:      utils.Set(append(a.Tags, b.Tags...)),
	}
}
func (op *AtomOperator) AtomToSentence(a *wcd.AtomicText) *wcd.AtomicSentence {
	url := ""
	tableHtml := ""
	if strings.HasPrefix(a.Text, consts.IMG_TAG_PREFIX) {
		url = strings.TrimSpace(strings.TrimPrefix(a.Text, consts.IMG_TAG_PREFIX))
	} else if strings.HasPrefix(a.Text, consts.TABLE_TAG_PREFIX) {
		tableHtml = strings.TrimSpace(strings.TrimPrefix(a.Text, consts.TABLE_TAG_PREFIX))
	}

	meta := &wcd.SentenceMeta{
		URL:       url,
		TableHTML: tableHtml,
	}
	return &wcd.AtomicSentence{
		Text:      a.Text,
		Atoms:     []*wcd.AtomicText{a},
		SegmentID: a.SegmentID,
		Meta:      meta,
		Tags:      a.Tags,
	}
}
func (op *AtomOperator) AtomAdd(a, b *wcd.AtomicText) *wcd.AtomicSentence {
	return &wcd.AtomicSentence{
		Text:      utils.Join(a.Text, b.Text),
		Atoms:     []*wcd.AtomicText{a, b},
		SegmentID: a.SegmentID,
	}
}
func (op *AtomOperator) AtomSplit(a *wcd.AtomicText, index int) (*wcd.AtomicText, *wcd.AtomicText) {
	if index < 0 || index >= len(a.Text) {
		return a, nil
	}
	s1, s2 := a.Text[:index], a.Text[index:]
	a = &wcd.AtomicText{
		Text:       s1,
		PositionID: a.PositionID,
		Xpath:      a.Xpath,
		Immutable:  a.Immutable,
		Tags:       a.Tags,
		Tail:       a.Tail,
		SegmentID:  a.SegmentID,
		Attrs:      a.Attrs,
	}
	b := &wcd.AtomicText{
		Text:       s2,
		PositionID: a.PositionID,
		Xpath:      a.Xpath,
		Immutable:  a.Immutable,
		Tags:       a.Tags,
		Tail:       a.Tail,
		SegmentID:  a.SegmentID,
		Attrs:      a.Attrs,
	}
	return a, b
}
