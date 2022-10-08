
# つくって学ぶ Kubebuilder

## CustomController の基礎

* Custom Resource を利用するためには、OpenAPI v3.0 形式で Custom Resource Definison(CRD) を書く必要がる
* k8s では、あるリソースの状態をチェックして何らかの処理を行うプログラムのことをコントローラーと呼ぶ。
  *  例えば Deployment リソースに基づいて ReplicaSet リソースを作成しているのも、コントローラーの一つと言える
* k8s には標準リソースに対応する数多くのコントローラーが存在することがわかる: https://github.com/kubernetes/kubernetes/tree/master/pkg/controller
* 標準で装備されているものと、カスタムリソースを対象としたコントローラーを、カスタムコントローラーと呼ぶ

---


* Reconciliation Loop
  * コントローラーのメインロジック。リソースに記述された状態を理想とし、システムの現在の状態と比較し、その差分がなくなるように調整する処理を実行し続ける。
    * これによって宣言的な状態を維持し続けるのだな
  * 冪等
    * Reconciliation Loop は冪等性を備えていなければならない
    * Pod が3つなら、何度 Reconcile が呼ばれても Pod は3つにしなければならない
      * 追加で3つつくったりしちゃだめよ

* トリガー: エッジドリブンとレベルドリブン
  * よく分かってない
  * 具体的な現場を見るのとDeployment 抽象の両方を見ないと行けないから発生している違い？

---

Controlloer は「あるべき状態でないときに、あるべき状態にするための処理をするもの」と考えればいい

## MarkdownViewController

* MarkdownViewController
  * ユーザーが用意した Markdown をレンダリングしてブラウザから閲覧できるようにサービスを提供するコントローラー
  * Markdown のレンダリングには mdBook を利用する
    * https://rust-lang.github.io/mdBook/

* ユーザーは MrakdownView カスタムリソースを作成する
* MarkdownViewResource
  * CustomResource に記述された Markdown を COnifgMap リソースとして作成
  * Markdown をレンダリングするための mdBook を Deployment リソースとして作成
  * mdBook にアクセスするための Service リソースを作成
* ユーザーは、作成されたサービスを経由して、レンダリングされた Markdown を閲覧できる

### MarkdownView CustomResource


```yml
apiVersion: view.zuetrope.github.io/v1
kind: MarkdownView
metadata:
  name: markdownview-sample
spec:
  markdowns:
    SUMMARY.md: |
      # Summary

      - [Page1](page1.md)
    page1.md: |
      # Page 1

      一ページ目のコンテンツです。
  replicas: 1
  viewerImage: "peaceiris/mdbook:latest"
```

```bash
 codes
├── 00_scaffold:  Kubebuilderで生成したコード
├── 10_tilt:      Tiltを利用した開発環境のセットアップを追加
├── 20_manifests: CRD, RBAC, Webhook用のマニフェストを生成
├── 30_client:    クライアントライブラリの利用例を追加
├── 40_reconcile: Reconcile処理、およびWebhookを実装
└── 50_completed: Finalizer, Recorder, モニタリングのコードを追加
```

---

# Kubebuilder

* `kubebuilder` とは、 CustomController のプロジェクトの雛形を自動作成するためのツール。
* ソースコードだけでなく、 Makefile や Dockerfile, 各種マニフェストなど数多くのファイルを生成する。

```bash

try_kubebuilder on  main [!?] on ☁️  (ap-northeast-1) 
❯ ./kubebuilder -h
CLI tool for building Kubernetes extensions and tools.

Usage:
  kubebuilder [flags]
  kubebuilder [command]

Examples:
The first step is to initialize your project:
    kubebuilder init [--plugins=<PLUGIN KEYS> [--project-version=<PROJECT VERSION>]]

<PLUGIN KEYS> is a comma-separated list of plugin keys from the following table
and <PROJECT VERSION> a supported project version for these plugins.

                              Plugin keys | Supported project versions
------------------------------------------+----------------------------
                base.go.kubebuilder.io/v3 |                          3
         declarative.go.kubebuilder.io/v1 |                       2, 3
  deploy-image.go.kubebuilder.io/v1-alpha |                          3
                     go.kubebuilder.io/v2 |                       2, 3
                     go.kubebuilder.io/v3 |                          3
               go.kubebuilder.io/v4-alpha |                          3
          grafana.kubebuilder.io/v1-alpha |                          3
       kustomize.common.kubebuilder.io/v1 |                          3
 kustomize.common.kubebuilder.io/v2-alpha |                          3

For more specific help for the init command of a certain plugins and project version
configuration please run:
    kubebuilder init --help --plugins=<PLUGIN KEYS> [--project-version=<PROJECT VERSION>]

Default plugin keys: "go.kubebuilder.io/v3"
Default project version: "3"


Available Commands:
  alpha       Alpha-stage subcommands
  completion  Load completions for the specified shell
  create      Scaffold a Kubernetes API or webhook
  edit        Update the project configuration
  help        Help about any command
  init        Initialize a new project
  version     Print the kubebuilder version

Flags:
  -h, --help                     help for kubebuilder
      --plugins strings          plugin keys to be used for this subcommand execution
      --project-version string   project version (default "3")

Use "kubebuilder [command] --help" for more information about a command.

try_kubebuilder on  main [!?] on ☁️  (ap-northeast-1) 
❯ 
```

subcommand。 本編では `init` と `create` を触っていく。

```bash

❯ ./kubebuilder -h | grep Available -A8
Available Commands:
  alpha       Alpha-stage subcommands
  completion  Load completions for the specified shell
  create      Scaffold a Kubernetes API or webhook
  edit        Update the project configuration
  help        Help about any command
  init        Initialize a new project
  version     Print the kubebuilder version


try_kubebuilder on  main [!?] on ☁️  (ap-northeast-1) 
❯ 

```

## プロジェクトの雛形作成

```bash
$ kubebuilder init --domain say3no.github.io --repo this/repo
```


## APIの雛形作成

```bash
$ kubebuilder create api --group view --version v1 --kind MarkdownView
```

生成するカスタムリソースの GVK 

* `--group`: リソースが属するグループ名を指定
* `--version`: 適切なバージョンを指定。今後使用が変わる可能性ｇありそうなら `v1alpha1` とかにするとか
* `--kind` : 作成するリソースの名前


---

- api
  - v1
    - `markdownview_types.go` <- `api/v1` 配下はこいつだけ弄るとおもってればおｋ
    - ...
    - ...
- cotroller
  - `markdownview_controller.go` <- CR Controller 本体
  - `suite_test.go`
- config
  - crd <- Custom Resource Definison
    - bases
      - `view.say3no.github.io_markdownviews.yaml`
    - `kustomization.yaml`
    - `kustomizeconfig.yaml`
    - patches
      - `cainjection_in_markdwonviews.yaml`
      - `webhook_in_markdownviews.yaml`
  - rbac
    - `role.yaml`
    - `markdownview_editor_role.yaml`
    - `markdownview_viewer_role.yaml`
  - samples
    - `view_v1_markdownview.yaml`


## Webhookの雛形作成

* k8s の拡張機能: admission webhook 
  * 特定のリソースおｗ作成/更新する際に Webhook API を呼び出し、バリデーションやリソースの書き換えを行うための機能
    * ansible の handler みたいな
* `kubebuilder` で生成可能な 3種類 の Webhook Option 。 主に使うのは上位2つかな
  * `--programmatic-validation`: リソースのバリデーションの Webhook
  * `--defaulting`: リソースのフィールドにデフォルト値を設定するための Webhook
  * `--conversion`: カスタムリソースのバージョンアップ時にリソースの変換をおこなうための Webhhok

```bash
$ kubebuilder create webhook --group view --verison v1 --kind MarkdownView --programmatic-validation --defaulting
try_kubebuilder/markdown-view on  main [!] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ ../kubebuilder create webhook --group view --version v1 --kind MarkdownView --programmatic-validation --defaulting
Writing kustomize manifests for you to edit...
Writing scaffold for you to edit...
api/v1/markdownview_webhook.go
Update dependencies:
$ go mod tidy
Running make:
$ make generate
/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
Next: implement your new Webhook and generate the manifests with:
$ make manifests
```

---


```bash
try_kubebuilder/markdown-view on  main [!?] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ make manifests 
/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

```

---

生成されたものたち

- api
  - v1
    - `markdownView_webhook.go` <- だいじ
    - `webhook_suite_test.go`
- config
  - certmanager <- Admission Webhook を利用するためには証明書が必要。そのための諸 yaml
    - `certificate.yaml`
    - `kustoization.yaml`
    - `kustomizeconfig.yaml`
  - default
    - `manager_webhook_patch.yaml`
    - `webhookcainjection_patch.yaml`
  - webhook <- マニフェストファイルのみなさん
    - `kustomization.yaml`
    - `kustomizeconfig.yaml`
    - `manifests.yaml`
    - `service.yaml`

### kustomization.yaml の編集

* 生成直後の状態では make manifests コマンドで マニフェストを生成しても、 Webhook機能が利用できるようにはなっていない
* こいつがコミットされた diff を見てくれ~~~


## カスタムコントローラーの動作確認


```bash
try_kubebuilder/markdown-view on  main via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ kind create cluster
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.25.2) 🖼 
 ✓ Preparing nodes 📦  
 ✓ Writing configuration 📜 
 ✓ Starting control-plane 🕹️ 
 ✓ Installing CNI 🔌 
 ✓ Installing StorageClass 💾 
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Thanks for using kind! 😊

try_kubebuilder/markdown-view on  main via 🐹 v1.19.2 on ☁️  (ap-northeast-1) took 1m 
❯ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml  

namespace/cert-manager created
customresourcedefinition.apiextensions.k8s.io/certificaterequests.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/certificates.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/challenges.acme.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/clusterissuers.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/issuers.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/orders.acme.cert-manager.io created
serviceaccount/cert-manager-cainjector created
serviceaccount/cert-manager created
serviceaccount/cert-manager-webhook created
configmap/cert-manager-webhook created
clusterrole.rbac.authorization.k8s.io/cert-manager-cainjector created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-issuers created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-clusterissuers created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-certificates created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-orders created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-challenges created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-ingress-shim created
clusterrole.rbac.authorization.k8s.io/cert-manager-view created
clusterrole.rbac.authorization.k8s.io/cert-manager-edit created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-approve:cert-manager-io created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-certificatesigningrequests created
clusterrole.rbac.authorization.k8s.io/cert-manager-webhook:subjectaccessreviews created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-cainjector created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-issuers created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-clusterissuers created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-certificates created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-orders created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-challenges created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-ingress-shim created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-approve:cert-manager-io created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-certificatesigningrequests created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-webhook:subjectaccessreviews created
role.rbac.authorization.k8s.io/cert-manager-cainjector:leaderelection created
role.rbac.authorization.k8s.io/cert-manager:leaderelection created
role.rbac.authorization.k8s.io/cert-manager-webhook:dynamic-serving created
rolebinding.rbac.authorization.k8s.io/cert-manager-cainjector:leaderelection created
rolebinding.rbac.authorization.k8s.io/cert-manager:leaderelection created
rolebinding.rbac.authorization.k8s.io/cert-manager-webhook:dynamic-serving created
service/cert-manager created
service/cert-manager-webhook created
deployment.apps/cert-manager-cainjector created
deployment.apps/cert-manager created
deployment.apps/cert-manager-webhook created
mutatingwebhookconfiguration.admissionregistration.k8s.io/cert-manager-webhook created
validatingwebhookconfiguration.admissionregistration.k8s.io/cert-manager-webhook created

try_kubebuilder/markdown-view on  main [!] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ kubectl get pod -n cert-manager
NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-cainjector-857ff8f7cb-bpq9h   1/1     Running   0          5m9s
cert-manager-d58554549-x5xf2               1/1     Running   0          5m9s
cert-manager-webhook-76fdf7c485-5d8rz      1/1     Running   0          5m9s

try_kubebuilder/markdown-view on  main [!] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ 
```


---

### コントローラーの動作確認

* arm は v3.8.7 がないっぽい

```bash
❯ make install
test -s /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/controller-gen || GOBIN=/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2
/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
test -s /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/kustomize || { curl -Ss "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s -- 3.8.7 /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin; }
Version v3.8.7 does not exist or is not available for darwin/arm64.
make: *** [/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/kustomize] Error 1

try_kubebuilder/markdown-view on  main via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ 
```

* https://qiita.com/nakamasato/items/5a4018c73d7d1e6da025

```bash
$ export KUSTOMIZE_VERSION=v4.2.0
$ make install
```

パスした

```bash
try_kubebuilder/markdown-view on  main [!] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ export KUSTOMIZE_VERSION=v4.2.0

try_kubebuilder/markdown-view on  main [!] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ make install
test -s /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/controller-gen || GOBIN=/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2
/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
test -s /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/kustomize || { curl -Ss "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s -- 4.2.0 /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin; }
{Version:kustomize/v4.2.0 GitCommit:d53a2ad45d04b0264bcee9e19879437d851cb778 BuildDate:2021-06-30T22:49:26Z GoOs:darwin GoArch:arm64}
kustomize installed to /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/kustomize
/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/markdownviews.view.say3no.github.io created

try_kubebuilder/markdown-view on  main [!] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) took 5s 
❯ ```

---

```bash
❯ make deploy
test -s /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/controller-gen || GOBIN=/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2
/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
cd config/manager && /Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/kustomize edit set image controller=controller:latest
/Users/sanosei/github.com/say3no/try_kubebuilder/markdown-view/bin/kustomize build config/default | kubectl apply -f -
namespace/markdown-view-system unchanged
customresourcedefinition.apiextensions.k8s.io/markdownviews.view.say3no.github.io unchanged
serviceaccount/markdown-view-controller-manager unchanged
role.rbac.authorization.k8s.io/markdown-view-leader-election-role unchanged
clusterrole.rbac.authorization.k8s.io/markdown-view-manager-role configured
clusterrole.rbac.authorization.k8s.io/markdown-view-metrics-reader unchanged
clusterrole.rbac.authorization.k8s.io/markdown-view-proxy-role unchanged
rolebinding.rbac.authorization.k8s.io/markdown-view-leader-election-rolebinding unchanged
clusterrolebinding.rbac.authorization.k8s.io/markdown-view-manager-rolebinding unchanged
clusterrolebinding.rbac.authorization.k8s.io/markdown-view-proxy-rolebinding unchanged
service/markdown-view-controller-manager-metrics-service unchanged
service/markdown-view-webhook-service unchanged
deployment.apps/markdown-view-controller-manager unchanged
certificate.cert-manager.io/markdown-view-serving-cert unchanged
issuer.cert-manager.io/markdown-view-selfsigned-issuer unchanged
mutatingwebhookconfiguration.admissionregistration.k8s.io/markdown-view-mutating-webhook-configuration configured
validatingwebhookconfiguration.admissionregistration.k8s.io/markdown-view-validating-webhook-configuration configured

try_kubebuilder/markdown-view on  main [!⇡] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ ```


```bash
❯  kubectl get pod -n markdown-view-system
NAME                                                READY   STATUS    RESTARTS   AGE
markdown-view-controller-manager-548cff7b6f-kqbjb   2/2     Running   0          4m59s

try_kubebuilder/markdown-view on  main [!] via 🐹 v1.19.2 on ☁️  (ap-northeast-1) 
❯ 
```

log がでている

```bash

❯ kubectl logs  -n markdown-view-system markdown-view-controller-manager-548cff7b6f-kqbjb -c manager -f
1.6652454396743307e+09  INFO    controller-runtime.metrics      Metrics server is starting to listen    {"addr": "127.0.0.1:8080"}
1.665245439674532e+09   INFO    controller-runtime.builder      Registering a mutating webhook  {"GVK": "view.say3no.github.io/v1, Kind=MarkdownView", "path": "/mutate-view-say3no-github-io-v1-markdownview"}
1.665245439674574e+09   INFO    controller-runtime.webhook      Registering webhook     {"path": "/mutate-view-say3no-github-io-v1-markdownview"}
1.6652454396745944e+09  INFO    controller-runtime.builder      Registering a validating webhook        {"GVK": "view.say3no.github.io/v1, Kind=MarkdownView", "path": "/validate-view-say3no-github-io-v1-markdownview"}
1.665245439674616e+09   INFO    controller-runtime.webhook      Registering webhook     {"path": "/validate-view-say3no-github-io-v1-markdownview"}
1.6652454396746435e+09  INFO    setup   starting manager
1.6652454396748865e+09  INFO    Starting server {"path": "/metrics", "kind": "metrics", "addr": "127.0.0.1:8080"}
1.6652454396749735e+09  INFO    Starting server {"kind": "health probe", "addr": "[::]:8081"}
1.6652454396748748e+09  INFO    controller-runtime.webhook.webhooks     Starting webhook server
1.6652454396750708e+09  INFO    controller-runtime.certwatcher  Updated current TLS certificate
I1008 16:10:39.675084       1 leaderelection.go:248] attempting to acquire leader lease markdown-view-system/952d835c.say3no.github.io...
1.6652454396751184e+09  INFO    controller-runtime.webhook      Serving webhook server  {"host": "", "port": 9443}
1.665245439675223e+09   INFO    controller-runtime.certwatcher  Starting certificate watcher
I1008 16:10:39.678102       1 leaderelection.go:258] successfully acquired lease markdown-view-system/952d835c.say3no.github.io
1.665245439678144e+09   DEBUG   events  markdown-view-controller-manager-548cff7b6f-kqbjb_af5d9842-822e-49cc-8ecf-baa6cfcb51b4 became leader    {"type": "Normal", "object": {"kind":"Lease","namespace":"markdown-view-system","name":"952d835c.say3no.github.io","uid":"5452541e-a083-4cc8-b2d6-60171684fefc","apiVersion":"coordination.k8s.io/v1","resourceVersion":"8172"}, "reason": "LeaderElection"}
1.6652454396785386e+09  INFO    Starting EventSource    {"controller": "markdownview", "controllerGroup": "view.say3no.github.io", "controllerKind": "MarkdownView", "source": "kind source: *v1.MarkdownView"}
1.6652454396785576e+09  INFO    Starting Controller     {"controller": "markdownview", "controllerGroup": "view.say3no.github.io", "controllerKind": "MarkdownView"}
1.665245439780741e+09   INFO    Starting workers        {"controller": "markdownview", "controllerGroup": "view.say3no.github.io", "controllerKind": "MarkdownView", "worker count": 1}
^C
```

別窓で イカ実行
```bash
❯ kubectl apply -f config/samples/view_v1_markdownview.yaml
markdownview.view.say3no.github.io/markdownview-sample created
```

200 を返していそう

```bash
...
1.665245439780741e+09   INFO    Starting workers        {"controller": "markdownview", "controllerGroup": "view.say3no.github.io", "controllerKind": "MarkdownView", "worker count": 1}
1.6652463090219948e+09  DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/mutate-view-say3no-github-io-v1-markdownview", "UID": "7bcedd20-cac1-42c5-ac6d-76103c0b3db2", "kind": "view.say3no.github.io/v1, Kind=MarkdownView", "resource": {"group":"view.say3no.github.io","version":"v1","resource":"markdownviews"}}
1.6652463090221906e+09  INFO    markdownview-resource   default {"name": "markdownview-sample"}
1.6652463090224829e+09  DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/mutate-view-say3no-github-io-v1-markdownview", "code": 200, "reason": "", "UID": "7bcedd20-cac1-42c5-ac6d-76103c0b3db2", "allowed": true}
1.665246309026078e+09   DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/validate-view-say3no-github-io-v1-markdownview", "UID": "cd347bf1-609f-4282-a0e2-787ca188c7a5", "kind": "view.say3no.github.io/v1, Kind=MarkdownView", "resource": {"group":"view.say3no.github.io","version":"v1","resource":"markdownviews"}}
1.6652463090261595e+09  INFO    markdownview-resource   validate create {"name": "markdownview-sample"}
1.6652463090261793e+09  DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/validate-view-say3no-github-io-v1-markdownview", "code": 200, "reason": "", "UID": "cd347bf1-609f-4282-a0e2-787ca188c7a5", "allowed": true}
```

### 開発の流れ

#### Controller の実装が変わったときは、以下のコマンドで docker build -> load to kind

```bash
$ make docker-build
$ kind load docker-image controller:latest
```

#### CRD に変更がある場合は、以下

```bash
$ make install
```

非互換な変換をした場合は、事前に `make uninstall`

#### CRD 以外のマニフェストファイルに変更がある場合

```bash
$ make deploy
```

#### 次のコマンドでカスタムコントローラの再起動

```bash
$ kubectl rollout restart -n markdown-view-system deployment markdown-view-controller-manager
```

### Tilt による効率的な開発

要するに変更を watch して action をするタスクランナーっぽい。
こっちは後回しでいいや

