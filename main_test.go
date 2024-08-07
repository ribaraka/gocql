//go:build cassandra || integration
// +build cassandra integration

package gocql

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	flag.Parse()

	cassandraVersion := flagCassVersion.String()[1:]

	jvmOpts := "-Dcassandra.test.fail_writes_ks=test -Dcassandra.custom_query_handler_class=org.apache.cassandra.cql3.CustomPayloadMirroringQueryHandler"
	if *clusterSize == 1 {
		// speeds up the creation of a single-node cluster.
		jvmOpts += " -Dcassandra.initial_token=0 -Dcassandra.skip_wait_for_gossip_to_settle=0"
	}

	env := map[string]string{
		"JVM_OPTS":                  jvmOpts,
		"CASSANDRA_SEEDS":           "cassandra1",
		"CASSANDRA_DC":              "datacenter1",
		"HEAP_NEWSIZE":              "100M",
		"MAX_HEAP_SIZE":             "256M",
		"CASSANDRA_RACK":            "rack1",
		"CASSANDRA_ENDPOINT_SNITCH": "GossipingPropertyFileSnitch",
		"CASS_VERSION":              cassandraVersion,
	}

	if *flagRunAuthTest {
		env["AUTH_TEST"] = "true"
	}

	fs := []testcontainers.ContainerFile{}
	if *flagRunSslTest {
		env["SSL_TEST"] = "true"

		fs = append(fs, []testcontainers.ContainerFile{
			{
				HostFilePath:      "./testdata/pki/.keystore",
				ContainerFilePath: "testdata/.keystore",
				FileMode:          0o777,
			},
			{
				HostFilePath:      "./testdata/pki/.truststore",
				ContainerFilePath: "testdata/.truststore",
				FileMode:          0o777,
			},
		}...)
	}

	fs = append(fs, testcontainers.ContainerFile{
		HostFilePath:      "update_container_cass_config.sh",
		ContainerFilePath: "/update_container_cass_config.sh",
		FileMode:          0o777,
	})

	networkRequest := testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: "cassandra",
		},
	}
	cassandraNetwork, err := testcontainers.GenericNetwork(ctx, networkRequest)
	if err != nil {
		log.Fatalf("Failed to create network: %s", err)
	}
	defer cassandraNetwork.Remove(ctx)

	// Function to create a Cassandra container (node)
	createCassandraContainer := func(number int) (string, error) {
		req := testcontainers.ContainerRequest{
			Image:        "cassandra:" + cassandraVersion,
			ExposedPorts: []string{"9042/tcp"},
			Env:          env,
			Files:        fs,
			Networks:     []string{"cassandra"},
			LifecycleHooks: []testcontainers.ContainerLifecycleHooks{{
				PostStarts: []testcontainers.ContainerHook{
					func(ctx context.Context, c testcontainers.Container) error {
						// wait for cassandra config.yaml to initialize
						time.Sleep(100 * time.Millisecond)

						code, _, err := c.Exec(ctx, []string{"bash", "./update_container_cass_config.sh"})
						if err != nil {
							return err
						}
						if code != 0 {
							return fmt.Errorf("script ./update_container_cass_config.sh exited with code %d", code)
						}
						return nil
					},
				},
			}},
			WaitingFor: wait.ForLog("Startup complete").WithStartupTimeout(2 * time.Minute),
			Name:       "cassandra" + strconv.Itoa(number),
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			return "", err
		}

		ip, err := container.ContainerIP(ctx)
		if err != nil {
			return "", err
		}

		return ip, nil
	}

	// collect cass nodes into a cluster
	*flagCluster = ""
	for i := 0; i < *clusterSize; i++ {
		ip, err := createCassandraContainer(i + 1)
		if err != nil {
			log.Fatalf("Failed to start Cassandra node %d: %v", i+1, err)
		}

		// if not the last iteration
		if i != *clusterSize-1 {
			ip += ","
		}

		*flagCluster += ip
	}

	if *flagRunAuthTest {
		// it requires additional time to properly build Cassandra with authentication.
		time.Sleep(10 * time.Second)
	}

	// run all tests
	code := m.Run()

	os.Exit(code)
}
