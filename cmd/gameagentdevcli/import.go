package main

import (
    "io"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
    Use:   "import <file>",
    Short: "将 YAML/JSON 导入到服务端，支持文件或标准输入",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        format, _ := cmd.Flags().GetString("format")
        dryRun, _ := cmd.Flags().GetBool("dry-run")
        reset, _ := cmd.Flags().GetBool("reset")
        var data []byte
        var err error
        if args[0] == "-" {
            data, err = io.ReadAll(os.Stdin)
            if err != nil {
                fail(err)
            }
        } else {
            data, err = os.ReadFile(args[0])
            if err != nil {
                fail(err)
            }
            if !cmd.Flags().Changed("format") {
                ext := filepath.Ext(args[0])
                if ext == ".json" {
                    format = "json"
                } else {
                    format = "yaml"
                }
            }
        }

        result, err := newClient().CreatorImport(format, string(data), reset, dryRun)
        if err != nil {
            fail(err)
        }
        printJSON(result)
    },
}

func init() {
    importCmd.Flags().String("format", "yaml", "当导入源为 - 时，指定 stdin 内容格式：yaml 或 json")
    importCmd.Flags().Bool("dry-run", false, "仅校验导入内容，不写入数据库")
    importCmd.Flags().Bool("reset", false, "导入前清空当前数据库")
}
