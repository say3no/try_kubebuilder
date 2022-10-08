
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

* Resource
  * Deployment 
  * ?

```


```

## APIの雛形作成

## Webhookの雛形作成

## カスタムコントローラーの動作確認
