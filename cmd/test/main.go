package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/boom/api"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/orb"
	orbconfig "github.com/caos/orbiter/internal/orb"
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/mntr"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

func main() {
	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	orbconfigPath := "/Users/benz/.orb/config"
	//monitor = monitor.Verbose()

	content, err := ioutil.ReadFile(orbconfigPath)
	if err != nil {
		fmt.Println(err.Error())
	}

	orbconfig := &orbconfig.Orb{}
	if err := yaml.Unmarshal(content, orbconfig); err != nil {
		fmt.Println(err.Error())
	}

	if orbconfig.URL == "" {
		fmt.Println("orbconfig has no URL configured")
	}

	if orbconfig.Repokey == "" {
		fmt.Println("orbconfig has no repokey configured")
	}

	if orbconfig.Masterkey == "" {
		fmt.Println("orbconfig has no masterkey configured")
	}

	ctx := context.Background()

	gitClient := git.New(ctx, monitor, "Orbiter", "orbiter@caos.ch", orbconfig.URL)
	if err := gitClient.Init([]byte(orbconfig.Repokey)); err != nil {
		panic(err)
	}

	path := "boom.grafana.admin.username"
	//path := ""
	//writeFilePath := "/Users/benz/.ssh/test.pub"
	writeFilePath := ""
	value := "admin"

	//read secret
	/*value, err := readSecret(monitor, gitClient, orbconfig, path)
	if err != nil {
		os.Stdout.Write([]byte(err.Error()))
	} else {
		os.Stdout.Write([]byte(value))
	}*/

	//change secret
	/*err = writeSecret(monitor, gitClient, orbconfig, path, value, writeFilePath, false)
	if err != nil {
		os.Stdout.Write([]byte(err.Error()))
	}*/

	value = "test"
	path = "boom.grafana.admin.password"
	//change secret
	err = writeSecret(monitor, gitClient, orbconfig, path, value, writeFilePath, false)
	if err != nil {
		os.Stdout.Write([]byte(err.Error()))
	}

	//read secret again
	/*value, err = readSecret(monitor, gitClient, orbconfig, path)
	if err != nil {
		os.Stdout.Write([]byte(err.Error()))
	} else {
		os.Stdout.Write([]byte(value))
	}*/
}

func readSecret(logger mntr.Monitor, gitClient *git.Client, orbconfig *orbconfig.Orb, path string) (string, error) {

	secretFunc := func(operator string) secret.Func {
		if operator == "boom" {
			return api.SecretFunc(orbconfig)
		} else if operator == "orbiter" {
			return orb.SecretsFunc(orbconfig)
		}
		return nil
	}

	return secret.Read(
		logger,
		gitClient,
		secretFunc,
		path)
}

func writeSecret(logger mntr.Monitor, gitClient *git.Client, orbconfig *orbconfig.Orb, path string, value, file string, stdin bool) error {
	s, err := key(value, file, stdin)
	if err != nil {
		return err
	}

	secretFunc := func(operator string) secret.Func {
		if operator == "boom" {
			return api.SecretFunc(orbconfig)
		} else if operator == "orbiter" {
			return orb.SecretsFunc(orbconfig)
		}
		return nil
	}

	if err := secret.Write(
		logger,
		gitClient,
		secretFunc,
		path,
		s); err != nil {
		panic(err)
	}
	return nil
}

func key(value string, file string, stdin bool) (string, error) {

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
		return "", errors.New("Key must be provided eighter by value or by file path or by standard input")
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

	key, err := readFunc()
	if err != nil {
		panic(err)
	}
	return string(key), err
}
