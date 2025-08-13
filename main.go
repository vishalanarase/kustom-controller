package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig  string
	metricsAddr string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint")
}

func main() {
	flag.Parse()

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Infof("Starting metrics server at %s", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			log.WithError(err).Fatal("Failed to start metrics server")
		}
	}()

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
		startTime := time.Now()
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			log.Error("Received non-pod object from watcher")
			errorsCount.Inc()
			continue
		}

		// Record metrics for all processed pods
		podsProcessed.WithLabelValues(pod.Namespace, string(event.Type)).Inc()

		// Skip non-default namespaces
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
				errorsCount.Inc()
			} else {
				log.Infof("Updated pod %s/%s with enforced resources", pod.Namespace, pod.Name)
			}
		case watch.Modified:
			log.Infof("Pod %s/%s modified", pod.Namespace, pod.Name)
			// Handle pod modification
		case watch.Deleted:
			log.Infof("Pod %s/%s deleted", pod.Namespace, pod.Name)
			// Handle pod deletion
		case watch.Error:
			log.Errorf("Watch error: %v", event.Object)
			errorsCount.Inc()
			return // Will trigger reconnect
		}

		processingTime.Observe(time.Since(startTime).Seconds())
	}

	log.Info("Watch channel closed")
}

// enforceResources sets default resource requests for a pod's containers if not already set.
func enforceResources(pod *corev1.Pod) {
	changesMade := false

	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]

		// Initialize Resources if nil
		if container.Resources.Requests == nil {
			container.Resources.Requests = make(corev1.ResourceList)
			changesMade = true
		}

		// Set CPU if not specified
		if _, exists := container.Resources.Requests[corev1.ResourceCPU]; !exists {
			container.Resources.Requests[corev1.ResourceCPU] = resource.MustParse("100m")
			changesMade = true
			log.Infof("Set default CPU request for container %s in pod %s/%s",
				container.Name, pod.Namespace, pod.Name)
		}

		// Set Memory if not specified
		if _, exists := container.Resources.Requests[corev1.ResourceMemory]; !exists {
			container.Resources.Requests[corev1.ResourceMemory] = resource.MustParse("128Mi")
			changesMade = true
			log.Infof("Set default memory request for container %s in pod %s/%s",
				container.Name, pod.Namespace, pod.Name)
		}
	}

	if changesMade {
		resourcesEnforced.Inc()
	}
}
