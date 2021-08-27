package orbctl_test

import (
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"
)

type flakyExitMatcher struct {
	types.GomegaMatcher
}

// FlakyExit is a custom matcher that wraps the gomega matcher returned by gexec.Exit
// and overwrites its MatchMayChangeInTheFuture to always return true. This ensures
// that an assertion with Eventually considers that a terminated command could change
// its return code if it's executed again.
func FlakyExit(code ...int) types.GomegaMatcher {
	return &flakyExitMatcher{
		gexec.Exit(code...),
	}
}

func (f *flakyExitMatcher) MatchMayChangeInTheFuture(_ interface{}) bool {
	return true
}
