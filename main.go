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
	pythonVersion string
	cwd           string

	infoLog  *log.Logger
	debugLog *log.Logger
	errorLog *log.Logger
)

func init() {
	flag.BoolVar(&verbose, "verbose", false, "详细输出")
	flag.BoolVar(&debug, "debug", false, "调试输出")
	flag.BoolVar(&listPython, "list-python", false, "显示可用 python 版本")
	flag.StringVar(&pythonVersion, "python", "3.12", "python version")

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

	infoLog.Println("✅️ Done")
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
	defer os.Remove(makePipPath)

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
