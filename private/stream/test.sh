#!/usr/bin/env bash

tests="none blake highway default"
for i in `seq 1 3`; do
    for test in $tests; do
       STORJ_HASH=$test time go test -run '^\QTestUpload\E$' -benchmem -cpuprofile $test-$i.pprof
    done
done
