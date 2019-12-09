package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/core/secret"
	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/internal/edge/watcher/cron"
	"github.com/caos/orbiter/internal/edge/watcher/immediate"
	"github.com/caos/orbiter/internal/kinds/orbiter"
	"github.com/caos/orbiter/internal/kinds/orbiter/adapter"
	"github.com/caos/orbiter/internal/kinds/orbiter/model"
	"github.com/caos/orbiter/logging"
	logcontext "github.com/caos/orbiter/logging/context"
	"github.com/caos/orbiter/logging/stdlib"
)

var (
	// Build arguments
	gitCommit = "none"
	gitTag    = "none"

	// orbctl
	repoURL        string
	repokey        string
	repokeyFile    string
	repokeyStdin   bool
	masterkey      string
	masterkeyFile  string
	masterkeyStdin bool
	rootCmd        = &cobra.Command{
		Use:   "orbctl",
		Short: "Interact with your orbs",
		Long: `orbctl launches orbiters and simplifies common tasks such as updating your kubeconfig.
	Participate in our community on https://github.com/caos/orbiter
	or visit our website on https://caos.ch`,
	}

	// takeoff
	verbose    bool
	recur      bool
	destroy    bool
	takeoffCmd = &cobra.Command{
		Use:   "takeoff",
		Short: "Launch an orbiter",
		Long:  "Ensures a desired state",
		RunE: func(cmd *cobra.Command, args []string) error {
			if recur && destroy {
				return errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
			}

			ctx, logger, gitClient, rk, mk, err := commonValues()
			if err != nil {
				return err
			}

			logger.WithFields(map[string]interface{}{
				"version": gitTag,
				"commit":  gitCommit,
				"destroy": destroy,
				"verbose": verbose,
				"repoURL": repoURL,
			}).Info("Orbiter is taking off")

			currentFile := "current.yml"
			secretsFile := "secrets.yml"
			configID := strings.ReplaceAll(strings.TrimSuffix(repoURL[strings.LastIndex(repoURL, "/")+1:], ".git"), "-", "")

			op := operator.New(&operator.Arguments{
				Ctx:         ctx,
				Logger:      logger,
				GitClient:   gitClient,
				MasterKey:   mk,
				DesiredFile: "desired.yml",
				CurrentFile: currentFile,
				SecretsFile: secretsFile,
				Watchers: []operator.Watcher{
					immediate.New(logger),
					cron.New(logger, "@every 30s"),
				},
				RootAssembler: orbiter.New(nil, nil, adapter.New(&model.Config{
					Logger:           logger,
					ConfigID:         configID,
					OrbiterVersion:   gitTag,
					NodeagentRepoURL: repoURL,
					NodeagentRepoKey: rk,
					CurrentFile:      currentFile,
					SecretsFile:      secretsFile,
					Masterkey:        mk,
				})),
			})

			iterations := make(chan *operator.IterationDone)
			if err := op.Initialize(); err != nil {
				panic(err)
			}

			go op.Run(iterations)

		outer:
			for it := range iterations {
				if destroy {
					if it.Error != nil {
						panic(it.Error)
					}
					return nil
				}

				if recur {
					if it.Error != nil {
						logger.Error(it.Error)
					}
					continue
				}

				if it.Error != nil {
					panic(it.Error)
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
				break
			}
			return nil
		},
	}

	// read secret
	readSecretCmd = &cobra.Command{
		Use:   "readsecret [name]",
		Short: "Decrypt and print to stdout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			_, logger, gitClient, _, mk, err := commonValues()
			if err != nil {
				return err
			}

			if err := gitClient.Clone(); err != nil {
				panic(err)
			}

			sec, err := gitClient.Read("secrets.yml")
			if err != nil {
				panic(err)
			}

			if err := secret.New(logger, sec, args[0], mk).Read(os.Stdout); err != nil {
				panic(err)
			}
			return nil
		},
	}

	// write secret
	value          string
	file           string
	stdin          bool
	writeSecretCmd = &cobra.Command{
		Use:   "writesecret [name]",
		Short: "Encrypt and push",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if stdin && (masterkeyStdin || repokeyStdin) {
				return errMultipleStdinKeys
			}

			s, err := key(value, file, stdin)
			if err != nil {
				return err
			}

			_, logger, gitClient, _, mk, err := commonValues()
			if err != nil {
				return err
			}

			if err := gitClient.Clone(); err != nil {
				panic(err)
			}

			sec, err := gitClient.Read("secrets.yml")
			if err != nil {
				panic(err)
			}

			if err := secret.New(logger, sec, args[0], mk).Write([]byte(s)); err != nil {
				panic(err)
			}

			if _, err := gitClient.UpdateRemoteUntilItWorks(&git.File{
				Path: "secrets.yml",
				Overwrite: func(o map[string]interface{}) ([]byte, error) {
					o[args[0]] = sec[args[0]]
					return yaml.Marshal(o)
				},
				Force: true,
			}); err != nil {
				panic(err)
			}
			return nil
		},
	}
)

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

var errMultipleStdinKeys = errors.New("Reading multiple keys from standard input does not work")

func commonValues() (context.Context, logging.Logger, *git.Client, string, string, error) {
	rk, mk, err := keys()
	if err != nil {
		return nil, nil, nil, "", "", err
	}

	logger := logcontext.Add(stdlib.New(os.Stdout))
	if verbose {
		logger = logger.Verbose()
	}

	ctx := context.Background()
	gitClient := git.New(ctx, logger, "Orbiter", repoURL)
	if err := gitClient.Init([]byte(rk)); err != nil {
		panic(err)
	}

	return ctx, logger, gitClient, rk, mk, nil
}

func keys() (string, string, error) {
	if masterkeyStdin && repokeyStdin {
		return "", "", errMultipleStdinKeys
	}

	rk, err := key(repokey, repokeyFile, repokeyStdin)
	if err != nil {
		return "", "", errors.Wrap(err, "repokey")
	}
	mk, err := key(masterkey, masterkeyFile, masterkeyStdin)
	return rk, mk, errors.Wrap(err, "masterkey")
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s %s\n", gitTag, gitCommit)
	rootCmd.PersistentFlags().StringVarP(&repoURL, "repourl", "g", "", "Use this orbs Git repo")
	rootCmd.MarkPersistentFlagRequired("repourl")
	rootCmd.PersistentFlags().StringVar(&repokey, "repokey", "", "SSH private key value for authenticating to orbs git repo")
	rootCmd.PersistentFlags().StringVarP(&repokeyFile, "repokey-file", "r", "", "SSH private key file for authenticating to orbs git repo")
	rootCmd.PersistentFlags().BoolVar(&repokeyStdin, "repokey-stdin", false, "Read SSH private key for authenticating to orbs git repo from standard input")
	rootCmd.PersistentFlags().StringVar(&masterkey, "masterkey", "", "Secret phrase value used for encrypting and decrypting secrets")
	rootCmd.PersistentFlags().StringVarP(&masterkeyFile, "masterkey-file", "m", "", "Secret phrase file used for encrypting and decrypting secrets")
	rootCmd.PersistentFlags().BoolVar(&masterkeyStdin, "masterkey-stdin", false, "Read Secret phrase used for encrypting and decrypting secrets from standard input")

	takeoffCmd.Flags().BoolVar(&verbose, "verbose", false, "Print debug levelled logs")
	takeoffCmd.Flags().BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	takeoffCmd.Flags().BoolVar(&destroy, "destroy", false, "Destroy everything and clean up")

	writeSecretCmd.Flags().StringVar(&value, "value", "", "Secret phrase value used for encrypting and decrypting secrets")
	writeSecretCmd.Flags().StringVarP(&file, "file", "f", "", "Secret phrase file used for encrypting and decrypting secrets")
	writeSecretCmd.Flags().BoolVar(&stdin, "stdin", false, "Read Secret phrase used for encrypting and decrypting secrets from standard input")

	rootCmd.AddCommand(takeoffCmd, readSecretCmd, writeSecretCmd)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("\x1b[0;31m%v\x1b[0m\n", r)))
			os.Exit(1)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

/*
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

func logger() logging.Logger {
	l := logcontext.Add(stdlib.New(os.Stdout))
	if verbose {
		l = l.Verbose()
	}
	return l
}
*/

/*
func main() {

	defer func() {
		if r := recover(); r != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("\x1b[0;31m%v\x1b[0m\n", r)))
			os.Exit(1)
		}
	}()

	verbose := flag.Bool("verbose", false, "Print logs for debugging")
	repoURL := flag.String("repourl", "", "Repository URL")
	recur := flag.Bool("recur", false, "Continously ensures the desired state")
	destroy := flag.Bool("destroy", false, "Destroys everything")
	addSecret := flag.String("addsecret", "", "Encrypts, encodes and writes the secret passed via STDIN at the given property key in ./secrets.yml")
	readSecret := flag.String("readsecret", "", "Decodes and decrypts the secret at the given property key in ./secrets.yml and writes it to STDOUT")

	flag.Parse()

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
*/
