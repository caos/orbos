package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Build arguments
	gitCommit = "none"
	gitTag    = "none"

	// orbctl
	repokey        string
	repokeyFile    string
	repokeyStdin   bool
	masterkey      string
	masterkeyFile  string
	masterkeyStdin bool
	rootCmd        = &cobra.Command{
		Use:   "orbctl [repo-url] [command]",
		Short: "Interact with your orbs",
		Long: `orbctl launches orbiters and simplifies common tasks such as updating your kubeconfig.
	Participate in our community on https://github.com/caos/orbiter
	or visit our website on https://caos.ch`,
		Args: cobra.ExactArgs(1),
	}

	// takeoff
	verbose    bool
	recur      bool
	destroy    bool
	takeoffCmd = &cobra.Command{
		Use:   "takeoff",
		Short: "Launch an orbiter",
		Long:  "Ensures a desired state",
	}

	// read secret
	readSecretCmd = &cobra.Command{
		Use:   "readsecret [name]",
		Short: "Decrypt and print to stdout",
		Args:  cobra.ExactArgs(1),
	}

	// write secret
	value          string
	file           string
	stdin          bool
	writeSecretCmd = &cobra.Command{
		Use:   "writesecret [name]",
		Short: "Encrypt and push",
		Args:  cobra.ExactArgs(1),
	}

	// kubeconfig
	kubeconfigCmd = &cobra.Command{
		Use:   "kubeconfig",
		Short: "Ensure your ~/.kube/config contains the orbs kubeconfig",
	}
)

func init() {
	rootCmd.Version = fmt.Sprintf("%s %s\n", gitTag, gitCommit)
	rootCmd.Flags().StringVar(&repokey, "repokey", "", "SSH private key value for authenticating to orbs git repo")
	rootCmd.Flags().StringVarP(&repokeyFile, "repokey-file", "", "r", "SSH private key file for authenticating to orbs git repo")
	rootCmd.Flags().BoolVar(&repokeyStdin, "repokey-stdin", false, "Read SSH private key for authenticating to orbs git repo from standard input")
	rootCmd.Flags().StringVar(&masterkey, "masterkey", "", "Secret phrase value used for encrypting and decrypting secrets")
	rootCmd.Flags().StringVarP(&masterkeyFile, "masterkey-file", "", "m", "Secret phrase file used for encrypting and decrypting secrets")
	rootCmd.Flags().BoolVar(&masterkeyStdin, "masterkey-stdin", false, "Read Secret phrase used for encrypting and decrypting secrets from standard input")

	takeoffCmd.Flags().BoolVar(&verbose, "verbose", false, "Print debug levelled logs")
	takeoffCmd.Flags().BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	takeoffCmd.Flags().BoolVar(&destroy, "destroy", false, "Destroy everything and clean up")

	writeSecretCmd.Flags().StringVar(&value, "value", "", "Secret phrase value used for encrypting and decrypting secrets")
	writeSecretCmd.Flags().StringVarP(&file, "file", "", "m", "Secret phrase file used for encrypting and decrypting secrets")
	writeSecretCmd.Flags().BoolVar(&stdin, "stdin", false, "Read Secret phrase used for encrypting and decrypting secrets from standard input")

	rootCmd.AddCommand(takeoffCmd, readSecretCmd, writeSecretCmd, kubeconfigCmd)
}

func main() {
	rootCmd.Execute()
}

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
