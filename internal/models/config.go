package models

// Config holds all configuration for the application.
// It includes AI settings and the prompts for different commands.
type Config struct {
	AI        AIConfig          `mapstructure:"ai" yaml:"ai"`
	Prompts   map[string]string `mapstructure:"prompts"`
	Detection DetectionConfig   `mapstructure:"detection"`
}

// DetectionConfig holds the configuration for project type detection.
type DetectionConfig struct {
	FileMap map[string]string `mapstructure:"filemap"`
}

// AIConfig holds the configuration for the AI service.
type AIConfig struct {
	APIKey string `mapstructure:"api_key" yaml:"api_key"`
	URL    string `mapstructure:"url" yaml:"url"`
	Model  string `mapstructure:"model" yaml:"model"`
}

// AppConfig is the global configuration instance.
var AppConfig Config

// --- OpenAI API Structures ---

// OpenAIRequest defines the structure for a request to the OpenAI API.
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message represents a single message in the chat history.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIStreamResponse defines the structure for a chunk in a streaming response.
type OpenAIStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
