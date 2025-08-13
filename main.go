package main

import (
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
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
	defer wiface.Stop()

	for event := range wiface.ResultChan() {

		pod := event.Object.(*corev1.Pod)
		if pod.Namespace != "default" {
			continue
		}

		switch event.Type {
		case watch.Added:
			log.Infof("Pod %s/%s added", pod.Namespace, pod.Name)
			// Handle pod addition
		case watch.Modified:
			log.Infof("Pod %s/%s modified", pod.Namespace, pod.Name)
			// Handle pod modification
		case watch.Deleted:
			log.Infof("Pod %s/%s deleted", pod.Namespace, pod.Name)
			// Handle pod deletion
		}
	}

	log.Info("Watch channel closed")
}
