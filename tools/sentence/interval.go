package sentence

import (
	"sort"
)

type Interval struct {
	Start int
	End   int
	ID    int
}

func FindInterval(coords int, intervals []Interval) (int, bool) {
	// 确保区间按起始坐标排序
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i].Start < intervals[j].Start
	})

	// 二分查找
	low, high := 0, len(intervals)-1
	for low <= high {
		mid := (low + high) / 2
		if coords < intervals[mid].Start {
			high = mid - 1
		} else if coords > intervals[mid].End {
			low = mid + 1
		} else {
			return intervals[mid].ID, true
		}
	}
	return -1, false // 未找到
}
