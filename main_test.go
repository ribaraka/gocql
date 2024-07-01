package gocql

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"os/exec"
	"testing"
	"time"
)

//const port = nat.Port("9042/tcp")
//
////	func TestMain(m *testing.M) {
////		ctx := context.Background()
////
////		cassandraContainer, err := cassandra.RunContainer(ctx,
////			testcontainers.WithImage("cassandra:4.0.8"),
////		)
////		if err != nil {
////			log.Fatalf("failed to start container: %s", err)
////		}
////		// Clean up the container
////		defer func() {
////			if err := cassandraContainer.Terminate(ctx); err != nil {
////				log.Fatalf("failed to terminate container: %s", err)
////			}
////		}()
////
////		*flagCluster, err = cassandraContainer.ConnectionHost(ctx)
////		if err != nil {
////			log.Fatalf("Failed to get container host: %s", err)
////		}
////
////		// Run the tests
////		code := m.Run()
////
////		// Exit with the test's exit code
////		os.Exit(code)
////	}

const port = nat.Port("9042/tcp")

func TestCassandraWithWaitStrategy2(t *testing.T) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "cassandra:4.0.8",
			ExposedPorts: []string{string(port)},
			Env: map[string]string{
				"CASSANDRA_SNITCH":          "GossipingPropertyFileSnitch",
				"JVM_OPTS":                  "-Dcassandra.skip_wait_for_gossip_to_settle=0 -Dcassandra.initial_token=0",
				"HEAP_NEWSIZE":              "128M",
				"MAX_HEAP_SIZE":             "1024M",
				"CASSANDRA_ENDPOINT_SNITCH": "GossipingPropertyFileSnitch",
				"CASSANDRA_DC":              "datacenter1",
			},
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      "working.sh",
					ContainerFilePath: "/working.sh",
					FileMode:          0o700,
				},
				{
					HostFilePath:      "./testdata/pki/.keystore",
					ContainerFilePath: "testdata/.keystore",
					FileMode:          0o700,
				},
				{
					HostFilePath:      "./testdata/pki/.truststore",
					ContainerFilePath: "testdata/.truststore",
					FileMode:          0o700,
				},
			},
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("failed to LAUNCH container: %s", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	//time.Sleep(1500 * time.Millisecond)
	code, _, err := container.Exec(ctx, []string{"bash", "./working.sh"})
	require.NoError(t, err)
	require.Zero(t, code)

	cmd := exec.Command("docker", "exec", container.GetContainerID(), "cat", "/etc/cassandra/cassandra.yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to check write_request_timeout_in_ms: %v\n", err)
	}

	fmt.Printf("CombinedOutput output: %s\n", string(output))

	//cmd := exec.Command("docker", "exec", container.GetContainerID(), "ls", "-a", "/testdata")
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//	log.Fatalf("Failed to check write_request_timeout_in_ms: %v\n", err)
	//}
	//
	//fmt.Printf("CombinedOutput output: %s\n", string(output))
	//
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	mappedPort, err := container.MappedPort(ctx, port)
	if err != nil {
		t.Fatal(err)
	}

	containerHost := host + ":" + mappedPort.Port()

	cluster := NewCluster(containerHost)

	session, err := createContaienrSessione(cluster, 10, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	var result string
	err = session.Query("SELECT cluster_name FROM system.local").Scan(&result)
	require.NoError(t, err)
	//fmt.Println("CLUSTER NAME IS ", result)
	assert.Equal(t, "MyCluster", result)

}

func createContaienrSessione(cluster *ClusterConfig, retries int, delay time.Duration) (*Session, error) {
	for i := 0; i < retries; i++ {
		session, err := cluster.CreateSession()
		if err == nil {
			return session, nil
		}
		log.Printf("Attempt %d: Unable to create session: %v", i+1, err)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("unable to create session after %d retries", retries)
}
