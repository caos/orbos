package main

import (
	"context"
	"fmt"
	"github.com/caos/orbos/internal/start"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"

	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/mntr"
)

func main() {
	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	orb, err := orb.ParseOrbConfig("/Users/benz/.orb/config")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = start.Boom(context.Background(), monitor, orb, true)
	fmt.Println(err.Error())
}
