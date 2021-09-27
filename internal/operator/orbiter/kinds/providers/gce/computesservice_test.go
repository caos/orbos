package gce_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	secret2 "github.com/caos/orbos/v5/pkg/secret"

	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/providers/gce"
	"github.com/caos/orbos/v5/internal/ssh"
	"github.com/caos/orbos/v5/mntr"
)

func TestComputeService(t *testing.T) {

	pool := &gce.Pool{
		OSImage:     "projects/centos-cloud/global/images/centos-7-v20200429",
		MinCPUCores: 2,
		MinMemoryGB: 4,
		StorageGB:   20,
	}

	private, public := ssh.Generate()

	jsonKey := os.Getenv("ORBOS_GCE_JSON_KEY")
	t.Log(jsonKey)
	if jsonKey == "" {
		t.Fatal("Environment variable ORBOS_GCE_JSON_KEY is empty")
	}

	svc := gce.NewMachinesService(
		mntr.Monitor{OnInfo: mntr.LogMessage},
		&gce.Spec{
			Verbose: false,
			JSONKey: &secret2.Secret{Value: jsonKey},

			Region: "europe-west1",
			Zone:   "europe-west1-b",
			SSHKey: &gce.SSHKey{
				Private: &secret2.Secret{Value: private},
				Public:  &secret2.Secret{Value: public},
			},
			Pools: map[string]*gce.Pool{
				"apool":       pool,
				"anotherpool": pool,
				"aThirdPool":  pool,
			},
		},
		"gce",
		"orbiter-elio",
	)

	machine, err := svc.Create("apool")
	if err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	if err := machine.ReadFile("/home/orbiter/.ssh/authorized_keys", buf); err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())

	if err := machine.WriteFile("/var/lib/orbiter/hier", bytes.NewReader([]byte("da")), 600); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	if err := machine.ReadFile("/var/lib/orbiter/hier", buf); err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())

	stdout, err := machine.Execute(nil, "sudo whoami")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(stdout))

	if _, err := svc.Create("apool"); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.ListPools(); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Create("anotherpool"); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.ListPools(); err != nil {
		t.Fatal(err)
	}

	aPool, err := svc.List("apool")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("apool:", sprint(aPool))

	anotherPool, err := svc.List("anotherpool")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("anotherpool:", sprint(anotherPool))
	if err != nil {
		t.Fatal(err)
	}

	if err := aPool[0].Remove(); err != nil {
		t.Fatal(err)
	}

	aPool, err = svc.List("apool")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("apool:", sprint(aPool))

	if err := aPool[0].Remove(); err != nil {
		t.Fatal(err)
	}

	aPool, err = svc.List("apool")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("apool:", sprint(aPool))

	if err := anotherPool[0].Remove(); err != nil {
		t.Fatal(err)
	}

	aPool, err = svc.List("anotherpool")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("anotherpool:", sprint(aPool))
}

func sprint(machines infra.Machines) string {
	info := make([]string, len(machines))
	for idx, machine := range machines {
		info[idx] = fmt.Sprintf("%s (%s)", machine.ID(), machine.IP())
	}
	return strings.Join(info, ", ")
}
