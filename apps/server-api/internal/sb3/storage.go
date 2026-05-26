package sb3

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

type Storage interface {
	Save(context.Context, string, []byte) (string, error)
	Read(context.Context, string) ([]byte, error)
}

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir}
}

func (s *LocalStorage) Save(_ context.Context, fileName string, rawSB3 []byte) (string, error) {
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return "", err
	}

	safeName := strings.ReplaceAll(filepath.Base(fileName), " ", "-")
	finalPath := filepath.Join(s.baseDir, randomPrefix()+"-"+safeName)
	if err := os.WriteFile(finalPath, rawSB3, 0o644); err != nil {
		return "", err
	}
	return finalPath, nil
}

func (s *LocalStorage) Read(_ context.Context, filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func randomPrefix() string {
	buffer := make([]byte, 6)
	if _, err := rand.Read(buffer); err != nil {
		return "sb3"
	}
	return hex.EncodeToString(buffer)
}
