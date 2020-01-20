// +build test integration

package core

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Vipers struct {
	Config  *viper.Viper
	Secrets *viper.Viper
}

func Config() func() *Vipers {
	configFlag := flag.String("config", "", "Configuration file path")
	secretsFlag := flag.String("secrets", "", "Secrets file path")
	var initializedConfig *Vipers
	return func() *Vipers {
		if initializedConfig != nil {
			fmt.Println("Returning already initialized config")
			return initializedConfig
		}

		if !flag.Parsed() {
			fmt.Println("Parsing flags")
			flag.Parse()
		}

		fmt.Println("Reading uninitialized config")

		if configFlag == nil || *configFlag == "" {
			panic(fmt.Errorf("--config flag not in program arguments %s", strings.Join(os.Args, " ")))
		}
		cfg, err := readViper(*configFlag)
		if err != nil {
			panic(err)
		}

		if secretsFlag == nil || *secretsFlag == "" {
			panic(fmt.Errorf("--secrets flag not in program arguments %s", strings.Join(os.Args, " ")))
		}
		secrets, err := readViper(*secretsFlag)
		if err != nil {
			panic(err)
		}

		initializedConfig = &Vipers{
			Config:  cfg,
			Secrets: secrets,
		}
		return initializedConfig
	}
}

func ProvidersUnderTest(initCallback func() *Vipers) []Provider {
	cfg := initCallback()
	return []Provider{
		Gce(cfg.Config.Sub("spec.providers.gce"), cfg.Secrets),
		//		Static(cfg.config.Sub("spec.providers.static"), cfg.secrets),
	}
}

func readViper(path string) (*viper.Viper, error) {

	cfg := viper.New()
	cfg.SetConfigFile(path)

	return cfg, cfg.ReadInConfig()
}
