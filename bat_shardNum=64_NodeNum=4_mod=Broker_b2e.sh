#!/bin/bash 

set -ex

PROJECT_DIR="/Users/xjb/Desktop/Huang Lab/BrokerChain/BlockEmulator/b2e-change/block-emulator-b2e"

rm -rf ./log
rm -rf ./record
rm -rf ./result
go build -o b2e main.go

run_cmd() {
    osascript -e "tell application \"Terminal\" \
        to do script \"cd '$PROJECT_DIR' && $1\""
}

./b2e -n 1 -N 4 -s 0 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 0 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 0 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 1 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 1 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 1 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 2 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 2 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 2 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 3 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 3 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 3 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 4 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 4 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 4 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 5 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 5 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 5 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 6 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 6 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 6 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 7 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 7 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 7 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 8 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 8 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 8 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 9 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 9 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 9 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 10 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 10 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 10 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 11 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 11 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 11 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 12 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 12 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 12 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 13 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 13 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 13 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 14 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 14 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 14 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 15 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 15 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 15 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 16 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 16 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 16 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 17 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 17 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 17 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 18 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 18 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 18 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 19 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 19 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 19 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 20 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 20 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 20 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 21 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 21 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 21 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 22 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 22 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 22 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 23 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 23 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 23 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 24 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 24 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 24 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 25 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 25 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 25 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 26 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 26 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 26 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 27 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 27 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 27 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 28 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 28 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 28 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 29 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 29 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 29 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 30 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 30 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 30 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 31 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 31 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 31 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 32 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 32 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 32 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 33 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 33 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 33 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 34 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 34 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 34 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 35 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 35 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 35 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 36 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 36 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 36 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 37 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 37 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 37 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 38 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 38 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 38 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 39 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 39 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 39 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 40 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 40 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 40 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 41 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 41 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 41 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 42 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 42 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 42 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 43 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 43 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 43 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 44 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 44 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 44 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 45 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 45 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 45 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 46 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 46 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 46 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 47 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 47 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 47 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 48 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 48 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 48 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 49 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 49 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 49 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 50 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 50 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 50 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 51 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 51 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 51 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 52 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 52 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 52 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 53 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 53 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 53 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 54 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 54 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 54 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 55 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 55 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 55 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 56 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 56 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 56 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 57 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 57 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 57 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 58 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 58 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 58 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 59 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 59 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 59 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 60 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 60 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 60 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 61 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 61 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 61 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 62 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 62 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 62 -S 64 -m 4 &

./b2e -n 1 -N 4 -s 63 -S 64 -m 4 &

./b2e -n 2 -N 4 -s 63 -S 64 -m 4 &

./b2e -n 3 -N 4 -s 63 -S 64 -m 4 &

run_cmd "./b2e -c -N 4 -S 64 -m 4 &"

for s in $(seq 0 63); do
    run_cmd "./b2e -n 0 -N 4 -s $s -S 64 -m 4"
done

# ./b2e -n 0 -N 4 -s 0 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 1 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 2 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 3 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 4 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 5 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 6 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 7 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 8 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 9 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 10 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 11 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 12 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 13 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 14 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 15 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 16 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 17 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 18 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 19 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 20 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 21 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 22 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 23 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 24 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 25 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 26 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 27 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 28 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 29 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 30 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 31 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 32 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 33 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 34 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 35 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 36 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 37 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 38 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 39 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 40 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 41 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 42 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 43 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 44 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 45 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 46 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 47 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 48 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 49 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 50 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 51 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 52 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 53 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 54 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 55 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 56 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 57 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 58 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 59 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 60 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 61 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 62 -S 64 -m 4 &

# ./b2e -n 0 -N 4 -s 63 -S 64 -m 4 &

