package client

import (
	"strings"
	"sync"

	"github.com/openebs/lib-csi/pkg/common/env"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// K8sMasterIPEnvironmentKey is the environment variable
	// key to provide kubernetes master IP address
	K8sMasterIPEnvironmentKey string = "OPENEBS_IO_K8S_MASTER"

	// KubeConfigEnvironmentKey is the environment variable
	// key to provide kubeconfig path
	KubeConfigEnvironmentKey string = "OPENEBS_IO_KUBE_CONFIG"
)

// getInClusterConfigFn is a typed function
// to abstract getting kubernetes incluster config
//
// NOTE:
//
//	typed function makes it simple to mock
type getInClusterConfigFn func() (*rest.Config, error)

// buildConfigFromFlagsFn is a typed function
// to abstract getting a kubernetes config from
// provided flags
//
// NOTE:
//
//	typed function makes it simple to mock
type buildConfigFromFlagsFn func(string, string) (*rest.Config, error)

// getKubeMasterIPFromENVFn is a typed function
// to abstract getting kubernetes master IP
// address from environment variable
//
// NOTE:
//
//	typed function makes it simple to mock
type getKubeMasterIPFromENVFn func(string) string

// getKubeConfigPathFromENVFn is a typed function to
// abstract getting kubernetes config path from
// environment variable
//
// NOTE:
//
//	typed function makes it simple to mock
type getKubeConfigPathFromENVFn func(string) string

// getKubeDynamicClientFn is a typed function to
// abstract getting dynamic kubernetes clientset
//
// NOTE:
//
//	typed function makes it simple to mock
type getKubeDynamicClientFn func(*rest.Config) (*dynamic.DynamicClient, error)

// getKubeClientsetFn is a typed function
// to abstract getting kubernetes clientset
//
// NOTE:
//
//	typed function makes it simple to mock
type getKubeClientsetFn func(*rest.Config) (*kubernetes.Clientset, error)

// Client provides Kubernetes client operations
type Client struct {
	// IsInCluster flags if this client points
	// to its own cluster
	IsInCluster bool

	// KubeConfigPath to get kubernetes clientset
	KubeConfigPath string

	// handle to get in-cluster config
	getInClusterConfig getInClusterConfigFn

	// handle to get kubernetes config
	// from flags
	buildConfigFromFlags buildConfigFromFlagsFn

	// handle to get kubernetes clienset
	getKubeClientset getKubeClientsetFn

	// handle to get kubernetes dynamic clientset
	getKubeDynamicClient getKubeDynamicClientFn

	// handle to get kubernetes master IP from
	// environment variable
	getKubeMasterIPFromENV getKubeMasterIPFromENVFn

	// handle to get kubernetes config path
	// from environment variable
	getKubeConfigPathFromENV getKubeConfigPathFromENVFn

	// qps optionally overrides the maximum queries per second the client
	// sends to the Kubernetes API server. Zero leaves the rest.Config
	// default in place (client-go applies QPS 5).
	qps float32

	// burst optionally overrides the maximum burst of queries above qps the
	// client allows. Zero leaves the rest.Config default in place (client-go
	// applies Burst 10).
	burst int
}

// OptionFn is a typed function to abstract
// any operation against the provided client
// instance
//
// NOTE:
//
//	This is the basic building block to create
//
// functional operations against the client
// instance
type OptionFn func(*Client)

// New returns a new instance of client
func New(opts ...OptionFn) *Client {
	c := &Client{}
	for _, o := range opts {
		o(c)
	}

	withDefaults(c)
	return c
}

var (
	instance *Client
	once     sync.Once
)

// Instance returns a singleton instance of
// this client
func Instance(opts ...OptionFn) *Client {
	once.Do(func() {
		instance = New(opts...)
	})

	return instance
}

// withDefaults sets the provided instance of
// client with necessary defaults
func withDefaults(c *Client) {
	for _, def := range defaultFns {
		def(c)
	}
}

var defaultFns = []OptionFn{
	withDefaultGetInClusterConfigFn(),
	withDefaultBuildConfigFromFlagsFn(),
	withDefaultGetKubeClientsetFn(),
	withDefaultGetKubeDynamicClientFn(),
	withDefaultGetKubeMasterIPFromENVFn(),
	withDefaultGetKubeConfigPathFromENVFn(),
}

// withDefaultGetInClusterConfigFn sets the default logic
// to get in-cluster config
func withDefaultGetInClusterConfigFn() OptionFn {
	return func(c *Client) {
		if c.getInClusterConfig == nil {
			c.getInClusterConfig = rest.InClusterConfig
		}
	}
}

// withDefaultBuildConfigFromFlagsFn sets the default logic
// to build config from flags
func withDefaultBuildConfigFromFlagsFn() OptionFn {
	return func(c *Client) {
		if c.buildConfigFromFlags == nil {
			c.buildConfigFromFlags = clientcmd.BuildConfigFromFlags
		}
	}
}

// withDefaultGetKubeClientsetFn sets the default logic
// to get kubernetes clientset
func withDefaultGetKubeClientsetFn() OptionFn {
	return func(c *Client) {
		if c.getKubeClientset == nil {
			c.getKubeClientset = kubernetes.NewForConfig
		}
	}
}

// withDefaultGetKubeDynamicClientFn sets the default logic
// to get kubernetes dynamic client instance
func withDefaultGetKubeDynamicClientFn() OptionFn {
	return func(c *Client) {
		if c.getKubeDynamicClient == nil {
			c.getKubeDynamicClient = dynamic.NewForConfig
		}
	}
}

// withDefaultGetKubeMasterIPFromENVFn sets the default logic
// to get kubernetes master IP address from environment
// variable
func withDefaultGetKubeMasterIPFromENVFn() OptionFn {
	return func(c *Client) {
		if c.getKubeMasterIPFromENV == nil {
			c.getKubeMasterIPFromENV = env.Get
		}
	}
}

// withDefaultGetKubeConfigPathFromENVFn sets the default logic
// to get kubeconfig path from environment variable
func withDefaultGetKubeConfigPathFromENVFn() OptionFn {
	return func(c *Client) {
		if c.getKubeConfigPathFromENV == nil {
			c.getKubeConfigPathFromENV = env.Get
		}
	}
}

// InCluster enables IsInCluster flag
func InCluster() OptionFn {
	return func(c *Client) {
		c.IsInCluster = true
	}
}

// WithKubeConfigPath sets kubeconfig path
// against this client instance
func WithKubeConfigPath(kubeConfigPath string) OptionFn {
	return func(c *Client) {
		c.KubeConfigPath = kubeConfigPath
	}
}

// WithQPS sets the maximum queries per second the client may send to the
// Kubernetes API server. A non-positive value is ignored, leaving the
// client-go default in place.
func WithQPS(qps float32) OptionFn {
	return func(c *Client) {
		if qps <= 0 {
			return
		}
		c.qps = qps
	}
}

// WithBurst sets the maximum burst of queries above the QPS limit the client
// may send to the Kubernetes API server. A non-positive value is ignored,
// leaving the client-go default in place.
func WithBurst(burst int) OptionFn {
	return func(c *Client) {
		if burst <= 0 {
			return
		}
		c.burst = burst
	}
}

// GetConfig returns Kubernetes config instance
// from the provided client
func GetConfig(c *Client) (*rest.Config, error) {
	if c == nil {
		return nil, errors.New("failed to get kubernetes config: nil client provided")
	}

	return c.GetConfigForPathOrDirect()
}

// GetConfigForPathOrDirect returns Kubernetes config
// instance from kubeconfig path or without it
func (c *Client) GetConfigForPathOrDirect() (config *rest.Config, err error) {
	if c.KubeConfigPath != "" {
		return c.ConfigForPath(c.KubeConfigPath)
	}

	return c.Config()
}

// ConfigForPath returns Kubernetes config instance
// based on KubeConfig path
func (c *Client) ConfigForPath(kubeConfigPath string) (config *rest.Config, err error) {
	return c.buildConfigFromFlags("", kubeConfigPath)
}

// Config returns Kubernetes config instance
// based on set criteria
func (c *Client) Config() (config *rest.Config, err error) {
	// IsInCluster flag holds the top most priority
	if c.IsInCluster {
		return c.getInClusterConfig()
	}

	// ENV holds second priority
	if strings.TrimSpace(c.getKubeMasterIPFromENV(K8sMasterIPEnvironmentKey)) != "" ||
		strings.TrimSpace(c.getKubeConfigPathFromENV(KubeConfigEnvironmentKey)) != "" {
		return c.getConfigFromENV()
	}

	// Defaults to InClusterConfig
	return c.getInClusterConfig()
}

func (c *Client) getConfigFromENV() (config *rest.Config, err error) {
	k8sMaster := c.getKubeMasterIPFromENV(K8sMasterIPEnvironmentKey)
	kubeConfig := c.getKubeConfigPathFromENV(KubeConfigEnvironmentKey)

	if strings.TrimSpace(k8sMaster) == "" &&
		strings.TrimSpace(kubeConfig) == "" {
		return nil, errors.Errorf(
			"failed to get kubernetes config: missing ENV: atleast one should be set: {%s} or {%s}",
			K8sMasterIPEnvironmentKey,
			KubeConfigEnvironmentKey,
		)
	}

	return c.buildConfigFromFlags(k8sMaster, kubeConfig)
}

// Clientset returns a new instance of Kubernetes clientset
func (c *Client) Clientset() (*kubernetes.Clientset, error) {
	config, err := c.GetConfigForPathOrDirect()
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to get kubernetes clientset: IsInCluster {%t}: KubeConfigPath {%s}",
			c.IsInCluster,
			c.KubeConfigPath,
		)
	}

	c.applyRateLimits(config)
	return c.getKubeClientset(config)
}

// applyRateLimits overrides the client-side QPS and burst on the provided
// config when they have been explicitly set via WithQPS/WithBurst. Zero
// values are left untouched so client-go applies its own defaults (QPS 5,
// Burst 10), keeping behavior unchanged unless an override is supplied. The
// one exception is a QPS-only override, which additionally needs an explicit
// burst to satisfy kubernetes.NewForConfig (see below).
func (c *Client) applyRateLimits(config *rest.Config) {
	if config == nil {
		return
	}
	if c.qps > 0 {
		config.QPS = c.qps
	}
	if c.burst > 0 {
		config.Burst = c.burst
	}
	// kubernetes.NewForConfig rejects a config with QPS > 0 and Burst <= 0,
	// so a QPS-only override would fail client construction. Fall back to the
	// burst client-go would otherwise have defaulted to, keeping the burst
	// dimension unchanged from the caller's point of view.
	if config.QPS > 0 && config.Burst <= 0 {
		config.Burst = rest.DefaultBurst
	}
}

// Dynamic returns a kubernetes dynamic client capable
// of invoking operations against kubernetes resources
func (c *Client) Dynamic() (dynamic.Interface, error) {
	config, err := c.GetConfigForPathOrDirect()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dynamic client")
	}

	return c.getKubeDynamicClient(config)
}
