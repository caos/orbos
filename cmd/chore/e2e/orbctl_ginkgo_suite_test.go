package e2e_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOrbctl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Orbctl Suite")
}
