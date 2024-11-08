#!/usr/bin/env bash

mockgen -source=x/poa/types/expected_keepers.go -package testutil -destination=x/poa/testutil/expected_keepers_mock.go
mockgen -source=x/poa/testutil/tx.go -package testutil -destination=x/poa/testutil/tx_mock.go
mockgen -source=x/poa/testutil/keys.go -package testutil -destination=x/poa/testutil/keys_mock.go
mockgen -source=x/poa/testutil/expected_msg_server.go -package testutil -destination=x/poa/testutil/expected_msg_server_mock.go