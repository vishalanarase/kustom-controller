package main

import (
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	log.Info("Create clientset from config")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to create clientset")
	}

	log.Info("Watch pods")
	wiface, err := clientset.CoreV1().Pods("").Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Fatal("Failed to watch pods")
	}

	for {
		select {
		case msg := <-wiface.ResultChan():
			log.Infof("Received event: %+v", msg)
		}
	}
}
