@echo off
rem    Coverage report for windows builds
rem
rem First general build
echo Building the source
go build
if ERRORLEVEL 1 GOTO buildFaild

echo Starting unit test coverage
go test . examples/... generated/... -coverprofile=cp.out

rem Output HTML coverage report (to coverage.html)
echo Creating HTML coverage report (coverage.html)
go tool cover -html=cp.out

rem Now show in default browser
echo Launching browser with results
rem start coverage.html

@echo Done
exit /B 0

:buildFailed
@echo Build failed