package uaa_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfauth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UAA Auth Suite")
}
