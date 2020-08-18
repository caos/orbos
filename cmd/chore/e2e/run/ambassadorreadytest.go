package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/yaml.v3"
)

func ambassadorReadyTest(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

	cmd, err := orbctl()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	cmd.Args = append(cmd.Args, "file", "print", "caos-internal/orbiter/current.yml")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return err
	}

	currentBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return err
	}

	current := struct {
		Providers struct {
			ProviderUnderTest struct {
				Current struct {
					Ingresses struct {
						Httpingress struct {
							Location     string
							Frontendport uint16
						}
					}
				}
			} `yaml:"provider-under-test"`
		}
	}{}

	if err := yaml.Unmarshal(currentBytes, &current); err != nil {
		return err
	}

	ep := current.Providers.ProviderUnderTest.Current.Ingresses.Httpingress
	resp, err := http.Get(fmt.Sprintf("https://%s:%d/ambassador/v0/check_ready", ep.Location, ep.Frontendport))
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New(resp.Status)
	}
	return nil
}
