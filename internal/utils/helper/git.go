package helper

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/caos/orbos/internal/git"
)

func CopyFolderToLocal(git *git.Client, tempDirectory, folderRelativePath string) error {
	folderPath := filepath.Join(tempDirectory, folderRelativePath)

	if err := RecreatePath(folderPath); err != nil {
		return err
	}

	files, err := git.ReadFolder(folderRelativePath)
	if err != nil {
		return err
	}

	for filename, file := range files {
		filePath := filepath.Join(folderPath, filename)
		err := ioutil.WriteFile(filePath, file, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
