package doc

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/dal/mongo"
	"github.com/DeepLangAI/wcd/tools/sentence"
	"github.com/DeepLangAI/wcd/utils"

	"github.com/antchfx/xmlquery"
	"github.com/beevik/etree"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/net/html"
)

type Document struct {
	xpathCache     map[string]string // position_id to xpath
	xmlQueryDoc    *xmlquery.Node
	ctx            context.Context
	rawHtml        string
	RuleStageGroup wcd.RuleStageGroupEnum

	ReservedNodes []*etree.Element
	Rule          *mongo.SiteRuleModel
	Doc           *etree.Document
	MaxPositionId int64
	Url           string
}

func Preprocess(htmlContent string, htmlUrl string, reservedNodeTags []string) (string, []*etree.Element, error) {
	// 将所有textarea标签替换为div标签
	// 如：https://tech.huanqiu.com/article/4Lfhd37YZdn?code=324
	htmlContent = replaceTextareaWithDiv(htmlContent)

	// 解析 HTML（自动修复结构）
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", nil, err
	}

	// 删除 <script> 标签
	//removeTags(doc, []string{
	//	"script", "style", "link", "noscript",
	//})
	tags := utils.Filter(consts.USELESS_TAGS_FOR_DOC, func(tag string) bool {
		return !utils.Contains(reservedNodeTags, tag)
	})
	reservedNodes := []*etree.Element{}
	removeTags(htmlUrl, doc, tags, &reservedNodes)

	// 输出修改后的 HTML
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", nil, err
	}
	bufHtml := buf.String()
	fixedHtml := fixHtml(bufHtml)
	return fixedHtml, reservedNodes, nil
}

func matchTag(n *html.Node, tag string, key string, matchFunc func(string) bool) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if n.Data != tag {
		return false
	}
	if utils.Any(n.Attr, func(attr html.Attribute) bool {
		return attr.Key == key && matchFunc(attr.Val)
	}) {
		return true
	}
	return false
}

// 递归遍历节点并删除 <tag> 标签
func removeTags(htmlUrl string, n *html.Node, tags []string, reservedNodes *[]*etree.Element) {
	attrs := map[string]string{}
	for _, attr := range n.Attr {
		attrs[attr.Key] = attr.Val
	}
	// 特殊case，不能删除aria-hidden为true的svg：
	// https://mp.weixin.qq.com/s?__biz=MzI3ODgwODA2MA%3D%3D&mid=2247535378&idx=1&sn=998d667c1fe0edd696fa321f7001806f&scene=0
	if attrs["aria-hidden"] == "true" {
		if n.Data != "svg" {
			n.Parent.RemoveChild(n)
		}
		return
	}
	if style := attrs["style"]; style != "" {
		if strings.Contains(style, "display: none") || strings.Contains(style, "display:none") {
			n.Parent.RemoveChild(n)
			return
		}
	}
	//nonVisible := utils.Any(n.Attr, func(attr html.Attribute) bool {
	//	if attr.Key != "style" {
	//		return false
	//	}
	//	value := attr.Val
	//	if strings.Contains(value, "display: none") {
	//		return true
	//	}
	//	if strings.Contains(value, "display:none") {
	//		return true
	//	}
	//	return false
	//})
	if n.Type == html.ElementNode {
		if n.Data == "iframe" {
			// 删除iframe标签内的内容
			for c := n.FirstChild; c != nil; {
				next := c.NextSibling
				c.Parent.RemoveChild(c)
				c = next
			}
			return
		}

		if utils.Contains(tags, n.Data) {
			// 包含math的script标签不删除，用于渲染公式
			if matchTag(n, "script", "src", func(val string) bool {
				return strings.Contains(val, "math")
			}) || matchTag(n, "link", "rel", func(val string) bool {
				return val == "stylesheet"
			}) || n.Data == "style" {
				elem := &etree.Element{
					Tag: n.Data,
					Attr: utils.Map(n.Attr, func(attr html.Attribute) etree.Attr {
						return etree.Attr{
							Key:   attr.Key,
							Value: attr.Val,
						}
					}),
				}
				if n.FirstChild != nil {
					elem.SetText(n.FirstChild.Data)
				}
				if href := elem.SelectAttrValue("href", ""); href != "" {
					href = utils.EnsureLinkAbsolute(href, htmlUrl)
					elem.CreateAttr("href", href)
				}
				*reservedNodes = append(*reservedNodes, elem)
			}
			n.Parent.RemoveChild(n)
			return
		}
		if utils.Any(n.Attr, func(attr html.Attribute) bool {
			if attr.Key == "class" {
				//if attr.Val == "header" {
				//	// 完全匹配，才是true
				//	return true
				//}
				if utils.Any([]string{
					"nav",
					"footer",
					"header", // 有可能误伤
				}, func(s string) bool {
					//return strings.Contains(attr.Val, s)
					return attr.Val == s
				}) {
					return true
				}
			}
			return false
		}) {
			n.Parent.RemoveChild(n)
			return
		}

		n.Attr = utils.Filter(n.Attr, func(attr html.Attribute) bool {
			return isValidAttrName(attr.Key)
		})
	} else if n.Type == html.CommentNode {
		n.Parent.RemoveChild(n)
		return
	}

	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		removeTags(htmlUrl, c, tags, reservedNodes)
		c = next
	}
}

func isValidAttrName(name string) bool {
	// 如果是纯数字，则不合法
	if _, err := strconv.Atoi(name); err == nil {
		return false
	}
	// 如果包含特殊字符，则不合法
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`).MatchString(name) {
		return false
	}

	return true
}

func replaceTextareaWithDiv(html string) string {
	// 匹配开始标签，如<textarea>或<textarea attr="value">
	startTagRegex := regexp.MustCompile(`(?i)<textarea\b([^>]*)>`)
	replaced := startTagRegex.ReplaceAllString(html, "<div$1>")

	// 匹配结束标签</textarea>
	endTagRegex := regexp.MustCompile(`(?i)</textarea>`)
	replaced = endTagRegex.ReplaceAllString(replaced, "</div>")

	return replaced
}

func fixHtml(htmlContent string) string {

	tags := []string{"img", "br", "hr", "input", "link", "meta", "area", "base", "col", "embed", "keygen", "param", "source", "track", "wbr"}
	// 定义正则表达式，匹配非闭合的单标签，但排除已经自闭合的标签
	re := regexp.MustCompile(fmt.Sprintf(`<(%v)((\s+([^>]*)>)|>)`, strings.Join(tags, "|")))

	// 替换匹配到的标签，添加自闭合的斜杠
	closedTags := re.ReplaceAllStringFunc(htmlContent, func(tag string) string {
		suffix := regexp.MustCompile(`(/\s*>)$`)
		// 检查是否已经是自闭合标签
		if suffix.MatchString(tag) {
			return tag
		}
		return tag[:len(tag)-1] + " />" // 在标签末尾添加 />
	})

	endTag := regexp.MustCompile(`\s*/\s*>`)
	closedTags = endTag.ReplaceAllString(closedTags, " />")

	return closedTags
}

func LoadDocumentFromSegmentResult(ctx context.Context, htmlStr string, url string, ruleStageGroup wcd.RuleStageGroupEnum) *Document {
	doc := etree.NewDocument()
	doc.ReadSettings = etree.ReadSettings{
		Permissive:             true,
		PreserveCData:          false,
		PreserveDuplicateAttrs: false,
		ValidateInput:          false,
		AutoClose:              xml.HTMLAutoClose,
	}
	err := doc.ReadFromString(htmlStr)
	if err != nil {
		hlog.CtxErrorf(ctx, "etree parse error: %v", err)
		return nil
	}

	d := &Document{
		ctx:            ctx,
		xpathCache:     map[string]string{},
		Url:            url,
		rawHtml:        htmlStr,
		RuleStageGroup: ruleStageGroup,
	}
	rule, err := d.MatchRule()
	if err != nil {
		hlog.CtxErrorf(ctx, "match rule error: %v", err)
		return nil
	}
	d.Rule = rule

	err = d.ResetHtml(doc)
	if err != nil {
		hlog.CtxErrorf(ctx, "reset html error: %v", err)
		return nil
	}
	d.CacheXpath()
	return d
}

func NewDocument(ctx context.Context, htmlStr string, url string, ruleStageGroup wcd.RuleStageGroupEnum) (*Document, error) {
	ruleMatcher := &Document{Url: url, RuleStageGroup: ruleStageGroup}
	rule, err := ruleMatcher.MatchRule()
	if err != nil {
		hlog.CtxErrorf(ctx, "match rule error: %v", err)
		return nil, err
	}

	rawHtmlStr := htmlStr
	reservedNodeTags := []string{}
	if rule != nil {
		reservedNodeTags = rule.ReservedNodes
	}
	htmlStr, reservedNodes, err := Preprocess(htmlStr, url, reservedNodeTags)
	if err != nil {
		hlog.CtxErrorf(ctx, "repair html error: %v", err)
		return nil, err
	}

	doc := etree.NewDocument()
	//autoClose := append(xml.HTMLAutoClose)
	doc.ReadSettings = etree.ReadSettings{
		Permissive:             true,
		PreserveCData:          false,
		PreserveDuplicateAttrs: false,
		ValidateInput:          false,
		AutoClose:              xml.HTMLAutoClose,
	}
	err = doc.ReadFromString(htmlStr)
	if err != nil {
		// 可以到https://tool.box3.cn/xml-validator.html查看htmlStr是否合法
		hlog.CtxErrorf(ctx, "etree parse error: %v", err)
		return nil, err
	}

	utils.CoreLog(ctx, utils.CoreNameSegment, utils.NodeNameSegmentPreProcess)
	d := &Document{
		ctx:            ctx,
		xpathCache:     map[string]string{},
		Url:            url,
		rawHtml:        rawHtmlStr,
		RuleStageGroup: ruleStageGroup,
		Rule:           rule,
		ReservedNodes:  reservedNodes,
	}
	utils.CoreLog(ctx, utils.CoreNameSegment, utils.NodeNameSegmentMatchRule)

	d.SetSkelton(doc)
	err = d.ResetHtml(doc)
	if err != nil {
		hlog.CtxErrorf(ctx, "reset html error: %v", err)
		return nil, err
	}
	d.CacheXpath()
	utils.CoreLog(ctx, utils.CoreNameSegment, utils.NodeNameSegmentSkelton)
	return d, nil
}

func (d *Document) ResetHtml(html *etree.Document) error {
	if html == nil || html.Root() == nil {
		return errors.New("html is nil")
	}
	htmlStr, err := html.WriteToString()
	regex := regexp.MustCompile(`position_id="(\d+)"`)
	found := regex.FindAllStringSubmatch(htmlStr, -1)
	maxPid := int64(0)
	for _, match := range found {
		if len(match) == 2 {
			pid, err := strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				hlog.CtxErrorf(d.ctx, "parse pid error: %v", err)
				continue
			}
			if pid != 0 {
				maxPid = max(pid, maxPid)
			}
		}
	}

	if err != nil {
		hlog.CtxErrorf(d.ctx, "etree write to string error: %v", err)
		return err
	}
	//xmlqueryDoc, err := xmlquery.Parse(strings.NewReader(htmlStr))
	xmlqueryDoc, err := xmlquery.ParseWithOptions(strings.NewReader(htmlStr), xmlquery.ParserOptions{
		Decoder: &xmlquery.DecoderOptions{
			Strict: false,
		},
	})
	if err != nil {
		hlog.CtxErrorf(d.ctx, "xmlquery parse error: %v", err)
		return err
	}
	d.xmlQueryDoc = xmlqueryDoc
	d.Doc = html
	d.MaxPositionId = maxPid
	return nil
}

// IsSimpleXPath 判断给定的 XPath 是否是简单的 XPath
func (d *Document) isSimpleXPath(xpath string) bool {
	// 定义正则表达式
	//正则比较复杂，可以去这里可视化查看：https://wangwl.net/static/projects/visualRegex#
	simpleXPathPattern := `^((//|/)(\*|\w+)(((\[\d+\])?(\[@\w+='\w+'\])?)|((\[@\w+='\w+'\])?(\[\d+\])?)))*?$`
	re := regexp.MustCompile(simpleXPathPattern)

	// 测试匹配
	return re.MatchString(xpath)
}

func (d *Document) RelativeXpath(elem *etree.Element, xpath string) []*etree.Element {
	pid := d.GetElemPositionId(elem)
	pidXpath := utils.GetPositionIdXpath(pid)
	queryElem := xmlquery.FindOne(d.xmlQueryDoc, pidXpath)
	elems, err := xmlquery.QueryAll(queryElem, xpath)

	if err != nil {
		hlog.CtxErrorf(d.ctx, "xmlquery query error: %v", err)
		return nil
	}
	elements := []*etree.Element{}
	for _, elem := range elems {
		pid := elem.SelectAttr(consts.KeyPositionId)
		element := d.Doc.FindElement(utils.GetPositionIdXpath(pid))
		if element != nil {
			elements = append(elements, element)
		}
	}
	return elements
}

func (d *Document) Xpath(xpath string) []*etree.Element {
	if d.isSimpleXPath(xpath) {
		//t := time.Now()
		elems := d.Doc.FindElements(xpath)
		//hlog.CtxDebugf(d.ctx, "xpath etree time: %v, xpath: %v", time.Since(t), xpath)
		return elems
	}
	//t := time.Now()
	elems, err := xmlquery.QueryAll(d.xmlQueryDoc, xpath)
	//hlog.CtxDebugf(d.ctx, "xpath xmlqquery time: %v, xpath: %v", time.Since(t), xpath)
	if err != nil {
		hlog.CtxErrorf(d.ctx, "xmlquery query error: %v", err)
		return nil
	}
	//t = time.Now()
	elements := []*etree.Element{}
	for _, elem := range elems {
		pid := elem.SelectAttr(consts.KeyPositionId)
		element := d.Doc.FindElement(fmt.Sprintf("//*[@%v='%v']", consts.KeyPositionId, pid))
		if element != nil {
			elements = append(elements, element)
		}
	}
	//hlog.CtxDebugf(d.ctx, "xpath etree time: %v, xpath: %v", time.Since(t), xpath)
	return elements
}

func (d *Document) XpathIter(xpath string, iterFunc func(element *etree.Element)) []*etree.Element {
	elements := d.Xpath(xpath)

	sort.Slice(elements, func(i, j int) bool {
		xpathI := d.GetElemXpath(elements[i])
		xpathJ := d.GetElemXpath(elements[j])
		return xpathI < xpathJ
	})
	for _, element := range elements {
		iterFunc(element)
	}
	return elements
}

func (d *Document) GetElemXpath(elem *etree.Element) string {
	if attr := elem.SelectAttr(consts.KeyPositionId); attr != nil {
		return d.xpathCache[attr.Value]
	}
	return ""
}
func (d *Document) GetElemPositionId(elem *etree.Element) int64 {
	if attr := elem.SelectAttr(consts.KeyPositionId); attr != nil {
		atoi, err := strconv.Atoi(attr.Value)
		if err != nil {
			return -1
		}
		return int64(atoi)
	}
	return -1
}

func (d *Document) IncreasePositionId() int64 {
	d.MaxPositionId += 1
	return d.MaxPositionId
}

func (d *Document) cacheXpath(node *etree.Element, rank int, path []string) {
	if attr := node.SelectAttr(consts.KeyPositionId); attr != nil {
		positionId := attr.Value
		if rank == 0 {
			path = append(path, fmt.Sprintf("%v", node.Tag))
		} else {
			path = append(path, fmt.Sprintf("%v[%v]", node.Tag, rank))
		}
		d.xpathCache[positionId] = fmt.Sprintf("/%v", strings.Join(path, "/"))
	}

	// 遍历子节点
	tagElements := map[string][]*etree.Element{}
	for _, child := range node.Child {
		switch child := child.(type) {
		case *etree.Element:
			if tagElements[child.Tag] == nil {
				tagElements[child.Tag] = []*etree.Element{}
			}
			tagElements[child.Tag] = append(tagElements[child.Tag], child)
		case *etree.CharData:
		}
	}
	for _, elems := range tagElements {
		if len(elems) == 1 {
			d.cacheXpath(elems[0], 0, path)
		} else {
			for i, elem := range elems {
				d.cacheXpath(elem, i+1, path)
			}
		}
	}
}

func (d *Document) CacheXpath() {
	d.cacheXpath(d.Doc.Root(), 0, []string{})
}

func (d *Document) SetSkelton(doc *etree.Document) {
	d.MaxPositionId = 0
	d.Traverse(doc.Root(), &TraverseParams{
		Traversefunc: func(node *etree.Element) {
			node.CreateAttr(consts.KeyPositionId, fmt.Sprintf("%v", d.IncreasePositionId()))
		},
	})
}

type TraverseParams struct {
	Traversefunc func(node *etree.Element)
	ElementFunc  func(node *etree.Element)
	CharDataFunc func(node *etree.CharData)
}

func (d *Document) Traverse(node *etree.Element, params *TraverseParams) {
	d.traverse(node, params, nil)
}

// 递归遍历节点
func (d *Document) traverse(node *etree.Element, params *TraverseParams, visited map[string]int) {
	if node == nil {
		return
	}
	if visited == nil {
		visited = map[string]int{}
	}

	// etree可能有bug，会导致无限递归，所以需要记录已经访问过的节点
	pid := node.SelectAttrValue(consts.KeyPositionId, "")
	if pid != "" {
		if _, ok := visited[pid]; ok {
			hlog.CtxErrorf(d.ctx, "traverse already visited")
			return
		}
	}
	visited[pid] = 1

	if params == nil {
		params = &TraverseParams{}
	}
	if params.Traversefunc != nil {
		params.Traversefunc(node)
	}
	// 遍历子节点
	for _, child := range node.Child {
		switch child := child.(type) {
		case *etree.Element:
			if params.ElementFunc != nil {
				params.ElementFunc(child)
			}
			d.traverse(child, params, visited) // 递归遍历子元素
		case *etree.CharData:
			if params.CharDataFunc != nil {
				params.CharDataFunc(child)
			}
		}
	}
}

func (d *Document) TraverseCheck(node *etree.Element, traverseFunc func(node *etree.Element) bool) bool {
	if node == nil {
		return false
	}
	if traverseFunc(node) {
		return true
	}
	// 遍历子节点
	for _, child := range node.ChildElements() {
		// 递归遍历子元素
		if d.TraverseCheck(child, traverseFunc) {
			return true
		}
	}
	return false
}

func (d *Document) TraverseSkipChild(node *etree.Element, traverseFunc func(node *etree.Element) bool) bool {
	// 当满足条件时，不再遍历子节点
	if node == nil {
		return false
	}
	if traverseFunc(node) {
		return true
	}
	// 遍历子节点
	for _, child := range node.ChildElements() {
		d.TraverseSkipChild(child, traverseFunc)
	}
	return false
}

func (d *Document) GetRawDocText(root *etree.Element) string {
	if root == nil {
		return ""
	}
	texts := []string{}
	if text := strings.TrimSpace(root.Text()); text != "" {
		texts = append(texts, root.Text())
	}
	for _, child := range root.ChildElements() {
		if text := strings.TrimSpace(d.GetRawDocTextWithTail(child)); text != "" {
			texts = append(texts, text)
		}
	}
	return strings.Join(texts, "\n")
}

// getRawDocText 提取文档中的文本内容，跳过指定标签和隐藏元素
func (d *Document) GetRawDocTextWithTail(root *etree.Element) string {
	skipChecker := func(elem *etree.Element) bool {
		skipTags := []string{"script", "style", "noscript", "iframe", "object"}
		hidden := false
		tagShouldSkip := false

		if style := elem.SelectAttrValue("style", ""); style != "" {
			style = strings.ReplaceAll(style, " ", "")
			hidden = hidden || strings.Contains(strings.ToLower(style), "display:none")
		}
		if class := elem.SelectAttrValue("class", ""); class != "" {
			class = strings.ReplaceAll(class, " ", "")
			names := utils.DictFromItems(
				utils.Map(
					utils.Filter(strings.Split(class, " "), func(name string) bool {
						return strings.TrimSpace(name) != ""
					}),
					func(name string) utils.DictItem[string, int] {
						return utils.DictItem[string, int]{Key: name, Value: 1}
					}),
			)
			hidden = hidden || names["hidden"] == 1
		}
		tagShouldSkip = utils.Contains(skipTags, elem.Tag)
		return tagShouldSkip || hidden
	}
	if root == nil {
		return ""
	}
	// 需要跳过的标签
	skipTags := []string{"script", "style", "noscript", "iframe", "object", "head"}
	if utils.Contains(skipTags, root.Tag) {
		return ""
	}

	tailStack := utils.Stack[*etree.Element]{}
	texts := []string{}
	if !skipChecker(root) {
		if text := strings.TrimSpace(root.Text()); text != "" {
			texts = append(texts, text)
		}
	}
	prefixTree := utils.NewPrefixTree()
	d.Traverse(root, &TraverseParams{
		ElementFunc: func(element *etree.Element) {
			nodeXpath := d.GetElemXpath(element)

			for !tailStack.IsEmpty() {
				peek, _ := tailStack.Peek()
				tailXpath := d.GetElemXpath(peek)
				if !strings.HasPrefix(nodeXpath, tailXpath) {
					tailStack.Pop()
					texts = append(texts, strings.TrimSpace(peek.Tail()))
				} else {
					break
				}
			}

			if prefixTree.HasPrefixOf(nodeXpath) {
				return
			}

			if !(skipChecker(element)) {
				if text := strings.TrimSpace(element.Text()); text != "" {
					texts = append(texts, text)
				}
			} else {
				prefixTree.Insert(nodeXpath)
			}
			if strings.TrimSpace(element.Tail()) != "" {
				tailStack.Push(element)
			}
		},
	})

	// 提取文本内容
	for !tailStack.IsEmpty() {
		tail, _ := tailStack.Pop()
		texts = append(texts, strings.TrimSpace(tail.Tail()))
	}

	texts = append(texts, strings.TrimSpace(root.Tail()))
	// 拼接文本内容
	return strings.TrimSpace(strings.Join(texts, "\n"))
}

func (d *Document) AddReservedNodes() error {
	var head *etree.Element
	if elems := d.Xpath("//head"); len(elems) == 0 {
		head = &etree.Element{Tag: "head"}
		d.Doc.Root().InsertChildAt(0, head)
	} else {
		head = elems[0]
	}

	for _, node := range d.ReservedNodes {
		head.AddChild(node)
	}
	return nil
}

func (d *Document) InsertMeta(articleMeta *wcd.ArticleMeta) {
	if articleMeta == nil {
		return
	}
	var head *etree.Element
	if elems := d.Xpath("//head"); len(elems) == 0 {
		head = &etree.Element{Tag: "head"}
		d.Doc.Root().InsertChildAt(0, head)
	} else {
		head = elems[0]
	}
	if articleMeta.Title != "" {
		elem := &etree.Element{Tag: "title"}
		elem.SetText(articleMeta.Title)
		head.InsertChildAt(0, elem)
	}
	if articleMeta.Author != "" {
		elem := &etree.Element{Tag: "meta"}
		elem.CreateAttr("name", "author")
		elem.CreateAttr("content", articleMeta.Author)
		head.InsertChildAt(0, elem)
	}
	if articleMeta.PublishTime != "" {
		elem := &etree.Element{Tag: "meta"}
		elem.CreateAttr("name", "pubtime")
		elem.CreateAttr("content", articleMeta.PublishTime)
		head.InsertChildAt(0, elem)
	}
	if articleMeta.GetSiteIcon() != "" {
		elem := &etree.Element{Tag: "link"}
		elem.CreateAttr("rel", "icon")
		elem.CreateAttr("href", articleMeta.GetSiteIcon())
		head.InsertChildAt(0, elem)
	}
	if articleMeta.GetSurfaceImage() != "" {
		elem := &etree.Element{Tag: "meta"}
		elem.CreateAttr("name", "og:image")
		elem.CreateAttr("content", articleMeta.GetSurfaceImage())
		head.InsertChildAt(0, elem)
	}
	if articleMeta.GetDescription() != "" {
		elem := &etree.Element{Tag: "meta"}
		elem.CreateAttr("name", "description")
		elem.CreateAttr("content", articleMeta.GetDescription())
		head.InsertChildAt(0, elem)
	}
}
func (d *Document) ToString() (string, error) {
	if d.Doc == nil {
		return "", nil
	}
	ws := etree.WriteSettings{
		CanonicalEndTags: true, // 不使用自闭合标签
		CanonicalText:    false,
		CanonicalAttrVal: false,
		AttrSingleQuote:  false,
	}
	d.Doc.WriteSettings = ws
	output, err := d.Doc.WriteToString()
	if err != nil {
		hlog.CtxErrorf(d.ctx, "failed to write doc to string: %v", err)
		return "", err
	}
	// 字符串替换（如替换 ></img> 为 />）
	closeTags := []string{"img", "br"}
	regex := regexp.MustCompile(fmt.Sprintf(`>(\s*?)</(%v)>`, strings.Join(closeTags, "|")))
	output = regex.ReplaceAllString(output, " />")

	output = utils.UnescapeHtml(output)
	return output, nil
}
func (d *Document) RemovElem(elem *etree.Element) error {
	parent := elem.Parent()
	if parent == nil {
		msg := fmt.Sprintf("elem has no parent: %v", elem)
		//hlog.CtxErrorf(d.ctx, msg)
		return errors.New(msg)
	}
	parent.RemoveChild(elem)
	return nil
}
func (d *Document) RemoveByXpath(xpath string) error {
	elems := d.Xpath(xpath)
	for _, elem := range elems {
		elem.Parent().RemoveChild(elem)
	}
	return nil
}
func (d *Document) ElemToAtom(elem *etree.Element, text string, tail bool, segmentId int, immutable bool) (*wcd.AtomicText, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("empty text")
	}
	xpath := d.GetElemXpath(elem)
	positionId := d.GetElemPositionId(elem)
	attrs := map[string]string{}
	skipAttrs := []string{
		consts.KeyPositionId,
		consts.KeyXpath,
	}
	for _, attr := range elem.Attr {
		if !utils.Contains(skipAttrs, attr.Key) {
			attrs[attr.Key] = attr.Value
		}
	}
	op := sentence.AtomOperator{}
	atom := op.NewAtom(
		text,
		int32(positionId),
		xpath,
		immutable,
		utils.XpathToTags(xpath),
		tail,
		int32(segmentId),
		attrs,
	)
	return atom, nil
}
func (d *Document) MatchRule() (*mongo.SiteRuleModel, error) {
	dal := mongo.SiteRuleModelDal
	rules, err := dal.FindManyByStageGroup(d.ctx, d.RuleStageGroup)
	if err != nil {
		return nil, err
	}
	sort.Slice(rules, func(i, j int) bool {
		return len(rules[i].Host) > len(rules[j].Host)
	})
	for _, rule := range rules {
		if rule.Match(d.Url) {
			return rule, nil
		}
	}
	return nil, nil
}

func (d *Document) GetRawHtmlStr() string {
	return d.rawHtml
}

func (d *Document) GetImages() []string {
	elems := d.Xpath("//img")
	imgUrls := []string{}
	for _, elem := range elems {
		for _, key := range consts.IMG_ATTRS {
			if value := elem.SelectAttrValue(key, ""); value != "" {
				if strings.HasPrefix(value, "http") {
					imgUrls = append(imgUrls, value)
				}
			}
		}
	}
	return imgUrls
}

func (d *Document) GetImagesWithPositionId() map[string]string {
	elems := d.Xpath("//img")
	imgUrls := map[string]string{}
	for _, elem := range elems {
		pid := elem.SelectAttrValue(consts.KeyPositionId, "")
		for _, key := range consts.IMG_ATTRS {
			if value := elem.SelectAttrValue(key, ""); value != "" {
				if strings.HasPrefix(value, "http") {
					imgUrls[value] = pid
				}
			}
		}
	}
	return imgUrls
}

func (d *Document) CheckHasText(elem *etree.Element) bool {
	txt := d.GetRawDocText(elem)
	return utils.Clean(txt) != ""
}
func (d *Document) CheckHasBr(elem *etree.Element) bool {
	return d.TraverseCheck(elem, func(subElem *etree.Element) bool {
		if subElem.Tag == "br" {
			return true
		}
		return false
	})
}

func (d *Document) elemIsValidImg(elem *etree.Element) bool {
	if elem.Tag == "svg" && len(elem.Child) > 0 {
		return true
	}
	if elem.Tag == "img" && utils.Any(consts.IMG_ATTRS, func(keyName string) bool {
		return elem.SelectAttrValue(keyName, "") != ""
	}) {
		return true
	}
	return false

}

func (d *Document) elemIsValidVideo(elem *etree.Element) bool {
	if elem.Tag == "video" {
		return true
	}
	return false
}

func (d *Document) CheckHasImg(elem *etree.Element) bool {
	return d.TraverseCheck(elem, func(subElem *etree.Element) bool {
		return d.elemIsValidImg(subElem)
	})
}
func (d *Document) CheckHasVideo(elem *etree.Element) bool {
	return d.TraverseCheck(elem, func(subElem *etree.Element) bool {
		return d.elemIsValidVideo(subElem)
	})
}

func (d *Document) GetMostTopElem(elem *etree.Element) *etree.Element {
	text := utils.Clean(d.GetRawDocText(elem))
	elemShouldSkip := map[string]int{}
	checkers := []func(elem *etree.Element) bool{
		d.elemIsValidImg,
		d.elemIsValidVideo,
	}
	for {
		if parent := elem.Parent(); parent != nil {
			textContent := d.GetRawDocText(parent)
			hasImg := false
			d.TraverseSkipChild(parent, func(node *etree.Element) bool {
				pid := node.SelectAttrValue(consts.KeyPositionId, "")
				if _, ok := elemShouldSkip[pid]; ok {
					return true
				}

				for _, checker := range checkers {
					if checker(node) {
						hasImg = true
						return true
					}
				}
				if pid != "" {
					elemShouldSkip[pid] = 1
				}
				return false
			})
			if hasImg {
				break
			}
			if utils.Clean(textContent) == text {
				elem = parent
			} else {
				break
			}
		} else {
			break
		}
	}
	return elem
}
