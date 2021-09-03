package disk

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	imageDirectory = "/tmp"
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
	Name string
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
		Name:      "",
	}
	return disk
}

// Generates a random image name for the backing file.
// the file name will be of the format fakeXXX, where X=[0-9]
func generateDiskImageName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomNumber := 100 + rand.Intn(899)
	imageName := "fake" + strconv.Itoa(randomNumber)
	return imageDirectory + "/" + imageName
}

// AttachDisk triggers a udev add event for the disk. If the disk is not present, the loop
// device is created and event is triggered
func (disk *Disk) AttachDisk() error {
	if disk.Name == "" {
		if err := disk.createLoopDevice(); err != nil {
			return err
		}
	}
	return nil
}

func (disk *Disk) createDiskImage() error {
	// no of blocks
	/*count := disk.Size / blockSize
	createImageCommand := "dd if=/dev/zero of=" + disk.imageName + " bs=" + strconv.Itoa(blockSize) + " count=" + strconv.Itoa(int(count))
	err := utils.RunCommand(createImageCommand)*/
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

func (disk *Disk) createLoopDevice() error {
	var err error
	if _, err = os.Stat(disk.imageName); err != nil {
		err = disk.createDiskImage()
		if err != nil {
			return err
		}
	}

	deviceName := getLoopDevName()
	devicePath := "/dev/" + deviceName
	// create the loop device using losetup
	createLoopDeviceCommand := "losetup " + devicePath + " " + disk.imageName
	err = RunCommandWithSudo(createLoopDeviceCommand)
	if err != nil {
		return fmt.Errorf("error creating loop device. Error : %v", err)
	}
	disk.Name = devicePath
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
	err := exec.Command(name, args...).Run() // #nosec G204
	if err != nil {
		return fmt.Errorf("run failed %s %v", cmd, err)
	}
	return err
}

func (disk *Disk) CreateFileSystem(fstype string) error {
	createMkfsCommand := "mkfs -t " + fstype + " " + disk.Name
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
	createMountCommand := "mount -o rw,pquota " + disk.Name + " " + path
	err := RunCommandWithSudo(createMountCommand)
	if err != nil {
		return fmt.Errorf("error mounting loop device. Error : %v", err)
	}
	disk.MountPoints = append(disk.MountPoints, path)
	return nil
}

func (disk *Disk) Unmount() error {
	var lastErr error = nil
	for _, mp := range disk.MountPoints {
		err := RunCommandWithSudo("umount " + mp)
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// DetachAndDeleteDisk detaches the loop device from the backing
// image. Also deletes the backing image and block device file in /dev
func (disk *Disk) DetachAndDeleteDisk() error {
	if disk.Name == "" {
		return fmt.Errorf("no such disk present for deletion")
	}
	detachLoopCommand := "losetup -d " + disk.Name
	err := RunCommandWithSudo(detachLoopCommand)
	if err != nil {
		return fmt.Errorf("cannot detach loop device. Error : %v", err)
	}
	deleteBackingImageCommand := "rm " + disk.imageName
	err = RunCommandWithSudo(deleteBackingImageCommand)
	if err != nil {
		return fmt.Errorf("could not delete backing disk image. Error : %v", err)
	}
	deleteLoopDeviceCommand := "rm " + disk.Name
	err = RunCommandWithSudo(deleteLoopDeviceCommand)
	if err != nil {
		return fmt.Errorf("could not delete loop device. Error : %v", err)
	}
	return nil
}
