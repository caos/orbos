// +build test unit

package adapter_test

/*
import (
	"os"
	"testing"

	"github.com/caos/infrop/internal/node-agent/executor"
	"github.com/caos/infrop/internal/edge/logger/context"
	"github.com/caos/infrop/internal/edge/logger/stdlib"
	"github.com/caos/infrop/internal/kinds/nodeagent/edge/rebooter/mock/noop"
	"github.com/caos/infrop/internal/pkg/repository/mock/dummy"
	"github.com/caos/infrop/internal/pkg/resolver/mock"
)

type ensureCurrentTestcase struct {
	desc    string
	desired map[string]interface{}
	current map[string]interface{}
	expect  func(map[string]interface{}) bool
}

func TestEnsureCurrent(t *testing.T) {
	testcases := []*ensureCurrentTestcase{
		&ensureCurrentTestcase{
			desc:    "Current version should align",
			current: map[string]interface{}{"rebootnotneeded": "0"},
			desired: map[string]interface{}{"rebootnotneeded": "2"},
			expect:  func(state map[string]interface{}) bool { return state["rebootnotneeded"] == "1" },
		},
		&ensureCurrentTestcase{
			desc:    "If dependencies need to be ensured, node must be marked unready",
			current: map[string]interface{}{"rebootnotneeded": "1", "ready": "true"},
			desired: map[string]interface{}{"rebootnotneeded": "2"},
			expect:  func(state map[string]interface{}) bool { return state["ready"] != true },
		},
		&ensureCurrentTestcase{
			desc:    "If no dependencies need to be ensured, node must be marked ready",
			current: map[string]interface{}{"rebootnotneeded": "1", "ready": "false"},
			desired: map[string]interface{}{"rebootnotneeded": "1"},
			expect:  func(state map[string]interface{}) bool { return state["ready"] == true },
		},
		&ensureCurrentTestcase{
			desc:    "Uninstalled software should be listed with an empty string",
			current: map[string]interface{}{"uninstalled": "1"},
			desired: map[string]interface{}{"uninstalled": "2"},
			expect:  func(state map[string]interface{}) bool { return state["uninstalled"] == "" },
		},
	}

	for testcasenumber, testcase := range testcases {
		executor := loadExecutor(testcase.current, testcase.desired)
		if err := executor.EnsureCurrent(); err != nil {
			t.Error(err)
		}
		if success := testcase.expect(testcase.current); !success {
			t.Errorf("%d) %s", testcasenumber+1, testcase.desc)
		}
	}
}

func loadExecutor(current map[string]interface{}, desired map[string]interface{}) *executor.Executor {
	return executor.New(context.Add(stdlib.New(os.Stdout)).Verbose(), dummy.New(current), dummy.New(desired), dummy.New(nil), mock.New(), noop.New())
}
*/
