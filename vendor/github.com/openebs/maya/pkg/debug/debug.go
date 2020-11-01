// +build debug

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

package debug

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"k8s.io/klog"
)

// LogBuildDetails logs the build details when the cspc-operator starts.
func LogBuildDetails() {
	klog.Info("This is a DEBUG build and should not be used in production")
}

// StartInjectionServer is a wrapper that starts a REST server that is used to inject errors in the debug build.
func StartInjectionServer() {
	go StartServer()
}

// StartServer starts the injection server.
func StartServer() {
	klog.Info("Starting Error Injection API Server...")
	RegisterRoutes()
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Println("Injection Server Listening on 8080 ...")
	server.ListenAndServe()
}

// RegisterRoutes registers routes for the injection server.
func RegisterRoutes() {
	http.HandleFunc("/", index)
	http.HandleFunc("/inject", inject)
}

// index is the handler of '/' route.
func index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Error Injection API Server Is Running!"))
}

// inject is the handler of 'inject' route
func inject(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		errorInjectionConfigInBytes, err := json.Marshal(EI)
		if err != nil {
			klog.Errorf("Failed to marshal error injection config: %s", err.Error())
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("Failed to marshal error injection config: %s", err.Error())))
			return
		}
		w.WriteHeader(200)
		w.Write(errorInjectionConfigInBytes)

	case "POST":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "Incorrect data or format in body")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("Incorrect data or format in body: %s", err.Error())))
			return
		}
		json.Unmarshal(reqBody, EI)

	default:
		fmt.Fprintf(w, "Method Not Allowed")
		w.WriteHeader(405)
		w.Write([]byte(fmt.Sprint("Method Not Allowed")))

	}
}

// CSPCs

// IsCSPCDeleteCollectionErrorInjected returns true if error is injected for
// CSPC delete collection.
func (ei *ErrorInjection) IsCSPCDeleteCollectionErrorInjected() bool {
	if ei.CSPCError.CRUDErrorInjection.InjectDeleteCollectionError == Inject {
		return true
	}
	return false
}

// IsCSPCDeleteErrorInjected returns true if error is injected for
// CSPC delete.
func (ei *ErrorInjection) IsCSPCDeleteErrorInjected() bool {
	if ei.CSPCError.CRUDErrorInjection.InjectDeleteError == Inject {
		return true
	}
	return false
}

// IsCSPCListErrorInjected returns true if error is injected for
// CSPC list.
func (ei *ErrorInjection) IsCSPCListErrorInjected() bool {
	if ei.CSPCError.CRUDErrorInjection.InjectListError == Inject {
		return true
	}
	return false
}

// IsCSPCGetErrorInjected returns true if error is injected for
// CSPC get.
func (ei *ErrorInjection) IsCSPCGetErrorInjected() bool {
	if ei.CSPCError.CRUDErrorInjection.InjectGetError == Inject {
		return true
	}
	return false
}

// IsCSPCCreateErrorInjected returns true if error is injected for
// CSPC create.
func (ei *ErrorInjection) IsCSPCCreateErrorInjected() bool {
	if ei.CSPCError.CRUDErrorInjection.InjectCreateError == Inject {
		return true
	}
	return false
}

// IsCSPCUpdateErrorInjected returns true if error is injected for
// CSPC update.
func (ei *ErrorInjection) IsCSPCUpdateErrorInjected() bool {
	if ei.CSPCError.CRUDErrorInjection.InjectUpdateError == Inject {
		return true
	}
	return false
}

// IsCSPCPatchErrorInjected returns true if error is injected for
// CSPC patch.
func (ei *ErrorInjection) IsCSPCPatchErrorInjected() bool {
	if ei.CSPCError.CRUDErrorInjection.InjectPatchError == Inject {
		return true
	}
	return false
}

// IsCSPIDeleteCollectionErrorInjected returns true if error is injected for
// CSPI delete collection.
func (ei *ErrorInjection) IsCSPIDeleteCollectionErrorInjected() bool {
	if ei.CSPIError.CRUDErrorInjection.InjectDeleteCollectionError == Inject {
		return true
	}
	return false
}

// IsCSPIDeleteErrorInjected returns true if error is injected for
// CSPI delete.
func (ei *ErrorInjection) IsCSPIDeleteErrorInjected() bool {
	if ei.CSPIError.CRUDErrorInjection.InjectDeleteError == Inject {
		return true
	}
	return false
}

// IsCSPIListErrorInjected returns true if error is injected for
// CSPI list.
func (ei *ErrorInjection) IsCSPIListErrorInjected() bool {
	if ei.CSPIError.CRUDErrorInjection.InjectListError == Inject {
		return true
	}
	return false
}

// IsCSPIGetErrorInjected returns true if error is injected for
// CSPI get.
func (ei *ErrorInjection) IsCSPIGetErrorInjected() bool {
	if ei.CSPIError.CRUDErrorInjection.InjectGetError == Inject {
		return true
	}
	return false
}

// IsCSPICreateErrorInjected returns true if error is injected for
// CSPI create.
func (ei *ErrorInjection) IsCSPICreateErrorInjected() bool {
	if ei.CSPIError.CRUDErrorInjection.InjectCreateError == Inject {
		return true
	}
	return false
}

// IsCSPIUpdateErrorInjected returns true if error is injected for
// CSPI update.
func (ei *ErrorInjection) IsCSPIUpdateErrorInjected() bool {
	if ei.CSPIError.CRUDErrorInjection.InjectUpdateError == Inject {
		return true
	}
	return false
}

// IsCSPIPatchErrorInjected returns true if error is injected for
// CSPI patch.
func (ei *ErrorInjection) IsCSPIPatchErrorInjected() bool {
	if ei.CSPIError.CRUDErrorInjection.InjectPatchError == Inject {
		return true
	}
	return false
}

// IsDeploymentDeleteCollectionErrorInjected returns true if error is injected for
// Deployment delete collection.
func (ei *ErrorInjection) IsDeploymentDeleteCollectionErrorInjected() bool {
	if ei.DeploymentError.CRUDErrorInjection.InjectDeleteCollectionError == Inject {
		return true
	}
	return false
}

// IsDeploymentDeleteErrorInjected returns true if error is injected for
// Deployment delete.
func (ei *ErrorInjection) IsDeploymentDeleteErrorInjected() bool {
	if ei.DeploymentError.CRUDErrorInjection.InjectDeleteError == Inject {
		return true
	}
	return false
}

// IsDeploymentListErrorInjected returns true if error is injected for
// Deployment list.
func (ei *ErrorInjection) IsDeploymentListErrorInjected() bool {
	if ei.DeploymentError.CRUDErrorInjection.InjectListError == Inject {
		return true
	}
	return false
}

// IsDeploymentGetErrorInjected returns true if error is injected for
// Deployment get.
func (ei *ErrorInjection) IsDeploymentGetErrorInjected() bool {
	if ei.DeploymentError.CRUDErrorInjection.InjectGetError == Inject {
		return true
	}
	return false
}

// IsDeploymentCreateErrorInjected returns true if error is injected for
// Deployment create.
func (ei *ErrorInjection) IsDeploymentCreateErrorInjected() bool {
	if ei.DeploymentError.CRUDErrorInjection.InjectCreateError == Inject {
		return true
	}
	return false
}

// IsDeploymentUpdateErrorInjected returns true if error is injected for
// Deployment update.
func (ei *ErrorInjection) IsDeploymentUpdateErrorInjected() bool {
	if ei.DeploymentError.CRUDErrorInjection.InjectUpdateError == Inject {
		return true
	}
	return false
}

// IsDeploymentPatchErrorInjected returns true if error is injected for
// Deployment patch.
func (ei *ErrorInjection) IsDeploymentPatchErrorInjected() bool {
	if ei.DeploymentError.CRUDErrorInjection.InjectPatchError == Inject {
		return true
	}
	return false
}

// IsZFSGetErrorInjected returns true if error is injected for ZFS get command
func (ei *ErrorInjection) IsZFSGetErrorInjected() bool {
	if ei.ZFSError.CRUDErrorInjection.InjectGetError == Inject {
		return true
	}
	return false
}

// IsZFSDeleteErrorInjected returns true if error is injected for ZFS delete command
func (ei *ErrorInjection) IsZFSDeleteErrorInjected() bool {
	if ei.ZFSError.CRUDErrorInjection.InjectDeleteError == Inject {
		return true
	}
	return false
}

// IsZFSCreateErrorInjected returns true if error is injected for ZFS create command
func (ei *ErrorInjection) IsZFSCreateErrorInjected() bool {
	if ei.ZFSError.CRUDErrorInjection.InjectCreateError == Inject {
		return true
	}
	return false
}

// IsCVRCreateErrorInjected returns true if error is injected for CVR create
// command
func (ei *ErrorInjection) IsCVRCreateErrorInjected() bool {
	return false
}

// IsCVRDeleteErrorInjected returns true if error is injected for CVR delete
// command
func (ei *ErrorInjection) IsCVRDeleteErrorInjected() bool {
	return false
}

// IsCVRGetErrorInjected returns true if error is injected for CVR get command
func (ei *ErrorInjection) IsCVRGetErrorInjected() bool {
	return false
}

// IsCVRUpdateErrorInjected returns true if error is injected for CVR update command
func (ei *ErrorInjection) IsCVRUpdateErrorInjected() bool {
	return false
}
