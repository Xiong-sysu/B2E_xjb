#!/bin/bash 

set -ex

PROJECT_DIR="/Users/xjb/Desktop/Huang Lab/BrokerChain/BlockEmulator/b2e-change/block-emulator-b2e"

rm -rf ./log
rm -rf ./record
rm -rf ./result
go build -o b2e main.go
# run_cmd() {
#     osascript -e "tell application \"Terminal\" to do script \"cd $(pwd); $1\""
# }

run_cmd() {
    osascript -e "tell application \"Terminal\" \
        to do script \"cd '$PROJECT_DIR' && $1\""
}

run_cmd "./b2e -n 1 -N 4 -s 0 -S 6 -m 4 &"

run_cmd "./b2e -n 2 -N 4 -s 0 -S 6 -m 4 &"

run_cmd "./b2e -n 3 -N 4 -s 0 -S 6 -m 4 &"

run_cmd "./b2e -n 1 -N 4 -s 1 -S 6 -m 4 &"

run_cmd "./b2e -n 2 -N 4 -s 1 -S 6 -m 4 &"

run_cmd "./b2e -n 3 -N 4 -s 1 -S 6 -m 4 &"

run_cmd "./b2e -n 1 -N 4 -s 2 -S 6 -m 4 &"

run_cmd "./b2e -n 2 -N 4 -s 2 -S 6 -m 4 &"

run_cmd "./b2e -n 3 -N 4 -s 2 -S 6 -m 4 &"

run_cmd "./b2e -n 1 -N 4 -s 3 -S 6 -m 4 &"

run_cmd "./b2e -n 2 -N 4 -s 3 -S 6 -m 4 &"

run_cmd "./b2e -n 3 -N 4 -s 3 -S 6 -m 4 &"

run_cmd "./b2e -n 1 -N 4 -s 4 -S 6 -m 4 &"

run_cmd "./b2e -n 2 -N 4 -s 4 -S 6 -m 4 &"

run_cmd "./b2e -n 3 -N 4 -s 4 -S 6 -m 4 &"

run_cmd "./b2e -n 1 -N 4 -s 5 -S 6 -m 4 &"

run_cmd "./b2e -n 2 -N 4 -s 5 -S 6 -m 4 &"

run_cmd "./b2e -n 3 -N 4 -s 5 -S 6 -m 4 &"

run_cmd "./b2e -c -N 4 -S 6 -m 4 > supervisor_A.txt  &"

run_cmd "./b2e -n 0 -N 4 -s 0 -S 6 -m 4 &"

run_cmd "./b2e -n 0 -N 4 -s 1 -S 6 -m 4 &"

run_cmd "./b2e -n 0 -N 4 -s 2 -S 6 -m 4 &"

run_cmd "./b2e -n 0 -N 4 -s 3 -S 6 -m 4 &"

run_cmd "./b2e -n 0 -N 4 -s 4 -S 6 -m 4 &"

run_cmd "./b2e -n 0 -N 4 -s 5 -S 6 -m 4 &"

