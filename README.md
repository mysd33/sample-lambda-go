# Private APIでのAPIGatewayを使ったLambda/GoのAWS SAMサンプルAP

## 構成イメージ
* API GatewayをPrivate APIで公開
    * VPC内にEC2で構築した、Bastionからアクセスする
* LambdaからDynamoDBやRDS AuroraへのDBアクセスを実現
    * LambdaはVPC内Lambdaとして、RDS Aurora（RDS Proxy経由）でのアクセスも可能としている
* AWS SDK for Go v2に対応
    * AWS SDKやX-Ray SDKの利用方法がv1の時と変更になっている

![構成イメージ](image/demo.png)

* X-Rayによる可視化
    * API Gateway、Lambdaにおいて、X-Rayによる可視化にも対応している
    * RDB(RDS Aurora)へのアクセス、DynamoDBへのアクセスのトレースにも対応

![X-Rayの可視化の例](image/xray-aurora.png)
![X-Rayの可視化の例2](image/xray-dynamodb.png)


* RDS Proxyの利用時の注意（ピン留め）
    * SQLを記載するにあたり、従来はプリペアドステートメントを使用するのが一般的であるが、RDS Proxyを使用する場合には、[ピン留め(Pinning)](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy-managing.html#rds-proxy-pinning)という現象が発生してしまう。その間、コネクションが切断されるまで占有されつづけてしまい再利用できず、大量のリクエストを同時に処理する場合にはコネクション枯渇し性能面に影響が出る恐れがある。
        * ピン留めが発生してるかについては、CloudWatch Logsでロググループ「/aws/rds/proxy/demo-rds-proxy」を確認して、以下のような文言が出ていないか確認するとよい。

        ```
        The client session was pinned to the database connection [dbConnection=…] for the remainder of the session. The proxy can't reuse this connection until the session ends. Reason: A parse message was detected.
        ```

    * 本サンプルAPのRDBアクセス処理では、プリペアドステートメントを使用しないよう実装することで、ピン留めが発生しないようにしている。注意点として、SQLインジェクションが起きないようにエスケープ処理を忘れずに実装している。

    * なお、本サンプルAPのようにX-Ray SDKでSQLトレースする場合、[xray.SQLContext関数を利用する](https://docs.aws.amazon.com/ja_jp/xray/latest/devguide/xray-sdk-go-sqlclients.html)際に確立するDBコネクションでピン留めが発生する。
        * xray.SQLContext関数を利用する際に、内部で発行されるSQL（"SELECT version(), current_user, current_database()"）がプリペアドステートメントを使用しているためピン留めが発生する。ピン留めの発生は回避できないとのこと。ただ、CloudWatchのRDS Proxyのログを見ても分かるが、直ちにコネクション切断されるため、ピン留めによる影響は小さいと想定される。（AWSサポート回答より）

## 事前準備
* ローカル環境に、AWS CLI、AWS SAM CLI、Docker環境が必要

## 1. IAMの作成
```sh
#cfnフォルダに移動
cd cfn
aws cloudformation validate-template --template-body file://cfn-iam.yaml
aws cloudformation create-stack --stack-name Demo-IAM-Stack --template-body file://cfn-iam.yaml --capabilities CAPABILITY_IAM
```

## 2. VPCおよびサブネット、InternetGateway等の作成
```sh
aws cloudformation validate-template --template-body file://cfn-vpc.yaml
aws cloudformation create-stack --stack-name Demo-VPC-Stack --template-body file://cfn-vpc.yaml
```

## 3. Security Groupの作成
```sh
aws cloudformation validate-template --template-body file://cfn-sg.yaml
aws cloudformation create-stack --stack-name Demo-SG-Stack --template-body file://cfn-sg.yaml
```

## 4. VPC Endpointの作成とプライベートサブネットのルートテーブル更新
* VPC内LambdaからDynamoDBへアクセスするためのVPC Endpointを作成
```sh
aws cloudformation validate-template --template-body file://cfn-vpe.yaml
aws cloudformation create-stack --stack-name Demo-VPE-Stack --template-body file://cfn-vpe.yaml
```
## 5. NAT Gatewayの作成とプライベートサブネットのルートテーブル更新
* VPC内Lambdaからインターネットに接続する場合に必要となる。
* hello-worldのサンプルAPでは、[https://checkip.amazonaws.com](https://checkip.amazonaws.com)へアクセスしに行くので、これを試す場合には作成が必要となる。

```sh
aws cloudformation validate-template --template-body file://cfn-ngw.yaml
aws cloudformation create-stack --stack-name Demo-NATGW-Stack --template-body file://cfn-ngw.yaml
```

## 6. RDS Aurora Serverless v2 for PostgreSQL、SecretsManager、RDS Proxy作成
* リソース作成に少し時間がかかる。(20分程度)
```sh
aws cloudformation validate-template --template-body file://cfn-rds.yaml
aws cloudformation create-stack --stack-name Demo-RDS-Stack --template-body file://cfn-rds.yaml --parameters ParameterKey=DBUsername,ParameterValue=postgres ParameterKey=DBPassword,ParameterValue=password
```

## 7. EC2(Bastion)の作成
* psqlによるRDBのテーブル作成や、APIGatewayのPrivate APIにアクセスするための踏み台を作成
```sh
aws cloudformation validate-template --template-body file://cfn-bastion-ec2.yaml
aws cloudformation create-stack --stack-name Demo-Bastion-Stack --template-body file://cfn-bastion-ec2.yaml
```

* 必要に応じてキーペア名等のパラメータを指定
    * 「--parameters ParameterKey=KeyPairName,ParameterValue=myKeyPair」

## 8. RDBのテーブル作成
* マネージドコンソールからEC2にセッションマネージャで接続し、Bastionにログインする。psqlをインストールし、DB接続する。
    * 以下参考に、Bastionにpsqlをインストールするとよい
        * https://techviewleo.com/how-to-install-postgresql-database-on-amazon-linux/
* DB接続後、ユーザテーブルを作成する。        
```sh
sudo amazon-linux-extras install -y epel

sudo tee /etc/yum.repos.d/pgdg.repo<<EOF
[pgdg14]
name=PostgreSQL 14 for RHEL/CentOS 7 - x86_64
baseurl=http://download.postgresql.org/pub/repos/yum/14/redhat/rhel-7-x86_64
enabled=1
gpgcheck=0
EOF

sudo yum makecache
sudo yum install -y postgresql14

#Auroraに直接接続
#CloudFormationのDemo-RDS-Stackスタックの出力「RDSClusterEndpointAddress」の値を参照
psql -h (Auroraのクラスタエンドポイント) -U postgres -d testdb    

#ユーザテーブル作成
CREATE TABLE IF NOT EXISTS m_user (user_id VARCHAR(50) PRIMARY KEY, user_name VARCHAR(50));
#ユーザテーブルの作成を確認
\dt
#いったん切断
\q

#RDS Proxyから接続しなおす
#CloudFormationのDemo-RDS-Stackスタックの出力「RDSProxyEndpoint」の値を参照
psql -h (RDS Proxyのエンドポイント) -U postgres -d testdb
#ユーザテーブルの存在を確認
\dt

```

## 9. DynamoDBのテーブル作成
* DynamoDBにTODOテーブルを作成する。
```sh
aws cloudformation validate-template --template-body file://cfn-dynamodb.yaml
aws cloudformation create-stack --stack-name Demo-DynamoDB-Stack --template-body file://cfn-dynamodb.yaml
```


## 10. AWS SAMでLambda/API Gatewayのデプロイ       
* SAMビルド    
```sh
# トップのフォルダに戻る
cd ..
sam build
# Windowsでもmakeをインストールすればmakeでいけます
make
```

* SAMデプロイ
```sh
# 1回目は
sam deploy --guided
# Windowsでもmakeをインストールすればmakeでいけます
make deploy_guided

# 2回目以降は
sam deploy
# Windowsでもmakeをインストールすればmakeでいけます

make deploy
```

* （参考）再度ビルドするとき
```sh
# .aws-sam配下のビルド資材を削除
rmdir /s /q .aws-sam
# ビルド
sam build

# Windowsでもmakeをインストールすればmakeでいけます
make clean
make
```


## 11. APの実行確認
* マネージドコンソールから、EC2(Bation)へSystems Manager Session Managerで接続して、curlコマンドで動作確認
    * 以下の実行例のURLを、sam deployの結果出力される実際のURLをに置き換えること

* hello-worldのAPI実行例    
```sh
curl https://5h5zxybd3c.execute-api.ap-northeast-1.amazonaws.com/Prod/hello

# 接続元Public IPアドレス（この例では、NAT Gatewayのもの）を返却
Hello, 18.180.139.158
```

* Userサービスでユーザ情報を登録するPOSTのAPI実行例
    * UserサービスはRDB(RDS Proxy経由でAuroraへ)アクセスするLambda/goのサンプルAP
```sh
curl -X POST -H "Content-Type: application/json" -d '{ "user_name" : "Taro"}' https://42b4c7bk9g.execute-api.ap-northeast-1.amazonaws.com/Prod/users

# 登録結果を返却
{"user_id":"99bf4d94-f6a4-11ed-85ec-be18af968bc1","user_name":"Taro"}
```

* Userサービスでユーザー情報を取得するGetのAPIの実行例（users/の後にPOSTのAPIで取得したユーザIDを指定）
```sh
curl https://42b4c7bk9g.execute-api.ap-northeast-1.amazonaws.com/Prod/users/99bf4d94-f6a4-11ed-85ec-be18af968bc1

# 対象のユーザ情報をRDBから取得し返却
{"user_id":"99bf4d94-f6a4-11ed-85ec-be18af968bc1","user_name":"Taro"}
```

* Todoサービスでやることリストを登録するPOSTのAPI実行例
    * TodoサービスはDynamoDBアクセスするLambda/goのサンプルAP
```sh
curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "ミルクを買う"}' https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/todo

# 登録結果を返却
{"todo_id":"04a14ad3-f6a5-11ed-b40f-f2ead45b980a","todo_title":"ミルクを買う"}
```

* Todoサービスでやること（TODO）を取得するGetのAPI実行例（todo/の後にPOSTのAPIで取得したTodo IDを指定）
```sh
curl https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/todo/04a14ad3-f6a5-11ed-b40f-f2ead45b980a

# 対象のやることをDyanamoDBから取得し返却
{"todo_id":"04a14ad3-f6a5-11ed-b40f-f2ead45b980a","todo_title":"ミルクを買う"}
```
## 12. SAMのCloudFormationスタック削除
* VPC内Lambdaが参照するHyperplane ENIの削除に最大20分かかるため、スタックの削除に時間がかかる。
```sh
sam delete
# Windowsでもmakeをインストールすればmakeでいけます
make delete
```

## 13. その他リソースのCloudFormationスタック削除
```sh
aws cloudformation delete-stack --stack-name Demo-Bastion-Stack
aws cloudformation delete-stack --stack-name Demo-DynamoDB-Stack
aws cloudformation delete-stack --stack-name Demo-RDS-Stack
aws cloudformation delete-stack --stack-name Demo-NATGW-Stack
aws cloudformation delete-stack --stack-name Demo-VPE-Stack
aws cloudformation delete-stack --stack-name Demo-SG-Stack
aws cloudformation delete-stack --stack-name Demo-VPC-Stack 
aws cloudformation delete-stack --stack-name Demo-IAM-Stack 
```

## ローカルでの実行確認
* AWS上でLambda等をデプロイしなくてもsam localコマンドを使ってローカル実行確認も可能である

* Postgres SQLのDockerコンテナを起動
```sh
docker run --name test-postgres -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres
#Postgresのコンテナにシェルで入って、psqlコマンドで接続
docker exec -i -t test-postgres /bin/bash
> psql -U postgres

# psqlで、testdbデータベースを作成
postgres> CREATE DATABASE testdb;
# testdbに切替
postgres> \c testdb
#ユーザテーブル作成
tesdb> CREATE TABLE IF NOT EXISTS m_user (user_id VARCHAR(50) PRIMARY KEY, user_name VARCHAR(50));
#ユーザテーブルの作成を確認
tesdb> \dt
#切断
tesdb> \q
```

* DynamoDB LocalのDockerコンテナを起動
```sh
cd dynamodb-local
docker-compose up
```

* dynamodb-adminでtodoテーブルを作成
    * [dynamodb-admin](https://github.com/aaronshaf/dynamodb-admin)をインストールし、起動する
        * [dynamodb-adminのインストール＆起動方法](https://github.com/aaronshaf/dynamodb-admin#use-as-globally-installed-app)

        ```
        dynamodb-admin
        ```

    * ブラウザで[http://localhost:8001/](http://localhost:8001/)にアクセスし「Create Table」ボタンをクリック    
    * 「Table Name」…「todo」、「Hash Attribute Name」…「todo_id」、「Hash Attribute Type」…「String」で作成

* sam localコマンドを実行
    * local-env.jsonファイルに、上書きする環境変数が記載されている

```sh
sam local start-api --env-vars local-env.json

# Windowsでもmakeをインストールすればmakeでいけます
make local_startapi 
```

* APの動作確認
```sh
curl http://127.0.0.1:3000/hello

curl -X POST -H "Content-Type: application/json" -d '{ "user_name" : "Taro"}' http://127.0.0.1:3000/users

curl http://127.0.0.1:3000/users/(ユーザID)

curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "Buy Milk"}' http://127.0.0.1:3000/todo

curl http://127.0.0.1:3000/todo/(TODO ID)
```

## godocの表示
* godocをインストール
```sh
go install golang.org/x/tools/cmd/godoc@latest     
```

* 使い方は、[godoc](https://pkg.go.dev/golang.org/x/tools/cmd/godoc)を参照のこと

* appフォルダでgodocコマンドを実行
```sh
cd app
godoc
```

* appbaseフォルダでgodocコマンドを実行    
```sh
cd app
godoc
```

- godoc起動中の状態で、[http://localhost:6060](http://localhost:6060)へアクセス
    - 「example.com/」の「appbase」パッケージは、[http://localhost:6060/pkg/example.com/appbase/](http://localhost:6060/pkg/example.com/appbase/)に表示される。
    - 「app」パッケージは、ほぼ全てが「internal」パッケージに配置しているため、デフォルトでは表示されない。m=allをクエリパラメータに指定して、[http://localhost:6060/pkg/app/?m=all](http://localhost:6060/pkg/app/?m=all)にアクセスするとよい。

## ソフトウェアフレームワーク
* 本サンプルアプリケーションでは、ソフトウェアフレームワーク実装例も同梱している。簡単のため、アプリケーションと同じプロジェクトでソース管理している。
* ソースコードはappbaseフォルダ配下にexample.com/appbaseパッケージとして格納されている。    
    * 本格的な開発を実施する場合には、業務アプリケーションと別のGitリポジトリとして管理し、参照するようにすべきであるが、ここでは、あえて同じプロジェクトに格納してノウハウを簡単に参考にしてもらいやすいようにしている。
* 各機能と実現方式は、以下の通り。

| 機能 | 機能概要と実現方式 | 拡張実装 | 拡張実装の格納パッケージ |
| ---- | ---- | ---- | ---- |
| オンラインAP制御 | APIの要求受信、ビジネスロジック実行、応答返却まで一連の定型的な処理を実行を制御する共通機能を提供する。AWS Lambda Go API Proxyを利用してginと統合し実現する。 | ○ | com.example/appbase/pkg/api<br/>com.example/appbase/pkg/domain |
| 入力チェック| APIのリクエストデータの入力チェックを実施する、ginのバインディング機能でgo-playground/validator/v10を使ったバリデーションを実現する。 | ○ | - |
| エラー（例外） | エラーコード（メッセージID）やメッセージを管理可能な共通的なビジネスエラー、システムエラー用のGoのErrorオブジェクトを提供する。 | ○ | com.example/appbase/pkg/errors |
| 集約例外ハンドリング | オンラインAP制御機能と連携し、エラー（例外）発生時、エラーログの出力、DBのロールバック、エラー画面やエラー電文の返却といった共通的なエラーハンドリングを実施する。 | ○ | com.example/appbase/pkg/interceptor |
| RDBアクセス | go標準のdatabase/sqlパッケージを利用しRDBへアクセスする。DB接続等の共通処理を個別に実装しなくてもよい仕組みとする。 | ○ | com.example/appbase/pkg/rdb |
| RDBトランザクション管理機能 | オンラインAP制御機能と連携し、サービスクラスの実行前後にRDBのトランザクション開始・終了を機能を提供する。 | ○ | com.example/appbase/pkg/rdb |
| DynamoDBアクセス | AWS SDKを利用しDynamoDBへアクセスする。 | - | - |
| 分散トレーシング（X-Ray） | AWS X-Rayを利用して、サービス間の分散トレーシング・可視化を実現する。実現には、AWS SAMのtemplate.ymlでの設定やSDKが提供する各withContextメソッドといった利用する。なお、Contextをメソッドの引数に引き渡さなくても取得できるようにグローバル変数で管理する。 | ○ | com.example/appbase/pkg/apcontext |
| ロギング | go.uber.org/zapの機能を利用し、プロファイルによって動作環境に応じたログレベルや出力先（ファイルや標準出力）、出力形式（タブ区切りやJSON）に切替可能とする。またメッセージIDをもとにログ出力可能な汎用的なAPIを提供する。 | ○ | com.example/appbase/pkg/logging |
| プロパティ管理 | spf13/viperの機能を利用し、APから環境依存のパラメータを切り出し、プロファイルによって動作環境に応じたパラメータ値に置き換え可能とする。 | ○ | com.example/appbase/pkg/config |

* 以下は、今後追加を検討中。

| 機能 | 機能概要と実現方式 | 拡張実装 | 拡張実装の格納パッケージ |
| ---- | ---- | ---- | ---- |
| DynamoDBトランザクション管理機能 | オンラインAP制御機能と連携し、サービスクラスの実行前後にRDBのトランザクション開始・終了を機能を提供する。 | ○ | com.example/appbase/pkg/dynamodb |
| メッセージ管理 | go標準のembededでログ等に出力するメッセージを設定ファイルで一元管理する。 | ○ | 未定 |
| API認証・認可| APIGatewayのCognitoオーサライザまたはLambdaオーサライザを利用し、APIの認証、認可を行う。 | ○ | 未定 |