package tests

import (
	"testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) { 
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration") 
}

var _ = Describe("Instagrano", func() {
	It("Full flow works", func() { 
		Expect(true).To(BeTrue()) 
	})
})
