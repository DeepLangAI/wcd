package extractor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/tools/doc"
)

type Extractor interface {
	Extract() (string, error)
}

func ExtractByXpath(xpath string, doc *doc.Document) (string, error) {
	if strings.Contains(xpath, consts.XpathSep) {
		parts := strings.Split(xpath, consts.XpathSep)
		for _, part := range parts {
			result, err := extractByXpath(part, doc)
			if err != nil {
				return "", err
			}
			if result != "" {
				return result, nil
			}
		}
	} else {
		return extractByXpath(xpath, doc)
	}
	return "", nil
}

func extractByXpath(xpath string, doc *doc.Document) (string, error) {
	propertyXpath := regexp.MustCompile(`/@(.*?)$`)
	propertyKey := ""

	functionXpath := regexp.MustCompile(`/([^/]+?)\(\)$`)
	function := ""

	if propertyXpath.MatchString(xpath) {
		propertyKey = propertyXpath.FindStringSubmatch(xpath)[1]
		xpath = propertyXpath.ReplaceAllString(xpath, "")
	} else if functionXpath.MatchString(xpath) {
		function = functionXpath.FindStringSubmatch(xpath)[1]
		xpath = functionXpath.ReplaceAllString(xpath, "")
	}

	result := ""
	coreRegex := `(\s|·)*`
	leftRegex := regexp.MustCompile(fmt.Sprintf(`^%s`, coreRegex))
	rightRegex := regexp.MustCompile(fmt.Sprintf(`%s$`, coreRegex))
	if elements := doc.Xpath(xpath); len(elements) > 0 {
		elem := elements[0]
		if propertyKey != "" {
			result = elem.SelectAttrValue(propertyKey, "")
		} else if function != "" {
			switch function {
			case "text":
				result = elem.Text()
			default:
				result = doc.GetRawDocText(elem)
			}
		} else {
			result = doc.GetRawDocText(elem)
		}
		result = leftRegex.ReplaceAllString(result, "")
		result = rightRegex.ReplaceAllString(result, "")
		return result, nil
	}
	return result, nil
}
