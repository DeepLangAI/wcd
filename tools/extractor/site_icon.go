package extractor

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/DeepLangAI/wcd/tools/doc"
	"github.com/DeepLangAI/wcd/utils"
)

type SiteIconExtractor struct {
	Doc    *doc.Document
	result string
	ctx    context.Context
}

func NewSiteIconExtractor(ctx context.Context, doc *doc.Document) *SiteIconExtractor {
	return &SiteIconExtractor{
		Doc: doc,
		ctx: ctx,
	}
}

func (t *SiteIconExtractor) Extract() (string, error) {
	candVals := []string{"icon", "shortcut icon"}
	links := t.Doc.Xpath("//link")
	if links != nil && len(links) > 0 {
		for _, link := range links {
			for _, attr := range link.Attr {
				val := strings.ToLower(attr.Value)
				if attr.Key == "rel" && utils.Any(candVals, func(candVal string) bool {
					return candVal == val
				}) {
					if href := link.SelectAttrValue("href", ""); href != "" {
						return utils.EnsureLinkAbsolute(href, t.Doc.Url), nil
					}
				}
			}
		}
	} else {
		parsedUrl, err := url.Parse(t.Doc.Url)
		if err != nil {
			return "", nil
		} else {
			return fmt.Sprintf("https://%s/favicon.ico", parsedUrl.Host), nil
		}
	}
	return "", nil

}
