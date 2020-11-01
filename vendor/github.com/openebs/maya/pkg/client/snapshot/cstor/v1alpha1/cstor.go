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

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openebs/maya/pkg/client/generated/cstor-volume-mgmt/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// constants
const (
	VolumeGrpcListenPort = 7777
	ProtocolVersion      = 1
)

//CommandStatus is the response from istgt for control commands
type CommandStatus struct {
	Response string `json:"response"`
}

//CreateSnapshot creates snapshots
func CreateSnapshot(ip, volName, snapName string) (*v1alpha1.VolumeSnapCreateResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Errorf("Unable to dial gRPC server on port %d error : %s", VolumeGrpcListenPort, err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunSnapCommandClient(conn)
	response, err := c.RunVolumeSnapCreateCommand(context.Background(),
		&v1alpha1.VolumeSnapCreateRequest{
			Version:  ProtocolVersion,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		return nil, errors.Errorf("Error when calling RunVolumeSnapCreateCommand: %s", err)
	}

	if response != nil {
		var responseStatus CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return nil, errors.Errorf("Snapshot create failed with error : %v", responseStatus.Response)
		}
	}

	return response, nil
}

//DestroySnapshot destroys snapshots
func DestroySnapshot(ip, volName, snapName string) (*v1alpha1.VolumeSnapDeleteResponse, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, VolumeGrpcListenPort), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Errorf("Unable to dial gRPC server on port error : %s", err)
	}
	defer conn.Close()

	c := v1alpha1.NewRunSnapCommandClient(conn)
	response, err := c.RunVolumeSnapDeleteCommand(context.Background(),
		&v1alpha1.VolumeSnapDeleteRequest{
			Version:  ProtocolVersion,
			Volume:   volName,
			Snapname: snapName,
		})

	if err != nil {
		return nil, errors.Errorf("Error when calling RunVolumeSnapDeleteCommand: %s", err)
	}

	if response != nil {
		var responseStatus CommandStatus
		json.Unmarshal(response.Status, &responseStatus)
		if strings.Contains(responseStatus.Response, "ERR") {
			return nil, errors.Errorf("Snapshot deletion failed with error : %v", responseStatus.Response)
		}
	}
	return response, nil
}
