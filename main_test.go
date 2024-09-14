//go:build cassandra || integration || tc
// +build cassandra integration tc

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
/*
 * Content before git sha 34fdeebefcbf183ed7f916f931aa0586fdaa1b40
 * Copyright (c) 2016, The Gocql authors,
 * provided under the BSD-3-Clause License.
 * See the NOTICE file distributed with this work for additional information.
 */

package gocql

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TChost struct {
	TC           testcontainers.Container
	Addr         string
	ID           string
	CountRestart int
}

var cassNodes = make(map[string]*TChost)
var networkName string

func TestMain(m *testing.M) {
	ctx := context.Background()

	flag.Parse()

	net, err := network.New(ctx)
	if err != nil {
		log.Fatal("cannot create network: ", err)
	}
	networkName = net.Name

	//collect cass nodes into a cluster
	*flagCluster = ""
	for i := 1; i <= *clusterSize; i++ {
		err = NodeUpTC(ctx, i)
		if err != nil {
			log.Fatalf("Failed to start Cassandra node %d: %v", i, err)
		}
	}

	// no need host_id for auth test
	if !*flagRunAuthTest {
		if err := assignCassNodeID(); err != nil {
			log.Fatalf("Failed to assign Cassandra node ID: %v", err)
		}
	}

	// run all tests
	code := m.Run()

	os.Exit(code)
}

func NodeUpTC(ctx context.Context, number int) error {
	cassandraVersion := flagCassVersion.String()[1:]

	jvmOpts := "-Dcassandra.test.fail_writes_ks=test -Dcassandra.custom_query_handler_class=org.apache.cassandra.cql3.CustomPayloadMirroringQueryHandler"
	if *clusterSize == 1 {
		// speeds up the creation of a single-node cluster.
		jvmOpts += " -Dcassandra.initial_token=0 -Dcassandra.skip_wait_for_gossip_to_settle=0"
	}

	env := map[string]string{
		"JVM_OPTS":                  jvmOpts,
		"CASSANDRA_SEEDS":           "node1",
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

	fs := []testcontainers.ContainerFile{
		{
			HostFilePath:      "./testdata/update_container_cass_config.sh",
			ContainerFilePath: "/update_container_cass_config.sh",
			FileMode:          0o777,
		},
	}

	if *flagRunSslTest {
		env["RUN_SSL_TEST"] = "true"
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

	req := testcontainers.ContainerRequest{
		Image:    "cassandra:" + cassandraVersion,
		Env:      env,
		Files:    fs,
		Networks: []string{networkName},
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{{
			PostStarts: []testcontainers.ContainerHook{
				func(ctx context.Context, c testcontainers.Container) error {
					// wait for cassandra config.yaml to initialize
					time.Sleep(100 * time.Millisecond)

					_, body, err := c.Exec(ctx, []string{"bash", "./update_container_cass_config.sh"})
					if err != nil {
						return err
					}

					data, _ := io.ReadAll(body)
					if ok := strings.Contains(string(data), "Cassandra configuration modified successfully."); !ok {
						return fmt.Errorf("./update_container_cass_config.sh didn't complete successfully %v", string(data))
					}

					return nil
				},
			},
		}},
		WaitingFor: wait.ForLog("Startup complete").WithStartupTimeout(2 * time.Minute),
		Name:       "node" + strconv.Itoa(number),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}

	cIP, err := container.ContainerIP(ctx)
	if err != nil {
		return err
	}

	if *flagRunAuthTest {
		// it requires additional time to properly build Cassandra with authentication.
		time.Sleep(10 * time.Second)
	}

	//hostID, err := getCassNodeID(ctx, container, cIP)
	//if err != nil {
	//	return err
	//}
	//fmt.Println("HostID IIIIissssssss", hostID)

	cassNodes[req.Name] = &TChost{
		TC:   container,
		Addr: cIP,
		//ID:   hostID,
	}

	if *clusterSize > number {
		*flagCluster += cIP + ","
	}

	return nil
}

//func getCassNodeID(ctx context.Context, container testcontainers.Container, ip string) (string, error) {
//	session, err := createCluster().CreateSession()
//	if err != nil {
//		return "", err
//	}
//	defer session.Close()
//
//	var hostID string
//	err = session.Query("SELECT host_id FROM system.local").Scan(&hostID)
//	if err != nil {
//		return "", fmt.Errorf("failed to query host node info: %v", err)
//	}
//
//	return hostID, nil
//}

func assignCassNodeID() error {
	cluster := createCluster()
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var hostIDMap = make(map[string]string)
	var localHostID, localIP string
	err = session.Query("SELECT host_id, rpc_address FROM system.local").Scan(&localHostID, &localIP)
	if err != nil {
		return fmt.Errorf("failed to query a host node: %v", err)
	}
	hostIDMap[localIP] = localHostID

	iter := session.Query("SELECT host_id, peer FROM system.peers").Iter()
	var peerHostID, peerIP string
	for iter.Scan(&peerHostID, &peerIP) {
		hostIDMap[peerIP] = peerHostID
	}

	if err := iter.Close(); err != nil {
		return fmt.Errorf("failed to query peer nodes info: %v", err)
	}

	for _, node := range cassNodes {
		id, ok := hostIDMap[node.Addr]
		if !ok {
			return fmt.Errorf("node %s not found in cassandra nodes", node.Addr)
		}

		node.ID = id
	}

	return nil
}

// restoreCluster is a helper function that ensures the cluster remains fully operational during topology changes.
// Commonly used in test scenarios where nodes are added, removed, or modified to maintain cluster stability and prevent downtime.
func restoreCluster(ctx context.Context) error {
	for _, container := range cassNodes {
		if running := container.TC.IsRunning(); running {
			continue
		}
		if err := container.TC.Start(ctx); err != nil {
			return fmt.Errorf("cannot start a container: %v", err)
		}

		container.CountRestart += 1

		err := wait.ForLog("Startup complete").
			WithStartupTimeout(30*time.Second).
			WithOccurrence(container.CountRestart+1).
			WaitUntilReady(ctx, container.TC)
		if err != nil {
			return fmt.Errorf("cannot wait until a start container: %v", err)
		}

		time.Sleep(30 * time.Second)
	}

	return nil
}

// getPool is a test helper designed to enhance readability by mocking the `func (p *policyConnPool) getPool(host *HostInfo) (pool *hostConnPool, ok bool)` method.
func getPool(p *policyConnPool, hostID string) (pool *hostConnPool, ok bool) {
	p.mu.RLock()
	pool, ok = p.hostConnPools[hostID]
	p.mu.RUnlock()
	return
}
