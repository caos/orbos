package main

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	_ "net/http/pprof"
)

func main() {
	file, err := ioutil.ReadFile("orbiter.yml")
	if err != nil {
		fmt.Println(err)
	}

	tree := new(tree.Tree)
	err = yaml.Unmarshal([]byte(file), tree)
	if err != nil {
		fmt.Println(err.Error())
	}
	desiredKind := &orb.DesiredV0{Common: tree.Common}

	if err := tree.Original.Decode(desiredKind); err != nil {
		fmt.Println(err.Error())
	}
	desiredKind.Common.Version = "v0"

	fmt.Println("eah")
	fmt.Println(*tree)
	fmt.Println(*desiredKind)

	secret.Masterkey = "WXHUIhhQsLkH2vcNSsKMkBEixgPC"
	sec := &secret.Secret{
		Encryption: "AES256",
		Encoding:   "Base64",
		Value:      "test",
	}

	yamlSecret, err := yaml.Marshal(sec)
	if err != nil {
		fmt.Println(err.Error())
	}

	secUn := new(secret.Secret)
	err = yaml.Unmarshal(yamlSecret, secUn)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(secUn)

	yamlSecret, err = yaml.Marshal(secUn)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(yamlSecret))
}
