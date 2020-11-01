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
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/klog"
)

//FileOperator operates on files
type FileOperator interface {
	Write(filename string, data []byte, perm os.FileMode) error
	Updatefile(fileName, updateVal, searchString string, perm os.FileMode) error
	GetLineDetails(filename, searchString string) (int, string, error)
	UpdateOrAppendMultipleLines(fileName string, keyUpdateValue map[string]string, perm os.FileMode) error
}

//RealFileOperator is used for writing the actual files without mocking
type RealFileOperator struct{}

// the real file operator for the actual program,
func (r RealFileOperator) Write(filename string, data []byte, perm os.FileMode) error {
	err := ioutil.WriteFile(filename, data, perm)
	if err != nil {
		klog.Errorf("Failed to write file: " + filename)
	}
	return err
}

// GetLineDetails return the line number and line content of matched string in file
func (r RealFileOperator) GetLineDetails(filename, searchString string) (int, string, error) {
	var line string
	var i int
	buffer, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return -1, "", errors.Wrapf(err, "failed to read %s file", filename)
	}
	lines := strings.Split(string(buffer), "\n")
	for i, line = range lines {
		if strings.Contains(line, searchString) {
			return i, line, nil
		}
	}
	return -1, "", nil
}

// Updatefile updates the line number with the given string
func (r RealFileOperator) Updatefile(fileName, updatedVal, searchString string, perm os.FileMode) error {
	buffer, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return errors.Wrapf(err, "failed to read %s file", fileName)
	}
	lines := strings.Split(string(buffer), "\n")
	for index, line := range lines {
		// If searchString found then update and return
		if strings.Contains(line, searchString) {
			lines[index] = updatedVal
			newbuffer := strings.Join(lines, "\n")
			klog.V(4).Infof("content in a file %s\n", lines)
			err = r.Write(fileName, []byte(newbuffer), perm)
			return err
		}
	}
	return errors.Errorf("failed to find %s in file %s", searchString, fileName)
}

// UpdateOrAppendMultipleLines will update or append multiple lines based on the
// given string
func (r RealFileOperator) UpdateOrAppendMultipleLines(fileName string,
	keyUpdateValue map[string]string, perm os.FileMode) error {
	var newLines []string
	var index int
	var line string
	buffer, err := ioutil.ReadFile(filepath.Clean(fileName))
	if err != nil {
		return errors.Wrapf(err, "failed to read %s file", fileName)
	}
	lines := strings.Split(string(buffer), "\n")
	// newLines pointing to lines reference
	newLines = lines
	if len(lines[len(lines)-1]) == 0 {
		// For replacing the NUL in the file
		newLines = lines[:len(lines)-1]
	}

	// TODO: We can split above read file into key value pairs and later we can
	// append with \n and update file
	// TODO: Use regular expresion to replace key value pairs
	// will be doing after current blockers
	for key, updatedValue := range keyUpdateValue {
		found := false
		for index, line = range newLines {
			if strings.HasPrefix(line, key) {
				newLines[index] = updatedValue
				found = true
				break
			}
		}
		// To remove particular line that matched with key
		if found && updatedValue == "" {
			newLines = append(newLines[:index], newLines[index+1:]...)
			continue
		}
		if found == false {
			newLines = append(newLines, updatedValue)
		}
	}
	newbuffer := strings.Join(newLines, "\n")
	klog.V(4).Infof("content in a file %s\n", newLines)
	err = r.Write(fileName, []byte(newbuffer), perm)
	return err
}

//TestFileOperator is used as a dummy FileOperator
type TestFileOperator struct{}

//Write is to mock write operation for FileOperator interface
func (r TestFileOperator) Write(filename string, data []byte, perm os.FileMode) error {
	return nil
}

//Updatefile is to mock Updatefile operation for FileOperator interface
func (r TestFileOperator) Updatefile(fileName, updateStorageVal, searchString string, perm os.FileMode) error {
	return nil
}

//GetLineDetails is to mock operation for FileOperator interface
func (r TestFileOperator) GetLineDetails(filename, searchString string) (int, string, error) {
	return -1, "", nil
}

// UpdateOrAppendMultipleLines is to mock operation for FileOperator interface
func (r TestFileOperator) UpdateOrAppendMultipleLines(
	fileName string,
	keyUpdateValue map[string]string,
	perm os.FileMode) error {
	return nil
}
