# Restore Velero backups

## Step 1: List backups

We will 'exec' into the Velero container to list our backups.

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

List the backups available.
```console
$ ./velero backup get

NAME                STATUS      ERRORS   WARNINGS   CREATED                         EXPIRES   STORAGE LOCATION   SELECTOR
my-localpv-backup   Completed   0        0          2021-09-04 01:13:36 +0000 UTC   29d       default            <none>
```

## Step 2: Create a restore

Restores don't overwrite already existing components with the same name. To replace already-existing components with the contents of the backup, you will have to delete.

Use the `--namespace-mappings [SOURCE_NAMESPACE]:[DESTINATION_NAMESPACE]` flag to restore to a different namespace.
```console
$ ./velero restore create my-localpv-restore --from-backup my-localpv-backup --restore-volumes=true
```

Verify the status of the restore and also the components that were restored.
```console
$ ./velero restore get
```

```console
$ exit
```
