[![Go Reference](https://pkg.go.dev/badge/iwdgo/sigintwindows.svg)](https://pkg.go.dev/iwdgo/sigintwindows)
[![Go Report Card](https://goreportcard.com/badge/github.com/iwdgo/sigintwindows)](https://goreportcard.com/report/github.com/iwdgo/sigintwindows)

# On Windows, sends a ctrl-break to a process

## How to

### Experiment

```
$ go get -d github.com/iwdgo/sigintwindows
$ cd <download path>
$ go test -v
=== RUN   TestSendCtrlBreak
    signal_windows_test.go:40: waiting 5 seconds before goroutine. No log to find.
    signal_windows_test.go:43: waiting 5 seconds in goroutine. Log displays unless interrupted.
2021/09/22 10:39:38 sub-process 55536 started
2021/09/22 10:39:48 graceful exit on interrupt
--- PASS: TestSendCtrlBreak (15.38s)
PASS
```

The output of the sub-process is in the `ctrlbreak.log` file.

Typing Ctrl-C should display error ` exit status 0xc000013a `

Exit code `0xC000013A` is the exit value `STATUS_CONTROL_C_EXIT` returned by the signal package.
[NTSTATUS](https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/596a1078-e883-4972-9bbc-49e60bebca55) says:

> {Application Exit by CTRL+C} The application terminated as a result of a CTRL+C.

### Use as a module

```
import "github.com/iwdgo/sigintwindows"

sigintwindows.SendCtrlBreak(<some pid>)
```


## Online

### Stackoverflow

https://stackoverflow.com/questions/45309984/signal-other-than-sigkill-not-terminating-process-on-windows  
https://stackoverflow.com/questions/55092139/gracefully-terminate-a-process-on-windows  

### Golang

https://github.com/golang/go/issues/29744

### About Ctrl-Break on Windows

https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/596a1078-e883-4972-9bbc-49e60bebca55  
https://docs.microsoft.com/en-us/windows/console/generateconsolectrlevent  
https://docs.microsoft.com/en-us/windows/console/ctrl-c-and-ctrl-break-signals  
https://docs.microsoft.com/en-us/windows/win32/procthread/process-creation-flags  

### Versions

`v0.2.2` Importable module  
`v0.1.0` Standalone experiment  
Standalone version of the test `TestCtrlBreak` of the `signal` package of golang.
