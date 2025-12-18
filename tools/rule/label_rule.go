package rule

import (
	"errors"
	"strings"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
)

type LabelRuleExecutor struct {
}

func NewLabelRuleExecutor() *LabelRuleExecutor {
	return &LabelRuleExecutor{}
}

func (l *LabelRuleExecutor) Execute(
	elem *etree.Element,
	rule consts.LabelRule,
	tail bool,
	text string,
) {
	if rule.CleanTag {
		if tail && elem.Tail() != "" {
			if strings.Contains(text, elem.Tail()) {
				elem.SetTail("")
			} else {
				elem.SetTail(strings.ReplaceAll(elem.Tail(), text, ""))
			}
		} else if elem.Text() != "" {
			if strings.Contains(text, elem.Text()) {
				elem.SetText("")
			} else {
				elem.SetText(strings.ReplaceAll(elem.Text(), text, ""))
			}
		}
		if utils.Contains([]string{"img", "svg", "table"}, elem.Tag) {
			l.dropNoiseImg(elem)
		}
	}
	if rule.NewTag != "" {
		elem.Tag = rule.NewTag
	}
	if rule.NewAttr != "" {
		elem.CreateAttr(rule.NewAttr, "")
	}
}

func (l *LabelRuleExecutor) dropNoiseImg(elem *etree.Element) error {
	parent := elem.Parent()
	if parent == nil {
		return errors.New("parent is nil")
	}
	parent.RemoveChild(elem)
	return nil
}
