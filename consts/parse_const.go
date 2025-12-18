package consts

import "time"

const KeyPositionId = "position_id"
const KeySubtree = "subtree"
const KeyXpath = "xpath"
const StyleAttr = "style"
const TagNameImg = "img"

var IMG_TAGS = map[string]string{
	"img":     "img",
	"map":     "img",
	"area":    "img",
	"canvas":  "img",
	"picture": "img",
	"figure":  "img",
	"svg":     "img",
}

var IMG_ATTRS = []string{"src", "_src", "data-src", "data-original", "data-lazy-bgimg", "data-lazy-src"}
var A_ATTRS = []string{"href", "data-href"}

var PARAGRAPH_TAGS = []string{
	"p",
	"span", // FIXME open this for labeling data
	"div",
}
var TABLE_TAGS = []string{
	"table",
}

var CHINESE_SENTENCE_STOP_SIGN = []string{
	"？", "！", "。", "…", "：", "；",
	":",
}
var ENGLISH_SENTENCE_STOP_SIGN = []string{
	"?", "!", ":", ";",
}

const END_PUNCT = `.!?。！？;；`

var SENTENCE_STOP_EXT = []string{
	"\"", "”", "’", "」", ")", "）",
}

var GOLD_TAG_MAPPING = map[string]string{
	"h1":         "h1",
	"h2":         "h2",
	"h3":         "h3",
	"h4":         "h4",
	"h5":         "h5",
	"h6":         "h6",
	"strong":     "strong",
	"b":          "b",
	"em":         "em",
	"hr":         "hr",
	"br":         "br",
	"img":        "img",
	"map":        "img",
	"area":       "img",
	"canvas":     "img",
	"picture":    "img",
	"figure":     "img",
	"svg":        "img",
	"figcaption": "figcaption",
	"table":      "table",
	"th":         "table",
	"tr":         "table",
	"td":         "table",
	"thead":      "table",
	"tbody":      "table",
	"tfoot":      "table",
	"col":        "table",
	"colgroup":   "table",
	"caption":    "caption",
	"menu":       "li",
	"ul":         "li",
	"ol":         "li",
	"li":         "li",
	"dl":         "li",
}

const IMG_TAG_PREFIX = "<img"
const TABLE_TAG_PREFIX = "<table"

// const VNODE_TAG_PREFIX = "deeplang-vnode-"
const VNODE_TAG_PREFIX = "dplv-"
const ATTR_PREFIX = "data-deeplang-"
const TAG_PREFIX = "deeplang-"

const (
	LABEL_NOISE       = "O"
	LABEL_CONTENT     = "content"
	LABEL_TITLE       = "article_title"
	LABEL_SUBTITLE    = "副标题"
	LABEL_AUTHOR      = "author"
	LABEL_SOURCE      = "source"
	LABEL_PUB_TIME    = "publish_time"
	LABEL_TITLE_EN    = "英文标题"
	LABEL_AUTHOR_EN   = "英文作者"
	LABEL_SOURCE_EN   = "英文来源"
	LABEL_INTRO       = "introduction"
	LABEL_ABSTRACT    = "abstract"
	LABEL_CATALOG     = "catalog"
	LABEL_TITLE_L1    = "title1"
	LABEL_TITLE_L2    = "title2"
	LABEL_TITLE_L3    = "title3"
	LABEL_TITLE_L4    = "title4"
	LABEL_TITLE_OTHER = "title5"
	LABEL_LEGEND      = "figure_title"
	LABEL_FIGURE      = "figure"
	LABEL_REFERENCE   = "reference"
	LABEL_OTHER       = "特殊内容待沟通"

	LABEL_PODCAST_SHOWNOTES = "podcast-shownotes"
	LABEL_PODCAST_SPEAKER   = "podcast-speaker"
	LABEL_PODCAST_TIME      = "podcast-time"
	LABEL_PODCAST_MESSAGE   = "podcast-message"
)
const (
	EDU_L1_LABEL_EDU_O = "EDU_O"
	EDU_L1_LABEL_BOT   = "BOT"
)

var AUTHOR_KEYWORDS = []string{
	"作者",
	"原创",
	"来源",
	"出品",
	"文",
	"责编",
	"责任编辑",
	"编辑",
	"撰文",
	"文案",
	"文字",
	"翻译",
	"报道",
	"记者",
	"校对",
	"设计制作",
	"设计",
	"审核",
	"美编",
	"ID",
	"排版",
	"热门专栏",
	"整理",
	"编导",
	"策划",
}

var PUBLISH_TIME_META = []string{ // publish-time will be put in <meta> for some standard website.
	// '//meta[starts-with(@property, "rnews:datePublished")]/@content', previous sample.
	`//meta[starts-with(@property, "rnews:datePublished")]`,
	`//meta[starts-with(@property, "article:published_time")]`,
	`//meta[starts-with(@property, "og:published_time")]`,
	`//meta[starts-with(@property, "og:release_date")]`,
	`//meta[starts-with(@itemprop, "datePublished")]`,
	`//meta[starts-with(@itemprop, "dateUpdate")]`,
	`//meta[starts-with(@name, "citation_date")]`,
	`//meta[starts-with(@name, "OriginalPublicationDate")]`,
	`//meta[starts-with(@name, "article_date_original")]`,
	`//meta[starts-with(@name, "og:time")]`,
	`//meta[starts-with(@name, "apub:time")]`,
	`//meta[starts-with(@name, "publication_date")]`,
	`//meta[starts-with(@name, "sailthru.date")]`,
	`//meta[starts-with(@name, "PublishDate")]`,
	`//meta[starts-with(@name, "publishdate")]`,
	`//meta[starts-with(@name, "PubDate")]`,
	`//meta[starts-with(@name, "pubtime")]`,
	`//meta[starts-with(@name, "_pubtime")]`,
	`//meta[starts-with(@name, "weibo: article:create_at")]`,
	`//meta[starts-with(@pubdate, "pubdate")]`,
}
var AUTHOR_META_KEYS = []string{
	"author", "article:author",
}

var DESCRIPTION_META_KEYS = []string{
	"description", "og:description",
}

var SURFACE_IMG_META_KEYS = []string{
	"og:image",
}
var TITLE_META_KEYS = []string{
	"og:title",
}

var DATETIME_SUBMATCH_PATTERN = []string{
	"var createTime = '(.+?)';", // 微信公众号文章
}
var DATETIME_PATTERN = []string{
	`(\d{4}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[0-1]?[0-9]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{4}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[2][0-3]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{4}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[0-1]?[0-9]:[0-5]?[0-9])`,
	`(\d{4}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[2][0-3]:[0-5]?[0-9])`,
	`(\d{4}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[1-24]\d时[0-60]\d分)([1-24]\d时)`,
	`(\d{2}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[0-1]?[0-9]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{2}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[2][0-3]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{2}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[0-1]?[0-9]:[0-5]?[0-9])`,
	`(\d{2}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[2][0-3]:[0-5]?[0-9])`,
	`(\d{2}[-|/|.]\d{1,2}[-|/|.]\d{1,2}\s*?[1-24]\d时[0-60]\d分)([1-24]\d时)`,
	`(\d{4}年\d{1,2}月\d{1,2}日\s*?[0-1]?[0-9]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{4}年\d{1,2}月\d{1,2}日\s*?[2][0-3]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{4}年\d{1,2}月\d{1,2}日\s*?[0-1]?[0-9]:[0-5]?[0-9])`,
	`(\d{4}年\d{1,2}月\d{1,2}日\s*?[2][0-3]:[0-5]?[0-9])`,
	`(\d{4}年\d{1,2}月\d{1,2}日\s*?[1-24]\d时[0-60]\d分)([1-24]\d时)`,
	`(\d{2}年\d{1,2}月\d{1,2}日\s*?[0-1]?[0-9]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{2}年\d{1,2}月\d{1,2}日\s*?[2][0-3]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{2}年\d{1,2}月\d{1,2}日\s*?[0-1]?[0-9]:[0-5]?[0-9])`,
	`(\d{2}年\d{1,2}月\d{1,2}日\s*?[2][0-3]:[0-5]?[0-9])`,
	`(\d{2}年\d{1,2}月\d{1,2}日\s*?[1-24]\d时[0-60]\d分)([1-24]\d时)`,
	`(\d{1,2}月\d{1,2}日\s*?[0-1]?[0-9]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{1,2}月\d{1,2}日\s*?[2][0-3]:[0-5]?[0-9]:[0-5]?[0-9])`,
	`(\d{1,2}月\d{1,2}日\s*?[0-1]?[0-9]:[0-5]?[0-9])`,
	`(\d{1,2}月\d{1,2}日\s*?[2][0-3]:[0-5]?[0-9])`,
	`(\d{1,2}月\d{1,2}日\s*?[1-24]\d时[0-60]\d分)([1-24]\d时)`,
	`(\d{4}[-|/|.]\d{1,2}[-|/|.]\d{1,2})`,
	`(\d{1,2}[-|/|.]\d{1,2}[-|/|.]\d{4})`,
	`(\d{2}[-|/|.]\d{1,2}[-|/|.]\d{1,2})`,
	`(\d{4}年\d{1,2}月\d{1,2}日)`,
	`(\d{2}年\d{1,2}月\d{1,2}日)`,
	`(\d{1,2}月\d{1,2}日)`,
	//`(([01]?[0-9])|(2[0-3])):([0-5][0-9])(:([0-5][0-9]))?`,
}

const TITLE_SPLIT_CHAR_PATTERN = `[-_|｜]`
const XpathSep = " | "
const LCS_TITLE_RATIO = 0.7

var DATETIME_FORMATS = []string{
	time.Layout,   // "01/02 03:04:05PM '06 -0700" // The reference time, in numerical order.
	time.ANSIC,    // "Mon Jan _2 15:04:05 2006"
	time.UnixDate, // "Mon Jan _2 15:04:05 MST 2006"
	time.RubyDate, // "Mon Jan 02 15:04:05 -0700 2006"
	time.RFC822,   // "02 Jan 06 15:04 MST"
	time.RFC822Z,  // "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	time.RFC850,   // "Monday, 02-Jan-06 15:04:05 MST"
	time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
	time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05+07:00",
	time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
	time.Kitchen,     // "3:04PM"

	time.Stamp,      // "Jan _2 15:04:05"
	time.StampMilli, // "Jan _2 15:04:05.000"
	time.StampMicro, // "Jan _2 15:04:05.000000"
	time.StampNano,  // "Jan _2 15:04:05.000000000"
	time.DateTime,   // "2006-01-02 15:04:05"
	"2006-01-02 15:04",
	time.DateOnly, // "2006-01-02"

	// "2" 的设计初衷是匹配无前导字符的日期，对 "06" 的解析是历史遗留的宽松行为。
	// 符号	含义					匹配示例	是否严格匹配
	// "2"	日期（无前导空格或零）	"6"		否
	// "_2"	日期前可能有空格		" 6"	是
	// "02"	日期前有零填充		"06"	是
	"Jan _2 2006", // 简写
	"Jan _2, 2006",
	"January _2 2006", // 全写
	"January _2, 2006",

	"2006/01/02",
	"2006/01/02 15:04:05",
	"2006/01/02 15:04",
	"2006年01月02日 15:04",        // 2025年03月06日 16:21
	"2006年01月02日 15:04:05",     // 2025年03月06日 16:21:00
	"2006年01月02日 15时04分",       // 2025年03月06日 16时21分
	"2006年01月02日 15时04分05秒",    // 2025年03月06日 16时21分00秒
	"2006-01-02T15:04:05-0700", // -0700是时区

	"2 January 2006",
	"2 January, 2006",
	"2 Jan 2006",
	"2 Jan, 2006",
	"02.01.2006", // 日、月、年
	"1/2/2006",   // 月、日、年
	"01/02/2006", // 月、日、年
}

const NormalizedDateTime = "2006-01-02 15:04:05"

const EmptyExtractXpath = "empty"

const TITLE_MAX_LENGTH = 40
