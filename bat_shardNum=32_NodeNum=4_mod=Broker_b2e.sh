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

./b2e -n 1 -N 4 -s 0 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 0 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 0 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 1 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 1 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 1 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 2 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 2 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 2 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 3 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 3 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 3 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 4 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 4 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 4 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 5 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 5 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 5 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 6 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 6 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 6 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 7 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 7 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 7 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 8 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 8 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 8 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 9 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 9 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 9 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 10 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 10 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 10 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 11 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 11 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 11 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 12 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 12 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 12 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 13 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 13 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 13 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 14 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 14 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 14 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 15 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 15 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 15 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 16 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 16 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 16 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 17 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 17 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 17 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 18 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 18 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 18 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 19 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 19 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 19 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 20 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 20 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 20 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 21 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 21 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 21 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 22 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 22 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 22 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 23 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 23 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 23 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 24 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 24 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 24 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 25 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 25 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 25 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 26 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 26 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 26 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 27 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 27 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 27 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 28 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 28 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 28 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 29 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 29 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 29 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 30 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 30 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 30 -S 32 -m 4 &

./b2e -n 1 -N 4 -s 31 -S 32 -m 4 &

./b2e -n 2 -N 4 -s 31 -S 32 -m 4 &

./b2e -n 3 -N 4 -s 31 -S 32 -m 4 &


run_cmd "./b2e -c -N 4 -S 32 -m 4 &"

for s in $(seq 0 31); do
    run_cmd "./b2e -n 0 -N 4 -s $s -S 32 -m 4"
done

# go run main.go -c -N 4 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 0 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 1 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 2 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 3 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 4 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 5 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 6 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 7 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 8 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 9 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 10 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 11 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 12 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 13 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 14 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 15 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 16 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 17 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 18 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 19 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 20 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 21 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 22 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 23 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 24 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 25 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 26 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 27 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 28 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 29 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 30 -S 32 -m 4 &

# go run main.go -n 0 -N 4 -s 31 -S 32 -m 4 &

