package tools

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/utils"

	"github.com/beevik/etree"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type Cleanner struct {
	Doc         *doc.Document
	CleaningDoc *doc.Document
	sentences   []*wcd.TextParseLabelSentence

	ctx context.Context
}

func NewCleaner(ctx context.Context, doc *doc.Document) *Cleanner {
	cleaner := &Cleanner{
		ctx: ctx,
		Doc: doc,
	}
	return cleaner
}
func (c *Cleanner) SetSentences(sentences []*wcd.TextParseLabelSentence) {
	c.sentences = sentences
}

func (c *Cleanner) Purify() error {
	htmlStr, err := c.Doc.ToString()
	if err != nil {
		hlog.CtxErrorf(c.ctx, "doc to string err: %v", err)
		return err
	}

	cleaningDoc := doc.LoadDocumentFromSegmentResult(c.ctx, htmlStr, c.Doc.Url, c.Doc.RuleStageGroup)
	if cleaningDoc == nil {
		return errors.New("cleaning doc is nil")
	}
	c.CleaningDoc = cleaningDoc

	// 以后可以尽量减少这里的去噪规则，尽量相信text-parse的结果
	paths := []func() error{
		c.CleanByCentroid,
		c.CleanBySiteRule,

		// 预处理时已经去掉了不可见节点，不再重复处理
		//c.CleanInvisibleNode,

		// 有些标签（如特殊的video）在切分过程中才能识别出来，所以不能提前去噪
		//c.CleanEmptyTag,

		c.CleanTinyNoise,
		c.CleanPotentialNoise,
		c.CleanNoiseImage,
		c.CleanNoiseLink,
		c.CleanMetaLink,
		c.CleanLinkBundle,
		c.CleanAuthorAvatar,
	}
	for _, path := range paths {
		t := time.Now()
		err := path()
		delta := time.Since(t)
		hlog.CtxInfof(c.ctx, "[cleaner purify] clean path: %s, cost: %s", getFunctionName(path), delta.String())
		if err != nil {
			hlog.CtxErrorf(c.ctx, "clean path: %s, err: %v", getFunctionName(path), err)
			return err
		}
	}

	c.Doc.ResetHtml(c.CleaningDoc.Doc)

	return nil
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (c *Cleanner) PostPurify() error {
	htmlStr, err := c.Doc.ToString()
	if err != nil {
		hlog.CtxErrorf(c.ctx, "doc to string err: %v", err)
		return err
	}

	cleaningDoc := doc.LoadDocumentFromSegmentResult(c.ctx, htmlStr, c.Doc.Url, c.Doc.RuleStageGroup)
	if cleaningDoc == nil {
		return errors.New("cleaning doc is nil")
	}
	c.CleaningDoc = cleaningDoc

	// 以后可以尽量减少这里的去噪规则，尽量相信text-parse的结果
	paths := []func() error{}
	if rule := c.CleaningDoc.Rule; rule != nil && rule.NoSemanticDenoise {
		// 无须语义去噪。因为语义去噪的标签不准
		paths = []func() error{
			c.CleanEmptyTag,
		}
	} else {
		paths = []func() error{
			c.CleanDuplicatedTitle,
			c.CleanImgBeyondCanvas,
			c.CleanEmptyTag,
		}
	}
	for _, path := range paths {
		t := time.Now()
		err := path()
		delta := time.Since(t)
		hlog.CtxInfof(c.ctx, "[cleaner post-purify] clean path: %v, cost: %s", getFunctionName(path), delta.String())
		if err != nil {
			return err
		}
	}

	c.Doc.ResetHtml(c.CleaningDoc.Doc)

	return nil
}

func (c *Cleanner) CleanByCentroid() error {
	if c.CleaningDoc.Rule != nil && len(c.CleaningDoc.Rule.Bodies) > 0 {
		return nil
	}
	if utils.Any([]string{
		"mp.weixin.qq.com", "mp.weixin.com",
	}, func(host string) bool {
		return strings.Contains(c.CleaningDoc.Url, host)
	}) {
		return nil
	}

	centroElems := []*etree.Element{}

	bodyElems := c.CleaningDoc.Xpath("//body")
	if len(bodyElems) == 0 {
		return nil
	}
	bodyElem := bodyElems[0]
	articleElems := c.CleaningDoc.RelativeXpath(bodyElem, ".//article")
	if len(articleElems) == 0 {
		hlog.CtxInfof(c.ctx, "no <article> found, use body as main content")
		centroElems = append(centroElems, bodyElem)
	} else {
		rawContent := c.CleaningDoc.GetRawDocText(c.CleaningDoc.Doc.Root())
		articleContent := strings.Join(
			utils.Map(articleElems, func(elem *etree.Element) string {
				return c.CleaningDoc.GetRawDocText(elem)
			}), "",
		)
		if float32(len(articleContent))/float32(len(rawContent)) > 0.6 {
			hlog.CtxInfof(c.ctx, "<article> content ratio is high, use article as main content")
			centroElems = append(centroElems, articleElems...)
		}
	}
	if len(centroElems) == 0 {
		if bodyElem == nil {
			return nil
		} else {
			centroElems = append(centroElems, bodyElem)
		}
	}

	html := etree.NewDocument()
	body := html.CreateElement("body")

	for _, elem := range centroElems {
		body.AddChild(elem)
	}
	err := c.CleaningDoc.ResetHtml(html)
	if err != nil {
		return err
	}
	return nil
}
func (c *Cleanner) CleanBySiteRule() error {
	if c.CleaningDoc.Rule == nil {
		return nil
	}
	rule := c.CleaningDoc.Rule

	for _, xpath := range rule.Noises {
		err := c.CleaningDoc.RemoveByXpath(xpath)
		if err != nil {
			return err
		}
	}

	html := etree.NewDocument()
	body := html.CreateElement("body")

	matchCount := 0
	for _, xpath := range rule.Bodies {
		elems := c.CleaningDoc.Xpath(xpath)
		for _, elem := range elems {
			body.AddChild(elem)
			matchCount += 1
		}
	}
	if matchCount > 0 {
		err := c.CleaningDoc.ResetHtml(html)
		if err != nil {
			return err
		}
	}
	if rule.BodyUseRuleOnly && len(rule.Bodies) > 0 {
		err := fmt.Errorf("body use rule only, not match any body, url: %v", c.CleaningDoc.Url)
		hlog.CtxErrorf(c.ctx, "clean by site rule err: %v", err)
		return err
	}
	return nil
}

func (c *Cleanner) CleanInvisibleNode() error {
	xpaths := []string{
		// 由于查询引擎限制，不能使用正则表达式
		"//*[contains(@style, 'display:none')]",
		"//*[contains(@style, 'display: none')]",
		"//*[@aria-hidden='true']", // 网页无障碍元素，一般用于隐藏元素
	}
	for _, xpath := range xpaths {
		elems := c.CleaningDoc.Xpath(xpath)
		for _, elem := range elems {
			//hlog.CtxDebugf(c.ctx, "remove invisible node: %v", c.CleaningDoc.GetElemPositionId(elem))
			err := c.CleaningDoc.RemovElem(elem)
			if err != nil {
				hlog.CtxErrorf(c.ctx, "remove elem err: %v", err)
				return err
			}
		}
	}
	return nil
}
func (c *Cleanner) CleanEmptyTag() error {
	t := time.Now()
	skipTags := map[string]int{
		//"br":    1,
		"img":   1,
		"svg":   1,
		"video": 1,
		//"math":  1,
	}
	totalSearchDuration := time.Duration(0)
	const SkipChild = true
	const ContinueChild = false
	checker := []func(element *etree.Element) bool{
		c.CleaningDoc.CheckHasText,
		//c.CleaningDoc.CheckHasBr,
		c.CleaningDoc.CheckHasImg,
		c.CleaningDoc.CheckHasVideo,
	}
	c.CleaningDoc.TraverseSkipChild(c.CleaningDoc.Doc.Root(), func(elem *etree.Element) bool {
		if skipTags[elem.Tag] == 1 {
			return true
		}

		t0 := time.Now()
		for _, check := range checker {
			if hasImportantItem := check(elem); hasImportantItem {
				totalSearchDuration += time.Since(t0)
				return false
			}
		}
		totalSearchDuration += time.Since(t0)

		err := c.CleaningDoc.RemovElem(elem)
		if err != nil {
			//hlog.CtxErrorf(c.ctx, "remove empty tag elem err: %v", err)
			return false
		}
		return true
	})
	hlog.CtxDebugf(c.ctx, "CleanEmptyTag cost: %v, total search duration: %v", time.Since(t), totalSearchDuration)
	return nil
}
func (c *Cleanner) CleanTinyNoise() error {
	for _, rule := range consts.TINY_NOISE_RULES {
		text, length := rule.Text, rule.Length
		c.CleaningDoc.XpathIter(fmt.Sprintf("//*[contains(text(), '%v')]", text), func(elem *etree.Element) {
			textContent := c.CleaningDoc.GetRawDocText(elem)
			if utf8.RuneCountInString(textContent) > length {
				return
			}
			hlog.CtxInfof(c.ctx, "remove tiny noise: %v, match: %v, text: %v", c.CleaningDoc.GetElemPositionId(elem), rule.Text, textContent)
			err := c.CleaningDoc.RemovElem(elem)
			if err != nil {
				hlog.CtxErrorf(c.ctx, "remove tiny noise elem err: %v", err)
				return
			}
		})
	}
	return nil
}
func (c *Cleanner) removePotentialNoiseIsSafe(elem *etree.Element, bodyText string) bool {
	if bodyText == "" {
		bodyText = c.CleaningDoc.GetRawDocText(c.CleaningDoc.Doc.Root())
	}
	elemText := c.CleaningDoc.GetRawDocText(elem)
	if float32(len(elemText))/float32(len(bodyText)) < 0.5 {
		return true
	}
	return false
}

func (c *Cleanner) CleanPotentialNoise() error {
	attrs := []string{"clsss", "id"}
	c.CleaningDoc.XpathIter("//*", func(elem *etree.Element) {
		attrStr := strings.Join(
			utils.Filter(
				utils.Map(elem.Attr, func(attr etree.Attr) string {
					return attr.Key
				}),
				func(s string) bool {
					if s == consts.KeyXpath {
						return false
					}
					return true
				},
			), " ",
		)
		noiseAttrResult := consts.RegexRule_NoiseAttr.FindString(attrStr)
		isNoise := false
		for _, attrKey := range attrs {
			attrValue := elem.SelectAttrValue(attrKey, "")
			if attrValue == "" {
				continue
			}
			unlikelyAttrResult := consts.RegexRule_UnlikelyCandidate.FindString(attrValue)
			negativeAttrResult := consts.RegexRule_Negative.FindString(attrValue)
			maybeOkAttrResult := consts.RegexRule_OkMaybeItsACandidate.FindString(attrValue)
			positiveAttrResult := consts.RegexRule_Positive.FindString(attrValue)
			if (unlikelyAttrResult != "" || negativeAttrResult != "" || noiseAttrResult != "") &&
				maybeOkAttrResult == "" && positiveAttrResult == "" || !utils.Contains(consts.CANT_DEL_TAGS, elem.Tag) &&
				c.removePotentialNoiseIsSafe(elem, "") {
				isNoise = true
				break
			}
		}
		if isNoise && noiseAttrResult != "" {
			hlog.CtxInfof(c.ctx, "remove potential noise: %v, match: %v, text: %v", c.CleaningDoc.GetElemPositionId(elem), noiseAttrResult, attrStr)
			c.CleaningDoc.RemovElem(elem)
		}
	})
	return nil
}
func (c *Cleanner) CleanNoiseImage() error {
	// 避免误伤头图
	if c.CleaningDoc.Rule != nil {
		hlog.CtxInfof(c.ctx, "skip clean noise image, rule: %v", c.CleaningDoc.Rule)
		return nil
	}
	c.CleaningDoc.XpathIter("//img", func(elem *etree.Element) {
		attrSrc := strings.Join(utils.Map(append(consts.IMG_ATTRS, "alt", "class"), func(key string) string {
			return elem.SelectAttrValue(key, "")
		}), " ")
		if match := consts.RegexRule_NegativeImg.FindString(attrSrc); match != "" {
			hlog.CtxInfof(c.ctx, "remove noise image: %v, match: %v, text: %v", c.CleaningDoc.GetElemPositionId(elem), match, attrSrc)
			err := c.CleaningDoc.RemovElem(elem)
			if err != nil {
				hlog.CtxErrorf(c.ctx, "remove noise image elem err: %v", err)
				return
			}
		}
		// 图片节点不需要子节点
		elem.Child = []etree.Token{}
	})
	return nil
}

func (c *Cleanner) CleanMetaLink() error {
	c.CleaningDoc.XpathIter("//link | //script", func(elem *etree.Element) {
		c.CleaningDoc.RemovElem(elem)
	})
	return nil
}

func (c *Cleanner) CleanNoiseLink() error {
	c.CleaningDoc.XpathIter("//a", func(elem *etree.Element) {
		txt := c.CleaningDoc.GetRawDocText(elem)
		if match := consts.RegexRule_NegativeLink.FindString(txt); match != "" {
			hlog.CtxInfof(c.ctx, "remove noise link: %v, match: %v, text: %v", c.CleaningDoc.GetElemPositionId(elem), match, txt)
			err := c.CleaningDoc.RemovElem(elem)
			if err != nil {
				hlog.CtxErrorf(c.ctx, "remove noise link elem err: %v", err)
			}
		}

	})
	return nil
}

func (c *Cleanner) checkIsLinkBundle(elem *etree.Element) (bool, map[int64]int, *etree.Element) {
	const TextRatio = float32(0.7)
	const MaxIterCnt = 3

	numLinks := 0
	currentTextContent := c.CleaningDoc.GetRawDocText(elem)
	currentTextContent = utils.Clean(currentTextContent)
	linkPids := map[int64]int{}

	c.CleaningDoc.Traverse(elem, &doc.TraverseParams{
		Traversefunc: func(node *etree.Element) {
			if checkIsLinkBundleTag(node.Tag) {
				numLinks += 1
				linkPids[c.CleaningDoc.GetElemPositionId(node)] = 1
			}
		},
	})

	for i := 1; i <= MaxIterCnt; i++ {
		parent := elem.Parent()
		if parent == nil {
			return false, nil, nil
		}

		parentNumLinks := 0
		parentLinkText := ""
		c.CleaningDoc.Traverse(parent, &doc.TraverseParams{
			Traversefunc: func(node *etree.Element) {
				if checkIsLinkBundleTag(node.Tag) {
					parentNumLinks += 1
					parentLinkText += c.CleaningDoc.GetRawDocText(node)
					linkPids[c.CleaningDoc.GetElemPositionId(node)] = 1
				}
			},
		})

		parentTextContent := c.CleaningDoc.GetRawDocText(parent)
		parentLinkText = utils.Clean(parentLinkText)
		parentTextContent = utils.Clean(parentTextContent)

		if parentNumLinks == numLinks {
			if i >= 2 && parentNumLinks >= 2 {
				// 可能误伤，需要确认
				return true, linkPids, parent
			}
			if len(parentTextContent) == len(currentTextContent) ||
				(float32(len(currentTextContent))/float32(len(parentTextContent)) >= TextRatio) {
				currentTextContent = parentTextContent
				elem = parent
				continue
			} else if float32(len(currentTextContent))/float32(len(parentTextContent)) < TextRatio {
				return false, nil, nil
			}
		} else if parentNumLinks > numLinks {
			if len(parentTextContent) == len(parentLinkText) ||
				float32(len(parentLinkText))/float32(len(parentTextContent)) >= TextRatio {
				// 如果父元素中的链接文本占比超过70%，则认为这是一个链接块
				return true, linkPids, parent
			}
			return false, nil, nil
		}
	}
	return false, nil, nil
}

func checkIsLinkBundleTag(tag string) bool {
	checkTags := map[string]int{
		"a": 1,
		//"img": 1,
	}
	_, ok := checkTags[tag]
	return ok
}

func (c *Cleanner) CleanLinkBundle() error {
	/*
		此函数用于清理链接块，如果链接块中包含的链接数超过一定比例，则删除该链接块
		链接块通常是指包含多个链接的元素，常被用在顶栏、侧栏之类的导航作用区域。例如：
		<div>
		  <a href="link1">link1</a>
		  <a href="link2">link2</a>
		  <a href="link3">link3</a>
		</div>
	*/

	// 如果有站点规则，则不进行清理
	if c.CleaningDoc.Rule != nil && (len(c.CleaningDoc.Rule.Bodies) > 0 || len(c.CleaningDoc.Rule.Noises) > 0) {
		return nil
	}
	visited := map[int64]int{}

	// 不用xpath，因为xpath太慢了。直接遍历所有a标签，然后检查是否是链接块
	c.CleaningDoc.Traverse(c.CleaningDoc.Doc.Root(), &doc.TraverseParams{
		Traversefunc: func(elem *etree.Element) {
			if elem.Parent() == nil {
				return
			}
			if !checkIsLinkBundleTag(elem.Tag) {
				return
			}

			pid := c.CleaningDoc.GetElemPositionId(elem)
			if visited[pid] > 0 {
				return
			}
			visited[pid] += 1
			if isBundle, linkPids, bundle := c.checkIsLinkBundle(elem); isBundle {
				c.CleaningDoc.RemovElem(bundle)
				for pid := range linkPids {
					visited[pid] += 1
				}
			}
		},
	})
	return nil
}

func (c *Cleanner) CleanAuthorAvatar() error {
	return nil
}

func (c *Cleanner) CleanDuplicatedTitle() error {
	return nil
}
func (c *Cleanner) CleanImgBeyondCanvas() error {
	sentences := c.sentences
	if sentences == nil {
		return nil
	}

	firstArticleSentencePid := int64(0)
	lastArticleSentencePid := int64(0)

	for _, sentence := range sentences {
		switch sentence.Label {
		case consts.LABEL_NOISE,
			consts.LABEL_AUTHOR,
			consts.LABEL_TITLE,
			consts.LABEL_PUB_TIME,
			consts.LABEL_SOURCE:
			continue
		}
		firstArticleSentencePid = int64(sentence.Atoms[0].PositionID)
		break
	}

	for i := len(sentences) - 1; i >= 0; i-- {
		sentence := sentences[i]
		if sentence.Label == consts.LABEL_NOISE {
			continue
		}
		lastArticleSentencePid = int64(sentence.Atoms[len(sentences[i].Atoms)-1].PositionID)
		break
	}

	if elems := c.CleaningDoc.Xpath(utils.GetPositionIdXpath(firstArticleSentencePid)); len(elems) > 0 {
		firstTopElem := c.CleaningDoc.GetMostTopElem(elems[0])
		if firstTopElem != nil {
			firstArticleSentencePid = c.CleaningDoc.GetElemPositionId(firstTopElem)
		}
	}

	if elems := c.CleaningDoc.Xpath(utils.GetPositionIdXpath(lastArticleSentencePid)); len(elems) > 0 {
		lastTopElem := c.CleaningDoc.GetMostTopElem(elems[0])
		if lastTopElem != nil {
			lastArticleSentencePid = c.CleaningDoc.GetElemPositionId(lastTopElem)
		}
	}
	// 都是0，说明没有检测的意义
	if firstArticleSentencePid+lastArticleSentencePid == 0 {
		return nil
	}
	if lastArticleSentencePid == 0 {
		lastArticleSentencePid = math.MaxInt64
	}

	c.CleaningDoc.XpathIter("//svg | //video", func(elem *etree.Element) {
		top := c.CleaningDoc.GetMostTopElem(elem)
		topPid := c.CleaningDoc.GetElemPositionId(top)
		if topPid < firstArticleSentencePid || topPid > lastArticleSentencePid {
			c.CleaningDoc.RemovElem(elem)
		}
	})

	return nil
}
