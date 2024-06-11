package machine

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	timeout = 10 * time.Second
)

var (
	k8sClient     client.Client
	eventRecorder record.EventRecorder
	cfg           *rest.Config
)

func TestMain(m *testing.M) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "vendor", "github.com", "openshift", "api", "machine", "v1beta1", "zz_generated.crd-manifests", "0000_10_machine-api_01_machines-CustomNoUpgrade.crd.yaml"),
			filepath.Join("..", "..", "..", "vendor", "github.com", "openshift", "api", "machine", "v1", "zz_generated.crd-manifests", "0000_10_control-plane-machine-set_01_controlplanemachinesets-CustomNoUpgrade.crd.yaml"),
			filepath.Join("..", "..", "..", "vendor", "github.com", "openshift", "api", "config", "v1", "zz_generated.crd-manifests", "0000_10_config-operator_01_infrastructures-CustomNoUpgrade.crd.yaml"),
		},
	}

	configv1.AddToScheme(scheme.Scheme)

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := testEnv.Stop(); err != nil {
			log.Fatal(err)
		}
	}()

	mgr, err := manager.New(cfg, manager.Options{
		Scheme: scheme.Scheme,
		Metrics: server.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	mgrCtx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := mgr.Start(mgrCtx); err != nil {
			log.Fatal(err)
		}
	}()
	defer cancel()

	k8sClient = mgr.GetClient()
	eventRecorder = mgr.GetEventRecorderFor("awscontroller")

	code := m.Run()
	os.Exit(code)
}
