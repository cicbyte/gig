package gitignore

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIClient 处理与gitignore API的交互
type APIClient struct {
	BaseURL string
}

// NewAPIClient 创建API客户端
func NewAPIClient() *APIClient {
	return &APIClient{
		BaseURL: "https://www.toptal.com/developers/gitignore/api/",
	}
}

// GetTemplate 获取指定语言的模板
func (c *APIClient) GetTemplate(languages []string) (string, error) {
	logger := GetLogger()

	if len(languages) == 0 {
		logger.Error("API请求失败: 未指定语言")
		return "", fmt.Errorf("未指定语言")
	}

	langParam := strings.Join(languages, ",")
	url := c.BaseURL + langParam
	logger.Debugf("请求API: %s", url)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		logger.Errorf("API请求失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Errorf("API请求返回非200状态码: %d", resp.StatusCode)
		return "", fmt.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 限制 1MB
	if err != nil {
		logger.Errorf("读取API响应失败: %v", err)
		return "", err
	}

	logger.Infof("成功从API获取模板，大小: %d字节", len(body))
	return string(body), nil
}

// GetSuggestionsFromAPI 从API获取建议
func (c *APIClient) GetSuggestionsFromAPI(currentContent string, languages []string) ([]string, error) {
	logger := GetLogger()
	logger.Debugf("从API获取建议，语言: %v", languages)

	apiContent, err := c.GetTemplate(languages)
	if err != nil {
		logger.Errorf("获取API模板失败: %v", err)
		return nil, err
	}

	existingRules := ExtractRules(currentContent)
	var suggestions []string

	templateRules := ParseRules(apiContent)
	logger.Debugf("API模板包含 %d 条规则", len(templateRules))

	for _, rule := range templateRules {
		if !existingRules[rule.Pattern] {
			suggestion := fmt.Sprintf("考虑添加 '%s' (来自 api 模板)。", rule.Pattern)
			suggestions = append(suggestions, suggestion)
		}
	}

	logger.Infof("从API生成 %d 条建议", len(suggestions))
	return suggestions, nil
}
