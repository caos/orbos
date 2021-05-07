package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/afiskon/promtail-client/promtail"

	"github.com/caos/orbos/internal/helpers"

	"gopkg.in/yaml.v3"
)

func ambassadorReadyTestFunc(logger promtail.Client) testFunc {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {
		cmd, err := orbctl()
		if err != nil {
			return err
		}

		buf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.Args = append(cmd.Args, "--gitops", "file", "print", "caos-internal/orbiter/current.yml")
		cmd.Stdout = buf
		cmd.Stderr = errBuf
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("reading orbiters current state failed: %w: %s", err, errBuf.String())
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
							Httpsingress struct {
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

		ep := current.Providers.ProviderUnderTest.Current.Ingresses.Httpsingress

		msg, err := helpers.Check("https", ep.Location, ep.Frontendport, "/ambassador/v0/check_ready", 200, false)
		if err != nil {
			return fmt.Errorf("ambassador ready check failed with message: %s: %w", msg, err)
		}
		logger.Infof(msg)
		return nil
	}
}
