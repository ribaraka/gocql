package gocql

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	cassandraContainer, err := cassandra.RunContainer(ctx,
		testcontainers.WithImage("cassandra:4.0.8"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	// Clean up the container
	defer func() {
		if err := cassandraContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	*flagCluster, err = cassandraContainer.ConnectionHost(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %s", err)
	}

	// Run the tests
	code := m.Run()

	// Exit with the test's exit code
	os.Exit(code)
}
