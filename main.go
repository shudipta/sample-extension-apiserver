package main

import (
	"os"
	"runtime"

	"sample-extension-apiserver/cmd"
	"k8s.io/apiserver/pkg/util/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}