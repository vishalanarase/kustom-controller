package main

import (
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

			enforceResources(pod)

			_, err := clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
			if err != nil {
				log.WithError(err).Errorf("Failed to update pod %s/%s with enforced resources", pod.Namespace, pod.Name)
			} else {
				log.Infof("Updated pod %s/%s with enforced resources", pod.Namespace, pod.Name)
			}
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

// enforceResources sets default resource requests for a pod's containers if not already set.
func enforceResources(pod *corev1.Pod) {
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Resources.Requests == nil {
			pod.Spec.Containers[i].Resources.Requests = corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			}
			log.Infof("Set default resources for container %s in pod %s/%s", pod.Spec.Containers[i].Name, pod.Namespace, pod.Name)
		}
	}
}
