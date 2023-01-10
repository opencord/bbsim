#!/bin/bash

# Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

usage()
{
    echo " "
    echo "Runs BBR against BBSim a number of times and output the results in a file name results.logs"
    echo " "
    echo "Usage: $0 [--onus|--pons|--runs]" >&2
    echo " "
    echo "   -o, --onus              Number of ONUs to emulate on each PON"
    echo "   -p, --pons              Number of PONs to emulate"
    echo "   -r, --runs              Number of runs to perform"
    echo " "
    echo "Example usages:"
    echo "   ./$0 -i -p 2 -o 32 -r 10"
    echo " "
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
        -h | --help)
                                usage
                                exit 0
                                ;;
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

