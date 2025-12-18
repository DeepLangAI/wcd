package consts

import "time"

type DbStatus int

const (
	StatusValid   = 0
	StatusDeleted = 1
)

type RuleStage int

const (
	RuleStageTesting RuleStage = 0
	RuleStageProd    RuleStage = 1
)
const (
	WorthType_Valueable = 1
	WorthType_404       = 2
	WorthType_NoContent = 3
)

const (
	CrawlHtmlTimeout     = 120 * time.Second
	TextParseReadTimeOut = 120 * time.Second
)

const (
	ActionType_Request  = "req"
	ActionType_Response = "resp"

	ActionCrawlHtml = "crawl-html"
	ActionTextParse = "text-parse"
)
