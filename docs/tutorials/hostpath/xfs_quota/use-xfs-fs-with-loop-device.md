# Create XFS filesystem at the basepath as loop device (if filesystem is not XFS)

If you don't have a device with XFS filesystem, you can use a loop device and create an XFS filesystem on it. This works even if your root filesystem is not XFS.

Create a 32MiB sparse file which will be formatted as XFS filesystem, mounted as loop device and exposed as the directory `/var/openebs/local`

1. Make sure library for managing xfs-fs is installed.
```console
sudo apt update
sudo apt-get install -y xfsprogs
#RHEL/Centos?
#sudo yum install -y xfsprogs
```

2. Make directory where mount will occur :-
```console
sudo mkdir -p /var/openebs/local
cd /var/openebs
```

3. Create sparse file of max size 32MiB using seek of max size 32MiB
```console
sudo dd if=/dev/zero of=xfs.32M bs=1 count=0 seek=32M
```

4. Format the sparse file in xfs format
```console
sudo mkfs -t xfs -q xfs.32M
```

5. Mount it as loop device with project quota enabled
```console
sudo mount -o loop,rw xfs.32M -o pquota /var/openebs/local
```