package gocql

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
	"log"
	"testing"
)

//func TestMain(m *testing.M) {
//	ctx := context.Background()
//
//	cassandraContainer, err := cassandra.RunContainer(ctx,
//		testcontainers.WithImage("cassandra:4.0.8"),
//	)
//	if err != nil {
//		log.Fatalf("failed to start container: %s", err)
//	}
//	// Clean up the container
//	defer func() {
//		if err := cassandraContainer.Terminate(ctx); err != nil {
//			log.Fatalf("failed to terminate container: %s", err)
//		}
//	}()
//
//	*flagCluster, err = cassandraContainer.ConnectionHost(ctx)
//	if err != nil {
//		log.Fatalf("Failed to get container host: %s", err)
//	}
//
//	// Run the tests
//	code := m.Run()
//
//	// Exit with the test's exit code
//	os.Exit(code)
//}

func TestCassandraWithConfigFile(t *testing.T) {
	ctx := context.Background()

	// Start Cassandra container
	cassandraContainer, err := cassandra.RunContainer(ctx,
		testcontainers.WithImage("cassandra:4.0.8"),
		//cassandra.WithConfigFile("config_408.yaml"),
		//cassandra.WithInitScripts("init-cassandra.sh"),
	)
	if err != nil {
		log.Fatalf("failed to start container1: %s", err)
	}
	defer func() {
		if err := cassandraContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	containerHost, err := cassandraContainer.ConnectionHost(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %s", err)
	}

	cluster := NewCluster(containerHost)

	session, err := cluster.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	// Execute a command to check the configuration value
	//cmd := exec.Command("docker", "exec", cassandraContainer.GetContainerID(), "cat", "/etc/cassandra/cassandra.yaml")
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//	log.Fatalf("Failed to check write_request_timeout_in_ms: %v\n", err)
	//}
	//
	//fmt.Printf("CombinedOutput output: %s\n", string(output))

	// One of the way to restart a cass-container
	//restart := exec.Command("docker", "restart", cassandraContainer.GetContainerID())
	//restartOutput, err := restart.CombinedOutput()
	//if err != nil {
	//	log.Fatalf("Failed to check write_request_timeout_in_ms: %v\n", err)
	//}
	//fmt.Printf("restartOutput output: %s\n", string(restartOutput))
	//

	var result string
	err = session.Query("SELECT cluster_name FROM system.local").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, "MyCluster", result)
}
