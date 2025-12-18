package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func CheckDarkByBrightness(r, g, b int) bool {
	// 感知亮度公式（Perceived Brightness）
	// ITU-R BT.601（Rec. 601） 是由国际电信联盟（ITU）发布的一个关于**标准清晰度电视（SDTV）**的视频编码标准。
	// 人眼对绿色最敏感，对红色次之，对蓝色最不敏感，因此加权系数是这样分配的。
	brightness := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return brightness < 128
}

func CheckHexIsDarkColor(hex string) bool {
	hex = strings.TrimPrefix(hex, "#")
	// 如果是3位简写格式，转换为6位
	if len(hex) == 3 {
		hex = fmt.Sprintf("%c%c%c%c%c%c",
			hex[0], hex[0],
			hex[1], hex[1],
			hex[2], hex[2])
	}

	if len(hex) != 6 {
		return false
	}

	r, err := strconv.ParseInt(hex[0:2], 16, 0)
	if err != nil {
		return false
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 0)
	if err != nil {
		return false
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 0)
	if err != nil {
		return false
	}

	return CheckDarkByBrightness(int(r), int(g), int(b))
}

func IsDarkColor(color string) bool {
	if color == "" {
		return false
	}
	if strings.Contains(strings.ToLower(color), "black") ||
		strings.Contains(strings.ToLower(color), "dark") {
		return true
	}
	regexRgb := regexp.MustCompile(`rgb\(\d+,\s*\d+,\s*\d+\)`)
	regexRgba := regexp.MustCompile(`rgba\(\d+,\s*\d+,\s*\d+\)`)
	regexDigits := regexp.MustCompile(`\d+`)
	const DarkThreshold = 128
	if regexRgb.MatchString(color) {
		dims := Map(regexDigits.FindAllString(color, -1), func(t string) int {
			i, err := strconv.ParseInt(t, 10, 64)
			if err != nil {
				return 0
			}
			return int(i)
		})
		if CheckDarkByBrightness(dims[0], dims[1], dims[2]) {
			return true
		}
		//if utils.Any(dims, func(d int) bool { return d < DarkThreshold }) {
		//	return true
		//}
	}

	if regexRgba.MatchString(color) {
		dims := Map(regexDigits.FindAllString(color, -1), func(t string) int {
			i, err := strconv.ParseInt(t, 10, 64)
			if err != nil {
				return 0
			}
			return int(i)
		})
		if CheckDarkByBrightness(dims[0], dims[1], dims[2]) {
			return true
		}
		//if utils.Any(dims[:len(dims)-1], func(d int) bool { return d < DarkThreshold }) {
		//	return true
		//}
	}
	regexHexColor := regexp.MustCompile(`#([0-9a-fA-F]{3}){1,2}`)
	if regexHexColor.MatchString(color) {
		if CheckHexIsDarkColor(color) {
			return true
		}
	}

	return false
}

func IsWhiteColor(color string) bool {
	if color == "" {
		return false
	}
	whiteColors := map[string]int{
		"#ffffff":     1,
		"#fff":        1,
		"white":       1,
		"transparent": 1,
	}
	if _, ok := whiteColors[color]; ok {
		return true
	}
	if strings.ReplaceAll(color, " ", "") == "rgb(255,255,255)" ||
		strings.ReplaceAll(color, " ", "") == "rgba(255,255,255,1)" {
		return true
	}

	return false
}
