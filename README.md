
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
