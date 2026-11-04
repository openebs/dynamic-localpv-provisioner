/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This code was taken from https://github.com/rancher/local-path-provisioner
and modified to work with the configuration options used by OpenEBS
*/

package app

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestConvertToK(t *testing.T) {
	type args struct {
		limit      string
		pvcStorage int64
	}
	tests := map[string]struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		"Missing limit grace": {
			args: args{
				limit:      "",
				pvcStorage: 5000000000,
			},
			want:    "0k",
			wantErr: false,
		},
		"Present limit with grace": {
			args: args{
				limit:      "0%",
				pvcStorage: 5000,
			},
			want:    "5k",
			wantErr: false,
		},
		"Present limit grace exceeding 100%": {
			args: args{
				limit:      "200%",
				pvcStorage: 5000000,
			},
			want:    "9766k",
			wantErr: false,
		},
		"Present limit grace with decimal%": {
			args: args{
				limit:      ".5%",
				pvcStorage: 1000,
			},
			want:    "1k", // the final result of limit can't be a float
			wantErr: false,
		},
		"Present limit grace with invalid pattern": {
			args: args{
				limit:      "10",
				pvcStorage: 10000,
			},
			want:    "",
			wantErr: true,
		},
		"Present limit grace with only %": {
			args: args{
				limit:      "%",
				pvcStorage: 10000,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToK(tt.args.limit, tt.args.pvcStorage)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToK() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertToK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLaunchPodUsesLogsAsTerminationMessageOnError(t *testing.T) {
	p := &Provisioner{
		kubeClient:  fake.NewSimpleClientset(),
		namespace:   "openebs",
		helperImage: "openebs/linux-utils",
	}
	config := podConfig{
		pOpts: &HelperPodOptions{
			name:               "pvc-1",
			cmdsForPath:        []string{"mkdir", "-m", "0777", "-p", "/data/pvc-1"},
			nodeAffinityLabels: map[string]string{k8sNodeLabelKeyHostname: "node-one"},
			serviceAccountName: "openebs-localpv-provisioner",
			imagePullPolicy:    corev1.PullIfNotPresent,
		},
		parentDir: "/var/openebs/local",
		volumeDir: "pvc-1",
		podName:   "init",
	}

	hPod, err := p.launchPod(context.Background(), config)
	if err != nil {
		t.Fatalf("launchPod() error = %v", err)
	}
	if got := hPod.Spec.Containers[0].TerminationMessagePolicy; got != corev1.TerminationMessageFallbackToLogsOnError {
		t.Errorf("TerminationMessagePolicy = %q, want %q", got, corev1.TerminationMessageFallbackToLogsOnError)
	}
}

func TestExitPod(t *testing.T) {
	const podName = "init-pvc-1"

	tests := map[string]struct {
		status corev1.PodStatus
		// wantErr lists the substrings expected in the returned error;
		// a nil wantErr expects no error.
		wantErr []string
	}{
		"helper pod succeeds": {
			status: corev1.PodStatus{Phase: corev1.PodSucceeded},
		},
		"helper pod fails with error output": {
			status: corev1.PodStatus{
				Phase: corev1.PodFailed,
				ContainerStatuses: []corev1.ContainerStatus{{
					Name: "local-path-init",
					State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
						ExitCode: 1,
						Reason:   "Error",
						Message:  "mkdir: can't create directory '/data/pvc-1': Read-only file system\n",
					}},
				}},
			},
			wantErr: []string{
				"pod init-pvc-1 has failed",
				"phase Failed",
				"container local-path-init terminated with exit code 1: Error: " +
					"mkdir: can't create directory '/data/pvc-1': Read-only file system",
			},
		},
		"helper pod is evicted": {
			status: corev1.PodStatus{
				Phase:   corev1.PodFailed,
				Reason:  "Evicted",
				Message: "The node was low on resource: ephemeral-storage.",
			},
			wantErr: []string{
				"pod init-pvc-1 has failed",
				"pod: Evicted: The node was low on resource: ephemeral-storage.",
			},
		},
		"helper pod cannot be scheduled": {
			status: corev1.PodStatus{
				Phase: corev1.PodPending,
				Conditions: []corev1.PodCondition{{
					Type:    corev1.PodScheduled,
					Status:  corev1.ConditionFalse,
					Reason:  corev1.PodReasonUnschedulable,
					Message: "0/1 nodes are available: 1 node(s) had untolerated taint.",
				}},
			},
			wantErr: []string{
				"create process timeout after 1 seconds, pod init-pvc-1 status: phase Pending",
				"not scheduled: Unschedulable: 0/1 nodes are available: 1 node(s) had untolerated taint.",
			},
		},
		"helper pod image cannot be pulled": {
			status: corev1.PodStatus{
				Phase: corev1.PodPending,
				ContainerStatuses: []corev1.ContainerStatus{{
					Name: "local-path-init",
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{
						Reason:  "ImagePullBackOff",
						Message: "Back-off pulling image \"openebs/linux-utils\"",
					}},
				}},
			},
			wantErr: []string{
				"create process timeout after 1 seconds",
				"container local-path-init waiting: ImagePullBackOff: Back-off pulling image \"openebs/linux-utils\"",
			},
		},
	}

	originalTimeout := CmdTimeoutCounts
	CmdTimeoutCounts = 1
	t.Cleanup(func() { CmdTimeoutCounts = originalTimeout })

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "openebs"},
				Spec:       corev1.PodSpec{RestartPolicy: corev1.RestartPolicyNever},
				Status:     tc.status,
			})
			p := &Provisioner{kubeClient: client, namespace: "openebs"}

			err := p.exitPod(context.Background(), podName)
			if tc.wantErr == nil && err != nil {
				t.Fatalf("exitPod() error = %v, want nil", err)
			}
			if tc.wantErr != nil && err == nil {
				t.Fatalf("exitPod() error = nil, want an error containing %q", tc.wantErr)
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("exitPod() error = %q, want it to contain %q", err, want)
				}
			}

			_, getErr := client.CoreV1().Pods("openebs").Get(context.Background(), podName, metav1.GetOptions{})
			if !k8serror.IsNotFound(getErr) {
				t.Errorf("expected the helper pod to be deleted, got error %v", getErr)
			}
		})
	}
}

func TestFormatReason(t *testing.T) {
	longOutput := strings.Repeat("x", maxHelperPodMessageLen+10)

	tests := map[string]struct {
		reason  string
		message string
		want    string
	}{
		"no reason or message": {
			want: "prefix",
		},
		"reason only": {
			reason: "Error",
			want:   "prefix: Error",
		},
		"multi-line message is collapsed into one line": {
			reason:  "Error",
			message: "\n+ set -e\n  mkdir: can't create directory\t'/data/pvc-1'\n\n",
			want:    "prefix: Error: + set -e mkdir: can't create directory '/data/pvc-1'",
		},
		"long message is capped": {
			message: longOutput,
			want:    "prefix: " + longOutput[:maxHelperPodMessageLen] + "...",
		},
		"long message is not capped inside a multi-byte character": {
			message: strings.Repeat("x", maxHelperPodMessageLen-1) + "é",
			want:    "prefix: " + strings.Repeat("x", maxHelperPodMessageLen-1) + "...",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := formatReason("prefix", tc.reason, tc.message); got != tc.want {
				t.Errorf("formatReason() = %q, want %q", got, tc.want)
			}
		})
	}
}
