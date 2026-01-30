package persistentvolume

import (
	"testing"
)

func TestListBuilderWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePVs  []string
		expectedPVLen int
	}{
		"PV set 1":  {[]string{}, 0},
		"PV set 2":  {[]string{"pv1"}, 1},
		"PV set 3":  {[]string{"pv1", "pv2"}, 2},
		"PV set 4":  {[]string{"pv1", "pv2", "pv3"}, 3},
		"PV set 5":  {[]string{"pv1", "pv2", "pv3", "pv4"}, 4},
		"PV set 6":  {[]string{"pv1", "pv2", "pv3", "pv4", "pv5"}, 5},
		"PV set 7":  {[]string{"pv1", "pv2", "pv3", "pv4", "pv5", "pv6"}, 6},
		"PV set 8":  {[]string{"pv1", "pv2", "pv3", "pv4", "pv5", "pv6", "pv7"}, 7},
		"PV set 9":  {[]string{"pv1", "pv2", "pv3", "pv4", "pv5", "pv6", "pv7", "pv8"}, 8},
		"PV set 10": {[]string{"pv1", "pv2", "pv3", "pv4", "pv5", "pv6", "pv7", "pv8", "pv9"}, 9},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIObjects(fakeAPIPVList(mock.availablePVs))
			if mock.expectedPVLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVLen, len(b.list.items))
			}
		})
	}
}

func TestListBuilderWithAPIObjects(t *testing.T) {
	tests := map[string]struct {
		availablePVs  []string
		expectedPVLen int
		expectedErr   bool
	}{
		"PV set 1": {[]string{}, 0, true},
		"PV set 2": {[]string{"pv1"}, 1, false},
		"PV set 3": {[]string{"pv1", "pv2"}, 2, false},
		"PV set 4": {[]string{"pv1", "pv2", "pv3"}, 3, false},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b, err := ListBuilderForAPIObjects(fakeAPIPVList(mock.availablePVs)).APIList()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if !mock.expectedErr && mock.expectedPVLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.availablePVs, len(b.Items))
			}
		})
	}
}

func TestListBuilderAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePVs  []string
		expectedPVLen int
		expectedErr   bool
	}{
		"PV set 1": {[]string{}, 0, true},
		"PV set 2": {[]string{"pv1"}, 1, false},
		"PV set 3": {[]string{"pv1", "pv2"}, 2, false},
		"PV set 4": {[]string{"pv1", "pv2", "pv3"}, 3, false},
		"PV set 5": {[]string{"pv1", "pv2", "pv3", "pv4"}, 4, false},
	}
	for name, mock := range tests {

		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b, err := ListBuilderForAPIObjects(fakeAPIPVList(mock.availablePVs)).APIList()
			if mock.expectedErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if err == nil && mock.expectedPVLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVLen, len(b.Items))
			}
		})
	}
}
