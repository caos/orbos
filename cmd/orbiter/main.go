package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v2"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/core/secret"
	"github.com/caos/orbiter/internal/edge/watcher/cron"
	"github.com/caos/orbiter/internal/edge/watcher/immediate"
	"github.com/caos/orbiter/internal/kinds/orbiter"
	"github.com/caos/orbiter/internal/kinds/orbiter/adapter"
	"github.com/caos/orbiter/internal/kinds/orbiter/model"
	"github.com/caos/orbiter/logging"
	logcontext "github.com/caos/orbiter/logging/context"
	"github.com/caos/orbiter/logging/stdlib"
)

var gitCommit string
var gitTag string

func main() {

	defer func() {
		if r := recover(); r != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("\x1b[0;31m%v\x1b[0m\n", r)))
			os.Exit(1)
		}
	}()

	verbose := flag.Bool("verbose", false, "Print logs for debugging")
	version := flag.Bool("version", false, "Print build information")
	repoURL := flag.String("repourl", "", "Repository URL")
	recur := flag.Bool("recur", false, "Continously ensures the desired state")
	destroy := flag.Bool("destroy", false, "Destroys everything")
	addSecret := flag.String("addsecret", "", "Encrypts, encodes and writes the secret passed via STDIN at the given property key in ./secrets.yml")
	readSecret := flag.String("readsecret", "", "Decodes and decrypts the secret at the given property key in ./secrets.yml and writes it to STDOUT")

	flag.Parse()

	if *version {
		fmt.Printf("%s %s\n", gitTag, gitCommit)
		os.Exit(0)
	}

	logger := logcontext.Add(stdlib.New(os.Stdout))
	if *verbose {
		logger = logger.Verbose()
	}

	masterkey := readSecretFile("masterkey")
	if readSecret != nil && *readSecret != "" {
		sec, err := initSecret(logger, *readSecret, masterkey)
		if err != nil {
			panic(err)
		}

		if err := sec.Read(os.Stdout); err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	if addSecret != nil && *addSecret != "" {

		sec, err := initSecret(logger, *addSecret, masterkey)
		if err != nil {
			panic(err)
		}

		value, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err) 
		}
		updatedSecrets, err := sec.Write(value)
		if err != nil {
			panic(err)
		}

		if err := ioutil.WriteFile("./secrets.yml", updatedSecrets, 0666); err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	logger.WithFields(map[string]interface{}{
		"version": gitTag,
		"commit":  gitCommit,
		"destroy": *destroy,
		"verbose": *verbose,
		"repoURL": *repoURL,
	}).Info("Orbiter is taking off")

	muxFlagsErr := errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
	if *recur && *destroy {
		panic(muxFlagsErr)
	}

	repoURLValue := *repoURL
	repoKey := readSecretFile("repokey")

	currentFile := "current.yml"
	secretsFile := "secrets.yml" 
	configID := strings.ReplaceAll(strings.TrimSuffix(repoURLValue[strings.LastIndex(repoURLValue, "/")+1:], ".git"), "-", "")

	op := operator.New(&operator.Arguments{
		Ctx:           context.Background(),
		Logger:        logger,
		MasterKey:     masterkey,
		RepoURL:       repoURLValue,
		DesiredFile:   "desired.yml",
		CurrentFile:   currentFile,
		SecretsFile:   secretsFile,
		DeploymentKey: repoKey,
		RepoCommitter: "Orbiter",
		Watchers: []operator.Watcher{
			immediate.New(logger),
			cron.New(logger, "@every 30s"),
		},
		RootAssembler: orbiter.New(nil, nil, adapter.New(&model.Config{
			Logger:           logger,
			ConfigID:         configID,
			OrbiterVersion:   gitTag,
			NodeagentRepoURL: *repoURL,
			NodeagentRepoKey: repoKey,
			CurrentFile:      currentFile,
			SecretsFile:      secretsFile,
			Masterkey:        masterkey,
		})),
	})

	iterations := make(chan *operator.IterationDone)
	if err := op.Initialize(); err != nil {
		panic(err)
	}

	go op.Run(iterations)

outer:
	for it := range iterations {
		if it.Error != nil {
			logger.Error(it.Error)
		}

		if *destroy {
			return
		}

		if !*recur {
			if it.Error != nil {
				return
			}
			statusReader := struct {
				Deps map[string]struct {
					Current struct {
						State struct {
							Status string
						}
					}
				}
			}{}
			yaml.Unmarshal(it.Current, &statusReader)
			for _, cluster := range statusReader.Deps {
				if cluster.Current.State.Status != "running" {
					continue outer
				}
			}
			return
		}
	}
}

func initSecret(logger logging.Logger, property string, masterkey string) (*secret.Secret, error) {
	secrets, err := ioutil.ReadFile("./secrets.yml")
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		secrets = make([]byte, 0)
	}
	return secret.New(logger, secrets, property, masterkey), nil
}

func readSecretFile(sec string) string {
	secretsPath := "/etc/orbiter/" + sec
	secret, err := ioutil.ReadFile(secretsPath)
	if err != nil {
		panic(fmt.Sprintf("secret not found at %s", secretsPath))
	}
	return string(secret)
}
