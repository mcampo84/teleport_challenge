#!/bin/bash

# Get the directory of the current script
SCRIPT_DIR=$(dirname "$0")

# Create directories
mkdir -p "$SCRIPT_DIR/../conf/server" "$SCRIPT_DIR/../conf/client"

# Generate CA key and certificate
openssl genpkey -algorithm RSA -out "$SCRIPT_DIR/../conf/ca.key"
openssl req -x509 -new -nodes -key "$SCRIPT_DIR/../conf/ca.key" -sha256 -days 365 -out "$SCRIPT_DIR/../conf/ca.crt" -subj "/C=US/ST=State/L=City/O=Organization/OU=OrgUnit/CN=example.com"

# Generate server key and CSR
openssl genpkey -algorithm RSA -out "$SCRIPT_DIR/../conf/server/server.key"
openssl req -new -key "$SCRIPT_DIR/../conf/server/server.key" -out "$SCRIPT_DIR/../conf/server/server.csr" -subj "/C=US/ST=State/L=City/O=Organization/OU=OrgUnit/CN=server.example.com"

# Sign server CSR with CA certificate
openssl x509 -req -in "$SCRIPT_DIR/../conf/server/server.csr" -CA "$SCRIPT_DIR/../conf/ca.crt" -CAkey "$SCRIPT_DIR/../conf/ca.key" -CAcreateserial -out "$SCRIPT_DIR/../conf/server/server.crt" -days 365 -sha256

# Generate client key and CSR
openssl genpkey -algorithm RSA -out "$SCRIPT_DIR/../conf/client/client.key"
openssl req -new -key "$SCRIPT_DIR/../conf/client/client.key" -out "$SCRIPT_DIR/../conf/client/client.csr" -subj "/C=US/ST=State/L=City/O=Organization/OU=OrgUnit/CN=client.example.com"

# Sign client CSR with CA certificate
openssl x509 -req -in "$SCRIPT_DIR/../conf/client/client.csr" -CA "$SCRIPT_DIR/../conf/ca.crt" -CAkey "$SCRIPT_DIR/../conf/ca.key" -CAcreateserial -out "$SCRIPT_DIR/../conf/client/client.crt" -days 365 -sha256
