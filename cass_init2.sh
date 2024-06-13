#!/bin/bash

# Path to the cassandra.yaml file inside the container
CASSANDRA_CONFIG="/etc/cassandra/cassandra.yaml"

# Path to the properties YAML file
#PROPERTIES_YAML="/docker-entrypoint-initdb.d/properties.yaml"

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
#update_property "key_cache_save_period" "14400"
#update_property "write_request_timeout" "399"
update_property "cluster_name" "MyCluster"
pkill -f cassandra
cassandra -R
#update_property "table_options.compression.class" "LZ4Compressor"


# Add more properties as needed
# update_property "another_property" "value"



# service cassandra restart