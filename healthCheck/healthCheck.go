package healthCheck

import (
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
)

type Kube struct {
	kubeconfig *string
}

type Status struct {
	Status string
}

func (k *Kube) checkDefaultNamespace(w http.ResponseWriter, r *http.Request) {
	config, err := clientcmd.BuildConfigFromFlags("", *k.kubeconfig)
	s := Status{"unavailable"}
	if err != nil {
		s.Status = "unavailable"
	} else {
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			s.Status = "unavailable"
		} else {
			_, err = clientset.CoreV1().Namespaces().Get("default", metav1.GetOptions{})
			if err != nil {
				s.Status = "unavailable"
			} else {
				s.Status = "available"
			}
		}
	}

	js, err := json.Marshal(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func HealthCheck(kubeconfig *string) {
	kube := Kube{kubeconfig: kubeconfig}
	http.HandleFunc("/health", kube.checkDefaultNamespace)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
