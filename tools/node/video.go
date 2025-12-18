package node

import (
	"context"
	"strings"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/tools/sentence"
	"github.com/beevik/etree"
)

type VideoNode struct {
	Elem *etree.Element
	Doc  *doc.Document
	Sik  *sentence.SegmentIdKeeper
	ctx  context.Context
}

func NewVideoNode(elem *etree.Element, doc *doc.Document, sik *sentence.SegmentIdKeeper, ctx context.Context) (*VideoNode, error) {
	n := &VideoNode{
		Elem: elem,
		Doc:  doc,
		Sik:  sik,
	}
	return n, nil
}

func (n *VideoNode) Match() bool {
	switch n.Elem.Tag {
	case "video":
		return true
	}
	if class := n.Elem.SelectAttrValue("class", ""); class != "" {
		// https://mp.weixin.qq.com/s?__biz=MzIzNjc1NzUzMw==&mid=2247779508&idx=1&sn=b713a3a6fc9188977f9d4bcb696d70cc&chksm=e9db9395ac314cf93370232f17d9fadb59c2a09291c30f4f3b4921b0335d599f91010a7ca32f
		// https://mp.weixin.qq.com/s?__biz=MzIzNjc1NzUzMw==&mid=2247779508&idx=1&sn=b713a3a6fc9188977f9d4bcb696d70cc&chksm=e9db9395ac314cf93370232f17d9fadb59c2a09291c30f4f3b4921b0335d599f91010a7ca32f
		if strings.Contains(class, "video_iframe") {
			return true
		}
		// https://www.newyorker.com/culture/infinite-scroll/sam-altman-and-jony-ive-will-force-ai-into-your-life
		if strings.Contains(class, "cne-video-embed") {
			return true
		}
		// https://www.theverge.com/news/677139/netflix-tudum-trailers-frankenstein-wake-up-dead-man
		if strings.Contains(class, "ytp-cued-thumbnail-overlay-image") {
			return true
		}
		// https://venturebeat.com/ai/qwenlong-l1-solves-long-context-reasoning-challenge-that-stumps-current-llms/
		if strings.Contains(class, "pbs__player") {
			return true
		}
		//https://www.cbsnews.com/video/what-to-know-about-major-supreme-court-rulings-still-ahead/
		if strings.Contains(class, "player__bg") {
			return true
		}
		// https://www.jonloomer.com/qvt/always-remove-these-placements/
		if class == "qvt-video-container" {
			return true
		}
		// https://www.universetoday.com/articles/how-many-exoplanets-are-hiding-in-dust?code=0616
		if class == "embed-container" {
			return true
		}
	}

	return false
}
func (n *VideoNode) Handle() ([]*wcd.AtomicText, error) {
	n.Elem.Tag = "video"
	n.Elem.Child = nil
	return make([]*wcd.AtomicText, 0), nil
}
