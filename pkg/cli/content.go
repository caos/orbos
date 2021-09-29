package cli

import (
	"errors"
	"github.com/caos/orbos/mntr"
	"io/ioutil"
	"os"
)

func Content(value string, file string, stdin bool) (val string, err error) {
	defer func() {
		if err != nil {
			err = mntr.ToUserError(err)
		}
	}()

	channels := 0
	if value != "" {
		channels++
	}
	if file != "" {
		channels++
	}
	if stdin {
		channels++
	}

	if channels != 1 {
		return "", errors.New("content must be provided eighter by value or by file path or by standard input")
	}

	if value != "" {
		return value, nil
	}

	readFunc := func() ([]byte, error) {
		return ioutil.ReadFile(file)
	}
	if stdin {
		readFunc = func() ([]byte, error) {
			return ioutil.ReadAll(os.Stdin)
		}
	}

	c, err := readFunc()
	if err != nil {
		panic(err)
	}
	return string(c), err
}
