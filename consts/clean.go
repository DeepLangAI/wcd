package consts

import (
	"regexp"
	"strings"
)

var CleanRegexItems = []struct {
	Pattern     *regexp.Regexp
	Replacement string
}{
	{regexp.MustCompile(`\s{255,}`), strings.Repeat(" ", 255)},
	{regexp.MustCompile(`\s*\n\s*`), "\n"},
	{regexp.MustCompile(`\t|[ \t]{2,}`), " "},
	{regexp.MustCompile(`<[^/>]+>[ \n\r\t]*</[^>]+>`), ""},
	{regexp.MustCompile(`‘`), "'"},
	{regexp.MustCompile(`’`), "'"},
	{regexp.MustCompile(`“`), `"`},
	{regexp.MustCompile(`”`), `"`},
	{regexp.MustCompile(`…`), "...."},
	{regexp.MustCompile(`—`), "-"},
	{regexp.MustCompile(`–`), "-"},
}

var USELESS_TAGS_FOR_DOC = []string{
	"header",
	"script", "style", "footer", "comment", "aside", "nav", "noscript",
	//"iframe",
	"symbol", "button", "input", "select",
	//"link",
}

type NoisePair struct {
	Left  string
	Right string
}

var NOISE_PAIRS = []NoisePair{
	{
		Left:  "【",
		Right: "】",
	},
	{
		Left:  "[",
		Right: "]",
	},
	{
		Left:  "「",
		Right: "」",
	},
	{
		Left:  "（",
		Right: "）",
	},
	{
		Left:  "(",
		Right: ")",
	},
}
