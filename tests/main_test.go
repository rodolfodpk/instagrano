package tests

import (
	"os"
	"testing"
)

var sharedContainers *TestContainers

func TestMain(m *testing.M) {
	// Setup: Create containers once
	sharedContainers = setupSharedTestContainers()

	// Run all tests
	exitCode := m.Run()

	// Teardown: Cleanup containers
	cleanupSharedTestContainers(sharedContainers)

	os.Exit(exitCode)
}
