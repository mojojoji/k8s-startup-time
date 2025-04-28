package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Response struct {
	StartupTime string `json:"startup_time"`
	Error       string `json:"error,omitempty"`
}

const (
	namespace         = "k8s-test"
	healthServerImage = "ghcr.io/mojojoji/k8s-startup-time-health-server:latest"
)

func measureStartupTime(clientset *kubernetes.Clientset) (time.Duration, error) {
	// Create deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "startup-test",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "startup-test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "startup-test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "health-server",
							Image: healthServerImage,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
		},
	}

	// Start timing
	startTime := time.Now()

	// Create deployment
	_, err := clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return 0, fmt.Errorf("error creating deployment: %v", err)
	}

	// Wait for pod to be ready and check health endpoint
	var podIP string
	for {
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: "app=startup-test",
		})
		if err != nil {
			log.Printf("Error listing pods: %v", err)
			time.Sleep(20 * time.Millisecond)
			continue
		}

		if len(pods.Items) > 0 && pods.Items[0].Status.PodIP != "" {
			podIP = pods.Items[0].Status.PodIP
			log.Printf("Got IP %s : %s", podIP, time.Since(startTime))
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Check health endpoint
	for {
		resp, err := http.Get(fmt.Sprintf("http://%s:8080/health", podIP))
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Printf("Got Healthy response %s : %s", podIP, time.Since(startTime))
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Calculate total time
	return time.Since(startTime), nil
}

func main() {
	// Get in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error getting in-cluster config: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	http.HandleFunc("/measure", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		startupTime, err := measureStartupTime(clientset)
		response := Response{}
		if err != nil {
			response.Error = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			response.StartupTime = startupTime.String()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	port := "8080"
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func int32Ptr(i int32) *int32 { return &i }
