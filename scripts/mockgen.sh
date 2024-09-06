#!/usr/bin/env bash

mockgen -source=x/poa/testutil/expected_keepers.go -package testutil -destination=x/poa/testutil/expected_keepers_mock.go