package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spiffe/go-spiffe/spiffe"
	"github.com/spiffe/spire/proto/spire/api/registration"
	"github.com/transferwise/spire-k8s-operator/pkg/controller/pod"
	"github.com/transferwise/spire-k8s-operator/pkg/controller/spiffeid"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"github.com/transferwise/spire-k8s-operator/pkg/apis"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	var spireHost string
	var trustDomain string
	var cluster string
	var enablePodController bool
	var podLabel string
	var podAnnotation string

	pflag.StringVar(&spireHost, "spire-server", "", "Host and port of the spire server to connect to")
	pflag.StringVar(&trustDomain, "trust-domain", "", "Spire trust domain to create IDs for")
	pflag.StringVar(&cluster, "cluster", "", "Cluster name as configured for psat attestor")
	pflag.BoolVar(&enablePodController, "enable-pod-controller", false, "Enable support for old controller style spiffe ID creation")
	pflag.StringVar(&podLabel, "pod-label", "", "Pod label to use for old auto-creation mechanism")
	pflag.StringVar(&podAnnotation, "pod-annotation", "", "Pod annotation to use for old auto-creation mechanism")

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()

	if len(spireHost) <= 0 {
		log.Error(fmt.Errorf("--spire-server flag must be provided"), "")
		os.Exit(1)
	}

	if len(cluster) <= 0 {
		log.Error(fmt.Errorf("--cluster flag must be provided"), "")
		os.Exit(1)
	}

	if len(trustDomain) <= 0 {
		log.Error(fmt.Errorf("--trust-domain flag must be provided"), "")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	reconcilerConfig := spiffeid.ReconcileSpiffeIdConfig{
		TrustDomain: trustDomain,
		Cluster:     cluster,
	}

	ctx := context.TODO()
	// Become the leader before proceeding
	err = leader.Become(ctx, "spire-k8s-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          "",
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	spireClient, err := ConnectSpire(spireHost)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	log.Info("Connected to spire server.")

	if err := spiffeid.Add(mgr, spireClient, reconcilerConfig); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if enablePodController {
		mode := pod.PodReconcilerModeServiceAccount
		value := ""
		if len(podLabel) > 0 {
			mode = pod.PodReconcilerModeLabel
			value = podLabel
		}
		if len(podAnnotation) > 0 {
			mode = pod.PodReconcilerModeAnnotation
			value = podAnnotation
		}
		podControllerConfig := pod.PodReconcilerConfig{
			TrustDomain: trustDomain,
			Mode:        mode,
			Value:       value,
		}
		if err := pod.Add(mgr, podControllerConfig); err != nil {
			log.Error(err, "")
			os.Exit(1)
		}
	}

	if err = serveCRMetrics(cfg); err != nil {
		log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
	}

	// Add to the below struct any other metrics ports you want to expose.
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	}
	// Create Service object to expose the metrics port(s).
	service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		log.Info("Could not create metrics Service", "error", err.Error())
	}

	// CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
	// necessary to configure Prometheus to scrape metrics from this operator.
	services := []*v1.Service{service}

	namespace, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		if err != k8sutil.ErrRunLocal {
			log.Error(err, "Failed to get my namespace")
			os.Exit(1)
		}
	} else {
		_, err = metrics.CreateServiceMonitors(cfg, namespace, services)
		if err != nil {
			log.Info("Could not create ServiceMonitor object", "error", err.Error())
			// If this operator is deployed to a cluster without the prometheus-operator running, it will return
			// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
			if err == metrics.ErrServiceMonitorNotPresent {
				log.Info("Install prometheus-operator in your cluster to create ServiceMonitor objects", "error", err.Error())
			}
		}
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

type SpiffeLogWrapper struct {
	delegate logr.Logger
}

func (slw SpiffeLogWrapper) Debugf(format string, args ...interface{}) {
	slw.delegate.V(1).Info(fmt.Sprintf(format, args))
}
func (slw SpiffeLogWrapper) Infof(format string, args ...interface{}) {
	slw.delegate.Info(fmt.Sprintf(format, args))
}
func (slw SpiffeLogWrapper) Warnf(format string, args ...interface{}) {
	slw.delegate.Info(fmt.Sprintf(format, args))
}
func (slw SpiffeLogWrapper) Errorf(format string, args ...interface{}) {
	slw.delegate.Info(fmt.Sprintf(format, args))
}

func ConnectSpire(serviceName string) (registration.RegistrationClient, error) {

	tlsPeer, err := spiffe.NewTLSPeer(spiffe.WithWorkloadAPIAddr("unix:///run/spire/sockets/agent.sock"), spiffe.WithLogger(SpiffeLogWrapper{log}))
	if err != nil {
		return nil, err
	}
	conn, err := tlsPeer.DialGRPC(context.TODO(), serviceName, spiffe.ExpectAnyPeer())
	if err != nil {
		return nil, err
	}
	spireClient := registration.NewRegistrationClient(conn)
	return spireClient, nil
}

// serveCRMetrics gets the Operator/CustomResource GVKs and generates metrics based on those types.
// It serves those metrics on "http://metricsHost:operatorMetricsPort".
func serveCRMetrics(cfg *rest.Config) error {
	// Below function returns filtered operator/CustomResource specific GVKs.
	// For more control override the below GVK list with your own custom logic.
	filteredGVK, err := k8sutil.GetGVKsFromAddToScheme(apis.AddToScheme)
	if err != nil {
		return err
	}
	// Get the namespace the operator is currently deployed in.
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}
	// To generate metrics in other namespaces, add the values below.
	ns := []string{operatorNs}
	// Generate and serve custom resource specific metrics.
	err = kubemetrics.GenerateAndServeCRMetrics(cfg, ns, filteredGVK, metricsHost, operatorMetricsPort)
	if err != nil {
		return err
	}
	return nil
}
