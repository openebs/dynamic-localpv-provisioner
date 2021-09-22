# Prerequisites for enabling XFS Quota on Hostpath LocalPV

## Install the 'xfsprogs' package

### Install on Ubuntu/Debian

Install the xfsprogs package on Ubuntu and Debian systems using the following commands:
```console
$ sudo apt-get update
$ sudo apt-get install -y xfsprogs
```

### Install on RHEL/CentOS

Install the xfsprogs package on RHEL/CentOS systems using the following commands:
```console
$ sudo yum install -y xfsprogs
$ #Using Fedora?
$ #sudo dnf install -y xfsprogs
```

## Mount filesystem using 'pquota' mount option

### Step 1:
Check to see if the filesystem of the hostpath directory is XFS. The default hostpath directory is '/var/openebs/local'. Use the following command to check if the filesystem is 'xfs' and to identify the device where the filesystem is stored.
```console
$ df -Th /var/openebs/local

Filesystem     Type  Size  Used Avail Use% Mounted on
/dev/nvme0n1p1 xfs   8.0G  959M  7.1G  12% /
```
>The above command may fail if the path does not yet exist. To work around this, let's check for the host-device name and the filesystem type the directory will have, if it was created. To do this, run the following script, which will run `df -Th` against /var/openebs/local, /var/openebs, /var, and so on...
>```console
>BASEPATH="/var/openebs/local"
>
>until OUTPUT=$(df -Th $BASEPATH 2> /dev/null)
>do
>BASEPATH=$(echo "$BASEPATH" | sed 's|\(.*\)/.*|\1|')
>done
>
>echo "PATH=${BASEPATH}"
>#Final output
>echo "$OUTPUT"
>```

### Step 2: 
Check if the existing mount options of the device we found in 'Step 1', already has 'pquota' or 'prjquota'. The sample command below uses the device '/dev/nvme0n1p1'. Execute the following command:
```console
$ sudo mount | grep "^/dev/nvme0n1p1"

/dev/nvme0n1p1 on / type xfs (rw,relatime,seclabel,attr2,inode64,noquota)
```
If the mount options include 'pquota' or 'prjquota', you can move over to [this doc](./enable-xfs-quota.md) to enable XFS Quota.

### Step 3:
In this step we will mount the device using the 'pquota' mount option. If the filesystem you're trying to mount with the pquota option is your root filesystem (mounted on '/'), then follow the instructions below. If the filesystem is a on a data disk, and not on the root filesystem, then move to the [data disk sub-section](#filesystem-on-data-disk).

#### **Root filesystem**

This will add/modify the rootflags option from your kernel's boot options.
Edit the file '/etc/default/grub'.
```console
$ sudo vi /etc/default/grub
```

Find the line with the variable "GRUB_CMDLINE_LINUX".
```console
GRUB_CMDLINE_LINUX="console=tty0 crashkernel=auto net.ifnames=0 console=ttyS0"
```

To the end of the string, add `rootflags=pquota`. If the 'rootflags' option is already set, add 'pquota' to the list of arguments.
```console
GRUB_CMDLINE_LINUX="console=tty0 crashkernel=auto net.ifnames=0 console=ttyS0 rootflags=pquota"
```

Locate your 'grub.cfg' file. You might find your grub.cfg file at any one of these paths, depending on the OS:
- /boot/grub2/grub.cfg 
- /boot/efi/EFI/ubuntu/grub.cfg
- /boot/efi/EFI/debian/grub.cfg
- /boot/efi/EFI/redhat/grub.cfg
- /boot/efi/EFI/centos/grub.cfg
- /boot/efi/EFI/fedora/grub.cfg

Create a backup copy of the existing grub.cfg. The sample commands below use the path '/boot/grub2/grub.cfg'. Replace the paths with your grub.cfg path.
```console
$ sudo cp /boot/grub2/grub.cfg /boot/grub2/grub.cfg.backup
```

Generate a new grub.cfg which will include the changes to mount options.
```console
$ sudo grub2-mkconfig -o /boot/grub2/grub.cfg
```

Reboot!
```console
$ sudo reboot
```

After the system reboots, you can check the mount options again to confirm.
```console
$ sudo mount | grep "^/dev/nvme0n1p1"
/dev/nvme0n1p1 on / type xfs (rw,relatime,seclabel,attr2,inode64,prjquota)
```

#### **Filesystem on data disk**

If your filesystem is not the root filesystem, follow the steps below.
Make a note of the 'Filesystem' and 'Mounted on' values from Step 1.
The sample command below uses a device at '/dev/nvme1n1' and a mount path at '/mnt/data'.

Unmount the filesystem on the data disk. Execute the following command:
```console
$ sudo umount /dev/nvme1n1
```

Mount the disk using 'pquota' mount option into the mount path.
> **Note:** 'pquota' is not usable with 'remount' mount option.
```console
$ sudo mount -o rw,pquota /dev/nvme1n1 /mnt/data
```

Check the mount options again to confirm.
```console
$ sudo mount | grep "^/dev/nvme1n1"
/dev/nvme1n1 on /mnt/data type xfs (rw,relatime,seclabel,attr2,inode64,prjquota)
```

You can make the changes persistent across reboots by adding the 'pquota' option to the /etc/fstab file at the data disk's entry.
```console
UUID=9cff3d69-3769-4ad9-8460-9c54050583f9 /mnt/data               xfs     defaults,pquota 0 0
```