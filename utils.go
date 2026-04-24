package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// PathExists 判断文件或文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		// 文件存在
		return true, nil
	}
	if os.IsNotExist(err) {
		// 文件不存在
		return false, nil
	}
	// 其他错误（如权限不足等）
	return false, err
}

// RunCommand 执行命令，输出直接透传到终端，不捕获
// 返回值为执行是否成功
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	// 核心：将子进程的 stdout 和 stderr 直接指向当前程序的 stdout 和 stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run 会等待命令执行完成
	return cmd.Run()
}

// RunCommand2 执行命令并返回标准输出
// 参数: name 是命令名 (如 "ls"), args 是参数列表 (如 "-l", "-a")
func RunCommand2(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	// Output() 会执行命令并捕获 stdout
	// 如果命令执行失败（退出码非0），err 不为 nil
	out, err := cmd.Output()
	if err != nil {
		// 如果需要查看 stderr 的具体内容，可以断言错误类型
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// exitErr.Stderr 包含了命令的 stderr 内容
			return "", fmt.Errorf("failed to execute: %s", string(exitErr.Stderr))
		}
		return "", err
	}

	// 去除首尾空格（通常命令输出末尾会有换行符）
	return strings.TrimSpace(string(out)), nil
}

func stringToFile(content string, path string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// Confirm 封装 survey 库的确认功能
// message: 向用户显示的提示信息
// defaultVal: 默认选项，true 表示默认为"是"，false 表示默认为"否"
// 返回值: 用户的确认结果 (bool) 和可能发生的错误 (error)
func Confirm(message string, defaultVal bool) (bool, error) {
	var result bool
	// 创建 Confirm 类型的提示对象
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultVal,
	}

	// 显示提示并获取用户输入
	// AskOne 函数会将用户的输入结果存入 result 变量中
	err := survey.AskOne(prompt, &result)
	if err != nil {
		return false, err
	}

	return result, nil
}
