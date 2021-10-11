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
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	// DiskImageSize is the default file size(1GB) used while creating backing image
	DiskImageSize       = 1073741824
	DiskImageNamePrefix = "openebs-disk-xfs_quota"
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
		imageName: "",
		DiskPath:  "",
	}
	return disk
}

func (disk *Disk) createDiskImage(imgDir string) error {
	f, err := ioutil.TempFile(imgDir, DiskImageNamePrefix+"-*.img")
	if err != nil {
		return fmt.Errorf("error creating disk image. Error : %v", err)
	}
	disk.imageName = f.Name()
	err = f.Truncate(disk.Size)
	if err != nil {
		return fmt.Errorf("error truncating disk image. Error : %v", err)
	}

	return nil
}

// CreateLoopDevice creates a loop device if the disk is not present
func (disk *Disk) CreateLoopDevice(imgDir string) error {
	if len(disk.DiskPath) != 0 {
		return fmt.Errorf("aborting disk creation. " +
			"Error: disk path is already set on Disk obj")
	}

	err := disk.createDiskImage(imgDir)
	if err != nil {
		return err
	}

	// create the loop device using losetup
	createLoopDeviceCommand := "losetup -P " + "/dev/" + getLoopDevName() + " " + disk.imageName + " --show"
	var devicePath []byte
	devicePath, err = RunCommand(createLoopDeviceCommand)
	if err != nil {
		return fmt.Errorf("error creating loop device. Error : %v", err)
	}
	// Trim trailing newline
	devicePath = devicePath[:len(devicePath)-1]

	disk.DiskPath = string(devicePath)

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

/*
// RunCommandWithSudo runs a command with sudo permissions
func RunCommandWithSudo(cmd string) error {
	return RunCommand("sudo " + cmd)
}
*/

// RunCommand runs a command on the host
func RunCommand(cmd string) ([]byte, error) {
	substring := strings.Fields(cmd)
	name := substring[0]
	args := substring[1:]
	stdout, err := exec.Command(name, args...).CombinedOutput() // #nosec G204
	if err != nil {
		return stdout, fmt.Errorf("run failed, cmd={%s}\nerror={%v}\noutput={%v}", cmd, err, stdout)
	}
	return stdout, err
}

func (disk *Disk) CreateFilesystem(fstype string) error {
	var mkfsCommand string
	switch fstype {
	case "ext4":
		mkfsCommand = "mkfs.ext4 -O quota -E quotatype=prjquota " + disk.DiskPath
	case "xfs":
		mkfsCommand = "mkfs.xfs " + disk.DiskPath
	default:
		return fmt.Errorf("error creating mkfs command. " +
			"Error: invalid filesystem type")
	}

	_, err := RunCommand(mkfsCommand)
	if err != nil {
		return fmt.Errorf("error creating fileystem on loop device. Error : %v", err)
	}
	return nil
}

func (disk *Disk) Wipefs() error {
	// List partitions
	fdiskCommand := "fdisk -o Device -l " + disk.DiskPath
	fdiskStdOut, err := RunCommand(fdiskCommand)
	if err != nil {
		return fmt.Errorf("error listing disk partitions. Error: %v", err)
	}

	// Wipe partitions
	stringsAndSubstrings := regexp.MustCompile("\n("+disk.DiskPath+".*)").
		FindAllStringSubmatch(string(fdiskStdOut), -1)
	for _, stringAndSubstrings := range stringsAndSubstrings {
		//Remove string match (which includes newline)
		// leave behind substring (without newline)
		substrings := stringAndSubstrings[1:]

		//This range will be always be of length 1
		for _, substring := range substrings {
			_, err := RunCommand("wipefs -fa " + substring)
			if err != nil {
				return fmt.Errorf("error removing filesystem "+
					"from partition %s. Error: %v",
					substring, err,
				)
			}
		}
	}

	// Wipe disk
	wipefsDeviceCommand := "wipefs -fa " + disk.DiskPath
	_, err = RunCommand(wipefsDeviceCommand)
	if err != nil {
		return fmt.Errorf("error wiping disk %s. Error: %v",
			disk.DiskPath, err,
		)
	}
	return nil
}

func MkdirAll(paths ...string) error {
	createMkdirCommand := "mkdir -p"
	for _, path := range paths {
		createMkdirCommand += " " + path
	}
	_, err := RunCommand(createMkdirCommand)
	if err != nil {
		return fmt.Errorf("error creating directory. Error : %v", err)
	}
	return nil
}

func (disk *Disk) PrjquotaMount(mountpoint string) error {
	createMountCommand := "mount -o rw,prjquota " + disk.DiskPath + " " + mountpoint
	_, err := RunCommand(createMountCommand)
	if err != nil {
		return fmt.Errorf("error mounting loop device. Error : %v", err)
	}
	disk.MountPoints = append(disk.MountPoints, mountpoint)
	return nil
}

func (disk *Disk) Unmount() []error {
	var lastErr []error
	for i := 0; i < len(disk.MountPoints); i++ {
		_, err := RunCommand("umount " + disk.MountPoints[i])
		if err != nil {
			lastErr = append(lastErr, err)
		} else {
			disk.MountPoints = disk.MountPoints[1:]
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
	_, err := RunCommand(detachLoopCommand)
	if err != nil {
		return fmt.Errorf("cannot detach loop device. Error : %v", err)
	}
	err = os.Remove(disk.imageName)
	if err != nil {
		return fmt.Errorf("could not delete backing disk image. Error : %v", err)
	}
	deleteLoopDeviceCommand := "rm " + disk.DiskPath
	_, err = RunCommand(deleteLoopDeviceCommand)
	if err != nil {
		return fmt.Errorf("could not delete loop device. Error : %v", err)
	}
	return nil
}

// PrepareDisk prepares the setup necessary for testing xfs hostpath quota
func PrepareDisk(imgDir, hostPath string) (Disk, error) {
	physicalDisk := NewDisk(DiskImageSize)

	err := MkdirAll(imgDir, hostPath)
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while making a new directory at {%s}, {%s}", imgDir, hostPath)
	}

	err = physicalDisk.CreateLoopDevice(imgDir)
	if err != nil {
		return physicalDisk, errors.Wrapf(err, "while creating loop back device with disk %+v", physicalDisk)
	}

	/*
		// Make xfs fs on the created loopback device
		err = physicalDisk.CreateFilesystem(fsType)
		if err != nil {
			return physicalDisk, errors.Wrapf(err, "while formatting the disk {%+v} with xfs fs", physicalDisk)
		}

		// Mount the xfs formatted loopback device
		err = physicalDisk.PrjquotaMount(hostPath)
		if err != nil {
			return physicalDisk, errors.Wrapf(err, "while mounting the disk with pquota option {%+v}", physicalDisk)
		}
	*/
	return physicalDisk, nil
}

// DestroyDisk performs performs the clean-up task after testing the features
func (disk *Disk) DestroyDisk(hostPath, imgDir string) error {
	var errs string
	// Unmount the disk
	err := disk.Unmount()
	if len(err) > 0 {
		for _, v := range err {
			errs = errs + v.Error() + ";"
		}
		return errors.Wrapf(errors.New(errs), "while unmounting the disk {%+v}", disk)
	}

	// Detach and delete the disk
	error := disk.DetachAndDeleteDisk()
	if err != nil {
		return errors.Wrapf(error, "while detaching and deleting the disk {%+v}", disk)
	}

	// Deleting the image directory and the hostpath directory
	_, error = RunCommand("rm -rf " + imgDir + " " + hostPath)
	if err != nil {
		return errors.Wrapf(error, "while deleting the directories {%s}, {%s}", imgDir, hostPath)
	}

	return nil
}
