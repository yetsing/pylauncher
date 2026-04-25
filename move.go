package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// MoveFileOrDir 移动文件或目录，支持跨磁盘分区
func MoveFileOrDir(src, dst string) error {
	// 1. 首先尝试直接使用 os.Rename (原子操作，最快)
	err := os.Rename(src, dst)
	if err == nil {
		return nil // 移动成功
	}

	// 2. 检查错误是否为跨设备/跨盘符
	// 在 Windows 上，这通常对应 syscall.Errno(17) 或特定的错误消息
	// 在 Linux 上对应 syscall.EXDEV
	if isCrossDeviceError(err) {
		infoLog.Println("📋 Use copy & delete cause of move cross device")
		return moveByCopyDelete(src, dst)
	}

	// 3. 其他错误直接返回
	return fmt.Errorf("failed to move: %w", err)
}

// isCrossDeviceError 判断错误是否为跨设备错误
func isCrossDeviceError(err error) bool {
	if err == nil {
		return false
	}
	// 尝试转换为 syscall.Errno 进行判断
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			// 17 是 Linux 下的 EXDEV，Windows 下通常也有对应的 Errno
			// 这里主要依靠错误类型判断，或者简单粗暴地判断错误信息
			return errors.Is(errno, syscall.EXDEV)
		}
	}
	// 兜底策略：检查错误信息中是否包含特定字符串（不推荐但有时有效）
	// 实际上，os.Rename 在 Windows 跨盘时通常直接返回错误，我们捕获后执行降级策略即可
	// 为了简化逻辑，只要 Rename 失败，且不是权限等问题，通常可以尝试降级，
	// 但最稳妥的是捕获到 Rename 失败后，直接判断是否需要 Copy+Delete。
	// 在这里，为了代码简洁，我们可以在 MoveFileOrDir 中直接捕获 Rename 失败并尝试 Copy+Delete，
	// 但为了严谨，我们保留这个判断逻辑。
	return true
}

// moveByCopyDelete 核心逻辑：复制然后删除
func moveByCopyDelete(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if srcInfo.IsDir() {
		return moveDirCopyDelete(src, dst)
	}
	return moveFileCopyDelete(src, dst)
}

// 移动文件：复制内容 -> 删除源文件
func moveFileCopyDelete(src, dst string) error {
	// 1. 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			errorLog.Println(err)
		}
	}(srcFile)

	// 2. 创建目标文件 (如果存在则覆盖，这符合移动语义)
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(dstFile *os.File) {
		err := dstFile.Close()
		if err != nil {
			errorLog.Println(err)
		}
	}(dstFile)

	// 3. 复制内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// 4. 同步权限 (可选，保持原文件权限)
	srcInfo, _ := os.Stat(src)
	err = os.Chmod(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	// 5. 关闭文件以确保数据写入磁盘
	dstFile.Close()
	srcFile.Close()

	// 6. 删除源文件
	return os.Remove(src)
}

// moveDirCopyDelete 优化版：直接复制，不再尝试 Rename
func moveDirCopyDelete(src, dst string) error {
	// 1. 创建目标目录
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// 2. 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 3. 递归复制所有子项（注意：这里直接调用复制函数，不经过 MoveFileOrDir）
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 【优化点】：直接判断是文件还是目录，分别调用复制函数
		// 避免再次进入 MoveFileOrDir 去做无谓的 os.Rename 尝试
		if entry.IsDir() {
			if err := moveDirCopyDelete(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	// 4. 所有文件复制完成后，再删除源目录及其内容
	// 注意：这里不能只删除空目录，要删除整个树
	return os.RemoveAll(src)
}

// 独立的文件复制函数（不含 Rename 逻辑）
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// 保留权限
	srcInfo, _ := os.Stat(src)
	return os.Chmod(dst, srcInfo.Mode())
}
