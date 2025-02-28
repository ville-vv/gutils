package ven

import (
	"fmt"
	"testing"
)

func TestCompareVersion(t *testing.T) {
	fmt.Println(CompareVersion("", "v1.0.0"))
	fmt.Println(CompareVersion("v0.0.0.1", ""))
	fmt.Println(CompareVersion("v1.0.0", "v1.0.0"))
	fmt.Println(CompareVersion("v1.0.0", "v1.0.0.1"))
	fmt.Println(CompareVersion("v1.0.0", "v1.0.0.1"))
	fmt.Println(CompareVersion("v1.0.0", "v1.0.0.1"))
	fmt.Println(CompareVersion("v1.2.0", "v1.0.2.1"))

	fmt.Println(IsGreaterEqual("v1.2.0", "v1.0.2.1"))
	fmt.Println(IsGreaterEqual("v1.2.0", "v1.2.0.0"))
	fmt.Println(IsGreaterEqual("v1.1.0", "v1.2.0.0"))
}

func TestVersionCompare(t *testing.T) {
	fmt.Println(IsSmaller("v1.2.8.99", "v1.2.9.0"))
}

func TestToNumber(t *testing.T) {
	tests := []struct {
		version   Version
		expected  uint64
		shouldErr bool
	}{
		// 正常情况
		{"v1.0", 1 << 48, false},
		{"v1.2", (1<<48 + 2<<32), false},
		{"v1.2.3", (1<<48 + 2<<32 + 3<<16), false},
		{"v1.2.3.4", (1<<48 + 2<<32 + 3<<16 + 4), false},

		// 边界情况
		{"v0", 0, true}, // 无有效版本号
		{"v9999.9999", (9999<<48 + 9999<<32), false},
		{"v9999.9999.9999", (9999<<48 + 9999<<32 + 9999<<16), false},
		{"v9999.9999.9999.9999", (9999<<48 + 9999<<32 + 9999<<16 + 9999), false},
		{"v10000.0", 0, true},   // 超出有效范围
		{"v1.1.1.1.1", 0, true}, // 超过 4 段

		// 错误版本号
		{"v1.a.3", 0, true},    // 包含非数字字符
		{"v1.2.x", 0, true},    // 包含字母字符
		{"v.a.b", 0, true},     // 无法解析
		{"v1.-1", 0, true},     // 负数不合法
		{"v1.2.3.-1", 0, true}, // 负数不合法
	}

	for _, test := range tests {
		t.Run(string(test.version), func(t *testing.T) {
			result := test.version.ToNumber()
			if test.shouldErr && result != 0 {
				t.Errorf("Expected error for version %s, but got %d", test.version, result)
			}
			if !test.shouldErr && result != test.expected {
				t.Errorf("For version %s, expected %d, but got %d", test.version, test.expected, result)
			}
		})
	}
}

func TestIsEqual(t *testing.T) {
	tests := []struct {
		version1 string
		version2 string
		expected bool
	}{
		// 相等的版本号
		{"v1.0", "1.0.0", true},
		{"v1.2", "1.2", true},
		{"v1.2.0", "1.2", true},
		{"1.0.0.0", "1.0", true},
		{"v2.0", "V2.0", true}, // 大小写 v 无关

		// 不相等的版本号
		{"v1.0", "1.1", false},
		{"v1.2", "1.3", false},
		{"v1.2.3", "1.2.4", false},
		{"v2.0", "1.9.9.9", false},
		{"v1.2.3.4", "1.2.3.5", false},

		// 边界情况
		{"", "v1.0", false},             // 空字符串
		{"v1.2.3.9999", "1.2.3", false}, // 版本号长短不一

		// 前缀和版本号不规范的情况
		{"1.0", "v1.0", true}, // 带前缀和不带前缀
		{"v1", "1.0.0", true}, // 简短版本号
	}

	for _, test := range tests {
		t.Run(test.version1+"=="+test.version2, func(t *testing.T) {
			result := IsEqual(test.version1, test.version2)
			if result != test.expected {
				t.Errorf("For version1: %s and version2: %s, expected %v but got %v",
					test.version1, test.version2, test.expected, result)
			}
		})
	}
}

func TestVersion_IsEqual(t *testing.T) {
	tests := []struct {
		version1 string
		version2 string
		expected bool
	}{
		// 相等的版本号
		{"v1.0", "1.0.0", true},
		{"v1.2.3", "1.2.3", true},
		{"v1.2.3.4", "1.2.3.4", true},
		{"v1.0", "V1.0", true},    // 大小写 v 无关
		{"v2.0", "2.0.0.0", true}, // 四段与两段相等情况
		{"1.2.3", "1.2.3", true},  // 不带前缀的版本号相等

		// 不相等的版本号
		{"v1.0", "1.1.0", false},
		{"v1.2.3", "1.2.4", false},
		{"v1.2.3.4", "1.2.3.5", false},
		{"v2.0", "1.9.9.9", false},
		{"v1.0.0", "1.0.0.1", false},

		// 边界情况
		{"", "v1.0", false},             // 空字符串
		{"v1.2.3.9999", "1.2.3", false}, // 长短不一的版本号
		{"1.0.0", "1.0.0.0", true},      // 三段与四段相等
		{"1.0.0.0.1", "1.0.0.0", false}, // 超过四段的无效版本

		{"v1.0.0200", "1.0.02", false}, // 超过四段的无效版本

		// 无效输入测试
		{"v1.10000.0", "1.10000.0", false}, // 超出范围的段
		{"v1.a.0", "1.0.0", false},         // 非数字字符
	}

	for _, test := range tests {
		t.Run(test.version1+"=="+test.version2, func(t *testing.T) {
			v := Version(test.version1)
			result := v.IsEqual(test.version2)
			if result != test.expected {
				t.Errorf("For version1: %s and version2: %s, expected %v but got %v",
					test.version1, test.version2, test.expected, result)
			}
		})
	}
}
