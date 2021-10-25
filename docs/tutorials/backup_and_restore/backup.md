# Backups using Velero and Restic

You can create backups of Hostpath LocalPV volumes using Velero and Restic. Follow the steps below:

## Step 1: Prepare object-storage

You will need an object-store to store your volume backups (Velero remote backup). You may use an AWS S3 bucket, a GCP storage bucket, or a MinIO instance for this.

In this guide, we'll use a minimal MinIO instance installation. [Click here](./velero/minio.md) for instructions on setting up your own minimal single-server MinIO.

> **Note**: Refer to the [official MinIO documentation](https://docs.min.io/docs/) for up-to-date instructions on setting up MinIO.

## Step 2: Install Velero with Restic

You will need Velero with Restic to create backups. In this guide, we'll use the above MinIO as out default backup storage location. [Click here](./velero/velero_with_restic.md) for instructions on setting up Velero with Restic and creating a backupstoragelocation object.

> **Note**: Refer to the [official Velero documentation](https://velero.io/docs/v1.6/restic/#setup-restic) for up-to-date instructions on setting up Velero with Restic.

## Step 3: Create backup

We will 'exec' into the Velero container's shell and run the following commands.

Get the Pod name for the Velero Pod running in 'velero' namespace.
```console
$ kubectl -n velero get pods

NAME                          READY   STATUS    RESTARTS   AGE
openebs-backup-minio-ss-0-0   1/1     Running   0          7h23m
restic-2xwsf                  1/1     Running   0          7h12m
velero-7dd57b857-2gd25        1/1     Running   0          7h12m
```

'Exec' into the Pod's velero container.
```console
$ kubectl -n velero exec -it velero-7dd57b857-2gd25 -c velero -- /bin/bash
```

Verify if the following command lists the backup-location 'default' as 'Available'.
```console
$ ./velero backup-location get

NAME      PROVIDER   BUCKET/PREFIX   PHASE       LAST VALIDATED                  ACCESS MODE   DEFAULT
default   aws        velero          Available   2021-09-04 01:05:06 +0000 UTC   ReadWrite     true
```

Create a backup. We will use the `--default-volumes-to-restic` to use the Restic plugin for volumes. Use the `--wait` flag to wait for the backup to complete or fail before the command returns.
```console
$ ./velero create backup my-localpv-backup --include-namespaces <app-namespace> --default-volumes-to-restic --wait

Backup request "my-localpv-backup" submitted successfully.
Waiting for backup to complete. You may safely press ctrl-c to stop waiting - your backup will continue in the background.
....
Backup completed with status: Completed. You may check for more information using the commands `velero backup describe my-localpv-backup` and `velero backup logs my-localpv-backup`.
```

Verify the status of the backup using the following command...
```console
$ ./velero backup get
NAME                STATUS      ERRORS   WARNINGS   CREATED                         EXPIRES   STORAGE LOCATION   SELECTOR
my-localpv-backup   Completed   0        0          2021-09-04 01:13:36 +0000 UTC   29d       default            <none>
```

```console
$ exit
```

For more information on using Velero, refer to the Velero documentation at [velero.io/docs](https://velero.io/docs).
