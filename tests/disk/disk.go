/*
Copyright 2021 The OpenEBS Authors

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

package disk

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	// DiskImageSize is the default file size(1GB) used while creating backing image
	DiskImageSize = 1073741824
)

// Disk has the attributes of a virtual disk which is emulated for integration
// testing.
type Disk struct {
	// Size in bytes
	Size int64
	// The backing image name
	// eg: /tmp/fake123
	imageName string
	// the disk name
	// eg: /dev/loop9002
	DiskPath string
	// mount point if any
	MountPoints []string
}

// NewDisk creates a Disk struct, with a specified size. Also the
// random disk image name is generated. The actual image is generated only when
// we try to attach the disk
func NewDisk(size int64) Disk {
	disk := Disk{
		Size:      size,
		imageName: generateDiskImageName(),
		DiskPath:  "",
	}
	return disk
}

// Generates a random image name for the backing file.
// the file name will be of the format fakeXXX, where X=[0-9]
func generateDiskImageName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomNumber := 100 + rand.Intn(899)
	imageName := "fake" + strconv.Itoa(randomNumber)
	return os.TempDir() + "/" + imageName
}

func (disk *Disk) createDiskImage() error {
	f, err := os.Create(disk.imageName)
	if err != nil {
		return fmt.Errorf("error creating disk image. Error : %v", err)
	}
	err = f.Truncate(disk.Size)
	if err != nil {
		return fmt.Errorf("error truncating disk image. Error : %v", err)
	}

	return nil
}

// CreateLoopDevice creates a loopback device if the disk is not present
func (disk *Disk) CreateLoopDevice() error {
	if disk.DiskPath == "" {
		var err error
		if _, err = os.Stat(disk.imageName); err != nil && os.IsNotExist(err) {
			err = disk.createDiskImage()
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		deviceName := getLoopDevName()
		devicePath := "/dev/" + deviceName
		// create the loop device using losetup
		createLoopDeviceCommand := "losetup " + devicePath + " " + disk.imageName + " --show"
		err = RunCommandWithSudo(createLoopDeviceCommand)
		if err != nil {
			return fmt.Errorf("error creating loop device. Error : %v", err)
		}
		disk.DiskPath = devicePath
	}
	return nil
}

// Generates a random loop device name. The name will be of the
// format loop9XXX, where X=[0-9]. The 4 digit numbering is chosen so that
// we get enough disks to be randomly generated and also it does not clash
// with the existing loop devices present in some systems.
func getLoopDevName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomNumber := 9000 + rand.Intn(999)
	diskName := "loop" + strconv.Itoa(randomNumber)
	return diskName
}

// RunCommandWithSudo runs a command with sudo permissions
func RunCommandWithSudo(cmd string) error {
	return RunCommand("sudo " + cmd)
}

// RunCommand runs a command on the host
func RunCommand(cmd string) error {
	substring := strings.Fields(cmd)
	name := substring[0]
	args := substring[1:]
	stdout, err := exec.Command(name, args...).CombinedOutput() // #nosec G204
	if err != nil {
		return fmt.Errorf("run failed, cmd={%s} error={%v} output={%v}\"", cmd, err, stdout)
	}
	return err
}

func (disk *Disk) CreateFileSystem(fstype string) error {
	createMkfsCommand := "mkfs -t " + fstype + " " + disk.DiskPath
	err := RunCommandWithSudo(createMkfsCommand)
	if err != nil {
		return fmt.Errorf("error creating fileystem on loop device. Error : %v", err)
	}
	return nil
}

func MkdirAll(path string) error {
	createMkdirCommand := "mkdir -p " + path
	err := RunCommandWithSudo(createMkdirCommand)
	if err != nil {
		return fmt.Errorf("error creating hostath directory. Error : %v", err)
	}
	return nil
}

func (disk *Disk) Mount(path string) error {
	createMountCommand := "mount -o rw,pquota " + disk.DiskPath + " " + path
	err := RunCommandWithSudo(createMountCommand)
	if err != nil {
		return fmt.Errorf("error mounting loop device. Error : %v", err)
	}
	disk.MountPoints = append(disk.MountPoints, path)
	return nil
}

func (disk *Disk) Unmount() []error {
	var lastErr []error
	for i := 0; i < len(disk.MountPoints); i++ {
		err := RunCommandWithSudo("umount " + disk.MountPoints[i])
		if err != nil {
			lastErr = append(lastErr, err)
		} else {
			disk.MountPoints = append(disk.MountPoints[:i], disk.MountPoints[i+1:]...)
			i-- // -1 as the slice just got shorter
		}
	}
	return lastErr
}

// DetachAndDeleteDisk detaches the loop device from the backing
// image. Also deletes the backing image and block device file in /dev
func (disk *Disk) DetachAndDeleteDisk() error {
	if disk.DiskPath == "" {
		return fmt.Errorf("no such disk present for deletion")
	}
	if len(disk.MountPoints) > 0 {
		return fmt.Errorf("the disk is still mounted at mountpoint(s): %+v", disk.MountPoints)
	}
	detachLoopCommand := "losetup -d " + disk.DiskPath
	err := RunCommandWithSudo(detachLoopCommand)
	if err != nil {
		return fmt.Errorf("cannot detach loop device. Error : %v", err)
	}
	err = os.Remove(disk.imageName)
	if err != nil {
		return fmt.Errorf("could not delete backing disk image. Error : %v", err)
	}
	deleteLoopDeviceCommand := "rm " + disk.DiskPath
	err = RunCommandWithSudo(deleteLoopDeviceCommand)
	if err != nil {
		return fmt.Errorf("could not delete loop device. Error : %v", err)
	}
	return nil
}

// PrepareDisk prepares the setup necessary for testing xfs hostpath quota
func PrepareDisk(fsType, hostPath string) (Disk, error) {
	physicalDisk := NewDisk(DiskImageSize)

	err := physicalDisk.CreateLoopDevice()
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while creating loop back device with disk %+v", physicalDisk)
	}

	// Make xfs fs on the created loopback device
	err = physicalDisk.CreateFileSystem(fsType)
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while formatting the disk {%+v} with xfs fs", physicalDisk)
	}

	err = MkdirAll(hostPath)
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while making a new directory {%s}", hostPath)
	}

	// Mount the xfs formatted loopback device
	err = physicalDisk.Mount(hostPath)
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while mounting the disk with pquota option {%+v}", physicalDisk)
	}

	return physicalDisk, nil
}

// DestroyDisk performs performs the clean-up task after testing the features
func DestroyDisk(physicalDisk Disk, hostPath string) error {
	var errs string
	// Unmount the disk
	err := physicalDisk.Unmount()
	if len(err) > 0 {
		for _, v := range err {
			errs = errs + v.Error() + ";"
		}
		return errors.Wrapf(errors.New(errs), "while unmounting the disk {%+v}", physicalDisk)
	}

	// Detach and delete the disk
	error := physicalDisk.DetachAndDeleteDisk()
	if err != nil {
		return errors.Wrapf(error, "while detaching and deleting the disk {%+v}", physicalDisk)
	}

	// Deleting the hostpath directory
	error = RunCommandWithSudo("rm -rf " + hostPath)
	if err != nil {
		return errors.Wrapf(error, "while deleting the mountpoint directory {%s}", hostPath)
	}

	return nil
}
