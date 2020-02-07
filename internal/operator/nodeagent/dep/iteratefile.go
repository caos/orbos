package dep

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func Manipulate(from io.Reader, to io.Writer, removeContaining, append []string, eachLine func(string) *string) error {

	scanner := bufio.NewScanner(from)

outer:
	for scanner.Scan() {
		line := scanner.Text()
		for _, remove := range removeContaining {
			if strings.Contains(line, remove) {
				continue outer
			}
		}
		if eachLine != nil {
			editLine := eachLine(line)
			if editLine == nil {
				continue
			}
			line = *editLine
		}

		if _, err := to.Write([]byte(line + "\n")); err != nil {
			return err
		}
	}

	if len(append) == 0 {
		return nil
	}

	_, err := to.Write([]byte(strings.Join(append, "\n") + "\n"))
	return err
}

func ManipulateFile(path string, removeContaining, append []string, eachLine func(string) *string) (err error) {
	tmpPath := path + ".tmp"

	if err := createTmpFile(path, tmpPath, removeContaining, append, eachLine); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func createTmpFile(path string, tmpPath string, removeContaining, append []string, eachLine func(string) *string) (err error) {

	closeFile := func(file *os.File) {
		closeErr := errors.Wrap(file.Close(), "closing file failed")
		if err == nil {
			err = closeErr
			return
		}
		if closeErr != nil {
			err = errors.Wrap(err, "closing file also failed:"+closeErr.Error())
		}
	}

	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	file, err := os.OpenFile(path, os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer closeFile(file)

	return Manipulate(file, tmpFile, removeContaining, append, eachLine)
}
