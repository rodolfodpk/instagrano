package tests

import (
	"testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInstagrano(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instagrano Test Suite")
}
