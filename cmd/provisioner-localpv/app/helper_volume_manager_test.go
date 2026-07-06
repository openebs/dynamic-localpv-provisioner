package app

import (
	"testing"
)

func TestNewLocalVolumeRequest(t *testing.T) {
	tests := map[string]struct {
		pOpts      *HelperPodOptions
		wantFsMode string
	}{
		"honors configured FilePermissions mode": {
			pOpts: &HelperPodOptions{
				name:   "pvc-1",
				path:   "/var/openebs/local/pvc-1",
				fsMode: "0750",
			},
			wantFsMode: "0750",
		},
		"honors restrictive FilePermissions mode": {
			pOpts: &HelperPodOptions{
				name:   "pvc-2",
				path:   "/var/openebs/local/pvc-2",
				fsMode: "0700",
			},
			wantFsMode: "0700",
		},
		"empty mode is left empty for the volume manager default": {
			pOpts: &HelperPodOptions{
				name:   "pvc-3",
				path:   "/var/openebs/local/pvc-3",
				fsMode: "",
			},
			wantFsMode: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := newLocalVolumeRequest(tc.pOpts)
			if req.FsMode != tc.wantFsMode {
				t.Errorf("newLocalVolumeRequest() FsMode = %q, want %q", req.FsMode, tc.wantFsMode)
			}
			if req.Name != tc.pOpts.name {
				t.Errorf("newLocalVolumeRequest() Name = %q, want %q", req.Name, tc.pOpts.name)
			}
			if req.Path != tc.pOpts.path {
				t.Errorf("newLocalVolumeRequest() Path = %q, want %q", req.Path, tc.pOpts.path)
			}
		})
	}
}
