package config

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cicbyte/gig/internal/defaults"
	"github.com/cicbyte/gig/internal/models"
	"github.com/spf13/viper"
)

type ConfigProcessor struct {
	appConfig *models.Config
}

func NewConfigProcessor(appConfig *models.Config) *ConfigProcessor {
	return &ConfigProcessor{appConfig: appConfig}
}

type ConfigItem struct {
	Key   string
	Value string
}

type ConfigShowResult struct {
	ConfigFile string
	Items      []ConfigItem
}

func (p *ConfigProcessor) Show(ctx context.Context) (*ConfigShowResult, error) {
	items := []ConfigItem{
		{Key: "ai.api_key", Value: viper.GetString("ai.api_key")},
		{Key: "ai.url", Value: viper.GetString("ai.url")},
		{Key: "ai.model", Value: viper.GetString("ai.model")},
	}

	cfgFile := viper.ConfigFileUsed()
	if cfgFile == "" {
		home, _ := os.UserHomeDir()
		cfgFile = filepath.Join(home, ".cicbyte", "gig", "config", "config.yaml")
	}

	return &ConfigShowResult{ConfigFile: cfgFile, Items: items}, nil
}

func (p *ConfigProcessor) Set(ctx context.Context, key, value string) error {
	validKeys := map[string]bool{
		"ai.api_key": true,
		"ai.url":     true,
		"ai.model":   true,
	}
	if !validKeys[key] {
		return fmt.Errorf("未知的配置项 '%s'，可用项：ai.api_key, ai.url, ai.model", key)
	}
	viper.Set(key, value)
	return p.Save(ctx)
}

func (p *ConfigProcessor) Reset(ctx context.Context, key string) error {
	defaultCfg := defaults.GetDefaultConfig()

	if key == "" {
		// 重置全部
		viper.Set("ai.api_key", defaultCfg.AI.APIKey)
		viper.Set("ai.url", defaultCfg.AI.URL)
		viper.Set("ai.model", defaultCfg.AI.Model)
		return p.Save(ctx)
	}

	mapping := map[string]string{
		"ai.api_key": defaultCfg.AI.APIKey,
		"ai.url":     defaultCfg.AI.URL,
		"ai.model":   defaultCfg.AI.Model,
	}
	if val, ok := mapping[key]; ok {
		viper.Set(key, val)
		return p.Save(ctx)
	}
	return fmt.Errorf("未知的配置项 '%s'，可用项：ai.api_key, ai.url, ai.model", key)
}

func (p *ConfigProcessor) Save(ctx context.Context) error {
	cfgFile := viper.ConfigFileUsed()
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("查找主目录时出错：%w", err)
		}
		cfgFile = filepath.Join(home, ".cicbyte", "gig", "config", "config.yaml")
	}

	if err := viper.WriteConfigAs(cfgFile); err != nil {
		return fmt.Errorf("写入配置文件时出错：%w", err)
	}
	fmt.Println("配置已保存至", cfgFile)
	return nil
}

func (p *ConfigProcessor) Edit(ctx context.Context) error {
	cfgFile := viper.ConfigFileUsed()
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("查找主目录时出错：%w", err)
		}
		cfgFile = filepath.Join(home, ".cicbyte", "gig", "config", "config.yaml")
	}

	// 确保文件存在
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return p.Save(ctx)
	}

	editor := os.Getenv("EDITOR")
	var cmd *exec.Cmd
	if editor != "" {
		parts := strings.Fields(editor)
		if len(parts) == 0 {
			switch runtime.GOOS {
			case "windows":
				cmd = exec.Command("notepad", cfgFile)
			case "darwin":
				cmd = exec.Command("open", "-t", cfgFile)
			default:
				cmd = exec.Command("xdg-open", cfgFile)
			}
		} else {
			cmd = exec.Command(parts[0], append(parts[1:], cfgFile)...)
		}
	} else {
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("notepad", cfgFile)
		case "darwin":
			cmd = exec.Command("open", "-t", cfgFile)
		default:
			cmd = exec.Command("xdg-open", cfgFile)
		}
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// MaskAPIKey 脱敏 API Key
func MaskAPIKey(key string) string {
	if key == "" {
		return "<未设置>"
	}
	if len(key) <= 5 {
		return "***"
	}
	return key[:2] + "***" + key[len(key)-3:]
}

// FormatValue 格式化配置值显示
func FormatValue(key, value string) string {
	if strings.HasSuffix(key, "api_key") {
		return MaskAPIKey(value)
	}
	if value == "" {
		return "<未设置>"
	}
	return value
}
