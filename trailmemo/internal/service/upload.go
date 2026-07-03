package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/trailmemo/internal/config"
)

type UploadService interface {
	UploadFile(file io.Reader, filename string, fileType string) (string, error)
	UploadAvatar(file io.Reader, filename string) (string, error)
	DeleteFile(fileURL string) error
}

type uploadService struct {
	uploadDir string // 上传目录
	maxSize   int64  // 最大文件大小
}

func NewUploadService() UploadService {
	cfg := config.Get()
	return &uploadService{
		uploadDir: cfg.Upload.Dir,
		maxSize:   int64(cfg.Upload.MaxSize) * 1024 * 1024,
	}
}

func (s *uploadService) UploadFile(file io.Reader, filename string, fileType string) (string, error) {
	if err := s.ensureUploadDir(); err != nil {
		return "", err
	}

	ext := filepath.Ext(filename) // 获取文件扩展名
	if ext == "" {
		ext = ".jpg"
	}

	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
	}

	if !allowedExts[strings.ToLower(ext)] {
		return "", errors.New("unsupported file type, only images are allowed")
	}
	// 生成新的文件名: 时间戳_文件类型_扩展名
	newFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), fileType, ext)
	subDir := filepath.Join(s.uploadDir, fileType) // 子目录

	if err := os.MkdirAll(subDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(subDir, newFilename) // 完整路径
	dst, err := os.Create(filePath)                // 创建文件
	if err != nil {
		return "", err
	}
	defer dst.Close() // 关闭文件

	// 复制文件内容到目标文件
	written, err := io.Copy(dst, file)
	if err != nil {
		return "", err
	}
	// 检查文件大小是否超过最大限制
	if written > s.maxSize {
		os.Remove(filePath)
		return "", errors.New("file size exceeds maximum limit")
	}
	// 返回相对路径
	relativePath := filepath.Join(fileType, newFilename)
	relativePath = strings.ReplaceAll(relativePath, "\\", "/")
	return relativePath, nil
}

func (s *uploadService) UploadAvatar(file io.Reader, filename string) (string, error) {
	return s.UploadFile(file, filename, "avatars")
}

func (s *uploadService) DeleteFile(fileURL string) error {
	if fileURL == "" {
		return nil
	}

	filePath := filepath.Join(s.uploadDir, fileURL) // 完整路径
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 如果文件不存在，直接返回 nil
		return nil
	}

	// 删除文件
	return os.Remove(filePath)
}

func (s *uploadService) ensureUploadDir() error {
	if _, err := os.Stat(s.uploadDir); os.IsNotExist(err) {
		return os.MkdirAll(s.uploadDir, 0755)
	}
	return nil
}
