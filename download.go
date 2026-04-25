package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadUnzipAndMove 执行下载、解压、移动的完整流程
func DownloadUnzipAndMove(url, targetPath string) error {
	// 1. 创建临时文件来存储下载的 zip
	// 使用 "" 作为第一个参数会让 Go 自动使用系统的临时目录（如 /tmp 或 %TEMP%）
	tempFile, err := os.CreateTemp("", "download-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempZipPath := tempFile.Name()

	// 确保函数退出时关闭文件并删除临时文件（即使出错也要清理）
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempZipPath)
	}()

	// 2. 下载文件
	infoLog.Printf("📥 Downloading %s ...", url)
	if err := downloadFile(url, tempFile); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// 3. 解压到临时目录
	// 我们先解压到一个随机的临时目录，确认无误后再移动
	tempExtractDir, err := os.MkdirTemp("", "extract-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for decompression: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempExtractDir) // 确保清理临时解压目录
	}()
	infoLog.Println("📦 Decompressing...")
	if err := unzipFile(tempZipPath, tempExtractDir); err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}

	// 4. 移动文件夹到指定位置
	// 注意：这里假设 zip 包里只有一个根文件夹。
	// 如果 zip 包里是散乱的文件，你需要调整这里的逻辑。

	// 查找解压后的子目录（通常 zip 包里会有一个主文件夹）
	// 这里简单处理：假设 tempExtractDir 下只有一个目录，或者我们直接把 tempExtractDir 里的内容视为目标
	// 如果你的 zip 结构是 app-1.0/bin, app-1.0/lib，你需要找到 app-1.0 这个路径
	sourcePath := findRootFolder(tempExtractDir)
	sourcePath = filepath.Join(sourcePath, "tools")

	if err := MoveFileOrDir(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to move directory: %w", err)
	}

	return nil
}

// downloadFile 将 url 的内容下载并写入到 dst 文件中
func downloadFile(url string, dst *os.File) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Status: %s", resp.Status)
	}

	// 使用 io.Copy 进行流式传输，避免大文件撑爆内存
	_, err = io.Copy(dst, resp.Body)
	return err
}

// unzipFile 将 zip 文件解压到 dest 目录
func unzipFile(zipPath, dest string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func(reader *zip.ReadCloser) {
		_ = reader.Close()
	}(reader)

	for _, file := range reader.File {
		// 安全检查：防止路径遍历攻击 (如 ../../../etc/passwd)
		if !strings.HasPrefix(filepath.Clean(file.Name), filepath.Base(dest)) {
			// 简单的校验逻辑，生产环境建议更严格
			if strings.Contains(file.Name, "..") {
				continue
			}
		}

		filePath := filepath.Join(dest, file.Name)

		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		err = func() error { // 打开 zip 内的文件
			srcFile, err := file.Open()
			if err != nil {
				return err
			}
			defer func(srcFile io.ReadCloser) {
				_ = srcFile.Close()
			}(srcFile)

			// 创建目标文件
			dstFile, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer func(dstFile *os.File) {
				_ = dstFile.Close()
			}(dstFile)

			// 复制内容
			if _, err := io.Copy(dstFile, srcFile); err != nil {
				return err
			}
			return nil
		}()

		if err != nil {
			return err
		}
	}
	return nil
}

// findRootFolder 简单查找目录下的第一个子文件夹
// 如果你的 zip 包结构特殊，可以修改此逻辑
func findRootFolder(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return dir // 出错直接返回原目录
	}

	// 如果目录下只有一个文件夹，返回该文件夹路径
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(dir, entries[0].Name())
	}
	// 否则返回解压的根目录（视需求而定）
	return dir
}
