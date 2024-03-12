# 実験6: Node のスケールアウト時の優先順位

kind で Kubernetes v1.29.2 のマルチノードクラスタを起動し、起動時に kind-worker-2 の Node に `foo:bar` の NoSchedule の Taint を付与しておきます。kind-worker の Node に Pod が起動できず、kind-worker-2 の Node がスケールアウトした状態を擬似的に再現します。(**あくまで再現なので挙動は全く同じではありません。**) kind-worker-2 の Taint を外して、Pending 状態だった Pod がスケジュールされる様子を観察します。

## 手順

kind で Kubernetes v1.29.2 のマルチノードクラスタを起動

```sh
kind create cluster --config kind.yaml
```

PriorityClass がデフォルト (`Priority: 0`)、system-cluster-critical (`Priority: 2000000000`), system-node-critical (`Priority: 2000001000`) の DaemonSet を起動

```sh
kubectl apply -R -f manifests/
```

kind-worker2 に Priority の低い順で Pod を 3 台起動


```sh
# Priority: 0
kubectl run nginx-default --image=nginx \
  --overrides='{"spec":{"nodeSelector":{"kubernetes.io/hostname":"kind-worker2"}}}'

# Priority: 2000000000
kubectl run nginx-cluster-critical --image=nginx \
  --overrides='{"spec":{"nodeSelector":{"kubernetes.io/hostname":"kind-worker2"},"priorityClassName":"system-cluster-critical"}}'


# Priority: 2000001000
kubectl run nginx-node-critical --image=nginx \
  --overrides='{"spec":{"nodeSelector":{"kubernetes.io/hostname":"kind-worker2"},"priorityClassName":"system-node-critical"}}'

# Priority: 0
kubectl run nginx-default-slow --image=nginx \
  --overrides='{"spec":{"nodeSelector":{"kubernetes.io/hostname":"kind-worker2"}}}'
```

ログを監視

```sh
stern -n kube-system kube-scheduler --exclude healthz --exclude lease \
    --exclude round_trippers --exclude "Watch close"  --exclude "client-side throttling"
```

kind-worker-2 の Node から Taint を削除

```sh
kubectl taint node kind-worker2 foo:NoSchedule-
```

kind のログをエクスポートして kind-worker2 の containerd のログを確認

```sh
kind export logs
```
