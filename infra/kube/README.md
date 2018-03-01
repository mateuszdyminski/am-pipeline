### Kafka

```
kubectl apply -f storage
kubectl apply -f rbac
kubectl apply -f zookeeper
kubectl apply -f kafka
```

### Elasticsearch

```
cd es

kubectl create -f es-discovery-svc.yaml
kubectl create -f es-svc.yaml
kubectl create -f es-master.yaml
kubectl rollout status -f es-master.yaml
kubectl create -f es-client.yaml
kubectl rollout status -f es-client.yaml
kubectl create -f es-data.yaml
kubectl rollout status -f es-data.yaml
```

### Pipeline

```
kubectl apply -f am
```