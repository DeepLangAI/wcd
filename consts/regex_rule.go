package consts

import "regexp"

var (
	RegexRule_OkMaybeItsACandidate = regexp.MustCompile("and|article|body|column|main|shadow|nickname")
	RegexRule_Positive             = regexp.MustCompile("article|artic|body|content|entry|hentry|main|page|pagination|post|text|blog|story|title")

	RegexRule_UnlikelyCandidate = regexp.MustCompile("community|disqus|extra|header|menu|remark|rss|agegate|pagination|pager|popup|tweet|twitter")
	RegexRule_Negative          = regexp.MustCompile("combx|comment|commnent|com-|contact|foot|footer|footnote|masthead|meta|outbrain|promo|related|scroll|shoutbox|sidebar|sponsor|shopping|tag|tool|recommend|recommon|search|crumb|disclaimer|relate|hot|share|pop_|side|qr_code|qr-code|qrcode|ad-break|extra|title-bar|video|navbar|erweima|data-ad|retop|wx-qr|nav-panel")
	RegexRule_NoiseAttr         = regexp.MustCompile("data-ad")
	RegexRule_NegativeImg       = regexp.MustCompile("avatar|logo|author|title|标题|weibo|wechat|weixin|icon|公众号|更多|关注|landing|loading")
	RegexRule_NegativeAvatarImg = regexp.MustCompile("avatar|author")
	//RegexRule_NegativeLink      = regexp.MustCompile("更多|more|详细|关注|aboutus|公众号|wechat|weibo")
	RegexRule_NegativeLink = regexp.MustCompile("更多|详细|关注|aboutus|公众号|wechat|weibo")

	RegexRule_ProfilImg         = regexp.MustCompile(`var hd_head_img = "(.*?)"`)
	RegexRule_AuthroDescription = regexp.MustCompile(`var profile_signature = "(.*?)"`)
	RegexRule_AuthorId          = regexp.MustCompile(`var biz = "(.*?)"`)
	RegexRule_AuthroName        = regexp.MustCompile(`window.name = "(.*?)"`)
)

var CANT_DEL_TAGS = []string{
	"html", "body", "article", "title",
	"li", "p", // 实验性质，如有问题可以删除
}
