package labels

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var _ Labels = (*Operator)(nil)

type Operator struct {
	model InternalOperator
}

func NoopOperator(product string) *Operator {
	return &Operator{model: InternalOperator{
		Version:               "unknown",
		InternalPartofProp:    InternalPartofProp{PartOf: product},
		InternalManagedByProp: InternalManagedByProp{ManagedBy: "CLI"},
		Major:                 "unknown",
	}}
}

func ForOperator(product, operator, version string) (*Operator, error) {

	if operator == "" || version == "" {
		return nil, errors.New("operator or version must not be nil")
	}

	return &Operator{model: InternalOperator{
		Version:               version,
		InternalPartofProp:    InternalPartofProp{PartOf: product},
		InternalManagedByProp: InternalManagedByProp{ManagedBy: operator},
		Major:                 major(version),
	}}, nil
}

func (l *Operator) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&l.model)
}

func MustForOperator(product, operator, version string) *Operator {
	o, err := ForOperator(product, operator, version)
	if err != nil {
		panic(err)
	}
	return o
}

func (l *Operator) Equal(r comparable) bool {
	if right, ok := r.(*Operator); ok {
		return l.model == right.model
	}
	return false
}

func (l *Operator) MarshalYAML() (interface{}, error) {
	return nil, errors.New("type *labels.Operator is not serializable")
}

func (l *Operator) Major() int8 {

	var unknown int8 = -1

	if l.model.Major == "unknown" {
		return unknown
	}

	m, err := strconv.Atoi(strings.TrimPrefix(l.model.Major, "v"))
	if err != nil {
		return unknown
	}

	if m > math.MaxInt8 {
		return unknown
	}

	return int8(m)
}

type InternalPartofProp struct {
	PartOf string `yaml:"app.kubernetes.io/part-of,omitempty"`
}

type InternalManagedByProp struct {
	ManagedBy string `yaml:"app.kubernetes.io/managed-by,omitempty"`
}

type InternalOperator struct {
	InternalManagedByProp `yaml:",inline"`
	Version               string `yaml:"app.kubernetes.io/version"`
	Major                 string `yaml:"caos.ch/major"`
	InternalPartofProp    `yaml:",inline"`
}

func major(version string) string {
	versionRegex := regexp.MustCompile("^v([0-9]+)\\.[0-9]+\\.[0-9]+$")
	matches := versionRegex.FindStringSubmatch(version)
	unknown := "unknown"
	if len(matches) != 2 {
		return unknown
	}

	m, err := strconv.Atoi(matches[1])
	if err != nil {
		return unknown
	}

	if m > math.MaxInt8 {
		return unknown
	}

	return fmt.Sprintf("v%d", m)
}
