package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/gig/cmd/add"
	"github.com/cicbyte/gig/cmd/completion"
	"github.com/cicbyte/gig/cmd/config"
	"github.com/cicbyte/gig/cmd/doctor"
	"github.com/cicbyte/gig/cmd/ignore"
	"github.com/cicbyte/gig/cmd/check"
	"github.com/cicbyte/gig/cmd/refactor"
	"github.com/cicbyte/gig/cmd/template"
	"github.com/cicbyte/gig/cmd/track"
	"github.com/cicbyte/gig/cmd/version"
	"github.com/cicbyte/gig/cmd/view"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/defaults"
	"github.com/cicbyte/gig/internal/models"
	"github.com/cicbyte/gig/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	cfgFile string
	genType string
)

var rootCmd = &cobra.Command{
	Use:   "gig",
	Short: "一个智能的 .gitignore 管理工具",
	Long: `gig 是一个命令行工具，旨在简化和增强您管理 .gitignore 文件的方式。

它提供了从模板生成、AI 智能重构到文件审计和问题诊断的全生命周期管理功能。
使用 gig，您可以轻松保持 .gitignore 文件的整洁、高效和最新。`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件 (默认为 $HOME/.cicbyte/gig/config/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&genType, "type", "t", "local", "模板源的类型 (local, github, ai)")

	rootCmd.AddCommand(add.GetAddCommand())
	rootCmd.AddCommand(ignore.GetIgnoreCommand())
	rootCmd.AddCommand(refactor.GetRefactorCommand())
	rootCmd.AddCommand(doctor.GetDoctorCommand())
	rootCmd.AddCommand(check.GetCheckCommand())
	rootCmd.AddCommand(track.GetTrackCommand())
	rootCmd.AddCommand(template.GetTemplateCommand())
	rootCmd.AddCommand(config.GetConfigCommand())
	rootCmd.AddCommand(version.GetVersionCommand())
	rootCmd.AddCommand(completion.GetCompletionCommand())
	rootCmd.AddCommand(view.GetViewCommand())
}

func initConfig() {
	home, err := utils.GetUserHomeDir()
	cobra.CheckErr(err)

	gigDir := filepath.Join(home, ".cicbyte", "gig")
	cfgDir := filepath.Join(gigDir, "config")
	templateDir := filepath.Join(gigDir, "template")
	promptsDir := filepath.Join(cfgDir, "prompts")

	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "创建配置目录失败：", err)
		return
	}
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "创建模板目录失败：", err)
		return
	}
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "创建提示词目录失败：", err)
		return
	}

	// --- 配置文件初始化 ---
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		defaultCfgPath := filepath.Join(cfgDir, "config.yaml")
		if !utils.FileExists(defaultCfgPath) {
			fmt.Fprintln(os.Stderr, "信息：正在创建默认配置文件于", defaultCfgPath)
			createDefaultConfigFile(defaultCfgPath)
		}
		viper.SetConfigFile(defaultCfgPath)
	}

	viper.SetEnvPrefix("GIG")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "读取配置文件时出错：", err)
	}

	// --- 解析 AI 配置 ---
	type PartialConfig struct {
		AI models.AIConfig `mapstructure:"ai" yaml:"ai"`
	}
	var partial PartialConfig
	if err := viper.Unmarshal(&partial); err != nil {
		fmt.Fprintln(os.Stderr, "解析配置时出错：", err)
		os.Exit(1)
	}
	models.AppConfig.AI = partial.AI

	// --- 读取 prompts ---
	defaultConfig := defaults.GetDefaultConfig()
	models.AppConfig.Prompts = make(map[string]string)
	files, err := os.ReadDir(promptsDir)
	if err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
				name := strings.TrimSuffix(f.Name(), ".md")
				content, err := utils.ReadFile(filepath.Join(promptsDir, f.Name()))
				if err == nil {
					models.AppConfig.Prompts[name] = string(content)
				}
			}
		}
	}
	// 补写缺失的 prompt 文件
	for name, content := range defaultConfig.Prompts {
		if _, exists := models.AppConfig.Prompts[name]; !exists {
			filePath := filepath.Join(promptsDir, name+".md")
			utils.WriteFile(filePath, content)
			models.AppConfig.Prompts[name] = content
		}
	}

	// --- 读取 detection.json ---
	detectionPath := filepath.Join(cfgDir, "detection.json")
	if !utils.FileExists(detectionPath) {
		fmt.Fprintln(os.Stderr, "信息：正在创建默认检测文件于", detectionPath)
		createDefaultDetectionFile(detectionPath)
	}
	jsonBytes, err := utils.ReadFile(detectionPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "读取 detection.json 时出错：", err)
		os.Exit(1)
	}
	if err := json.Unmarshal([]byte(jsonBytes), &models.AppConfig.Detection); err != nil {
		fmt.Fprintln(os.Stderr, "解析 detection.json 时出错：", err)
		os.Exit(1)
	}

	// --- 初始化内置模板 ---
	defaults.InitBundledTemplates(templateDir)

	// --- 设置全局配置 ---
	common.AppConfigModel = &models.AppConfig
}

func createDefaultConfigFile(path string) {
	defaultConfig := defaults.GetDefaultConfig()
	type PartialConfig struct {
		AI models.AIConfig `yaml:"ai"`
	}
	partial := PartialConfig{
		AI: defaultConfig.AI,
	}
	yamlBytes, err := yaml.Marshal(&partial)
	if err != nil {
		fmt.Fprintln(os.Stderr, "创建默认配置时出错：", err)
		return
	}
	if err := utils.WriteFile(path, string(yamlBytes)); err != nil {
		fmt.Fprintln(os.Stderr, "写入默认配置文件时出错：", err)
	}

	promptsDir := filepath.Join(filepath.Dir(path), "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "创建提示词目录失败：", err)
		return
	}
	for name, content := range defaultConfig.Prompts {
		filePath := filepath.Join(promptsDir, name+".md")
		utils.WriteFile(filePath, content)
	}
}

func createDefaultDetectionFile(path string) {
	defaultConfig := defaults.GetDefaultConfig()
	jsonBytes, err := json.MarshalIndent(defaultConfig.Detection, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "创建默认检测时出错：", err)
		return
	}
	if err := utils.WriteFile(path, string(jsonBytes)); err != nil {
		fmt.Fprintln(os.Stderr, "写入默认检测文件时出错：", err)
	}
}
