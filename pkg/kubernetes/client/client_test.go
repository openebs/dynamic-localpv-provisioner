package client

import (
	"testing"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func fakeGetClientsetOk(c *rest.Config) (*kubernetes.Clientset, error) {
	return &kubernetes.Clientset{}, nil
}

func fakeGetClientsetErr(c *rest.Config) (*kubernetes.Clientset, error) {
	return nil, errors.New("fake error")
}

func fakeInClusterConfigOk() (*rest.Config, error) {
	return &rest.Config{}, nil
}

// fakeInClusterConfigWithHost returns a config complete enough for the real
// kubernetes.NewForConfig to accept, so tests can exercise clientset
// construction rather than a mock of it.
func fakeInClusterConfigWithHost() (*rest.Config, error) {
	return &rest.Config{Host: "https://127.0.0.1:6443"}, nil
}

func fakeInClusterConfigErr() (*rest.Config, error) {
	return nil, errors.New("fake error")
}

func fakeBuildConfigFromFlagsOk(kubemaster string, kubeconfig string) (*rest.Config, error) {
	return &rest.Config{}, nil
}

func fakeBuildConfigFromFlagsErr(kubemaster string, kubeconfig string) (*rest.Config, error) {
	return nil, errors.New("fake error")
}

func fakeGetKubeConfigPathOk(e string) string {
	return "fake"
}

func fakeGetKubeConfigPathNil(e string) string {
	return ""
}

func fakeGetKubeMasterIPOk(e string) string {
	return "fake"
}

func fakeGetKubeMasterIPNil(e string) string {
	return ""
}

func fakeGetDynamicClientSetOk(c *rest.Config) (*dynamic.DynamicClient, error) {
	return dynamic.NewForConfig(c)
}

func fakeGetDynamicClientSetErr(c *rest.Config) (*dynamic.DynamicClient, error) {
	return nil, errors.New("fake error")
}

func TestNewInCluster(t *testing.T) {
	c := New(InCluster())
	if !c.IsInCluster {
		t.Fatalf("test failed: expected IsInCluster as 'true' actual '%t'", c.IsInCluster)
	}
}

func TestConfig(t *testing.T) {
	tests := map[string]struct {
		isInCluster        bool
		kubeConfigPath     string
		getInClusterConfig getInClusterConfigFn
		getKubeMasterIP    getKubeMasterIPFromENVFn
		getKubeConfigPath  getKubeConfigPathFromENVFn
		getConfigFromENV   buildConfigFromFlagsFn
		isErr              bool
	}{
		"t1": {true, "", fakeInClusterConfigOk, nil, nil, nil, false},
		"t2": {true, "", fakeInClusterConfigErr, nil, nil, nil, true},
		"t3": {false, "", fakeInClusterConfigErr, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, true},
		"t4": {false, "", fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, false},
		"t5": {false, "fakeKubeConfigPath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, false},
		"t6": {false, "", nil, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t7": {false, "", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t8": {false, "fakeKubeConfigPath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, true},
		"t9": {false, "fakeKubeConfigpath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				IsInCluster:              mock.isInCluster,
				KubeConfigPath:           mock.kubeConfigPath,
				getInClusterConfig:       mock.getInClusterConfig,
				getKubeMasterIPFromENV:   mock.getKubeMasterIP,
				getKubeConfigPathFromENV: mock.getKubeConfigPath,
				buildConfigFromFlags:     mock.getConfigFromENV,
			}
			_, err := c.Config()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestGetConfigFromENV(t *testing.T) {
	tests := map[string]struct {
		getKubeMasterIP   getKubeMasterIPFromENVFn
		getKubeConfigPath getKubeConfigPathFromENVFn
		getConfigFromENV  buildConfigFromFlagsFn
		isErr             bool
	}{
		"t1": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, true},
		"t2": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t3": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, false},
		"t4": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, false},
		"t5": {fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, true},
		"t6": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsErr, true},
		"t7": {fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, true},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				getKubeMasterIPFromENV:   mock.getKubeMasterIP,
				getKubeConfigPathFromENV: mock.getKubeConfigPath,
				buildConfigFromFlags:     mock.getConfigFromENV,
			}
			_, err := c.getConfigFromENV()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestGetConfigFromPathOrDirect(t *testing.T) {
	tests := map[string]struct {
		kubeConfigPath     string
		getConfigFromFlags buildConfigFromFlagsFn
		getInClusterConfig getInClusterConfigFn
		isErr              bool
	}{
		"T1": {"", fakeBuildConfigFromFlagsErr, fakeInClusterConfigOk, false},
		"T2": {"fake-path", fakeBuildConfigFromFlagsOk, fakeInClusterConfigErr, false},
		"T3": {"fake-path", fakeBuildConfigFromFlagsErr, fakeInClusterConfigOk, true},
		"T4": {"", fakeBuildConfigFromFlagsOk, fakeInClusterConfigErr, true},
		"T5": {"fake-path", fakeBuildConfigFromFlagsErr, fakeInClusterConfigErr, true},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				KubeConfigPath:           mock.kubeConfigPath,
				buildConfigFromFlags:     mock.getConfigFromFlags,
				getInClusterConfig:       mock.getInClusterConfig,
				getKubeMasterIPFromENV:   fakeGetKubeMasterIPNil,
				getKubeConfigPathFromENV: fakeGetKubeConfigPathNil,
			}
			_, err := c.GetConfigForPathOrDirect()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestClientset(t *testing.T) {
	tests := map[string]struct {
		isInCluster            bool
		kubeConfigPath         string
		getInClusterConfig     getInClusterConfigFn
		getKubeMasterIP        getKubeMasterIPFromENVFn
		getKubeConfigPath      getKubeConfigPathFromENVFn
		getConfigFromENV       buildConfigFromFlagsFn
		getKubernetesClientset getKubeClientsetFn
		isErr                  bool
	}{
		"t10": {true, "", fakeInClusterConfigOk, nil, nil, nil, fakeGetClientsetOk, false},
		"t11": {true, "", fakeInClusterConfigOk, nil, nil, nil, fakeGetClientsetErr, true},
		"t12": {true, "", fakeInClusterConfigErr, nil, nil, nil, fakeGetClientsetOk, true},

		"t21": {false, "", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathNil, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t22": {false, "", nil, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t23": {false, "", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
		"t24": {false, "fake-path", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, fakeGetClientsetOk, true},
		"t25": {false, "", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetErr, true},
		"t26": {false, "fakePath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsErr, fakeGetClientsetOk, true},

		"t30": {false, "", fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, fakeGetClientsetOk, false},
		"t31": {false, "", fakeInClusterConfigOk, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, fakeGetClientsetErr, true},
		"t32": {false, "", fakeInClusterConfigErr, fakeGetKubeMasterIPNil, fakeGetKubeConfigPathNil, nil, nil, true},
		"t33": {false, "fakePath", nil, fakeGetKubeMasterIPOk, fakeGetKubeConfigPathOk, fakeBuildConfigFromFlagsOk, fakeGetClientsetOk, false},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				IsInCluster:              mock.isInCluster,
				KubeConfigPath:           mock.kubeConfigPath,
				getInClusterConfig:       mock.getInClusterConfig,
				getKubeMasterIPFromENV:   mock.getKubeMasterIP,
				getKubeConfigPathFromENV: mock.getKubeConfigPath,
				buildConfigFromFlags:     mock.getConfigFromENV,
				getKubeClientset:         mock.getKubernetesClientset,
			}
			_, err := c.Clientset()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error actual '%s'", name, err)
			}
		})
	}
}

func TestDynamic(t *testing.T) {
	tests := map[string]struct {
		getKubeMasterIP               getKubeMasterIPFromENVFn
		getInClusterConfig            getInClusterConfigFn
		getKubernetesDynamicClientSet getKubeDynamicClientFn
		kubeConfigPath                string
		getConfigFromENV              buildConfigFromFlagsFn
		getKubeConfigPath             getKubeConfigPathFromENVFn
		isErr                         bool
	}{
		"t1": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetOk, "fake-path", fakeBuildConfigFromFlagsOk, fakeGetKubeConfigPathNil, false},
		"t2": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetErr, "fake-path", fakeBuildConfigFromFlagsOk, fakeGetKubeConfigPathOk, true},
		"t3": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetOk, "fake-path", fakeBuildConfigFromFlagsErr, fakeGetKubeConfigPathOk, true},
		"t4": {fakeGetKubeMasterIPOk, fakeInClusterConfigOk, fakeGetDynamicClientSetOk, "", fakeBuildConfigFromFlagsOk, fakeGetKubeConfigPathOk, false},
		"t5": {fakeGetKubeMasterIPOk, fakeInClusterConfigErr, fakeGetDynamicClientSetErr, "", fakeBuildConfigFromFlagsOk, fakeGetKubeConfigPathOk, true},
		"t6": {fakeGetKubeMasterIPNil, fakeInClusterConfigOk, fakeGetDynamicClientSetErr, "", fakeBuildConfigFromFlagsErr, fakeGetKubeConfigPathNil, true},
		"t7": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetOk, "", fakeBuildConfigFromFlagsErr, fakeGetKubeConfigPathNil, true},
		"t8": {fakeGetKubeMasterIPNil, fakeInClusterConfigErr, fakeGetDynamicClientSetErr, "", fakeBuildConfigFromFlagsErr, fakeGetKubeConfigPathNil, true},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				getKubeMasterIPFromENV:   mock.getKubeMasterIP,
				KubeConfigPath:           mock.kubeConfigPath,
				getInClusterConfig:       mock.getInClusterConfig,
				buildConfigFromFlags:     mock.getConfigFromENV,
				getKubeConfigPathFromENV: mock.getKubeConfigPath,
				getKubeDynamicClient:     mock.getKubernetesDynamicClientSet,
			}
			_, err := c.Dynamic()
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error but got '%v'", name, err)
			}
		})
	}
}

func TestConfigForPath(t *testing.T) {
	tests := map[string]struct {
		kubeConfigPath    string
		getConfigFromPath buildConfigFromFlagsFn
		isErr             bool
	}{
		"T1": {"", fakeBuildConfigFromFlagsErr, true},
		"T2": {"fake-path", fakeBuildConfigFromFlagsOk, false},
	}
	for name, mock := range tests {
		name, mock := name, mock // pin It
		t.Run(name, func(t *testing.T) {
			c := &Client{
				KubeConfigPath:       mock.kubeConfigPath,
				buildConfigFromFlags: mock.getConfigFromPath,
			}
			_, err := c.ConfigForPath(mock.kubeConfigPath)
			if mock.isErr && err == nil {
				t.Fatalf("test '%s' failed: expected error actual no error", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("test '%s' failed: expected no error but got '%v'", name, err)
			}
		})
	}
}

func TestInstance(t *testing.T) {
	c := Instance()
	if c == nil {
		t.Fatalf("test failed: expected non nil client instance got nil")
	}
}

func TestWithQPS(t *testing.T) {
	tests := map[string]struct {
		qps      float32
		expected float32
	}{
		"positive value is set":     {qps: 50, expected: 50},
		"fractional value is set":   {qps: 7.5, expected: 7.5},
		"zero value is ignored":     {qps: 0, expected: 0},
		"negative value is ignored": {qps: -1, expected: 0},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c := New(WithQPS(mock.qps))
			if c.qps != mock.expected {
				t.Fatalf("test '%s' failed: expected qps '%v' actual '%v'", name, mock.expected, c.qps)
			}
		})
	}
}

func TestWithBurst(t *testing.T) {
	tests := map[string]struct {
		burst    int
		expected int
	}{
		"positive value is set":     {burst: 100, expected: 100},
		"zero value is ignored":     {burst: 0, expected: 0},
		"negative value is ignored": {burst: -1, expected: 0},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c := New(WithBurst(mock.burst))
			if c.burst != mock.expected {
				t.Fatalf("test '%s' failed: expected burst '%v' actual '%v'", name, mock.expected, c.burst)
			}
		})
	}
}

func TestApplyRateLimits(t *testing.T) {
	tests := map[string]struct {
		qps           float32
		burst         int
		nilConfig     bool
		expectedQPS   float32
		expectedBurst int
	}{
		// Unset overrides must leave the config's zero values untouched so
		// client-go applies its own defaults (QPS 5, Burst 10).
		"unset leaves config defaults": {qps: 0, burst: 0, expectedQPS: 0, expectedBurst: 0},
		// A QPS-only override must still yield a positive burst, since
		// kubernetes.NewForConfig rejects QPS > 0 with Burst <= 0.
		"qps override only defaults burst": {qps: 50, burst: 0, expectedQPS: 50, expectedBurst: rest.DefaultBurst},
		"burst override only":              {qps: 0, burst: 100, expectedQPS: 0, expectedBurst: 100},
		"both overridden":                  {qps: 50, burst: 100, expectedQPS: 50, expectedBurst: 100},
		"nil config is a no-op":            {qps: 50, burst: 100, nilConfig: true},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			c := &Client{qps: mock.qps, burst: mock.burst}
			if mock.nilConfig {
				// Should not panic.
				c.applyRateLimits(nil)
				return
			}
			config := &rest.Config{}
			c.applyRateLimits(config)
			if config.QPS != mock.expectedQPS {
				t.Fatalf("test '%s' failed: expected QPS '%v' actual '%v'", name, mock.expectedQPS, config.QPS)
			}
			if config.Burst != mock.expectedBurst {
				t.Fatalf("test '%s' failed: expected Burst '%v' actual '%v'", name, mock.expectedBurst, config.Burst)
			}
		})
	}
}

// TestClientsetAppliesRateLimits pins the call to applyRateLimits inside
// Clientset. It deliberately uses the real getKubeClientset
// (kubernetes.NewForConfig) so that both halves are covered: that the
// overrides reach the rest.Config, and that the resulting config is one
// client-go will actually accept.
func TestClientsetAppliesRateLimits(t *testing.T) {
	tests := map[string]struct {
		opts          []OptionFn
		expectedQPS   float32
		expectedBurst int
	}{
		// QPS-only is the case client-go rejects when burst is left at zero.
		"qps only":   {opts: []OptionFn{WithQPS(50)}, expectedQPS: 50, expectedBurst: rest.DefaultBurst},
		"burst only": {opts: []OptionFn{WithBurst(100)}, expectedQPS: 0, expectedBurst: 100},
		"both":       {opts: []OptionFn{WithQPS(50), WithBurst(100)}, expectedQPS: 50, expectedBurst: 100},
		"neither":    {opts: nil, expectedQPS: 0, expectedBurst: 0},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			var got *rest.Config
			c := New(mock.opts...)
			c.IsInCluster = true
			c.getInClusterConfig = fakeInClusterConfigWithHost
			// Wrap the default clientset builder rather than replacing it, so
			// a config client-go would reject still surfaces as an error.
			buildClientset := c.getKubeClientset
			c.getKubeClientset = func(config *rest.Config) (*kubernetes.Clientset, error) {
				got = config
				return buildClientset(config)
			}

			if _, err := c.Clientset(); err != nil {
				t.Fatalf("test '%s' failed: unexpected error building clientset: %v", name, err)
			}
			if got.QPS != mock.expectedQPS {
				t.Fatalf("test '%s' failed: expected QPS '%v' actual '%v'", name, mock.expectedQPS, got.QPS)
			}
			if got.Burst != mock.expectedBurst {
				t.Fatalf("test '%s' failed: expected Burst '%v' actual '%v'", name, mock.expectedBurst, got.Burst)
			}
		})
	}
}
