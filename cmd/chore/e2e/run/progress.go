package main

import (
	"math"
	"strings"
	"time"
)

func printProgress(settings programSettings, step uint8, started time.Time, timeout time.Duration) {
	elapsed := int(math.Round(float64(time.Now().Sub(started)) / float64(timeout) * 100))
	left := 100 - elapsed

	logProgress := settings.logger.Infof
	if elapsed > 85 {
		logProgress = settings.logger.Warnf
	}
	logProgress("%s step %d timeout status %s [%s%s] %s (%d%%)\n",
		settings.orbID,
		step,
		started.Format("15:04:05"),
		strings.Repeat("#", elapsed),
		strings.Repeat("_", left),
		started.Add(timeout).Format("15:04:05"), elapsed)
}
