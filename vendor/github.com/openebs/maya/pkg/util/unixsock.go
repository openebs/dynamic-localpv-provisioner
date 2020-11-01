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

package util

import (
	"io"
	"net"
	"strings"

	"k8s.io/klog"
)

// IstgtUctlUnxpath is the storage path for the UNIX domain socket from istgt
const (
	IstgtUctlUnxpath = "/var/run/istgt_ctl_sock"
	EndOfLine        = "\r\n"
	IstgtHeader      = "iSCSI Target Controller version"
)

//IsResponseEOD will detect if the data coming from UNIX pipe is completely received
func IsResponseEOD(resp []string, cmd string) bool {
	return (len(resp) != 0 && !strings.HasPrefix(resp[len(resp)-1], IstgtHeader) &&
		!strings.HasPrefix(resp[len(resp)-1], cmd))
}

//Reader reads the response from unix domain socket
func Reader(r io.Reader, cmd string) []string {
	resp := []string{}
	//collect bytes into fulllines buffer till the end of line character is reached
	fulllines := []byte{}
	for {
		buf := make([]byte, 1024)
		n, err := r.Read(buf[:])
		if n > 0 {
			klog.Infof("Client got: %s", string(buf[0:n]))
			fulllines = append(fulllines, buf[0:n]...)
			if strings.HasSuffix(string(fulllines), EndOfLine) {
				lines := strings.Split(string(fulllines), EndOfLine)
				for _, line := range lines {
					if len(line) != 0 {
						klog.Infof("Appending line to resp: %s", line)
						resp = append(resp, line+EndOfLine)
					}
				}
				//clear the fulllines buffer once the response lines are appended to the response
				fulllines = nil
			}
			if IsResponseEOD(resp, cmd) {
				klog.Infof("Breaking out of loop for line: %s", resp[len(resp)-1])
				break
			}
		}
		if err != nil {
			klog.Errorf("Read error : %v", err)
			break
		}
		buf = nil
	}
	klog.Infof("response : %v", resp)
	return resp
}

//Writer writes a command to unix domain socket
func Writer(w io.Writer, msg string) error {
	_, err := w.Write([]byte(msg))
	if err != nil {
		klog.Fatalf("Write error: %s", err)
	} else {
		klog.Infof("Client sent: %s", msg)
	}
	return err
}

//UnixSock operates on unix domain sockets
type UnixSock interface {
	SendCommand(cmd string) ([]string, error)
}

//RealUnixSock is used for sending data through real unix domain sockets
type RealUnixSock struct{}

//SendCommand for the real unix sock for the actual program,
func (r RealUnixSock) SendCommand(cmd string) ([]string, error) {
	c, err := net.Dial("unix", IstgtUctlUnxpath)
	if err != nil {
		klog.Fatal("Dial error", err)
	}
	err = Writer(c, cmd+EndOfLine)
	if err != nil {
		c.Close()
		return nil, err
	}
	resp := Reader(c, cmd)
	c.Close()
	return resp, err
}

//TestUnixSock is used as a dummy UnixSock
type TestUnixSock struct{}

//SendCommand for the real unix sock for the actual program,
func (r TestUnixSock) SendCommand(cmd string) ([]string, error) {
	return nil, nil
}
