package gocql

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
	"log"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

// AFTER RESTART PORT IS REMAPPED AND EVEN IF YOU SHOW THE RIGHT PORT IT JUST CANT CREATE A SESSION. ERRORS:
// 2024/06/18 19:15:48 bugged_cases_test.go:96: Attempt 3: Unable to create session: gocql: unable to create session:
// unable to discover protocol version: read tcp 127.0.0.1:41058->127.0.0.1:33285: read: connection reset by peer
//
// 2024/06/18 19:15:50 bugged_cases_test.go:96: Attempt 4: Unable to create session: gocql: unable to create session:
// unable to discover protocol version: dial tcp 127.0.0.1:33285: connect: connection refused
func TestCassandraSessione(t *testing.T) {
	ctx := context.Background()

	cassandraContainer, err := cassandra.RunContainer(ctx,
		testcontainers.WithImage("cassandra:4.0.8"),
		cassandra.WithInitScripts("cass_init_origin.sh"),
	)
	if err != nil {
		log.Fatalf("failed to start container1: %s", err)
	}
	defer func() {
		if err := cassandraContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	restart := exec.Command("docker", "restart", cassandraContainer.GetContainerID())
	restartOutput, err := restart.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to check write_request_timeout_in_ms: %v\n", err)
	}
	fmt.Printf("restartOutput output: %s\n", string(restartOutput))

	session, err := createSessione3(ctx, cassandraContainer, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}

	var result string
	err = session.Query("SELECT cluster_name FROM system.local").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, "MyCluster", result)
}

func createSessione3(ctx context.Context, c *cassandra.CassandraContainer, retries int, delay time.Duration) (*Session, error) {
	for i := 0; i < retries; i++ {

		host, err := c.Host(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get host %s", err)
		}

		newpport, err := c.MappedPort(ctx, port)
		if err != nil {
			return nil, fmt.Errorf("unable to get  mapped port %s", err)
		}

		result, err := IncrementString(newpport.Port())
		if err != nil {
			return nil, fmt.Errorf("unable to increment a string mapped port %s", err)

		}

		//newpport = nat.Port(newpport.Int() + 1))
		//containerHost := host + ":" + newpport.Port()
		containerHost := host + ":" + result

		//containerHost, err := c.ConnectionHost(ctx)
		//if err != nil {
		//	log.Fatalf("Failed to get container host: %s", err)
		//}

		fmt.Println("ConnectionHost", containerHost)

		cluster := NewCluster(containerHost)

		session, err := cluster.CreateSession()
		if err == nil {
			return session, nil
		}
		log.Printf("Attempt %d: Unable to create session: %v", i+1, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("unable to create session after %d retries", retries)
}

func IncrementString(s string) (string, error) {
	// Convert string to integer
	i, err := strconv.Atoi(s)
	if err != nil {
		return "", err
	}

	// Add one to the integer
	i++

	// Convert the integer back to string
	return strconv.Itoa(i), nil
}

// WORKING CASE BUT WITH WORKAROUNDs
func TestRunning(t *testing.T) {
	ctx := context.Background()

	// Start Cassandra container
	cassandraContainer, err := cassandra.RunContainer(ctx,
		testcontainers.WithImage("cassandra:4.0.8"),
		cassandra.WithInitScripts("cass_init_origin.sh"),
	)
	if err != nil {
		log.Fatalf("failed to start container1: %s", err)
	}
	defer func() {
		if err := cassandraContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	//Stop the Cassandra container
	err = cassandraContainer.Stop(ctx, nil)
	require.NoError(t, err)

	// Start the Cassandra container again
	// + there's need to be addded a num num_tokens property to 1(i.e. #update_property "num_tokens" "1")
	// due to a bug with cass container restarting.
	err = cassandraContainer.Start(ctx)
	require.NoError(t, err)

	//	Execute a command to check the configuration value
	cmd := exec.Command("docker", "exec", cassandraContainer.GetContainerID(), "cat", "/etc/cassandra/cassandra.yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to check write_request_timeout_in_ms: %v\n", err)
	}

	fmt.Printf("CombinedOutput output: %s\n", string(output))

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
}
