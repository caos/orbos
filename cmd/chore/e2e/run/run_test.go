package main

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/afiskon/promtail-client/promtail"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/pkg/orb"
)

func Test(t *testing.T) {

	orbconfig := prefixedEnv("ORBCONFIG")

	run(t, programSettings{
		clusterkey:  "k8s",
		providerkey: "providerundertest",
		tag:         prefixedEnv("TAG"),
		download:    boolEnv(t, prefixedEnv("DOWNLOAD")),
		cleanup:     boolEnv(t, prefixedEnv("CLEANUP")),
		from:        parseUint8(t, prefixedEnv("FROM")),
		orbconfig:   orbconfig,
		orbID:       orbID(t, orbconfig),
		logger:      &goTestLogger{t: t, level: promtail.DEBUG},
	})
}

func prefixedEnv(env string) string {
	return os.Getenv("ORBOS_E2E_" + env)
}

func parseUint8(t *testing.T, val string) uint8 {
	parsed, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		t.Fatal(err)
	}
	return uint8(parsed)
}

func boolEnv(t *testing.T, val string) bool {
	if val == "" {
		return false
	}
	value, err := strconv.ParseBool(val)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func orbID(t *testing.T, orbconfig string) string {
	orbCfg, err := orb.ParseOrbConfig(helpers.PruneHome(orbconfig))
	if err != nil {
		t.Fatal(err)
	}

	orb.IsComplete(orbCfg)
	if err != nil {
		t.Fatal(err)
	}

	return strings.ToLower(strings.Split(strings.Split(orbCfg.URL, "/")[1], ".")[0])
}

var _ promtail.Client = (*goTestLogger)(nil)

type goTestLogger struct {
	t     *testing.T
	level promtail.LogLevel
}

func (g *goTestLogger) Debugf(format string, args ...interface{}) {
	if g.level <= promtail.DEBUG {
		g.t.Logf(format, args...)
	}
}

func (g *goTestLogger) Infof(format string, args ...interface{}) {
	if g.level <= promtail.INFO {
		g.t.Logf(format, args...)
	}
}

func (g *goTestLogger) Warnf(format string, args ...interface{}) {
	if g.level <= promtail.WARN {
		g.t.Logf(format, args...)
	}
}

func (g *goTestLogger) Errorf(format string, args ...interface{}) {
	if g.level <= promtail.ERROR {
		g.t.Errorf(format, args...)
	}
}

func (g *goTestLogger) Shutdown() {}
