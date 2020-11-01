/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PreUpgradeAction is of string type
// Once the preupgrade checks are done, one of the below actions will be taken
type PreUpgradeAction string

const (
	// DisableReconciler is an action to add the label DisableReconciler
	// Note: Not used as part of 1.1 preupgrade
	DisableReconciler PreUpgradeAction = "DisableReconciler"

	// Continue the preupgrade tasks
	Continue PreUpgradeAction = "Continue"

	// Abort running preupgrade tasks
	Abort PreUpgradeAction = "Abort"
)

// 1.1 preupgrade need to contine only if SPC doesn't have SPCFinalizer on it
func (Spc *SPC) getPreUpgradeAction() (PreUpgradeAction, error) {
	if Spc.HasFinalizer(SPCFinalizer) {
		return Abort, nil
	}
	if !Spc.Object.DeletionTimestamp.IsZero() {
		return Abort, nil
	}
	return Continue, nil
}

type preUpgradeFn func(*SPC) error

// Table of functions that need to be executed for particular PreUpgradeAction
var performPreUpgradeFn = map[PreUpgradeAction]preUpgradeFn{
	DisableReconciler: noop,
	Continue:          performPreUpgrade,
	Abort:             noop,
}

func noop(*SPC) error {
	return nil
}

func (Spc *SPC) performPreUpgradeOnAssociatedBDCs() error {
	return Spc.addSPCFinalizerOnAssociatedBDCs()
}

// As part of 1.1 preupgrade,
// Set finalizers on all BDCs of this SPC
// set finalizer on SPC
func performPreUpgrade(Spc *SPC) error {
	err := Spc.performPreUpgradeOnAssociatedBDCs()
	if err != nil {
		return err
	}
	_, err = Spc.AddFinalizer(SPCFinalizer)
	return err
}

// Perform 1.1 preupgrade tasks for given SPC
// get preupgrade fn to execute
// Execute the fn
func (Spc *SPC) preUpgrade() error {
	res, err := Spc.getPreUpgradeAction()
	if err != nil {
		return err
	}
	err = performPreUpgradeFn[res](Spc)
	return err
}

// PreUpgrade performs 1.1 preupgrade tasks for all SPCs
func PreUpgrade() error {
	spcList, _ := NewKubeClient().List(metav1.ListOptions{})
	for _, obj := range spcList.Items {
		obj := obj
		Spc := SPC{&obj}
		err := Spc.preUpgrade()
		if err != nil {
			return err
		}
	}
	return nil
}
