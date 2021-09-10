# Deploy single-server MinIO with Hostpath LocalPV storage

Follow the instructions below to create a minimal MinIO deployment on a Kubernetes cluster. This setup creates a single MinIO server with 4 volumes, and is not recommended for use in production environments.

**Prerequisites:**
1. Kubernetes v1.19 or above
2. kubectl CLI

## Step 1: Create namespace

We will create the MinIO Tenant in the namespace which we will use for Velero.

```console
$ kubectl create namespace velero
```

## Step 2: Create MinIO Operator and MinIO Tenant using Helm Chart

We will use the MinIO Helm chart to create the MinIO Operator and also a single-server MinIO tenant. We will use 'minio' and 'minio123' as the default Access Key and Secret Key respectively. This is a minimal install, refer to the [MinIO helm chart documentation](https://github.com/minio/operator/blob/master/helm/minio-operator/README.md) for more options.

```console
$ #helm repo add minio https://operator.min.io/
$ #helm repo update
$ helm install minio-operator minio/minio-operator \
	--namespace minio-operator \
	--create-namespace \
	--set tenants[0].name="openebs-backup-minio" \
	--set tenants[0].namespace="velero" \
	--set tenants[0].certificate.requestAutoCert=false \
	--set tenants[0].pools[0].servers=1 \
	--set tenants[0].pools[0].volumesPerServer=4 \
	--set tenants[0].pools[0].size=10Gi \
	--set tenants[0].pools[0].storageClassName="openebs-hostpath" \
	--set tenants[0].secrets.enabled=true \
	--set tenants[0].secrets.name="openebs-backup-minio-creds-secret" \
	--set tenants[0].secrets.accessKey="minio" \
	--set tenants[0].secrets.secretKey="minio123"
```

Verify if the MinIO Operator Pod and the MinIO Console Pod got created.
```console
$ kubectl get pods -n minio-operator

NAME                                     READY   STATUS    RESTARTS   AGE
minio-operator-67bc7fc5d6-sb2wq          1/1     Running   0          120m
minio-operator-console-59db5db85-k2qbw   1/1     Running   0          120m
```

Verify the status of the Secrets, Tenant objects and Pods in the 'velero' namespace.
```console
$ kubectl get tenants,secrets,pods -n velero

NAME                                       STATE         AGE
tenant.minio.min.io/openebs-backup-minio   Initialized   6h47m

NAME                                       TYPE                                  DATA   AGE
secret/default-token-kqlr8                 kubernetes.io/service-account-token   3      6h49m
secret/openebs-backup-minio-creds-secret   Opaque                                2      6h41m
secret/operator-tls                        Opaque                                1      6h41m
secret/operator-webhook-secret             Opaque                                3      6h41m

NAME                              READY   STATUS    RESTARTS   AGE
pod/openebs-backup-minio-ss-0-0   1/1     Running   0          6h40m
```

You can use the 'mc' MinIO client to create buckets in the MinIO object store.
