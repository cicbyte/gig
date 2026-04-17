package gitignore

import (
	"bufio"
	"strings"
)

// ParseRules 将.gitignore内容解析为规则列表
func ParseRules(content string) []Rule {
	var rules []Rule
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		
		rules = append(rules, Rule{
			Pattern: trimmedLine,
			LineNum: i + 1,
			Raw:     line,
		})
	}
	
	return rules
}

// BuildRuleMap 创建规则查找映射
func BuildRuleMap(content string) map[string]int {
	ruleMap := make(map[string]int)
	scanner := bufio.NewScanner(strings.NewReader(content))
	
	for i := 1; scanner.Scan(); i++ {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ruleMap[line] = i
		}
	}
	
	return ruleMap
}

// ExtractRules 从内容中提取有效规则（非空行和非注释）
func ExtractRules(content string) map[string]bool {
	existingRules := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(content))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			existingRules[line] = true
		}
	}
	
	return existingRules
}