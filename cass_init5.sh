#!/bin/bash

# Path to the cassandra.yaml file inside the container
CASSANDRA_CONFIG="/etc/cassandra/cassandra.yaml"

# Function to update a property in the cassandra.yaml file
update_property() {
  local property=$1
  local value=$2
  local root_property=${property%%.*}
  local nested_property=${property#*.}

  if grep -q "^${property}:" "$CASSANDRA_CONFIG"; then
    # If the property exists, update its value
    sed -i "s|^\(${property}:\).*|\1 ${value}|" "$CASSANDRA_CONFIG"
    echo "Updated $property to $value"
  else
    if [[ "$property" == *"."* ]]; then
      # If it's a nested property
      if grep -q "^${root_property}:" "$CASSANDRA_CONFIG"; then
        if grep -q "^    ${nested_property}:" "$CASSANDRA_CONFIG"; then
          sed -i "/^${root_property}:/,/^[^ ]/ s|^\(    ${nested_property}:\).*|\1 ${value}|" "$CASSANDRA_CONFIG"
        else
          # Add nested property under existing root property
          awk -v root="$root_property" -v prop="$nested_property" -v val="$value" '
          $0 ~ "^"root":" {
            print $0
            print "    "prop": "val
            next
          }
          { print $0 }
          ' "$CASSANDRA_CONFIG" > tmpfile && mv tmpfile "$CASSANDRA_CONFIG"
        fi
        echo "Added $property with value $value"
      else
        # Add new root property with nested property
        echo -e "${root_property}:\n    ${nested_property}: ${value}" >> "$CASSANDRA_CONFIG"
        echo "Added $property with value $value"
      fi
    else
      # If it's a root-level property, add it directly
      echo "${property}: ${value}" >> "$CASSANDRA_CONFIG"
      echo "Added $property with value $value"
    fi
  fi
}

# Function to configure Cassandra based on the version
configure_cassandra() {
  local VERSION="4.0.8"
  local keypath="testdata"
  local conf=(
    "client_encryption_options.enabled:true"
    "client_encryption_options.keystore:$keypath/.keystore"
    "client_encryption_options.keystore_password:cassandra"
    "client_encryption_options.require_client_auth:true"
    "client_encryption_options.ololo:tetest"
    "client_encryption_options.truststore:$keypath/.truststore"
    "client_encryption_options._truststore:$keypath/.truststore"
    "client_encryption_options.truststore1:$keypath/.truststore"
    "client_encryption_options.truststore_password_:cassandra"
    "client_encryption_options.protocol: tcp"
    "concurrent_reads:2"
    "concurrent_writes:2"
    "write_request_timeout_in_ms:5000"
    "read_request_timeout_in_ms:5000"
  )

  if [[ $VERSION == 3.*.* ]]; then
    conf+=(
      "rpc_server_type:sync"
      "rpc_min_threads:2"
      "rpc_max_threads:2"
      "enable_user_defined_functions:true"
      "enable_materialized_views:true"
    )
  elif [[ $VERSION == 4.0.* ]]; then
    conf+=(
      "enable_user_defined_functions:true"
      "enable_materialized_views:true"
    )
  else
    conf+=(
      "user_defined_functions_enabled:true"
      "materialized_views_enabled:true"
    )
  fi

  for setting in "${conf[@]}"; do
    IFS=":" read -r property value <<< "$setting"
    update_property "$property" "$value"
  done
}

# Configure Cassandra
configure_cassandra

echo "Cassandra configuration modified successfully."
