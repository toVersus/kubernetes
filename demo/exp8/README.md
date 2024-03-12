# 実験8: Balloon Pod

余剰ノードを確保する Ballon Pod の手法では kube-scheduler の Preemption を活用しています。Preemption の処理と Pod の優先度の関係を追いかけてみます。

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

デフォルトの優先度の Pod をノードの割り当て可能リソースギリギリで起動

```sh
kubectl run nginx --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx","image":"nginx","resources":{"requests":{"cpu":"7"}}}]}}'
```

ログを監視

```sh
stern -n kube-system kube-scheduler --exclude healthz --exclude lease \
    --exclude round_trippers --exclude "Watch close"  --exclude "client-side throttling"
```

同じノードに優先度の高い Pod をノードの割り当て可能リソースギリギリで起動

```sh
kubectl run nginx-critical --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx","image":"nginx","resources":{"requests":{"cpu":"7"}}}],"priorityClassName":"system-cluster-critical"}}'
```

既存の Pod を削除してお掃除

```sh
kubectl delete pods nginx-critical
```

`maxUnavailable: 0` の PDB が設定されたデフォルトの優先度の Deployment をデプロイ

```sh
kubectl apply -R -f manifests/
```

同じノードに優先度の高い Pod をノードの割り当て可能リソースギリギリで起動

```sh
kubectl run nginx-critical --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx","image":"nginx","resources":{"requests":{"cpu":"7"}}}],"priorityClassName":"system-cluster-critical"}}'
```
