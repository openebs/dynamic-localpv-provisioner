# Install Velero with Restic

Follow the steps below to install Velero with Restic. We will use the velero-plugin-for-aws plugin to create remote backups to a MinIO instance. The MinIO instance used in this guide, is in the 'velero' namespace. This is the same instance which we have configured in [this guide](./minio.md).

## Step 1: Create dedicated MinIO bucket

We will use the 'mc' MinIO client to create a bucket to store our volume backups.

Execute the following kubectl command to run the 'minio/mc' container in the MinIO instance's namespace. We will execute the commands in a TTY shell inside the container.
```console
$ kubectl -n velero run minio-client --image=minio/mc --rm -it --command -- /bin/sh
```
You should be inside the container and able to run commands in the shell. Run the following commands to configure the MinIO client and create the bucket 'velero'.

```console
$ mc config host add velero http://minio.velero.svc:80 'minio' 'minio123'

mc: Configuration written to `/root/.mc/config.json`. Please update your access credentials.
mc: Successfully created `/root/.mc/share`.
mc: Initialized share uploads `/root/.mc/share/uploads.json` file.
mc: Initialized share downloads `/root/.mc/share/downloads.json` file.
Added `velero` successfully.
```

```console
$ mc mb -p velero/velero

Bucket created successfully `velero/velero`.
```

```console
$ exit
```

## Step 2: Create a file to store the MinIO credentials

We will use pass the MinIO access key and secret key to Velero using a file called 'minio-credentials'.

```console
$ cat << EOF > ./minio-credentials
[default]
aws_access_key_id=minio
aws_secret_access_key=minio123
EOF
```

## Step 3: Install Velero and Restic using the Velero Helm Chart

Install Velero using the vmware-tanzu/velero helm chart. We are creating a release with minimal options, for more instructions on setting up Velero, refer to the [Velero helm chart documentation](https://github.com/vmware-tanzu/helm-charts/blob/main/charts/velero/README.md).

```console
$ #helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts
$ #helm repo update
$ helm install velero vmware-tanzu/velero \
	--namespace velero \
	--create-namespace \
	--set-file credentials.secretContents.cloud=$(pwd)/minio-credentials \
	--set configuration.provider="aws" \
	--set configuration.backupStorageLocation.name="default" \
	--set configuration.backupStorageLocation.bucket="velero" \
	--set configuration.backupStorageLocation.config.region="minio" \
	--set configuration.backupStorageLocation.config.s3ForcePathStyle="minio" \
	--set configuration.backupStorageLocation.config.s3Url="http://minio.velero.svc:80" \
	--set backupsEnabled=true \
	--set snapshotsEnabled=false \
	--set deployRestic=true \
	--set initContainers[0].name=velero-plugin-for-aws \
	--set initContainers[0].image=velero/velero-plugin-for-aws:latest \
	--set initContainers[0].volumeMounts[0].mountPath=/target \
	--set initContainers[0].volumeMounts[0].name=plugins
```

Verify if the Velero and Restic components got created.
```console
$ kubectl get secrets,backupstoragelocations,pods -n velero

NAME                                       TYPE                                  DATA   AGE
secret/default-token-kqlr8                 kubernetes.io/service-account-token   3      7h17m
secret/openebs-backup-minio-creds-secret   Opaque                                2      7h9m
secret/operator-tls                        Opaque                                1      7h9m
secret/operator-webhook-secret             Opaque                                3      7h9m
secret/sh.helm.release.v1.velero.v1        helm.sh/release.v1                    1      6h57m
secret/velero                              Opaque                                1      6h57m
secret/velero-restic-credentials           Opaque                                1      6h56m
secret/velero-server-token-6r4sq           kubernetes.io/service-account-token   3      6h57m

NAME                                      PHASE       LAST VALIDATED   AGE     DEFAULT
backupstoragelocation.velero.io/default   Available   26s              6h57m   true

NAME                              READY   STATUS    RESTARTS   AGE
pod/openebs-backup-minio-ss-0-0   1/1     Running   0          7h9m
pod/restic-2xwsf                  1/1     Running   0          6h57m
pod/velero-7dd57b857-2gd25        1/1     Running   0          6h57m
```

You can use the 'velero' CLI tool to use Velero if your MinIO service is exposed (NodePort or LocalBalancer service) and is reachable from your shell. 
We will use Velero from inside the 'velero' container's shell.
