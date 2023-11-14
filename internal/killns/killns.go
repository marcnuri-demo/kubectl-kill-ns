package killns

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strings"
)

func LoadKubeConfig() (*rest.Config, error) {
	home := homedir.HomeDir()
	kubeConfig := filepath.Join(home, ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", kubeConfig)
}

func KillNamespace(kubeConfig *rest.Config, namespace string) {
	if len(strings.TrimSpace(namespace)) == 0 {
		panic("namespace not provided")
	}
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}
	errDelete := client.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
	ns, errGet := client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err := errors.Join(errDelete, errGet); err != nil {
		panic(err)
	}
	if ns.Spec.Finalizers == nil || len(ns.Spec.Finalizers) == 0 {
		return
	}
	ns.Spec.Finalizers = []corev1.FinalizerName{}
	ns.ObjectMeta.ResourceVersion = ""
	// https://github.com/kubernetes/kubernetes/issues/77086#issuecomment-486840718
	_, errUpdate := client.CoreV1().Namespaces().Finalize(context.TODO(), ns, metav1.UpdateOptions{})
	if errUpdate != nil {
		panic(errUpdate)
	}
}
