package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"unicode/utf8"

	"golang.org/x/crypto/blake2b"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"net/url"
	"regexp"
	"strings"

	"github.com/DeepLangAI/wcd/consts"
)

func isAlpha(str string) bool {
	return regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(str)
}

func endsWithAlpha(str string) bool {
	return regexp.MustCompile(`[a-zA-Z]+$`).MatchString(str)
}
func startsWithAlpha(str string) bool {
	return regexp.MustCompile(`^[a-zA-Z]+`).MatchString(str)
}

func Join(a, b string) string {
	return a + b
	// todo 沟通对其他调用方没影响之后再改
	// 如果a、b都是纯英文字母，则中间添加空格
	if endsWithAlpha(a) && startsWithAlpha(b) {
		return a + " " + b
	}
	return a + b
}

// PrefixTree 定义了一个前缀树结构
type PrefixTree struct {
	root map[string]interface{}
}

// NewPrefixTree 创建一个新的前缀树
func NewPrefixTree() *PrefixTree {
	return &PrefixTree{
		root: make(map[string]interface{}),
	}
}

// Insert 向前缀树中插入一个单词
func (pt *PrefixTree) Insert(word string) {
	node := pt.root
	for _, char := range word {
		charStr := string(char)
		if _, ok := node[charStr]; !ok {
			node[charStr] = make(map[string]interface{})
		}
		node = node[charStr].(map[string]interface{})
	}
	node["end"] = true
}

// HasPrefixOf 检查前缀树中是否有某个单词的前缀
func (pt *PrefixTree) HasPrefixOf(word string) bool {
	if word == "" {
		return false
	}
	node := pt.root
	for _, char := range word {
		charStr := string(char)
		if _, ok := node[charStr]; !ok {
			return false
		}
		node = node[charStr].(map[string]interface{})
		if _, ok := node["end"]; ok {
			return true
		}
	}
	return false
}

func GetStyleMap(style string) map[string]string {
	styleMap := make(map[string]string)
	split := strings.Split(style, ";")
	for _, s := range split {
		items := strings.Split(s, ":")
		if len(items) == 2 {
			key, val := strings.TrimSpace(items[0]), strings.TrimSpace(items[1])
			styleMap[key] = val
		}
	}
	return styleMap
}

func StyleMapToString(styleMap map[string]string) string {
	return strings.Join(MapDict(styleMap, func(key, value string) string {
		return fmt.Sprintf("%v: %v", key, value)
	}), ";")
}

func TrimLeftCommon(str, cut string) string {
	// 找到两个字符串的最长公共前缀
	for len(str) > 0 && len(cut) > 0 && str[0] == cut[0] {
		str = str[1:]
		cut = cut[1:]
	}
	return str
}

func GetUrlHost(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		items := strings.Split(url, "/")
		if len(items) > 2 {
			return items[2]
		}
		return ""
	}
	return ""
}

func EnsureLinkAbsolute(link, htmlUrl string) string {
	// link是待处理的链接，可能是相对链接，url是当前页面的链接

	// 如果link已经是绝对链接，则直接返回
	if strings.HasPrefix(link, "http") {
		return link
	}
	if strings.HasPrefix(link, "//") {
		return "http:" + link
	}
	if strings.HasPrefix(link, "://") {
		return "http" + link
	}

	// 如果link是相对根目录的链接，则将其转换为绝对链接。用主机名加上相对路径
	baseUrl, err := url.Parse(htmlUrl)
	if err != nil {
		return link
	}
	relativePath, err := url.Parse(link)
	if err != nil {
		return link
	}

	fullUrl := baseUrl.ResolveReference(relativePath)
	return fullUrl.String()
}
func ExtractUrlHost(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Host
}

func Clean(value string) string {
	for _, item := range consts.CleanRegexItems {
		value = item.Pattern.ReplaceAllString(value, item.Replacement)
	}
	return strings.TrimSpace(value)
}
func RemoveSpace(value string) string {
	return regexp.MustCompile(`[\s\p{Zs}]+`).ReplaceAllString(value, "")
}

// LongestCommonSubstring，计算最长连续子串
func LCS(a, b string) string {
	m := len(a)
	n := len(b)
	maxLen := 0   // 记录最长子串长度
	endIndex := 0 // 记录最长子串在a中的结束位置

	// 初始化DP表，dp[i][j]表示以a[i-1]和b[j-1]结尾的公共子串长度
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	// 构建DP表
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				// 更新最大值
				if dp[i][j] > maxLen {
					maxLen = dp[i][j]
					endIndex = i - 1 // 记录a中的结束位置
				}
			} else {
				dp[i][j] = 0 // 不连续时重置为0
			}
		}
	}

	// 提取最长子串
	if maxLen == 0 {
		return ""
	}
	return a[endIndex-maxLen+1 : endIndex+1]
}

func HashWithVersion(str string) string {
	// 创建一个新的 BLAKE2b 哈希
	hasher, err := blake2b.New256([]byte(consts.VERSION))
	if err != nil {
		panic(err)
	}

	// 更新哈希
	hasher.Write([]byte(str))

	// 计算哈希值
	hash := hasher.Sum(nil)

	// 返回哈希值的前 maxLength 个字符
	maxLength := 8
	return hex.EncodeToString(hash)[:maxLength]
}

func GetLcsString(full, sub string) string {
	fullClean := Clean(full)
	subClean := Clean(sub)

	lcs := LCS(fullClean, subClean)
	idx := strings.Index(fullClean, lcs)
	lcsRunes := []rune(lcs)
	fullRunes := []rune(full)
	fullCleanRunes := []rune(fullClean)

	runeIndex := 0
	byteCount := 0
	for i, r := range fullCleanRunes {
		runeBytes := len(string(r))
		if byteCount+runeBytes > idx {
			runeIndex = i
			break
		}
		byteCount += runeBytes
	}

	end := min(runeIndex+len(lcsRunes), len(fullRunes))
	if runeIndex >= end {
		return ""
	}
	return string(fullRunes[runeIndex:end])
}

func ReplaceLinks(originalString string, replaceMap map[string]string) string {
	// 替换字符串中的链接
	if len(replaceMap) == 0 {
		return originalString
	}

	quotedLinks := MapDict(replaceMap, func(originImgUrl string, _ string) string {
		return regexp.QuoteMeta(originImgUrl)
	})
	sort.Slice(quotedLinks, func(i, j int) bool {
		return len(quotedLinks[i]) > len(quotedLinks[j]) // 按链接长度从大到小排序，避免短链接覆盖长链接
	})

	// 创建一个模式，用于匹配原链接
	re := regexp.MustCompile(strings.Join(quotedLinks, "|"))

	// 替换链接
	return re.ReplaceAllStringFunc(originalString, func(match string) string {
		replace := replaceMap[match]
		if replace == "" {
			return match
		}
		return replace
	})
}

func IsOssImage(imgUrl string) bool {
	if imgUrl == "" {
		return false
	}
	return strings.Contains(imgUrl, "oss-cn-zhangjiakou.aliyuncs.com")
}
func UrlIsPdf(link string) bool {
	if re := regexp.MustCompile(`api.lingowhale.com/api/plugin/file/download\?file_id=.*`); re.MatchString(link) {
		return true
	}
	if strings.HasSuffix(link, ".pdf") {
		return true
	}
	if strings.Contains(link, "arxiv.org/pdf") {
		return true
	}
	return false
}

func IsLingowhalePdf(pdfUrl string) bool {
	if re := regexp.MustCompile(`api.lingowhale.com/api/plugin/file/download\?file_id=.*`); re.MatchString(pdfUrl) {
		return true
	}
	if strings.Contains(pdfUrl, "oss-cn-zhangjiakou.aliyuncs.com") {
		return true
	}
	if re := regexp.MustCompile(`internal-api-drive-stream.feishu.cn/space/api/box/stream/download/authcode/\?code=.*`); re.MatchString(pdfUrl) {
		return true
	}
	//if strings.Contains(pdfUrl, "arxiv.org/pdf") {
	//	return true
	//}
	return false
}

type UnionFind[T comparable] struct {
	parent map[T]T
	rank   map[T]int
}

func NewUnionFind[T comparable]() *UnionFind[T] {
	return &UnionFind[T]{
		parent: make(map[T]T),
		rank:   make(map[T]int),
	}
}

func (uf *UnionFind[T]) Find(p T) T {
	if _, exists := uf.parent[p]; !exists {
		uf.parent[p] = p
		uf.rank[p] = 1
	}
	if uf.parent[p] != p {
		uf.parent[p] = uf.Find(uf.parent[p]) // 路径压缩
	}
	return uf.parent[p]
}

func (uf *UnionFind[T]) Union(p, q T) {
	rootP := uf.Find(p)
	rootQ := uf.Find(q)

	if rootP != rootQ {
		// 按秩合并
		if uf.rank[rootP] > uf.rank[rootQ] {
			uf.parent[rootQ] = rootP
		} else if uf.rank[rootP] < uf.rank[rootQ] {
			uf.parent[rootP] = rootQ
		} else {
			uf.parent[rootQ] = rootP
			uf.rank[rootP]++
		}
	}
}

func (uf *UnionFind[T]) GetRoots() []T {
	roots := make(map[T]struct{})
	for node := range uf.parent {
		root := uf.Find(node)
		roots[root] = struct{}{}
	}
	return KeysOfMap(roots)
}

func UnescapeHtml(output string) string {
	output = strings.ReplaceAll(output, "&amp;", "&")
	output = strings.ReplaceAll(output, "&nbsp;", " ")
	//output = html.UnescapeString(output)
	// https://xz.aliyun.com/news/18218 这个网站用了自定义的特殊标签，需要把前缀去掉变成正常的普通标签
	output = strings.ReplaceAll(output, "<ne-", "<")
	output = strings.ReplaceAll(output, "</ne-", "</")
	//output = strings.ReplaceAll(output, "&quot;", "\"")
	return output
}

// 尝试多种编码进行解码
func TryReadText(content []byte) string {
	// 先尝试 UTF-8
	if utf8.Valid(content) {
		decoded := string(content)
		if len(decoded) > 0 {
			return decoded
		}
	}

	// 如果不是 UTF-8，依次尝试 GBK, GB2312（等价于 HZGB2312）, GB18030
	encodings := []encoding.Encoding{
		simplifiedchinese.GBK,
		simplifiedchinese.HZGB2312,
		simplifiedchinese.GB18030,
	}

	for _, enc := range encodings {
		reader := transform.NewReader(bytes.NewReader(content), enc.NewDecoder())
		decoded, err := io.ReadAll(reader)
		if err == nil {
			if len(string(decoded)) > 0 {
				return string(decoded)
			}
		}
	}

	// 所有尝试失败则返回空
	return ""
}

func StrToMd5(str string) string {
	hasher := md5.New()
	io.WriteString(hasher, str) // 写入字符串到MD5哈希器

	// 获取哈希值
	hashBytes := hasher.Sum(nil)
	hashStr := fmt.Sprintf("%x", hashBytes) // 转换为16进制字符串
	return hashStr
}

func FirstNRunes(s string, n int) string {
	runes := []rune(s)
	if n > len(runes) {
		n = len(runes)
	}
	return string(runes[:n])
}
