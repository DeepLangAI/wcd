package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/tools/formatter"
	"github.com/DeepLangAI/wcd/tools/rule"
	"github.com/DeepLangAI/wcd/tools/rule/node_rule"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/beevik/etree"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type Formatter struct {
	ctx    context.Context
	doc    *doc.Document
	labels []*wcd.TextParseLabelSentence
}

func NewFormatter(ctx context.Context, doc *doc.Document, labels []*wcd.TextParseLabelSentence) *Formatter {
	formatter := &Formatter{
		ctx:    ctx,
		doc:    doc,
		labels: labels,
	}
	return formatter
}

func (f *Formatter) PreFormat() {
	paths := []func(){
		f.formatByNodeRules,
		f.tireHeadings,
	}
	for _, path := range paths {
		t := time.Now()
		path()
		delta := time.Since(t)
		hlog.CtxInfof(f.ctx, "[formatter pre-format] format path: %v, cost: %s", getFunctionName(path), delta.String())
	}
}
func (f *Formatter) PostFormat() {
	paths := []func(){}
	if rule := f.doc.Rule; rule != nil && rule.NoSemanticDenoise {
		// 无须语义去噪。因为语义去噪的标签不准
		paths = []func(){
			f.formatByStyle,
		}
	} else {
		paths = []func(){
			f.formatByLabelRules,
			f.formatBySubTree,
			f.formatByStyle,
		}
	}
	for _, path := range paths {
		t := time.Now()
		path()
		delta := time.Since(t)
		hlog.CtxInfof(f.ctx, "[formatter post-format] format path: %v, cost: %s", getFunctionName(path), delta.String())
	}
}

// todo 降低耗时
func (f *Formatter) formatByLabelRules() {
	labelToRule := map[string]consts.LabelRule{}
	labelRuleExecutor := rule.NewLabelRuleExecutor()

	for _, rule := range consts.LABEL_RULES {
		labelToRule[rule.Label] = rule
	}
	for _, sentence := range f.labels {
		label := sentence.Label
		text := sentence.Text
		if _, ok := labelToRule[label]; !ok {
			continue
		}
		rule := labelToRule[label]
		for _, atom := range sentence.Atoms {
			xpath := utils.GetPositionIdXpath(atom.PositionID)
			elems := f.doc.Xpath(xpath)
			for _, elem := range elems {
				labelRuleExecutor.Execute(elem, rule, atom.Tail, text)
			}
		}
	}
}

// todo 降低耗时
func (f *Formatter) formatByNodeRules() {
	rules := []node_rule.NodeRule{
		node_rule.NewStyleNodeRule(),
		node_rule.NewLinkNodeRule(f.doc.Url),
		node_rule.NewDivNodeRule(),
		node_rule.NewImgNodeRule(f.doc.Url),
	}
	for _, rule := range rules {
		f.doc.XpathIter(rule.GetXpath(), func(elem *etree.Element) {
			rule.Act(elem)
		})
	}
}

func (f *Formatter) formatBySubTree() {
	formatter := formatter.NewSubtreeFormatter(f.ctx, f.doc, f.labels)
	formatter.Format()
}

func (f *Formatter) tireHeadings() {
	headingXpaths := strings.Join(
		utils.Map(utils.Range(1, 7, 1), func(i int) string {
			return fmt.Sprintf("//h%d", i)
		}),
		"",
	)
	elems := f.doc.Xpath(headingXpaths)
	sort.Slice(elems, func(i, j int) bool {
		return elems[i].Tag[1:] < elems[j].Tag[1:]
	})
	baseLevel := 1
	for i, elem := range elems {
		if i == 0 {
			elem.Tag = fmt.Sprintf("h%d", baseLevel)
		} else {
			if elem.Tag[1:] != elems[i-1].Tag[1:] {
				baseLevel += 1
			}
		}
		elem.Tag = fmt.Sprintf("h%d", baseLevel)
	}

}

func (f *Formatter) formatByStyle() {
	styleFormatter := formatter.NewStyleFormatter(f.ctx, f.doc)
	f.doc.ResetHtml(f.doc.Doc)
	styleFormatter.Format()
}
