package main

import (
	"context"
	"fmt"
	deploy "k8s/deploy"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
)

func main() {
	var (
		client           *kubernetes.Clientset
		deploymentLabels map[string]string
		err              error
	)

	ctx := context.Background()
	// Accessing kubeconfig
	if client, err = deploy.GetClient(); err != nil {
		log.Fatalf("Error in client: %s \n", err)
		os.Exit(1)
	}

	// Deploying App
	if deploymentLabels, err = deploy.DeployApp(ctx, client); err != nil {
		log.Fatalf("Error in deployApp: %s \n", err)
		os.Exit(1)
	}
	fmt.Printf("Deployment of the app is finished check pods. Deployed with labels: %+v\n", deploymentLabels)

	// Pod Status Check
	if err := deploy.PodStatus(ctx, client, deploymentLabels); err != nil {
		log.Fatalf("Error in podStatus: %s \n", err)
		os.Exit(1)
	}
}
