#!/bin/bash

# Path to the cassandra.yaml file inside the container
CASSANDRA_CONFIG="/etc/cassandra/cassandra.yaml"

# Function to update a property in the cassandra.yaml file
update_property() {
  local property=$1
  local value=$2
  if grep -q "^${property}:" "$CASSANDRA_CONFIG"; then
    # If the property exists, update its value
    sed -i "s/^${property}: .*/${property}: ${value}/" "$CASSANDRA_CONFIG"
    echo "Updated $property to $value"
  else
    # If the property does not exist, add it
    echo "${property}: ${value}" >> "$CASSANDRA_CONFIG"
    echo "Added $property with value $value"
  fi
}

# Update desired properties
#update_property "write_request_timeout_in_ms" "700"
#update_property "enable_user_defined_functions" "true"
#update_property "num_tokens" "1"
update_property "cluster_name" "MyCluster"


#update_property "concurrent_reads" "2"
#update_property "concurrent_writes": "2"
#update_property "write_request_timeout_in_ms": "5000"
#update_property "read_request_timeout_in_ms": "5000"
#update_property "user_defined_functions_enabled: true"

# Add more properties as needed
# update_property "another_property" "value"

echo "Cassandra configuration modified successfully."
