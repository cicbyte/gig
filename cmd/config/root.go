package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/gig/internal/common"
	"github.com/cicbyte/gig/internal/logic/config"
	"github.com/spf13/cobra"
)

func GetConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "管理应用程序配置。",
		Long: `管理 gig 的配置设置。

不带参数时显示当前所有配置项。
支持子命令 list/set/edit/reset 进行精细操作。
也支持直接模式：gig config <key> <value>。`,
		Args: cobra.RangeArgs(0, 2),
		Run: func(cmd *cobra.Command, args []string) {
			processor := config.NewConfigProcessor(common.AppConfigModel)

			if len(args) == 0 {
				// 展示当前配置
				result, err := processor.Show(context.Background())
				if err != nil {
					fmt.Fprintf(os.Stderr, "错误：%v\n", err)
					return
				}
				fmt.Printf("当前配置 (%s)\n", result.ConfigFile)
				fmt.Println(strings.Repeat("─", 50))
				for _, item := range result.Items {
					fmt.Printf("  %-14s%s\n", item.Key, config.FormatValue(item.Key, item.Value))
				}
			} else {
				// 两个参数：当作 set
				if err := processor.Set(context.Background(), args[0], args[1]); err != nil {
					fmt.Fprintf(os.Stderr, "错误：%v\n", err)
				}
			}
		},
	}

	cmd.AddCommand(getConfigListCommand())
	cmd.AddCommand(getConfigSetCommand())
	cmd.AddCommand(getConfigEditCommand())
	cmd.AddCommand(getConfigResetCommand())
	return cmd
}

func getConfigListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "以表格形式展示所有配置项。",
		Run: func(cmd *cobra.Command, args []string) {
			processor := config.NewConfigProcessor(common.AppConfigModel)
			result, err := processor.Show(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误：%v\n", err)
				return
			}
			fmt.Printf("当前配置 (%s)\n", result.ConfigFile)
			fmt.Println(strings.Repeat("─", 50))
			for _, item := range result.Items {
				fmt.Printf("  %-14s%s\n", item.Key, config.FormatValue(item.Key, item.Value))
			}
		},
	}
}

func getConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "设置配置项的值。",
		Long:  "设置指定配置项的值。可用项：ai.api_key, ai.url, ai.model",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			processor := config.NewConfigProcessor(common.AppConfigModel)
			if err := processor.Set(context.Background(), args[0], args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "错误：%v\n", err)
			}
		},
	}
}

func getConfigEditCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "用系统编辑器打开配置文件。",
		Run: func(cmd *cobra.Command, args []string) {
			processor := config.NewConfigProcessor(common.AppConfigModel)
			if err := processor.Edit(context.Background()); err != nil {
				fmt.Fprintf(os.Stderr, "错误：%v\n", err)
			}
		},
	}
}

func getConfigResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset [key]",
		Short: "重置配置项为默认值。",
		Long:  "重置指定配置项为默认值。不传 key 则重置全部。可用项：ai.api_key, ai.url, ai.model",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			processor := config.NewConfigProcessor(common.AppConfigModel)
			var key string
			if len(args) > 0 {
				key = args[0]
			}
			if err := processor.Reset(context.Background(), key); err != nil {
				fmt.Fprintf(os.Stderr, "错误：%v\n", err)
			}
		},
	}
}
