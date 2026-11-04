package app

import (
	"context"
	"errors"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	pvController "sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"
)

// TestProvisionSelectedNodeLookup verifies the ProvisioningState returned when
// the selected node cannot be fetched. A node that no longer exists (for
// example, scaled down after the scheduler picked it but before the volume was
// created) must return ProvisioningReschedule so the library removes the
// volume.kubernetes.io/selected-node annotation and the scheduler can pick a
// different node. Any other lookup error stays ProvisioningFinished so the
// claim is retried against the same node.
func TestProvisionSelectedNodeLookup(t *testing.T) {
	opts := pvController.ProvisionOptions{
		PVC: &v1.PersistentVolumeClaim{
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			},
		},
		SelectedNodeName: "deleted-node",
	}

	t.Run("selected node not found reschedules", func(t *testing.T) {
		p := Provisioner{kubeClient: fake.NewSimpleClientset()}

		pv, state, err := p.Provision(context.TODO(), opts)

		if err == nil {
			t.Fatal("expected an error for a missing node, got nil")
		}
		if pv != nil {
			t.Fatalf("expected no PV, got %v", pv)
		}
		if state != pvController.ProvisioningReschedule {
			t.Fatalf("expected ProvisioningReschedule, got %v (err: %v)", state, err)
		}
	})

	t.Run("other node lookup errors do not reschedule", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		client.PrependReactor("get", "nodes", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("apiserver unavailable")
		})
		p := Provisioner{kubeClient: client}

		_, state, err := p.Provision(context.TODO(), opts)

		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if state != pvController.ProvisioningFinished {
			t.Fatalf("expected ProvisioningFinished, got %v (err: %v)", state, err)
		}
	})
}
