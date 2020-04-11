package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	//_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientConfig first tries to get a config object which uses the service account kubernetes gives to pods,
// if it is called from a process running in a kubernetes environment.
// Otherwise, it tries to build config from a default kubeconfig filepath if it fails, it fallback to the default config.
// Once it get the config, it returns the same.
func GetClientConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Unable to create config. Error: %+v", err)
		err1 := err
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			err = fmt.Errorf("InClusterConfig as well as BuildConfigFromFlags Failed. Error in InClusterConfig: %+v\nError in BuildConfigFromFlags: %+v", err1, err)
			return nil, err
		}
	}

	return config, nil
}

// GetClientset first tries to get a config object which uses the service account kubernetes gives to pods,
// if it is called from a process running in a kubernetes environment.
// Otherwise, it tries to build config from a default kubeconfig filepath if it fails, it fallback to the default config.
// Once it get the config, it creates a new Clientset for the given config and returns the clientset.
func GetClientset() (*kubernetes.Clientset, error) {
	config, err := GetClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		err = fmt.Errorf("Failed creating clientset. Error: %+v", err)
		return nil, err
	}

	return clientset, nil
}

// PrettyString returns the prettified string of the interface supplied. (If it can)
func PrettyString(in interface{}) string {
	jsonStr, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		err := fmt.Errorf("Unable to marshal, Error: %+v", err)
		if err != nil {
			fmt.Printf("Unable to marshal, Error: %+v\n", err)
		}
		return fmt.Sprintf("%+v", in)
	}
	return string(jsonStr)
}

func main() {
	clientset, err := GetClientset()
	if err != nil {
		panic(err)
	}

	// build the pod defination we want to deploy
	podObject := getPodObject()

	// now create the pod in kubernetes cluster using the clientset

	pod, err := clientset.CoreV1().Pods(podObject.Namespace).Create(podObject)
	if err != nil {
		fmt.Errorf("cannot create pod")
	}

	fmt.Println("Pod ", pod.Name, "created successfully...")

	fmt.Println("Listing pods..")

	pods, err := clientset.CoreV1().Pods("default").List(metav1.ListOptions{})

	for _, pod := range pods.Items {
		fmt.Println("Pod Name: ", pod.Name)
		fmt.Printf(PrettyString(pod))
		fmt.Println()
		fmt.Println(strings.Repeat("*", 80))
	}

	fmt.Println("Pods listed")

}

func getPodObject() *core.Pod {

	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "demo",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "busybox",
					Image:           "busybox",
					ImagePullPolicy: core.PullIfNotPresent,
					Command: []string{
						"sleep",
						"3600",
					},
				},
			},
		},
	}
}
