#!/bin/bash

usage()
{
    echo "usage: ./bbr.sj [-o onus -p pons -r runs]"
}

run()
{
    echo "Running with: ${ONU} Onus ${PON} Pons for ${RUN} times" >> results.logs
    for i in {0..10}
    do
        echo "RUN Number: $i"
        docker rm -f bbsim
        DOCKER_RUN_ARGS="-pon ${PON} -onu ${ONU}" make docker-run
        sleep 5
        ./bbr -pon $PON -onu $ONU 2>&1 | tee bbr.logs
        docker logs bbsim 2>&1 | tee bbsim.logs
        echo "RUN Number: $i" >> results.logs
        cat bbr.logs | grep Duration | awk '{print $5}' >> results.logs
    done
}

ONU=1
PON=1
RUN=10

while [ "$1" != "" ]; do
    case $1 in
        -o | --onus )           ONU=$2
                                ;;
        -p | --pons )           PON=$2
                                ;;
        -r | --runs )           RUN=$2
                                ;;
        --) # End of all options
                                shift
                                break
                                ;;
    esac
    shift
done

run

