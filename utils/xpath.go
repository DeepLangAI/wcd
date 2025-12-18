package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/DeepLangAI/wcd/consts"
)

// 处理xpath，只保留至最深的段落标签
func Pdepth(xpath string) string {
	htags := []string{}
	for i := 1; i <= 6; i++ {
		htags = append(htags, fmt.Sprintf("h%d", i))
	}
	tags := []string{"p", "div", "tr", "li", "section", "figcaption", "dd", "dt", "br"}
	tags = append(tags, htags...)

	// 构建正则表达式
	tagPattern := strings.Join(tags, "|")
	pattern := fmt.Sprintf(`(.*/(%s)(\[\d+\])?)(/.*)?$`, tagPattern)

	// 编译正则表达式
	re := regexp.MustCompile(pattern)

	// 替换匹配的内容
	modifiedXpath := re.ReplaceAllString(xpath, "$1")

	return modifiedXpath
}

// 将 XPath 转换为重要标签集合
func XpathToTags(xpath string) []string {
	if xpath == "" {
		return []string{}
	}

	// REGEX_XPATH_IDX 是一个正则表达式，用于匹配下标部分
	var REGEX_XPATH_IDX = regexp.MustCompile(`\[\d*\]`)
	// 移除下标部分
	xpath = REGEX_XPATH_IDX.ReplaceAllString(xpath, "")

	// 分割 XPath 并映射标签
	tags := map[string]int{}
	for _, item := range strings.Split(xpath, "/") {
		if mappedTag, ok := consts.GOLD_TAG_MAPPING[item]; ok {
			tags[mappedTag] = 1
		}
	}

	tagsSet := KeysOfMap(tags)
	sort.Strings(tagsSet) // 按照字母顺序排序
	return tagsSet
}

func GetPositionIdXpath(positionId any) string {
	return fmt.Sprintf("//*[@%v='%v']", consts.KeyPositionId, positionId)
}

func SplitXpath(xpath string) []string {
	if xpath == "" {
		return nil
	}
	regex := regexp.MustCompile(`\[\d+\]`)
	return Map(strings.Split(xpath, "/"), func(item string) string {
		return regex.ReplaceAllString(item, "")
	})
}
