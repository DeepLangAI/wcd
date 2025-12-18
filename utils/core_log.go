package utils

import (
	"context"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const (
	NodeBegin = "begin"
	NodeDone  = "done"

	CoreNameSegment           = "seg"
	NodeNameSegmentPreProcess = "pre_process"
	NodeNameSegmentMatchRule  = "match_rule"
	NodeNameSegmentSkelton    = "skelton"
	NodeNameSegmentPreDistill = "pre_distill"
	NodeNameSegmentSegment    = "seg"
	NodeNameSegmentPreFormat  = "pre_format"

	CoreNameLabel    = "label"
	CoreNameCrawlImg = "crawl_img"

	CoreNamePostDistill                 = "post_distill"
	NodeNamePostDistillFormat           = "post_format"
	NodeNamePostDistillDistill          = "post_distill"
	NodeNamePostDistillRenderHtml       = "render_html"
	NodeNamePostDistillCheckWorthless   = "check_worthless"
	NodeNamePostDistillGetExistSentence = "exist_sentence"
	NodeNamePostDistillGetText          = "get_text"
	NodeNamePostDistillGetImg           = "get_img"
)

func CoreLog(ctx context.Context, coreName string, nodeName string) {
	hlog.CtxInfof(ctx, "wcd core link core_name: %s, node_name: %s", coreName, nodeName)
}
