#!/bin/bash

TestRestartAfterTermination() {
    ./foreman >> /dev/null &
    sleep 0.2

    pid=$(ps | grep "sleep" | awk '{print $1}')
    [[ ! -z $pid ]] && kill -SIGINT $pid
    sleep 0.2
    
    restartPid=$(ps | grep "sleep" | awk '{print $1}')
    if [[ -z $restartPid ]]; then
        echo "TestRestartAfterTermination: TEST FAILED"
        Clean
        exit 1
    else
        echo "TestRestartAfterTermination: TEST PASSED"
    fi
}

TestTerminateRunOnceService() {
    ./foreman >> /dev/null &
    sleep 0.2

    pid=$(ps | grep "redis-server" | awk '{print $1}')
    [[ ! -z $pid ]] && kill -SIGINT $pid
    sleep 0.2

    restartPid=$(ps | grep "redis-server" | awk '{print $1}')
    if [[ -z $restartPid ]]; then
        echo "TestTerminateRunOnceService: TEST PASSED"
    else
        echo "TestTerminateRunOnceService: TEST FAILED"
        Clean
        exit 1
    fi
}

TestTerminationOnBrockenDependency() {
    ./foreman >> /dev/null &
    sleep 0.2

    pid=$(ps | grep "redis-server" | awk '{print $1}')
    [[ ! -z $pid ]] && kill -SIGINT $pid
    sleep 0.2

    pingPid=$(ps | grep "ping" | awk '{print $1}')
    if [[ -z $sleepPid ]]; then
        echo "TestTerminationOnBrockenDependency: TEST PASSED"
    else
        echo "TestTerminationOnBrockenDependency: TEST FAILED"
        Clean
        exit 1
    fi

    Clean
}

Clean() {
    foreman=$(ps | grep "foreman" | awk '{print $1}')
    kill -SIGTERM $foreman
}

go build -o foreman foreman.go service_manager.go services_graph.go procfile_parser.go

TestRestartAfterTermination
TestTerminateRunOnceService
TestTerminationOnBrockenDependency

rm ./foreman

echo "status ok: TEST PASSED"
exit 0
