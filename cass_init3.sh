#!/bin/bash

# Path to the cassandra.yaml file inside the container
CASSANDRA_CONFIG="/etc/cassandra/cassandra.yaml"


# Ensure that the necessary environment variables are set
#if [ -z "$CASSANDRA_VERSION" ] || [ -z "$KEYPATH" ]; then
#  echo "Error: CASSANDRA_VERSION and KEYPATH environment variables must be set."
#  exit 1
#fi


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

# Function to configure Cassandra based on the version
configure_cassandra() {
  local VERSION="4.0.8"
  local keypath="$(pwd)/testdata/pki"
  local conf=(
    "client_encryption_options.enabled: true"
    "client_encryption_options.keystore: $keypath/.keystore"
    "client_encryption_options.keystore_password: cassandra"
    "client_encryption_options.require_client_auth: true"
    "client_encryption_options.truststore: $keypath/.truststore"
    "client_encryption_options.truststore_password: cassandra"
    "concurrent_reads: 2"
    "concurrent_writes: 2"
    "write_request_timeout_in_ms: 5000"
    "read_request_timeout_in_ms: 5000"
  )

  if [[ $VERSION == 3.*.* ]]; then
    conf+=(
      "rpc_server_type: sync"
      "rpc_min_threads: 2"
      "rpc_max_threads: 2"
      "enable_user_defined_functions: true"
      "enable_materialized_views: true"
    )
  elif [[ $VERSION == 4.0.* ]]; then
    conf+=(
      "enable_user_defined_functions: true"
      "enable_materialized_views: true"
    )
  else
    conf+=(
      "user_defined_functions_enabled: true"
      "materialized_views_enabled: true"
    )
  fi

  for setting in "${conf[@]}"; do
    IFS=": " read -r property value <<< "$setting"
    update_property "$property" "$value"
  done
}

# Configure Cassandra
#update_property "num_tokens" "1"
configure_cassandra

echo "Cassandra configuration modified successfully."
