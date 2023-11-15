package killns

import (
	"context"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/tools/setup-envtest/env"
	"sigs.k8s.io/controller-runtime/tools/setup-envtest/remote"
	"sigs.k8s.io/controller-runtime/tools/setup-envtest/store"
	"sigs.k8s.io/controller-runtime/tools/setup-envtest/versions"
	"sigs.k8s.io/controller-runtime/tools/setup-envtest/workflows"
	"testing"
)

func setupEnvTest() *envtest.Environment {
	envTestDir, err := store.DefaultStoreDir()
	if err != nil {
		panic(err)
	}
	envTest := &env.Env{
		FS:  afero.Afero{Fs: afero.NewOsFs()},
		Out: os.Stdout,
		Client: &remote.Client{
			Bucket: "kubebuilder-tools",
			Server: "storage.googleapis.com",
		},
		Platform: versions.PlatformItem{
			Platform: versions.Platform{
				OS:   runtime.GOOS,
				Arch: runtime.GOARCH,
			},
		},
		Version: versions.AnyVersion,
		Store:   store.NewAt(envTestDir),
	}
	envTest.CheckCoherence()
	workflows.Use{}.Do(envTest)
	versionDir := envTest.Platform.Platform.BaseName(*envTest.Version.AsConcrete())
	return &envtest.Environment{
		BinaryAssetsDirectory: filepath.Join(envTestDir, "k8s", versionDir),
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestKillNamespaceMissingNamespace(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("KillNamespace should have panicked")
		}
		if r.(string) != "namespace not provided" {
			t.Errorf("KillNamespace panicked with unexpected error: %s", r)
		}
	}()
	KillNamespace(&rest.Config{}, "")
}

func TestKillNamespaceMissingConfig(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("KillNamespace should have panicked")
		}
		if r.(string) != ".kube/config not provided" {
			t.Errorf("KillNamespace panicked with unexpected error: %s", r)
		}
	}()
	KillNamespace(nil, "test")
}

func TestKillNamespace(t *testing.T) {
	envTest := setupEnvTest()
	envTestConfig, err := envTest.Start()
	if err != nil {
		t.Errorf("Error starting test environment: %s", err)
		return
	}
	defer func() {
		if stopErr := envTest.Stop(); stopErr != nil {
			panic(stopErr)
		}
	}()
	t.Run("With no namespace", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("KillNamespace should have panicked")
			}
			if r.(*errors.StatusError).ErrStatus.Message != "namespaces \"i-dont-exist\" not found" {
				t.Errorf("KillNamespace panicked with unexpected error: %s", r)
			}
		}()
		KillNamespace(envTestConfig, "i-dont-exist")
	})
	t.Run("With existent namespace", func(t *testing.T) {
		client, _ := kubernetes.NewForConfig(envTestConfig)
		client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "plain",
			},
		}, metav1.CreateOptions{})
		KillNamespace(envTestConfig, "plain")
		_, err := client.CoreV1().Namespaces().Get(context.TODO(), "plain", metav1.GetOptions{})
		if err.Error() != "namespaces \"plain\" not found" {
			t.Errorf("Namespace should have been deleted, but it wasn't")
		}
	})
	t.Run("With existent namespace with finalizer", func(t *testing.T) {
		client, _ := kubernetes.NewForConfig(envTestConfig)
		client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "finalizer",
			},
			Spec: corev1.NamespaceSpec{Finalizers: []corev1.FinalizerName{"kubernetes"}},
		}, metav1.CreateOptions{})
		KillNamespace(envTestConfig, "finalizer")
		_, err := client.CoreV1().Namespaces().Get(context.TODO(), "finalizer", metav1.GetOptions{})
		if err.Error() != "namespaces \"finalizer\" not found" {
			t.Errorf("Namespace should have been deleted, but it wasn't")
		}
	})
}
