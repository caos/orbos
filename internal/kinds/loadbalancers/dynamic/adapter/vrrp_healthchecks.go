package adapter

import (
	"fmt"
	"strings"

	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic/model"
)

func vrrpHealthChecksScript(transport []model.Source) string {
	return "/usr/local/bin/health " + strings.Join((vrrpHealthCheckArgs(transport)), " ")
}

func vrrpHealthCheckArgs(transport []model.Source) []string {

	if deriveAny(func(src model.Source) bool {
		return len(src.Destinations) > 0
	}, transport) {
		return []string{stringifyVRRPHealthChecksArg(model.Port(30000), model.HealthChecks{
			Protocol: "http",
			Path:     "/healthz",
			Code:     200,
		})}
	}

	return deriveFmapSourceVRRPHealthChecks(func(src model.Source) string {
		return stringifyVRRPHealthChecksArg(src.SourcePort, *src.HealthChecks)
	}, transport)
}

func stringifyVRRPHealthChecksArg(port model.Port, hc model.HealthChecks) string {
	return fmt.Sprintf(`%d@%s://127.0.0.1:%d%s`, hc.Code, hc.Protocol, port, hc.Path)
}
