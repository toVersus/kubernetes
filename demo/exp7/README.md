# 実験7: Pod のホストポートの重複

同一 Node にスケジュールされる Pod でホストポートが重複した場合に、Pod がスケジュールできない様子を観察します。

## 手順

kind で Kubernetes v1.29.2 のマルチノードクラスタを起動

```sh
kind create cluster --config kind.yaml
```

ホストポートの番号に `10080` を設定して Pod を起動

```sh
kubectl run nginx --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx","image":"nginx","ports":[{"containerPort":80,"hostPort":10080,"protocol":"TCP"}]}]}}'
```

ログを監視

```sh
stern -n kube-system kube-scheduler --exclude healthz --exclude lease \
    --exclude round_trippers --exclude "Watch close"  --exclude "client-side throttling"
```

ホストポートの番号に `10080` を設定して別の Pod を起動

```sh
kubectl run nginx-invalid --image=nginx \
  --overrides='{"spec":{"containers":[{"name":"nginx","image":"nginx","ports":[{"containerPort":80,"hostPort":10080,"protocol":"TCP"}]}]}}'
```
