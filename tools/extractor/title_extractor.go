package extractor

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type TitleExtractor struct {
	Doc    *doc.Document
	result string
	ctx    context.Context
}

func NewTitleExtractor(ctx context.Context, doc *doc.Document) *TitleExtractor {
	return &TitleExtractor{
		Doc: doc,
		ctx: ctx,
	}
}

func (t *TitleExtractor) tryGetFirstLine(originTitle string) string {
	lines := strings.Split(originTitle, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func (t *TitleExtractor) tryUnEscapeGetFirstLine(originTitle string) (string, error) {
	unquote := originTitle
	if strings.Contains(originTitle, "\\n") {
		unquote = strings.ReplaceAll(originTitle, "\\n", "\n")
	}

	if firstLine := t.tryGetFirstLine(unquote); firstLine != "" {
		return firstLine, nil
	}
	return "", errors.New("no line found")
}

func (t *TitleExtractor) Extract() (string, error) {
	if t.result != "" {
		return t.result, nil
	}
	var title string
	var err error

	nodes := []func() (string, error){
		t.bySiteRule,
		// t.byHTagAntTitle,
		t.byTitle,
		t.byMeta,
		t.byHTag,
	}
	for i, node := range nodes {
		title, err = node()
		if err != nil {
			hlog.CtxErrorf(t.ctx, "extract title failed, err: %v", err)
			return "", err
		}
		title = strings.TrimSpace(title)
		if title != "" {
			if i != 0 {
				// 站点规则取出来的标题，不进行去噪处理
				title = t.removeUnNeedParts(title)
			}
			if firstLine, err := t.tryUnEscapeGetFirstLine(title); err == nil {
				return firstLine, nil
			}
			return title, nil
		}
	}
	return title, nil
}

func (t *TitleExtractor) bySiteRule() (string, error) {
	titleXpath := ""
	if t.Doc.Rule != nil {
		titleXpath = t.Doc.Rule.Title
	}
	if titleXpath == "" {
		return "", nil
	}
	title, err := ExtractByXpath(titleXpath, t.Doc)

	if err != nil {
		hlog.CtxErrorf(t.ctx, "extract title failed, err: %v", err)
		return "", err
	}
	return title, nil
}

func (t *TitleExtractor) removeUnNeedParts(title string) string {
	for {
		hasNoise := false
		for _, noisePair := range consts.NOISE_PAIRS {
			if strings.HasPrefix(title, noisePair.Left) && strings.HasSuffix(title, noisePair.Right) {
				hasNoise = true
				title = strings.TrimPrefix(title, noisePair.Left)
				title = strings.TrimSuffix(title, noisePair.Right)
				break
			}
		}
		if !hasNoise {
			break
		}
	}

	parts := regexp.MustCompile(consts.TITLE_SPLIT_CHAR_PATTERN).Split(title, -1)
	if len(parts) > 0 {
		partTitle := strings.Join(parts[:len(parts)-1], "")
		if float32(len(partTitle))/float32(len(title)) > consts.LCS_TITLE_RATIO {
			return partTitle
		}
	}
	return title
}

func (t *TitleExtractor) byTitle() (string, error) {
	title, err := ExtractByXpath("//title", t.Doc)
	if err != nil {
		hlog.CtxErrorf(t.ctx, "extract title failed, err: %v", err)
		return "", err
	}
	return title, nil
}

func (t *TitleExtractor) byHTag() (string, error) {
	xpath := strings.Join(
		utils.Map(utils.Range(0, 6, 1), func(t int) string {
			return fmt.Sprintf("//h%v", t+1)
		}),
		" | ",
	)
	title, err := ExtractByXpath(xpath, t.Doc)
	if err != nil {
		return "", err
	}
	return title, nil
}

func (t *TitleExtractor) byHTagAntTitle() (string, error) {
	titleText, err := t.byTitle()
	if err != nil {
		return "", err
	}
	if titleText == "" {
		return "", nil
	}
	titleTextRaw := titleText

	xpath := strings.Join(utils.Map(utils.Range(0, 6, 1), func(t int) string {
		return fmt.Sprintf("//h%v", t+1)
	}), " | ")
	result := ""
	for _, elem := range t.Doc.Xpath(xpath) {
		elemText := t.Doc.GetRawDocText(elem)
		lcs := utils.GetLcsString(titleTextRaw, elemText)
		if len(lcs) > len(result) {
			result = lcs
		}
	}
	if float32(len(result))/float32(len(titleTextRaw)) >= consts.LCS_TITLE_RATIO {
		return result, nil
	}
	return titleTextRaw, nil
}

func (t *TitleExtractor) byMeta() (string, error) {
	me := NewMetaExtractor(t.ctx, t.Doc)
	metaContent := me.Extract()

	for _, key := range consts.TITLE_META_KEYS {
		if val, ok := metaContent[key]; ok {
			return val, nil
		}
	}
	return "", nil

}
