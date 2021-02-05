package kubernetes

import (
	"reflect"
	"testing"

	"github.com/caos/orbos/mntr"
	v1 "k8s.io/api/core/v1"
)

func Test_reconcileTaints(t *testing.T) {
	type args struct {
		node v1.Node
		pool Pool
	}
	someTaintKey := "someKey"
	someTaintEffect := v1.TaintEffectNoSchedule
	someDesiredTaint := Taint{
		Key:    someTaintKey,
		Effect: someTaintEffect,
	}
	someNodeTaint := v1.Taint{
		Key:    someTaintKey,
		Effect: someTaintEffect,
	}

	node := func(taint ...v1.Taint) v1.Node {
		return v1.Node{Spec: v1.NodeSpec{Taints: append([]v1.Taint{}, taint...)}}
	}
	pool := func(taint ...Taint) Pool {
		taints := Taints(append([]Taint{}, taint...))
		return Pool{Taints: &taints}
	}

	nodePtr := func(node v1.Node) *v1.Node {
		return &node
	}

	tests := []struct {
		name string
		args args
		want *v1.Node
	}{
		{
			name: "It should add configured taints",
			args: args{
				node: v1.Node{},
				pool: pool(someDesiredTaint),
			}, want: nodePtr(node(someNodeTaint)),
		},
		{
			name: "It should leave the taints as they are if the taints property is nil",
			args: args{
				node: node(someNodeTaint),
				pool: Pool{},
			}, want: nil,
		},
		{
			name: "It should remove existing taints if the empty slice is passed",
			args: args{
				node: node(someNodeTaint),
				pool: pool(),
			}, want: nodePtr(node()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reconcileTaints(tt.args.node, tt.args.pool, mntr.Monitor{})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reconcileTaints() got = %v, want %v", got, tt.want)
			}
		})
	}
}
