#!/usr/bin/env bash
#
#  Coverage report for windows builds
#
#First general build
echo "Building the source"
go build || exit $?

echo "Starting unit test coverage"
go test . examples/... generated/... -coverprofile=cp.out

# Output HTML coverage report (to coverage.html)
echo "Creating HTML coverage report (coverage.html)"
go tool cover -html=cp.out

# Now show in default browser
echo "Launching browser with results"
