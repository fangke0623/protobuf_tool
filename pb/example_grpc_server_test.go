package pb

import (
	"testing"
)

// TestGetPublishOptionFromMethodName 测试GetPublishOptionFromMethodName方法
func TestGetPublishOptionFromMethodName(t *testing.T) {
	// 测试用例：已知方法名和预期的publish值
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{"GetExample方法", "GetExample", true},
		{"CreateExample方法", "CreateExample", false},
		{"UpdateExample方法", "UpdateExample", true},
		{"DeleteExample方法", "DeleteExample", false},
		{"ListExamples方法", "ListExamples", true},
	}

	// 运行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测试的函数
			result := GetPublishOptionFromMethodName(tt.method)
			
			// 验证结果是否符合预期
			if result != tt.expected {
				t.Errorf("%s: 预期 publish=%v, 实际 publish=%v", tt.name, tt.expected, result)
			}
		})
	}
}

// TestGetPublishOptionFromMethodName_UnknownMethod 测试未知方法名的情况
func TestGetPublishOptionFromMethodName_UnknownMethod(t *testing.T) {
	// 调用被测试的函数，使用一个不存在的方法名
	result := GetPublishOptionFromMethodName("UnknownMethod")
	
	// 验证结果是否为默认值true
	if result != true {
		t.Errorf("未知方法: 预期 publish=true, 实际 publish=%v", result)
	}
}

// TestGetPublishOptionFromMethodName_EmptyMethod 测试空方法名的情况
func TestGetPublishOptionFromMethodName_EmptyMethod(t *testing.T) {
	// 调用被测试的函数，使用空字符串作为方法名
	result := GetPublishOptionFromMethodName("")
	
	// 验证结果是否为默认值true
	if result != true {
		t.Errorf("空方法名: 预期 publish=true, 实际 publish=%v", result)
	}
}
