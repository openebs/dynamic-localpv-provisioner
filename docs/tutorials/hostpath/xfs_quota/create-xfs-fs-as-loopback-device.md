# Create XFS filesystem at the basepath as loopback device (if filesystem is not XFS)

Create a 32Mb disk file which will be formatted as xfs filesystem, mounted as loopback device and exposed as the directory `/var/openebs/local`

1. make sure library for managing xfs-fs is installed
```console
sudo apt update
sudo apt-get install -y xfsprogs
```

2. make directory where mount will occur :-
```console
sudo mkdir -p /var/openebs/local
cd /var/openebs
```

3. create sparse file of max size 32Mb using seek of max size 32Mb
```console
sudo dd if=/dev/zero of=xfs.32M bs=1 count=0 seek=32M
```

4. format the sparse file in xfs format
```console
sudo mkfs -t xfs -q xfs.32M
```

5. mount it as loopback device with project quota enabled
```console
sudo mount -o loop,rw xfs.32M -o pquota /var/openebs/local
```