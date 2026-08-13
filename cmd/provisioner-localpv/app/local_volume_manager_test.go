/*
Copyright 2026 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"testing"
	"time"
)

// rootPath is rejected by ExtractPaths, so the operation returns before running
// any command. The mutex is taken first, which is what these tests exercise.
const rootPath = "/rootvolume"

func TestNewProvisionerSharesLocalVolumeManager(t *testing.T) {
	p := &Provisioner{}
	p.localVolumeManager = NewLocalVolumeManager()

	if p.localVolumeManager == nil {
		t.Fatal("expected the provisioner to hold a volume manager")
	}
}

func TestCreateVolumeWaitsForManagerLock(t *testing.T) {
	vm := NewLocalVolumeManager()
	vm.mu.Lock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = vm.CreateVolume(context.Background(), &VolumeRequest{
			Name: "pvc-1",
			Path: rootPath,
		}, false)
	}()

	select {
	case <-done:
		vm.mu.Unlock()
		t.Fatal("CreateVolume returned while the manager's lock was held")
	case <-time.After(200 * time.Millisecond):
	}

	vm.mu.Unlock()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("CreateVolume did not return after the lock was released")
	}
}

// A provisioner used to build a new manager, and therefore a new mutex, for
// every request, so local operations were never serialized against each other.
// Both helpers must go through the manager held by the provisioner.
func TestLocalHelpersUseTheProvisionerManager(t *testing.T) {
	tests := []struct {
		name string
		call func(p *Provisioner) error
	}{
		{
			name: "create",
			call: func(p *Provisioner) error {
				return p.createVolumeLocally(context.Background(), &HelperPodOptions{
					name: "pvc-1",
					path: rootPath,
				}, false)
			},
		},
		{
			name: "delete",
			call: func(p *Provisioner) error {
				return p.deleteVolumeLocally(context.Background(), &HelperPodOptions{
					name: "pvc-1",
					path: rootPath,
				})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &Provisioner{localVolumeManager: NewLocalVolumeManager()}
			p.localVolumeManager.mu.Lock()

			done := make(chan struct{})
			go func() {
				defer close(done)
				_ = tc.call(p)
			}()

			select {
			case <-done:
				p.localVolumeManager.mu.Unlock()
				t.Fatalf("%sVolumeLocally did not wait for the provisioner's volume manager", tc.name)
			case <-time.After(200 * time.Millisecond):
			}

			p.localVolumeManager.mu.Unlock()

			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("call did not return after the lock was released")
			}
		})
	}
}
