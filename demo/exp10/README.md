# 実験10: SchedulingGate による Pod のスケジュールの抑止

SchedulingGate を利用することで Pod のスケジュールを静かに抑止することができます。

## 手順

kind で Kubernetes v1.29.2 のマルチノードクラスタを起動

```sh
kind create cluster --config kind.yaml
```

PodSchedulingReadiness の Feature Gate が有効になっているか確認 (`1` になっていたら有効化されている)

```sh
kubectl get --raw /metrics | grep kubernetes_feature_enabled
kubectl get --raw /metrics | grep kubernetes_feature_enabled | grep PodSchedulingReadiness
```

ログを監視

```sh
stern -n kube-system kube-scheduler --exclude healthz --exclude lease \
    --exclude round_trippers --exclude "Watch close"  --exclude "client-side throttling"
```

SchedulingGate を設定した Pod を作成

```sh
kubectl run nginx --image=nginx \
  --overrides='{"spec":{"schedulingGates":[{"name":"foo.3-shake.com"}]}}'
```

Pod の状態を確認

```sh
kubectl get pods
```

Pod のイベントを発火

```sh
kubectl label pods nginx foo=bar
```

SchedulingGate を後から追加することはできない

```sh
kubectl patch pod nginx --type='json' -p='[{"op":"add","path":"/spec/schedulingGates/1","value":{"name":"bar.3-shake.com"}}]'
```

SchedulingGate を削除

```sh
kubectl patch pod nginx --type='json' -p='[{"op":"remove","path":"/spec/schedulingGates/0"}]'
```

Pod を一旦削除

```sh
kubectl delete pod nginx
```

SchedulingGate を設定していない Pending 状態の Pod を作成

```sh
kubectl run nginx --image=nginx \
  --overrides='{"spec":{"nodeSelector":{"foo":"bar"}}}'
```

SchedulingGate を後から設定することはできない

```sh
kubectl patch pod nginx --type='json' -p='[{"op":"add","path":"/spec/schedulingGates","value":[{"name":"foo.3-shake.com"}]}]'
```
