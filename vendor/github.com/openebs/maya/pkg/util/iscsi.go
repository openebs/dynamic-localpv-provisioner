// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// CheckForIscsi is blocking call for checking status of istgt in cstor-istgt container.
package util

import (
	"time"

	"k8s.io/klog"
)

const (
	IstgtConfPath        = "/usr/local/etc/istgt/istgt.conf"
	IstgtStatusCmd       = "STATUS"
	IstgtRefreshCmd      = "REFRESH"
	IstgtReplicaCmd      = "REPLICA"
	IstgtExecuteQuietCmd = "-q"
	ReplicaStatus        = "Replica status"
	WaitTimeForIscsi     = 3 * time.Second
	// IstgtResizeCmd holds the command to trigger resize
	IstgtResizeCmd = "RESIZE"
	IstgtDRFCmd    = "DRF"
)

func CheckForIscsi(UnixSockVar UnixSock) {
	for {
		_, err := UnixSockVar.SendCommand(IstgtStatusCmd)
		if err != nil {
			time.Sleep(WaitTimeForIscsi)
			klog.Warningf("Waiting for istgt... err : %v", err)
			continue
		}
		break
	}
}
