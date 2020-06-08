package helper

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func GetAbsPath(pathParts ...string) (string, error) {

	filePath := filepath.Join(pathParts...)
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "Error while getting absolute path for %s", filePath)
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
