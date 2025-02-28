package ven

import (
	"strconv"
	"strings"
)

const (
	equal   = 0
	greater = 1
	smaller = -1
)

type Version string

func (sel Version) Split() []string {
	str := strings.TrimPrefix(strings.ToLower(string(sel)), "v")
	return strings.Split(str, ".")
}

func (sel Version) ToNumber() uint64 {
	// 去掉前缀 'v'，并按 '.' 分割成段
	parts := sel.Split()
	if len(parts) > 4 {
		return 0
	}
	var num uint64 = 0
	var shift uint = 48 // 初始偏移量，从高位开始，每段占 16 位
	for _, part := range parts {
		// 将每一段转换为数值
		value, err := strconv.Atoi(part)
		if err != nil {
			return 0
		}
		if value < 0 || value > 9999 {
			return 0
		}
		// 将该段的数值按位移操作加到最终数值中
		num += uint64(value) << shift
		shift -= 16 // 每段占用 16 位，下一段偏移量减少 16 位
	}
	return num
}

func (sel Version) IsEqual(str string) bool {
	v1 := sel.ToNumber()
	v2 := Version(str).ToNumber()
	return v1 == v2 && v1 != 0
}

func CompareVersion(version1, version2 string) int {
	// 去掉版本号中的 v 或 V
	v1 := Version(version1).Split()
	v2 := Version(version2).Split()
	// 找出较长的版本号
	maxLength := len(v1)
	if len(v2) > maxLength {
		maxLength = len(v2)
	}
	// 比较每一部分的版本号
	for i := 0; i < maxLength; i++ {
		var num1, num2 int
		if i < len(v1) {
			num1, _ = strconv.Atoi(v1[i])
		}
		if i < len(v2) {
			num2, _ = strconv.Atoi(v2[i])
		}
		if num1 > num2 {
			return greater
		} else if num1 < num2 {
			return smaller
		}
	}
	return equal
}

// IsGreaterEqual 大于等于
func IsGreaterEqual(version1, version2 string) bool {
	return CompareVersion(version1, version2) != smaller
}

// IsSmaller 小于
func IsSmaller(version1, version2 string) bool {
	return CompareVersion(version1, version2) == smaller
}

// IsEqual 等于
func IsEqual(version1, version2 string) bool {
	return CompareVersion(version1, version2) == equal
}

// IsGreater 大于
func IsGreater(version1, version2 string) bool {
	return CompareVersion(version1, version2) == greater
}

// IsSmallerEqual 小于等于
func IsSmallerEqual(version1, version2 string) bool {
	return CompareVersion(version1, version2) != greater
}
