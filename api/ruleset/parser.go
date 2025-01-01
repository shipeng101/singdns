package ruleset

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Parser is the interface for rule set parsers
type Parser interface {
	Parse(data []byte) ([]string, error)
}

// JSONParser parses sing-box rule set (.json) format
type JSONParser struct{}

// ListParser parses list format (one rule per line)
type ListParser struct{}

// NewParser creates a new parser based on the file type
func NewParser(fileType string) Parser {
	switch strings.ToLower(fileType) {
	case "json":
		return &JSONParser{}
	case "list":
		return &ListParser{}
	default:
		// 如果类型未指定或不支持，默认使用列表解析器
		return &ListParser{}
	}
}

// Parse parses .json format data
func (p *JSONParser) Parse(data []byte) ([]string, error) {
	// 先尝试解析标准 JSON 格式
	var rules struct {
		Rules []string `json:"rules"`
	}

	if err := json.Unmarshal(data, &rules); err != nil {
		// 如果解析失败，尝试直接解析字符串数组
		var stringArray []string
		if err := json.Unmarshal(data, &stringArray); err != nil {
			// 如果还是失败，尝试使用列表解析器
			listParser := &ListParser{}
			return listParser.Parse(data)
		}
		return stringArray, nil
	}

	return rules.Rules, nil
}

// Parse parses list format data
func (p *ListParser) Parse(data []byte) ([]string, error) {
	lines := strings.Split(string(data), "\n")
	var rules []string

	for _, line := range lines {
		// 去除空白字符
		line = strings.TrimSpace(line)
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 如果是域名规则，确保格式正确
		if strings.Contains(line, ".") {
			// 移除可能的前缀（如 "domain:"）
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					line = strings.TrimSpace(parts[1])
				}
			}
			rules = append(rules, line)
		}
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("no valid rules found in the data")
	}

	return rules, nil
}

// DownloadRuleSet downloads rule set from URL
func DownloadRuleSet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download rule set: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download rule set: status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read rule set data: %v", err)
	}

	return data, nil
}
