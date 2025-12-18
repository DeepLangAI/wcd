package utils

import (
	"context"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"math"
	"os/exec"
	"sort"
	"strings"
)

func Any[T any](elems []T, success func(T) bool) bool {
	for _, elem := range elems {
		if success(elem) {
			return true
		}
	}
	return false
}

func All[T any](elems []T, success func(T) bool) bool {
	for _, elem := range elems {
		if !success(elem) {
			return false
		}
	}
	return true
}

func IndexOf(slice []string, str string) int {
	for i, s := range slice {
		if s == str {
			return i
		}
	}
	return -1
}

func Set[T comparable](values []T) []T {
	cache := map[T]int{}
	for _, v := range values {
		cache[v] = 1
	}
	return KeysOfMap(cache)
}

func KeysOfMap[T comparable, V any](dict map[T]V) []T {
	keys := make([]T, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	return keys
}

func ValuesOfMap[T comparable, V any](dict map[T]V) []V {
	values := make([]V, 0, len(dict))
	for _, v := range dict {
		values = append(values, v)
	}
	return values
}

// Contains reports whether v is present in s.
func Contains[S ~[]E, E comparable](s S, v E) bool {
	return Index(s, v) >= 0
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
func Index[S ~[]E, E comparable](s S, v E) int {
	for i := range s {
		if v == s[i] {
			return i
		}
	}
	return -1
}

// Round 四舍五入函数
func Round(f float64) float64 {
	if f < 0 {
		return math.Ceil(f - 0.5)
	}
	return math.Floor(f + 0.5)
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func Min[T Ordered](x T, y ...T) T {
	for _, v := range y {
		if v < x {
			x = v
		}
	}
	return x
}

func Map[T, R any](elems []T, f func(T) R) []R {
	result := make([]R, len(elems))
	for i, elem := range elems {
		result[i] = f(elem)
	}
	return result
}
func MapDict[K comparable, V, R any](data map[K]V, f func(K, V) R) []R {
	result := make([]R, 0, len(data))
	for k, v := range data {
		result = append(result, f(k, v))
	}
	return result
}

func Filter[T any](data []T, f func(T) bool) []T {
	// 留下f为true的元素
	result := make([]T, 0, len(data))
	for _, v := range data {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

type DictItem[K comparable, V any] struct {
	Key   K
	Value V
}

func DictFromItems[K comparable, V any](items []DictItem[K, V]) map[K]V {
	data := make(map[K]V)
	for _, item := range items {
		data[item.Key] = item.Value
	}
	return data
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func Range[N Number](start, end, step N) []N {
	result := make([]N, 0, int((end-start)/step)+1)
	for i := start; i < end; i += step {
		result = append(result, i)
	}
	return result
}

type CmpFunc[T any] func(a, b T) int

// Sort 函数对切片进行排序，支持多个比较函数
func Sort[T comparable](data []T, cmpFuncs ...CmpFunc[T]) {
	sort.Slice(data, func(i, j int) bool {
		for _, cmp := range cmpFuncs {
			result := cmp(data[i], data[j])
			if result < 0 {
				return true
			} else if result > 0 {
				return false
			}
			// 如果 result == 0，继续使用下一个比较函数
		}
		// 如果所有比较函数都返回 0，则认为两个元素相等
		return false
	})
}

func NewAuthContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, consts.HeaderAuthorization, "dummy")
	return ctx
}

func ExecShellCmd(cmd string) (string, error) {
	shell := "sh"
	flag := "-c"

	// 创建命令对象
	command := exec.Command(shell, flag, cmd)

	// 执行并捕获合并输出（stdout + stderr）
	output, err := command.CombinedOutput()
	result := strings.TrimSpace(string(output))

	// 返回标准化结果
	return result, err
}
