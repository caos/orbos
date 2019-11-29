package test

import (
	"testing"
)

func TestKubernetes(t *testing.T) {
	//	phase(t, dayOne, immediately) // clean environment
	//	phase(t, dayOne, never) // bootstrapping
	phase(t, dayTwo, never) // upscaling, updating
	// TODO: implement testPhase(t, dayThree, cleanupOnFailure)     // downscaling, upgrading, downdating TODO: implement
	// TODO: implement testPhase(t, dayFour, alwaysCleanup)      // move things around, mutate ids etc
}

type cleanupMode int

const (
	never cleanupMode = iota
	onFailure
	always
	immediately
)

func phase(t *testing.T, tester func(t *testing.T) (func(), func() error), cleanupMode cleanupMode) {
	if t.Failed() {
		return
	}
	phaseTest, phaseCleanup := tester(t)
	if cleanupMode == immediately {
		cleanup(t, phaseCleanup)
		return
	}

	phaseTest()

	if t.Failed() {
		if cleanupMode == onFailure || cleanupMode == always {
			cleanup(t, phaseCleanup)
		}
		t.Fatal("phase failed")
	}
	t.Log("Phase succeeded")
	if cleanupMode == always {
		cleanup(t, phaseCleanup)
	}
}

func cleanup(t *testing.T, cleanup func() error) {
	t.Log("Cleaning up")
	if err := cleanup(); err != nil {
		t.Fatal("cleaning up failed:", err)
	}
}
