#!/bin/bash
CASSANDRA_CONFIG=/etc/cassandra/cassandra.yaml

# Modify specific properties in the cassandra.yaml file
sed -i 's/cluster_name:.*$/cluster_name: "MyCluster"/' $CASSANDRA_CONFIG

# Restart Cassandra service to apply changes
pkill -f cassandra
cassandra -R


echo "Cassandra configuration modified successfully."
