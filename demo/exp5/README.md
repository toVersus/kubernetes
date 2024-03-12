# 実験5: Scheduling Framework と Scheduler plugin

## 事前準備

kind で Kubernetes をセルフビルド

```sh
# Kubernetes と kube-scheduler-simulator のソースコードをクローン
git clone -b demo/kube-scheduler-code-reading git@github.com:toVersus/kubernetes.git
git clone -b kubernetes-v1.29.2 git@github.com:toVersus/kube-scheduler-simulator.git

cd kubernetes
kind build node-image .
```

## 手順

kind でセルフビルドした Kubernetes のマルチノードクラスタを起動

```sh
kind create cluster --config kind.yaml
```

kube-scheduler に Pod の annotation を更新する権限を付与

```sh
kubectl patch clusterrole system:kube-scheduler \
  --type='json' -p='[{"op": "add", "path": "/rules/0", "value":{ "apiGroups": [""], "resources": ["pods"], "verbs": ["update"]}}]'
```

ログを監視

```sh
stern -n kube-system kube-scheduler --exclude healthz --exclude lease \
    --exclude round_trippers --exclude "Watch close" --exclude "client-side throttling"
```

Pod を起動

```sh
kubectl run nginx --image=nginx
```

Pod の annotation を確認

```sh
kubectl get pods nginx -oyaml
```
