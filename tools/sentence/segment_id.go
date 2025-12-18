package sentence

import "github.com/DeepLangAI/wcd/utils"

type SegmentIdKeeper struct {
	xpathCache map[string]int
}

func NewSegmentIdKeeper() *SegmentIdKeeper {
	return &SegmentIdKeeper{
		xpathCache: make(map[string]int),
	}
}

func (s *SegmentIdKeeper) GetSegmentdId(xpath string) int {
	paraXpath := utils.Pdepth(xpath)
	if _, ok := s.xpathCache[paraXpath]; !ok {
		s.xpathCache[paraXpath] = len(s.xpathCache) + 1
	}
	return s.xpathCache[paraXpath]
}
