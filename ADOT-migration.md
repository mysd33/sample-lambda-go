# X-Ray SDK/DaemonからADOTへの移行
* AWS X-Ray 用の SDK と Daemon は2026年2月25日にメンテナンスモードに入り、2027年2月25日にサポート終了となるため、ADOT(AWS Distro for OpenTelemetry) への移行に対応した。
* 割と大がかりな移行作業だったため、移行の際の参考情報をまとめている。
* X-Rayの実際のトレースの表示については、[README](README.md#adotx-rayによるトレース情報の可視化)を参照    

## 1. Before/After
* タグを切って差分で確認できるようにしている
    * [Before（X-Ray SDK/Daemon）](https://github.com/mysd33/sample-lambda-go/releases/tag/xray-sdk)
    * [After（ADOT）](https://github.com/mysd33/sample-lambda-go/releases/tag/adot)
    * [差分比較](https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot)

## 2. 移行時の参考情報

* Go関連
    * [AWS X-Ray: Migrate to OpenTelemetry Go](https://docs.aws.amazon.com/xray/latest/devguide/manual-instrumentation-go.html)   
    * [AWS X-Ray: Migrate to OpenTelemetry Go - AWS SDK for Go v2 instrumentation](https://docs.aws.amazon.com/xray/latest/devguide/manual-instrumentation-go.html#aws-sdk-instrumentation)  
        * DynamoDB、S3、SQS等、AWS SDK for Goを利用したAWSサービスへのアクセスのトレースのための計装方法
    * [AWS X-Ray: Migrate to OpenTelemetry Go - Instrumenting outgoing HTTP calls](https://docs.aws.amazon.com/xray/latest/devguide/manual-instrumentation-go.html#http-client-instrumentation)  
        * [otelhttp - NewTransport](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp#NewTransport)というHTTP通信のOpenTelemetry対応のためのGoのライブラリを利用した、HTTPクライアントのトレースのための計装方法             
    * [otelsql](https://github.com/XSAM/otelsql)
        * SQLのOpenTelemetry対応のためのGoのライブラリ。現状、GoのOpenTelemetryの公式リポジトリにはSQL用のライブラリがないため、サードパーティ製のライブラリを利用している。
    * [otelmongo](go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo)
        * DocumentDB(MongoDB)のOpenTelemetry対応のためのGoのライブラリ。[MongoDBのGoドライバー](https://github.com/mongodb/mongo-go-driver)の[V2移行](https://github.com/mongodb/mongo-go-driver/blob/master/docs/migration-2.0.md)も必要となる。
    * [Using the AWS Distro for OpenTelemetry Go SDK](https://aws-otel.github.io/docs/getting-started/go-sdk/manual-instr)
        * ADOT公式のドキュメント
    * [OpenTelemetry Go Documentation](https://opentelemetry.io/ja/docs/languages/go/)
        * OpenTelemetry公式のGoのドキュメント。[Getting Started](https://opentelemetry.io/ja/docs/languages/go/getting-started/)等を参考

* Lambda/Go関連
    * [AWS X-Ray: Migrate to OpenTelemetry Go - Lambda manual instrumentation](https://docs.aws.amazon.com/xray/latest/devguide/manual-instrumentation-go.html#lambda-instrumentation)   
        * Lambdaのmain関数で、lambda.Start関数を呼び出す前に、OpenTelemetryの計装コードを手動で埋め込む方法が記載されている。
    * [AWS Distro for OpenTelemetry Lambda](https://aws-otel.github.io/docs/getting-started/lambda)
        * ただし、上のリンクに記載された最新の最適化されたアプローチは、ADOT CollectorのLambdaレイヤーの[サポートランタイム](https://aws-otel.github.io/docs/getting-started/lambda#supported-runtimes)にOS専用ランタイム(OS-only Runtime provided.al2023)がないため、Goの場合はまだ以下のレガシーアプローチをとる必要がありそう。
    * [AWS Distro for OpenTelemetry Lambda Support For Go(the legacy approach)](https://aws-otel.github.io/docs/getting-started/lambda/lambda-go)
        * Lambda/Goだと、このサイトに従いレガシーアプローチにより提供されるLambdaLayerを使用して移行した。
    * [OpenTelemetry AWS Lambda Instrumentation for Golang](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda#section-readme)
    * [Recommended Configurations for OpenTelemetry AWS Lambda Instrumentation with AWS X-Ray](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig#section-readme)

## 3. 実装の移行作業のポイント
### 3.1. モジュールの追加
* goのバージョンアップを実施しておく。以下の情報から1.23以上にするのがよいと思われる。
    * [OpenTelemetryの公式サイト](https://opentelemetry.io/ja/docs/languages/go/getting-started/#prerequisites)にはGo1.23以上と書かれている
    * [ADOTの公式サイト](https://aws-otel.github.io/docs/getting-started/go-sdk/manual-instr#requirements)にはGo1.19以上と書かれている

* go getでいろいろなモジュールを必要がある。
    * ソフトウェアフレームワーク(appbase)の例
        * `go get`で追加して、コード修正後、`go mod tidy`
        * go.modの例
            * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/go.mod
    * AP(app)の例
        * 基本、ソフトウェアフレームワーク側に隠蔽されるので、go mod tidyするとindirectで参照する形になる
        * go.modの例
            * https://github.com/mysd33/sample-lambda-go/blob/adot/app/go.mod
　
* Go関連（OpenTelemetry SDK）

```sh
go get go.opentelemetry.io/contrib/propagators/aws/xray
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/otel/sdk/resource
go get go.opentelemetry.io/otel/sdk/trace
go get go.opentelemetry.io/otel/sdk/metric
go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
```

* AWS SDK関連(otelawsライブラリ)

```sh
go get go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws
```
　
* Labmda関連(otellambdaライブラリ)
    
```sh
go get -u go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda
go get -u go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig
```

* HTTP関連（otelhttpライブラリ）
    
```sh
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
```
　
* DocumentDB（MongoDB）関連（otelmongoライブラリ）
    * 以下はMongoDriver V2用なので、V1→V2移行が必要。（V1用のライブラリを使う方法もある）

```sh
go get go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo
```
　
* Aurora（RDB）関連（otelsqlライブラリ）
    * 現在は、Opentelemtryのcontribのライブラリでは提供されていないようで、それをポートしたプロジェクトを使用している。

```sh　
go get github.com/XSAM/otelsql
```

### 3.2. 外部リソースの呼び出しコードの修正
* いずれのケースの変更も難しくはない。

* AWS SDKでの呼び出し
    * 以前のX-Ray SDKでのコードを削除して、「otelaws.AppendMiddlewares(&cfg.APIOptions)」を指定するようにする。
        * DynamoDB
            * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/pkg/dynamodb/dynamodb.go#L46-L48
            * 差分
                * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-b11d1783a1b4aa01e3c160f168dacd13e787fd0fcecac3c6e15d6005e476e4e7
        * S3
            * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/pkg/objectstorage/objectstorage.go#L213-L215
            * 差分
                * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-5d832e78d288943c3710446b4d71c5cd3442cee5ac81c6f7be7b57ffd1c25100
        * SQS
            * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/pkg/async/async.go#L73-L75
            * 差分
                * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-793191c234ec6f6473dd27fbe2ef1b1463263be8faaf88ea133bcc3fe1dc6312

* HTTP通信
    * 以前GetやPostメソッドにあった、X-Ray SDKでのコードを削除して
    * http.Clientを作成するときに、「Transport: otelhttp.NewTransport(http.DefaultTransport)」を指定するようにする。
    * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/pkg/httpclient/httpclient.go#L80-L84
    * 差分
        * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-38608920822691ca7965fff3ef70d8afcbbf94fbccb873f8da28b2a58bf299f8
* DocumentDB（MongoDB）
    * mongo.Connectするときのオプションとして、SetMonitor(otelmongo.NewMonitor())を指定するようにする。
    * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/pkg/documentdb/documentdb.go#L116-L119
    * 差分
        * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-b0f4ab11e9a0726587072ca2ece0b13472c9bcadd1d893466aa308fcbab2c398
　
* Aurora（RDB SQL）    
    * 以前のX-RaySDKを使ってコネクション作成していたコードを削除して、otelsql.Openでコネクションを作成するようにする。
    * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/pkg/rdb/transaction_manager.go#L130-L142
    * 差分
        * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-c3bba235c6a75de2765be69a94951f491730136812a47392248217151cf4dd6a        

### 3.3. Lambdaのmain関数のコード修正
* main関数でOpenTelemetryでのトレースを開始するコードが必要である。
* どれもmain関数に書くのは同じコードなので、ソフトウェアフレームワークで共通化した、otel.StartLambda関数を作成している。
* なお、注意点として、sam local等のローカル実行の場合は、ADOTのコードを呼び出すとADOTのCollectorがX-Rayをずっと探しに行こうとしてもアクセスできないことでLambdaがタイムアウトエラーになってしまうので、環境変数`ENV`を見てローカル実行のときには、ADOTのコードを呼ばないように通常のlambda.Startに切り替えている。
    * 後述のtemplate.yaml側のLambdaLayerの有効・無効の切り替えも合わせて実施　
    * ソフトウェアフレームワーク(appbase/pkg/otelパッケージ)
        * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/pkg/otel/otel_lambda.go#L26
    * ソフトウェアフレームワーク化により、業務AP側のmain関数は従来のlambda.Start関数をotel.StartLambda関数に変更すれば済むので、修正影響が極小化されている。
        * 業務AP（API Gatewayトリガ）:main関数 
            * https://github.com/mysd33/sample-lambda-go/blob/adot/app/cmd/todo/main.go#L40
            * 差分
                * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-8c4d1aedce9da260399b20accca5f3904b392df44e3c8c55cbc51e47ff5a38fb
        * 業務AP（SQSトリガ）:main関数
            * API Gatewayトリガと全く同じ修正
            * https://github.com/mysd33/sample-lambda-go/blob/adot/app/cmd/todo-async/main.go#L32
            * 差分
                * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-c8c9de4c18a5d7d14cae6670c4bf2cbde7eb60394b0259746972a85560f8d5df

### 3.4. template.yml、CloudFormation、samconfig.toml、Makefile（LambdaLayerの設定）などの修正
* 従来のX-Ray SDKの場合、LambdaにはX-Rayデーモンの設定が不要であったが、ADOTの場合は、ADOT CollectorのLambda Layerを追加する必要がある。
    * ある種、[ECSでADOT Collectorをサイドカーコンテナにする](https://github.com/mysd33/ecs-on-fargate-adot-cfn-demo)のと同じ。

* template.yamlの例
    * https://github.com/mysd33/sample-lambda-go/blob/adot/template.yaml#L59
    * https://github.com/mysd33/sample-lambda-go/blob/adot/template.yaml#L116
    * 差分
        * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-1363ef5ce8886100842332c97163aad7934237e1fe49b5d40422b45fdc30f38e

* X-RayのVPCEndpointの追加
    * VPC内LambdaであってもX-RaySDK/DaemonだとVPC Endpointは不要だったが、ADOTにするとX-RayのVPC Endpointが必要になる。
    * https://github.com/mysd33/sample-lambda-go/blob/adot/cfn/cfn-vpe.yaml#L103
    * 差分
        * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-9ed197f55c53ff9eabd8328a275330896a4aa033bb40a5fe75ce3c703e56261a

* LambdaLayerの有効・無効の切り替え
    * sam local等のローカル実行の場合に、ADOTのLamdaLayerを定義してしまうと、ADOTのCollectorがX-Rayへアクセスを試みてしまい探してもないため、Lambdaがタイムアウトエラーになってしまう。
    * samconfig.tomlの仕組みと組み合わせて、sam localするときに `--config-env local`オプションをつけて実行すると、template.yamlのパラメータ「`LambdaLayersEEnabled`」が「`false`」になるようにして、この場合は、LambdaLayerが有効化されないようにしている。    
        * sampconfig.tomlの例
            * https://github.com/mysd33/sample-lambda-go/blob/adot/samconfig.toml#L47
            * 差分
                * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-9ed197f55c53ff9eabd8328a275330896a4aa033bb40a5fe75ce3c703e56261a

    * Makefileにも、sam local実行するときのオプションにも付与してあげると、楽になる。
        * Makefileの例
            * https://github.com/mysd33/sample-lambda-go/blob/adot/Makefile#L69C47-L69C66
            * 差分
                * https://github.com/mysd33/sample-lambda-go/compare/xray-sdk...adot#diff-9ed197f55c53ff9eabd8328a275330896a4aa033bb40a5fe75ce3c703e56261a

### 4. 懸念事項
1. OpenTelemetry SDK関連ライブラリをgo getしようとしても、cockroachdb/errorsが依存するgoogle.golang.org/genprotoのライブラリが競合してしまい、go getに失敗してしまう問題が発生
    * おそらくOpenTelemetryと、cockroachdb/errorsが参照しているgenprotoのバージョンが異なるのか、genprotoが旧バージョンと新バージョンで分割形態が変わったため競合してしまう、go getに失敗してしまい、モジュール追加ができない問題が発生した。
    * 試行錯誤した結果、回避策として、go.modで、google.golang.org/genprotoの特定のバージョンをexcludeしたが、これが、一番いいやり方かはわからない。最終的には、cockroachdb/errorsのバージョンアップを待つまでの暫定的な回避策としている。
        * https://github.com/mysd33/sample-lambda-go/blob/adot/appbase/go.mod#L6
1. go言語＝OS専用ランタイム（カスタムランタイム：provided.al2023）が、最新のADOTの推奨実装方法に対応しておらず、レガシーアプローチをとるしか無さそう
    * [AWS Distro for OpenTelemetry Lambda](https://aws-otel.github.io/docs/getting-started/lambda)に記載された最新の最適化されたアプローチは、ADOT Collectorの[Lambdaレイヤーのサポートランタイム](https://aws-otel.github.io/docs/getting-started/lambda#supported-runtimes)に、OS専用ランタイム(OS-only Runtime provided.al2023)がないため、Goの場合はまだ以下のレガシーアプローチをとる必要がある。
    * [AWS Distro for OpenTelemetry Lambda Support For Go(the legacy approach)](https://aws-otel.github.io/docs/getting-started/lambda/lambda-go)の手順に従い、レガシーアプローチにより提供されるLambdaLayerを使用して移行した。
    * 今後、Goでも最新のADOTの実装手順に対応したら、template.yamlなどの実装を変更する必要がありそう。
1. コールドスタートの処理時間が、かなり遅くなっていた気がする
    * APは変わっていないのに、以前に比べてコールドスタートが体感3～5倍遅くなっている気がする。実際、X-Rayのトレースの帯もそれくらいの時間、長くなっています。
　正直、あまり最近使っていないかったので、各種ライブラリを最新バージョンアップしているのでその影響もあるのかもですが、
　ADOTのCollectorがサイドカーに追加されたことで、その初期化で遅くなっているんじゃないかと思ったりもしています。
　ADOT移行のときには、単性能検証・テスト・メモリの調整（必要に応じて複合性能テストも）必要になるんじゃないかと思いました。

1. X-Rayのトレースマップの見栄えが悪くなる
    * 今後改善されるかもしれないが、X-Ray SDKと比較してトレースマップの見栄えが悪くなってしまい、DynamoDB、SQS、S3固有のアイコンが表示されず、汎用のDBアイコンや、歯車になってしまう。
    * 以前のX-Ray SDKの場合とのトレースの図のキャプチャを比較は[README](README.md#adotx-rayによるトレース情報の可視化)を参照。

1. AWS Lambda Go API Proxyのアーカイブ化と今後
    * ADOT自体とは直接関係ないが、[AWS Lambda Go API Proxy](https://github.com/awslabs/aws-lambda-go-api-proxy)は、2025年5月22日にアーカイブ化されてしまった。* 当該Issue(https://github.com/awslabs/aws-lambda-go-api-proxy/issues/143)の記載から、Lambda Web Adapter(https://github.com/aws/aws-lambda-web-adapter)の利用が有力な選択肢の一つと考えられるが、[Golang gin in Zip example](https://github.com/aws/aws-lambda-web-adapter/tree/main/examples/gin-zip)のサンプルコードを見ると、main関数はginを起動しlambda.Start関数を使わない実装になってしまうため、前述のLambdaのmain関数へのADOTの手動計装(https://docs.aws.amazon.com/xray/latest/devguide/manual-instrumentation-go.html#lambda-instrumentation)ができず、ADOTと実装の相性が悪いように見える。Lambda Web Adapterを使う場合、otelgin(https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin)を使えば、Lambdaの手動計装を実装せずともうまくトレースできるか？等、動作確認してみる必要があると思われる（が、この場合もコールドスタート影響が気になる）