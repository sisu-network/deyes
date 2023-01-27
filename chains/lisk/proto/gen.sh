#!/bin/bash

protoc -I=. --go_out=.. ./transfer_data.proto
