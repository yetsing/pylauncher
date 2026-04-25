package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	verbose       bool
	debug         bool
	listPython    bool
	gui           bool
	versionFlag   bool
	platform      string
	pythonVersion string
	cwd           string

	infoLog  *log.Logger
	debugLog *log.Logger
	errorLog *log.Logger

	cliPlatforms = map[string]string{
		"x86":   "cli-32.exe",
		"win32": "cli-32.exe",
		"x64":   "cli-64.exe",
		"arm64": "cli-arm64.exe",
	}
	guiPlatforms = map[string]string{
		"x86":   "gui-32.exe",
		"win32": "gui-32.exe",
		"x64":   "gui-64.exe",
		"arm64": "gui-arm64.exe",
	}

	CmdVersion = "0.1.0"
	GitCommit  string
	BuildTime  string
)

func init() {
	flag.BoolVar(&verbose, "verbose", false, "verbose mode")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.BoolVar(&listPython, "list-python", false, "show available python version")
	flag.BoolVar(&gui, "gui", false, "use gui launcher.exe")
	flag.BoolVar(&versionFlag, "version", false, "show version")
	flag.StringVar(&pythonVersion, "python", "3.12", "python version")
	flag.StringVar(&platform, "platform", "x64", "platform(values: x86 x64 arm64)")

	flag.Parse()

	// 初始化日志记录器
	infoLog = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	errorLog = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)

	if verbose {
		debugLog = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)
	} else {
		debugLog = log.New(io.Discard, "", 0) // 丢弃输出
	}

	if debug {
		// debug 模式会输出更详细信息
		debugLog = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Llongfile)
	}
}

func main() {
	if versionFlag {
		fmt.Printf("Version: %s (%s %s)\n", CmdVersion, GitCommit, BuildTime)
		return
	}

	platforms := cliPlatforms
	if gui {
		platforms = guiPlatforms
	}
	exeName, exists := platforms[platform]
	if !exists {
		errorLog.Fatalf("platform expected one of [x86, x64, arm64], but got %q", platform)
	}

	cwd, err := os.Getwd()
	if err != nil {
		errorLog.Fatal(err)
	}

	versions, err := getPythonVersionList()
	if err != nil {
		errorLog.Fatal(err)
	}
	if listPython {
		for _, version := range versions {
			fmt.Println(version)
		}
		return
	}

	matchVersions := make([]string, 8)
	for _, version := range versions {
		if strings.HasPrefix(version, pythonVersion) {
			matchVersions = append(matchVersions, version)
		}
	}
	if len(matchVersions) == 0 {
		errorLog.Fatalf("Can't find Python version match %s", pythonVersion)
	}
	pythonVersion = matchVersions[len(matchVersions)-1]

	infoLog.Printf("⏳ Installing Python %s", pythonVersion)
	venvPath := filepath.Join(cwd, "python3")
	pythonExecutable := filepath.Join(venvPath, "python.exe")
	installPython(pythonExecutable, pythonVersion)

	err = os.Mkdir(filepath.Join(venvPath, "Scripts"), 0755)
	if err != nil && !os.IsExist(err) {
		errorLog.Fatal(err)
	}

	infoLog.Println("🔨 Make pip wrapper")
	makePipWrapper(pythonExecutable)

	infoLog.Println("🔨 Make activate script")
	makeActivateScript()

	infoLog.Println("🔨 Make entrypoint")
	url := fmt.Sprintf("https://github.com/yetsing/pylauncher/releases/download/LauncherV1.0.0/%s", exeName)
	file, err := os.OpenFile("launcher.exe", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	err = downloadFile(url, file)
	if err != nil {
		errorLog.Printf("⚠️ Failed to download launcher.exe from github: %v", err)
		infoLog.Println("🔄 Try download launcher.exe from gitee")
		url = fmt.Sprintf("https://raw.giteeusercontent.com/ayeqing/yq-file-storage/raw/master/LauncherV1.0.0/%s", exeName)
		err = downloadFile(url, file)
		if err != nil {
			errorLog.Fatal(err)
		}
	}
	entrypointName := "main.py"
	if gui {
		entrypointName = "main.pyw"
	}
	// O_CREATE: 如果文件不存在则创建
	// O_EXCL: 与 O_CREATE 配合使用，如果文件已存在，则 OpenFile 会返回错误
	file, err = os.OpenFile(entrypointName, os.O_CREATE|os.O_EXCL|os.O_RDONLY, 0644)
	if err != nil && !os.IsExist(err) {
		errorLog.Fatal(err)
	}
	_ = file.Close()

	infoLog.Println("✅️ Done")
	infoLog.Printf("🎯 entrypoint %s", entrypointName)
}

func installPython(executable string, version string) {
	exists, err := PathExists(executable)
	if err != nil {
		errorLog.Fatal(err)
	}
	gotVersion := ""
	if exists {
		gotVersion, err = getPythonVersion(executable)
		if err != nil {
			errorLog.Fatal(err)
		}
	}
	if gotVersion == version {
		infoLog.Printf("🦘 Skip python because it already exists")
		return
	}

	if exists {
		msg := fmt.Sprintf("Install new %s (remove exist %s)?", version, gotVersion)
		override, err := Confirm(msg, false)
		if err != nil {
			errorLog.Fatal(err)
		}
		if !override {
			return
		}
	}

	url := fmt.Sprintf("https://www.nuget.org/api/v2/package/python/%s", version)
	err = DownloadUnzipAndMove(url, filepath.Join(cwd, "python3"))
	if err != nil {
		errorLog.Fatal(err)
	}
}

func makePipWrapper(pythonExecutable string) {
	pipExecutable := filepath.Join(cwd, "pip_wrapper", "bin", "pip.exe")
	exists, err := PathExists(pipExecutable)
	if err != nil {
		errorLog.Fatal(err)
	}
	if exists {
		infoLog.Println("🦘 Skip pip wrapper because it already exists")
		return
	}

	err = RunCommand(pythonExecutable, "-m", "ensurepip")
	if err != nil {
		errorLog.Fatal(err)
	}

	err = os.MkdirAll(filepath.Join(cwd, "pip_wrapper", "scripts"), 0755)
	if err != nil && !os.IsExist(err) {
		errorLog.Fatal(err)
	}

	err = os.MkdirAll(filepath.Join(cwd, "pip_wrapper", "bin"), 0755)
	if err != nil && !os.IsExist(err) {
		errorLog.Fatal(err)
	}

	pipPath := filepath.Join(cwd, "pip_wrapper", "scripts", "pip.py")
	err = stringToFile(pipCode, pipPath)
	if err != nil {
		errorLog.Fatal(err)
	}

	makePipPath := filepath.Join(cwd, "pip_wrapper", "makepip.py")
	err = stringToFile(makePipCode, makePipPath)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(makePipPath)

	err = RunCommand(pythonExecutable, makePipPath)
	if err != nil {
		errorLog.Fatal(err)
	}
}

func makeActivateScript() {
	activateCmdPath := filepath.Join(cwd, "activate.cmd")
	err := stringToFile(activateCmd, activateCmdPath)
	if err != nil {
		errorLog.Fatal(err)
	}

	activatePs1Path := filepath.Join(cwd, "activate.ps1")
	err = stringToFile(activatePs1, activatePs1Path)
	if err != nil {
		errorLog.Fatal(err)
	}
}
