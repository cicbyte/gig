package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cicbyte/gig/internal/models"
	"github.com/spf13/viper"
)

type AIUtil struct{}

var AI = AIUtil{}

// 通用流式AI请求，返回完整内容
func (AIUtil) StreamChat(promptType string, vars ...string) (string, error) {
	apiKey := viper.GetString("ai.api_key")
	if apiKey == "" {
		return "", fmt.Errorf("AI API 密钥未配置。请运行 'gig config' 或设置 GIG_AI_API_KEY")
	}
	apiURL := viper.GetString("ai.url")
	if !strings.HasSuffix(apiURL, "/chat/completions") {
		apiURL = strings.TrimRight(apiURL, "/") + "/chat/completions"
	}
	model := viper.GetString("ai.model")

	promptTpl := models.AppConfig.Prompts[promptType]
	anyVars := make([]any, len(vars))
	for i, v := range vars {
		anyVars[i] = v
	}
	prompt := fmt.Sprintf(promptTpl, anyVars...)

	reqBody, err := json.Marshal(models.OpenAIRequest{
		Model:    model,
		Messages: []models.Message{{Role: "user", Content: prompt}},
		Stream:   true,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 限制 1MB
		return "", fmt.Errorf("AI service returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var fullContent strings.Builder
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if strings.TrimSpace(data) == "[DONE]" {
				break
			}
			var streamResp models.OpenAIStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err == nil {
				if len(streamResp.Choices) > 0 {
					content := streamResp.Choices[0].Delta.Content
					fmt.Print(content)
					fullContent.WriteString(content)
				}
			}
		}
	}
	return fullContent.String(), nil
}

// 建议流式返回切片
func (AIUtil) StreamSuggestions(promptType string, vars ...string) ([]string, error) {
	content, err := AI.StreamChat(promptType, vars...)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(content), "\n"), nil
}
