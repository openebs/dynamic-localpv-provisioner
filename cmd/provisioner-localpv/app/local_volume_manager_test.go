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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// pathUnderRoot sits directly under "/", which ExtractPaths rejects, so an
// operation using it returns before running any command. The mutex is taken
// first, which is what these tests exercise.
const pathUnderRoot = "/rootvolume"

// The local volume manager used to be a pointer that only NewProvisioner ever
// set, so a Provisioner built any other way dereferenced nil on its first local
// operation. Holding the manager by value makes the zero Provisioner usable.
func TestZeroValueProvisionerRunsLocalOperations(t *testing.T) {
	p := &Provisioner{}
	opts := &HelperPodOptions{name: "pvc-1", path: pathUnderRoot}

	// Both calls are expected to fail path validation. The point is that they
	// return an error at all instead of panicking.
	if err := p.createVolumeLocally(context.Background(), opts, false); err == nil {
		t.Error("createVolumeLocally: expected the path to be rejected")
	}
	if err := p.deleteVolumeLocally(context.Background(), opts); err == nil {
		t.Error("deleteVolumeLocally: expected the path to be rejected")
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
			Path: pathUnderRoot,
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
					path: pathUnderRoot,
				}, false)
			},
		},
		{
			name: "delete",
			call: func(p *Provisioner) error {
				return p.deleteVolumeLocally(context.Background(), &HelperPodOptions{
					name: "pvc-1",
					path: pathUnderRoot,
				})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &Provisioner{}
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

// TestConcurrentLocalVolumeOperationsSerialize is the regression test for issue
// #359: because every request built its own manager, and therefore its own
// mutex, concurrent provisioning on one node ran CreateVolume and ApplyQuota at
// the same time. The XFS project-ID allocation reads the highest existing ID and
// adds one, so racing callers picked the same ID and several volumes ended up
// sharing a single project.
//
// The local volume manager shells out for its real work, so the test puts stub
// commands ahead of it on PATH. That measures the actual critical section -
// including the part inside exec - without touching the host filesystem.
func TestConcurrentLocalVolumeOperationsSerialize(t *testing.T) {
	const workers = 8

	binDir := t.TempDir()
	callLog := filepath.Join(t.TempDir(), "calls.log")

	// Each stub brackets a short sleep with a marker, so two operations that
	// overlap show up as two "enter" lines with no "exit" between them.
	stub := "#!/bin/sh\n" +
		"echo enter >> " + callLog + "\n" +
		"sleep 0.05\n" +
		"echo exit >> " + callLog + "\n"
	for _, name := range []string{"mkdir", "sh"} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(stub), 0o755); err != nil {
			t.Fatalf("writing the %s stub: %v", name, err)
		}
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	p := &Provisioner{}

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			opts := &HelperPodOptions{
				name: fmt.Sprintf("pvc-%d", i),
				path: fmt.Sprintf("/var/openebs/local/pvc-%d", i),
			}
			<-start
			// Creates and deletes take the same mutex, so they must serialize
			// against each other as well as among themselves.
			var err error
			if i%2 == 0 {
				err = p.createVolumeLocally(context.Background(), opts, false)
			} else {
				err = p.deleteVolumeLocally(context.Background(), opts)
			}
			if err != nil {
				t.Errorf("local operation %d: %v", i, err)
			}
		}(i)
	}
	close(start)
	wg.Wait()

	raw, err := os.ReadFile(callLog)
	if err != nil {
		t.Fatalf("no stub command ever ran, so nothing was measured: %v", err)
	}

	depth, maxDepth, entered := 0, 0, 0
	for _, marker := range strings.Fields(string(raw)) {
		switch marker {
		case "enter":
			entered++
			if depth++; depth > maxDepth {
				maxDepth = depth
			}
		case "exit":
			depth--
		}
	}

	if entered != workers {
		t.Fatalf("expected %d stub invocations, got %d: the local volume manager "+
			"no longer shells out, so this test is measuring nothing", workers, entered)
	}
	if maxDepth != 1 {
		t.Errorf("local volume operations overlapped: %d ran at once, want 1", maxDepth)
	}
}
