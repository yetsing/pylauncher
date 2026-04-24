package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

func getPythonVersion(executable string) (string, error) {
	output, err := RunCommand2(executable, "--version")
	if err != nil {
		return "", err
	}
	z := strings.Split(output, " ")
	if len(z) < 2 {
		return "", fmt.Errorf("could not parse python version: %s", output)
	}
	v := z[1]

	// convert version, like 3.13.0rc3 => 3.13.0-rc3
	var sb strings.Builder
	found := false
	for _, c := range v {
		if !found && !((c >= '0' && c <= '9') || c == '.') {
			sb.WriteRune('-')
			found = true
		}
		sb.WriteRune(c)
	}
	return sb.String(), nil
}

func getPythonVersionList() ([]string, error) {
	url := "https://nuget.azure.cn/v3-flatcontainer/python/index.json"
	debugLog.Printf("Fetching Python version list from %s", url)
	jsonData, err := fetchURL(url)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		log.Fatal(err)
	}

	// 获取 versions 字段
	versionsInterface, ok := result["versions"]
	if !ok {
		return nil, fmt.Errorf("versions field not found")
	}

	// 类型断言为 slice
	versionsSlice, ok := versionsInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("versions is not an array")
	}

	// 转换为字符串数组
	var versionsRaw []string
	for _, v := range versionsSlice {
		versionStr, ok := v.(string)
		if !ok {
			log.Fatal("version is not a string")
		}
		versionsRaw = append(versionsRaw, versionStr)
	}

	// 1. 将字符串转换为 version.Version 对象
	versions := make([]*version.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, err := version.NewVersion(raw)
		if err != nil {
			errorLog.Printf("warning: can not parse %q: %v\n", raw, err)
			continue
		}
		versions[i] = v
	}

	// 2. 使用 sort.Sort 进行排序
	// version.Collection 实现了 sort.Interface 接口
	sort.Sort(version.Collection(versions))

	// 3. 转回 []string
	versionList := make([]string, len(versions))
	for i, ver := range versions {
		versionList[i] = ver.Original()
	}
	return versionList, nil
}

// fetchURL 是一个辅助函数，用于发送 HTTP 请求
func fetchURL(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间，防止卡死
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置 User-Agent，模拟浏览器访问，防止被网站拦截
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, StatusCode: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
