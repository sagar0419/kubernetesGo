package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/scheme"
)

// Creating access for kubeconfig which is in ~/.kube folder
func GetClient() (*kubernetes.Clientset, error) {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// Deploying App on the Kubernetes cluster
func DeployApp(ctx context.Context, client *kubernetes.Clientset) (map[string]string, error) {
	var deployment *v1.Deployment

	appFile, err := os.ReadFile("deploy.yaml")
	if err != nil {
		return nil, fmt.Errorf("appFile cannot read, an error is there %v \n", err)
	}

	obj, groupVersionKind, err := scheme.Codecs.UniversalDeserializer().Decode(appFile, nil, nil)

	switch obj.(type) {
	case *v1.Deployment:
		deployment = obj.(*v1.Deployment)
	default:
		return nil, fmt.Errorf("Unrecognised type: %s \n", groupVersionKind)
	}

	_, err = client.AppsV1().Deployments("default").Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		deployResponse, err := client.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{}) // if Deployments("") empty then it will select the default namespace. You can also create the namespace.
		if err != nil {
			return nil, fmt.Errorf("deployResponse error %v \n", err)
		}
		return deployResponse.Spec.Template.Labels, nil
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("deployResponse get error %v \n", err)
	}

	deployResponse, err := client.AppsV1().Deployments("default").Update(ctx, deployment, metav1.UpdateOptions{}) // if Deployments("") empty then it will select the default namespace. You can also create the namespace.
	if err != nil {
		return nil, fmt.Errorf("deployResponse update error %v \n", err)
	}
	return deployResponse.Spec.Template.Labels, nil
}

// Checking the status of the pods

func PodStatus(ctx context.Context, client *kubernetes.Clientset, podlabels map[string]string) error {
	for {
		validateLabel, err := labels.ValidatedSelectorFromSet(podlabels)
		if err != nil {
			return fmt.Errorf("error in pods validateLabel  %v \n", err)
		}

		podlist, err := client.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
			LabelSelector: validateLabel.String(),
		})
		if err != nil {
			return fmt.Errorf("error in listing pods %v \n", podlist)
		}
		podsRunning := 0
		for _, pods := range podlist.Items {
			if pods.Status.Phase == "Running" {
				podsRunning++
			}
		}
		if podsRunning > 0 && podsRunning == len(podlist.Items) {
			fmt.Printf("Waiting for the pods to start Pod status running %d /%d \n", podsRunning, len(podlist.Items))
			break
		}
		fmt.Printf("Waiting for the pods to start Pod status running %d /%d \n", podsRunning, len(podlist.Items))
		time.Sleep(1 * time.Second)
	}
	return nil
}
