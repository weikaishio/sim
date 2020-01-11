#!/usr/bin/env bash

go build
./client_test -svr_addr=127.0.0.1:8910