package main

import (
	"context"
	"github.com/caos/orbos/internal/git"
	orbc "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/secret/operators"
	"github.com/caos/orbos/internal/stores/github"
	"github.com/caos/orbos/mntr"
	"math/rand"
	"os"
)

func main() {

	github.ClientID = "584b9ee26948a7fc9152"
	github.ClientSecret = "18b602505f8e020446007899252e85372934ff0c"
	github.Key = RandStringBytes(32)

	//ReadAndWriteSecretWithoutKey()
	ReadAndWriteSecretWithKey()
}

func ReadAndWriteSecretWithoutKey() {
	orbconfig := "/Users/benz/.orb/boom-test-config"
	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	orbStruct, err := orbc.ParseOrbConfig(orbconfig)
	if err != nil {
		os.Exit(1)
	}

	ctx := context.Background()

	gitClient := git.New(ctx, monitor, "orbos", "orbos@caos.ch")

	if err := gitClient.Configure(orbStruct.URL, []byte(orbStruct.Repokey)); err != nil {
		os.Exit(1)
	}

	if err := gitClient.Clone(); err != nil {
		os.Exit(1)
	}
	//path := ""

	/*secrets["consoleenvironmentjson"] = conf.ConsoleEnvironmentJSON
	secrets["serviceaccountjson"] = conf.Tracing.ServiceAccountJSON
	secrets["keys"] = conf.Secrets.Keys
	secrets["googlechaturl"] = conf.Notifications.GoogleChatURL
	secrets["twiliosid"] = conf.Notifications.Twilio.SID
	secrets["twilioauthtoken"] = conf.Notifications.Twilio.AuthToken
	secrets["emailappkey"] = conf.Notifications.Email.AppKey*/

	path := "zitadel.consoleenvironmentjson"
	s := "admin"
	if err := secret.Write(
		monitor,
		gitClient,
		path,
		s,
		operators.GetAllSecretsFunc()); err != nil {
		panic(err)
	}

	value, err := secret.Read(
		monitor,
		gitClient,
		path,
		operators.GetAllSecretsFunc())

	if err != nil {
		panic(err)
	}
	if _, err := os.Stdout.Write([]byte(value)); err != nil {
		panic(err)
	}
}

func ReadAndWriteSecretWithKey() {
	orbconfig := "/Users/benz/.orb/stefan-orbos-gce"
	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	orbStruct, err := orbc.ParseOrbConfig(orbconfig)
	if err != nil {
		os.Exit(1)
	}
	ctx := context.Background()

	gitClient := git.New(ctx, monitor, "orbos", "orbos@caos.ch")

	if err := gitClient.Configure(orbStruct.URL, []byte(orbStruct.Repokey)); err != nil {
		os.Exit(1)
	}

	if err := gitClient.Clone(); err != nil {
		os.Exit(1)
	}

	path := ""
	//path := "orbiter.orbos-benz.kubeconfig"
	s := "admin"
	if err := secret.Write(
		monitor,
		gitClient,
		path,
		s,
		operators.GetAllSecretsFunc()); err != nil {
		panic(err)
	}

	value, err := secret.Read(
		monitor,
		gitClient,
		path,
		operators.GetAllSecretsFunc())

	if err != nil {
		panic(err)
	}
	if _, err := os.Stdout.Write([]byte(value)); err != nil {
		panic(err)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
