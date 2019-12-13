package adapter

/*
import (
	"fmt"
	"strings"

	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic/model"
)

func vrrpHealthChecksScript(cmp infra.Compute, transport []model.Source) string {

	selfIP, err := cmp.InternalIP()
	if err != nil {
		panic(err)
	}
	return "/usr/local/bin/health " + strings.Join((vrrpHealthCheckArgs(*selfIP, transport)), " ")
}

func vrrpHealthCheckArgs(selfIP string, transport []model.Source) []string {

	if deriveAny(func(src model.Source) bool {
		return len(src.Destinations) > 0
	}, transport) {
		return []string{stringifyVRRPHealthChecksArg("127.0.0.1", model.Port(29999), model.HealthChecks{
			Protocol: "http",
			Path:     "/ready",
			Code:     200,
		})}
	}

	return deriveFmapSourceVRRPHealthChecks(func(src model.Source) string {
		return stringifyVRRPHealthChecksArg(selfIP, src.SourcePort, *src.HealthChecks)
	}, transport)
}

func stringifyVRRPHealthChecksArg(ip string, port model.Port, hc model.HealthChecks) string {
	return fmt.Sprintf(`%d@%s://%s:%d%s`, hc.Code, ip, hc.Protocol, port, hc.Path)
}
*/
