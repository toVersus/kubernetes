# 実験9: NodeResourcesFit の戦略の違い

NodeResourcesFit の LeastAllocated と MostAllocated の戦略の違いを見ていきます。

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

Pod を CPU 要求 1 core で作成

```sh
kubectl run nginx-1 --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx-1","image":"nginx","resources":{"requests":{"cpu":"1"}}}]}}'
```

ログを監視

```sh
stern -n kube-system kube-scheduler --exclude healthz --exclude lease \
    --exclude round_trippers --exclude "Watch close"  --exclude "client-side throttling"
```

別の Pod を CPU 要求 1 core で作成

```sh
kubectl run nginx-2 --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx-2","image":"nginx","resources":{"requests":{"cpu":"1"}}}]}}'
```

Pod の annotation から `scheduler-simulator/score-result` を確認

```sh
kubectl get pods nginx-2 -oyaml
```

各ノードで実行した Score plugin の得点は以下の通り

```json
{
  "kind-worker": {
    "ImageLocality": "0",
    "NodeAffinity": "0",
    "NodeResourcesBalancedAllocation": "72",
    "NodeResourcesFit": "65",
    "TaintToleration": "0",
    "VolumeBinding": "0"
  },
  "kind-worker2": {
    "ImageLocality": "0",
    "NodeAffinity": "0",
    "NodeResourcesBalancedAllocation": "89",
    "NodeResourcesFit": "83",
    "TaintToleration": "0",
    "VolumeBinding": "0"
  },
  "kind-worker3": {
    "ImageLocality": "0",
    "NodeAffinity": "0",
    "NodeResourcesBalancedAllocation": "87",
    "NodeResourcesFit": "79",
    "TaintToleration": "0",
    "VolumeBinding": "0"
  }
}
```

|                   | kind-worker   | kind-worker2  | kind-worker3  |
| ----------------- | ------------- | ------------- |---------------|
|Allocatable CPU    | 2000          | 5000          | 8000          |
|Requested CPU      | 1200          | 1200          | 2200          |
|Allocatable Memory | 5505789952    | 5505789952    | 5505789952    |
|Requested Memory   | 471859200     | 471859200     | 681574400     |
|Score              | 65            | 83            | 79            |


```
# kind-worker
{((2000 - 1200) * 100 * 1 / 2000) + ((5505789952 - 471859200) * 100 * 1 / 5505789952)} / (1 + 1) = 65

# kind-worker2
{((5000 - 1200) * 100 * 1 / 5000) + ((5505789952 - 471859200) * 100 * 1 / 5505789952)} / (1 + 1) = 83

# kind-worker3
{((8000 - 2200) * 100 * 1 / 8000) + ((5505789952 - 681574400) * 100 * 1 / 5505789952)} / (1 + 1) = 79
```

kind のクラスタを削除

```sh
kind delete cluster
```

Kubernetes のソースコードを修正

```diff
diff --git a/cmd/kube-scheduler/scheduler.go b/cmd/kube-scheduler/scheduler.go
index d630b24f465..7acb6c5af7e 100644
--- a/cmd/kube-scheduler/scheduler.go
+++ b/cmd/kube-scheduler/scheduler.go
@@ -27,7 +27,7 @@ import (
 )

 func main() {
-       command, cancelFn, err := debuggablescheduler.NewSchedulerCommand()
+       command, cancelFn, err := debuggablescheduler.NewSchedulerCommand(debuggablescheduler.WithSchedulerConfig("/etc/kubernetes/scheduler-config.yaml"))
        if err != nil {
                panic(err)
        }
```

kind で Kubernetes をセルフビルド

```sh
# Kubernetes と kube-scheduler-simulator のソースコードをクローン
git clone -b demo/kube-scheduler-code-reading git@github.com:toVersus/kubernetes.git
git clone -b kubernetes-v1.29.2 git@github.com:toVersus/kube-scheduler-simulator.git

cd kubernetes
kind build node-image . --image kindest/node:v1.29.2-custom-scheduler-config
```

NodeResourcesFit の戦略に MostAllocated が設定されたプロファイルを含んだ kube-scheduler の設定ファイルを読み込ませた kind クラスタを起動

```sh
kind create cluster --config kind-gke-optimize-utilization.yaml
```

kube-scheduler に Pod の annotation を更新する権限を付与

```sh
kubectl patch clusterrole system:kube-scheduler \
  --type='json' -p='[{"op": "add", "path": "/rules/0", "value":{ "apiGroups": [""], "resources": ["pods"], "verbs": ["update"]}}]'
```

Pod のスケジューラに `gke.io/optimize-utilization-scheduler` を指定して、CPU 要求 1 core で作成

```sh
kubectl run nginx-1 --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx-1","image":"nginx","resources":{"requests":{"cpu":"1"}}}],"schedulerName":"gke.io/optimize-utilization-scheduler"}}'
```

ログを監視

```sh
stern -n kube-system kube-scheduler --exclude healthz --exclude lease \
    --exclude round_trippers --exclude "Watch close"  --exclude "client-side throttling"
```

Pod のスケジューラに `gke.io/optimize-utilization-scheduler` を指定して、別の Pod を CPU 要求 1 core で作成

```sh
kubectl run nginx-2 --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx-2","image":"nginx","resources":{"requests":{"cpu":"1"}}}],"schedulerName":"gke.io/optimize-utilization-scheduler"}}'
```

Pod の annotation から `scheduler-simulator/score-result` を確認

```sh
kubectl get pods nginx-2 -oyaml
```

各ノードで実行した Filter plugin の得点は以下の通り

```json
{
  "kind-control-plane": {
    "NodeName": "passed",
    "NodeUnschedulable": "passed",
    "TaintToleration": "node(s) had untolerated taint {node-role.kubernetes.io/control-plane: }"
  },
  "kind-worker": {
    "NodeName": "passed",
    "NodeResourcesFit": "Insufficient cpu",
    "NodeUnschedulable": "passed",
    "TaintToleration": "passed"
  },
  "kind-worker2": {
    "NodeName": "passed",
    "NodeResourcesFit": "passed",
    "NodeUnschedulable": "passed",
    "TaintToleration": "passed"
  },
  "kind-worker3": {
    "NodeName": "passed",
    "NodeResourcesFit": "passed",
    "NodeUnschedulable": "passed",
    "TaintToleration": "passed"
  }
}
```

各ノードで実行した Score plugin の得点は以下の通り

```json
{
  "kind-worker2": {
    "ImageLocality": "0",
    "NodeAffinity": "0",
    "NodeResourcesBalancedAllocation": "89",
    "NodeResourcesFit": "16",
    "TaintToleration": "0",
    "VolumeBinding": "0"
  },
  "kind-worker3": {
    "ImageLocality": "0",
    "NodeAffinity": "0",
    "NodeResourcesBalancedAllocation": "93",
    "NodeResourcesFit": "11",
    "TaintToleration": "0",
    "VolumeBinding": "0"
  }
}
```

|                   | kind-worker   | kind-worker2  | kind-worker3  |
| ----------------- | ------------- | ------------- |---------------|
|Allocatable CPU    | 2000          | 5000          | 8000          |
|Requested CPU      | 2200          | 1200          | 1200          |
|Allocatable Memory | 5505789952    | 5505789952    | 5505789952    |
|Requested Memory   | 471859200     | 471859200     | 471859200     |
|Score              | -             | 16            | 11            |


```
# kind-worker2
{(1200 * 100 * 1 / 5000) + (471859200 * 100 * 1 / 5505789952)} / (1 + 1) = 16

# kind-worker3
{(1200 * 100 * 1 / 8000) + (471859200 * 100 * 1 / 5505789952)} / (1 + 1) = 11
```
