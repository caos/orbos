package helper

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetAbsPath(pathParts ...string) (string, error) {

	filePath := filepath.Join(pathParts...)
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("error while getting absolute path for %s: %w", filePath, err)
	}
	return absFilePath, nil
}

func RecreatePath(pathParts ...string) error {

	absPath, err := GetAbsPath(pathParts...)
	if err != nil {
		return err
	}

	if err = os.RemoveAll(absPath); err != nil {
		return err
	}

	return os.MkdirAll(absPath, os.ModePerm)
}
