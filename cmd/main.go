package main

import (
	"flag"
	"fmt"
	"github.com/marcnuri-demo/kubectl-kill-ns/internal/killns"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
			os.Exit(1)
		}
	}()
	kubeConfig, err := killns.LoadKubeConfig()
	if err != nil {
		fmt.Println(".kube/config not found", err)
		os.Exit(1)
	}
	flag.Parse()
	killns.KillNamespace(kubeConfig, flag.Arg(0))
}
