package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
}

func main() {
	flag.Parse()

	log.Info("Build config from kubeconfig")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.WithError(err).Fatal("Failed to build config")
	}
}
