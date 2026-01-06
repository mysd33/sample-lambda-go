# Lambda/GoのAWS SAMサンプルAP
## 構成イメージ
* オンラインリアルタイム処理方式
    * API GatewayをトリガにLambda実行
    * フロントエンドは、Regional Public APIで公開し、バックエンドはPrivate APIで公開
        * バックエンドは、動作確認用にVPC内にEC2で構築したBastionからのアクセスにも対応
    * LambdaからDynamoDBやRDS Aurora、DocumentDBといったDBアクセスへのアクセスを実現
    * LambdaはVPC内Lambdaとして、RDS Aurora（RDS Proxy経由）、DocumentDBへのアクセスも可能としている

* ディレード処理方式
    * Lambdaから、SQSへのアクセスし、非同期処理の実行依頼を実現
    * SQSトリガにLambda実行
    * 標準キュー、FIFOキューの両方に対応
    * DynamoDBのトランザクション管理機能を利用した、メッセージ送達とDynamoDBトランザクションの整合性を担保する仕組みを実装

* 純バッチ処理方式
    * EventBridgeによるスケジュール起動によりLambda実行
    * ジョブのフロー制御が必要な場合は、EventBridgeからStep Functionsを起動し、各ステートでLambda関数を実行する
    * TODO: サンプルは未実装

* 本サンプルAPのソフトウェアアーキテクチャの図は以下の通り。

    ![ソフトウェアアーキテクチャ](image/architecture.png)


* LambdaからAWS SDKを用いたDynamoDB、SQS、S3等の各種AWSリソースへのアクセスに対応
    * AWS SDK for Go v2に対応した実装
        * v2では、AWS SDKやX-Ray SDKの利用方法がv1の時とAPIがかなり変更されている    

![構成イメージ](image/demo.png)

* Lambda間の呼び出しイメージ
    * フロントエンド（BFF:Backend For Frontend）からバックエンドの各サービスへアクセスする、SQSを介してディレード実行するという呼び出し関係になっている
    * User API、Todo APIサービスはバックエンドサービス扱いで、bationから直接アクセスできるようにもなっている

![呼び出しイメージ](image/demo2.png)

* AppConfigによる設定の外部化
    * [AppConfig](https://docs.aws.amazon.com/ja_jp/appconfig/latest/userguide/what-is-appconfig.html)を使用し、APから外部管理された設定の取得、AppConfig機能を使ったデプロイに対応している。
    * マネージドなLambdaレイヤにより提供される[AppConfig Agent Lambdaエクステンション](https://docs.aws.amazon.com/ja_jp/appconfig/latest/userguide/appconfig-integration-lambda-extensions.html)を使って、LambdaアプリケーションからAppConfigの設定をキャッシュするとともに、アプリケーションの再デプロイ不要で設定変更を反映することができる。

![AppConfigイメージ](image/demo3.png)

* X-Rayによる可視化
    * API Gateway、Lambdaにおいて、X-Rayによる可視化にも対応している
    * RDB(RDS Aurora)、DynamoDBへのアクセス、REST API、SQSの呼び出しのトレースにも対応
    * RDB(Aurora)アクセスの可視化の例
        ![X-Rayの可視化の例](image/xray-aurora.png)
    * DynamoDBアクセスの可視化の例
        ![X-Rayの可視化の例2](image/xray-dynamodb.png)
    * REST APIの呼び出しの可視化の例
        ![X-Rayの可視化の例3](image/xray-bff.png)
    * SQS、S3等の呼び出しの可視化の例
        ![X-Rayの可視化の例4](image/xray-sqs-delayed.png)
    * DocumentDBの呼び出しは、mongo-go-driverがX-Ray SDKに対応していないとのことで、未実施。
        * [issue](https://github.com/aws/aws-xray-sdk-go/issues/348)
        
* RDS Proxyの利用時の注意
    * ピン留め
        * SQLを記載するにあたり、従来はプリペアドステートメントを使用するのが一般的であるが、RDS Proxyを使用する場合には、[ピン留め(Pinning)](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy-managing.html#rds-proxy-pinning)という現象が発生してしまう。その間、コネクションが切断されるまで占有されつづけてしまい再利用できず、大量のリクエストを同時に処理する場合にはコネクション枯渇し性能面に影響が出る恐れがある。
            * ピン留めが発生してるかについては、CloudWatch Logsでロググループ「/aws/
            rds/proxy/demo-rds-proxy」を確認して、以下のような文言が出ていないか確認するとよい。

            ```
            The client session was pinned to the database connection [dbConnection=…] for the remainder of the session. The proxy can't reuse this connection until the session ends. Reason: A parse message was detected.
            ```

        * 本サンプルAPのRDBアクセス処理では、プリペアドステートメントを使用しないよう実装することで、ピン留めが発生しないようにしている。この際、注意点として、SQLインジェクションが起きないようにエスケープ処理を忘れずに実装している。

        * なお、本サンプルAPのようにX-Ray SDKでSQLトレースする場合、[xray.SQLContext関数を利用する](https://docs.aws.amazon.com/ja_jp/xray/latest/devguide/xray-sdk-go-sqlclients.html)際に確立するDBコネクションでピン留めが発生する。
            * xray.SQLContext関数を利用する際に、内部で発行されるSQL（"SELECT version(), current_user, current_database()"）がプリペアドステートメントを使用しているためピン留めが発生する。ピン留めの発生は回避できないとのこと。ただ、CloudWatchのRDS Proxyのログを見ても分かるが、直ちにコネクション切断されるため、ピン留めによる影響は小さいと想定される。（AWSサポート回答より）
    * RDBトランザクションの利用
        * RDS Proxy経由で接続する場合、１つのトランザクション内での呼び出しは、同じコネクションを使用する。auto commit無効の場合は、トランザクションが終了（commit/rollback）するまで、接続の再利用は行われない。このため、前述のプリペアドステートメントによるピン留めを過度に気にする必要はないとも思える。いずれにしても、ピン留めを理解した上で、トランザクションとRDBコネクションの管理を統制することが大事である。
            * https://pages.awscloud.com/rs/112-TZM-766/images/EV_amazon-rds-aws-lambda-update_Jul28-2020_RDS_Proxy.pdf
	            * pp.12-13

## 事前準備
* ローカル環境に、Go、AWS CLI、AWS SAM CLI、Docker環境が必要です。
    * Goのバージョンは1.22.6を使用しています。

## 1. IAMの作成
```sh
#cfnフォルダに移動
cd cfn
aws cloudformation validate-template --template-body file://cfn-iam.yaml
aws cloudformation create-stack --stack-name Demo-IAM-Stack --template-body file://cfn-iam.yaml --capabilities CAPABILITY_NAMED_IAM
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
* API GatewayのPrivate APIのためのVPC Endpointや、VPC内LambdaからDynamoDB、SQS、AppConfig等へアクセスするためのVPC Endpointを作成
```sh
aws cloudformation validate-template --template-body file://cfn-vpe.yaml
aws cloudformation create-stack --stack-name Demo-VPE-Stack --template-body file://cfn-vpe.yaml
```
## 5. NAT Gatewayの作成とプライベートサブネットのルートテーブル更新
* VPC内Lambdaからインターネットに接続する場合に必要となる。
* hello-worldのサンプルAPでは[https://checkip.amazonaws.com](https://checkip.amazonaws.com)へアクセスしに行くためのみに必要なので、もしhello-worldのサンプルAPの確認が不要な場合は、作成不要。

```sh
aws cloudformation validate-template --template-body file://cfn-ngw.yaml
aws cloudformation create-stack --stack-name Demo-NATGW-Stack --template-body file://cfn-ngw.yaml
```

## 6. SecretsManagerの作成
* Aurora、DocumentDBの認証情報をSecretsManagerに作成する。
```sh
aws cloudformation validate-template --template-body file://cfn-secrets.yaml
aws cloudformation create-stack --stack-name Demo-SM-Stack --template-body file://cfn-secrets.yaml --parameters ParameterKey=AuroraDBUsername,ParameterValue=postgres ParameterKey=DocDBUsername,ParameterValue=root
```
* SecretsManagerが生成したパスワード等を確認しておく
```sh
# Aurora
aws secretsmanager get-secret-value --secret-id Demo-RDS-Secrets
# DocomentDB
aws secretsmanager get-secret-value --secret-id Demo-DocDB-Secrets
```

## 7. RDS Aurora Serverless v2 for PostgreSQL、RDS Proxy作成
* リソース作成に少し時間がかかる。(20分程度)
```sh
aws cloudformation validate-template --template-body file://cfn-rds.yaml
aws cloudformation create-stack --stack-name Demo-RDS-Stack --template-body file://cfn-rds.yaml
```

## 8. DocumentDB作成
* リソース作成に少し時間がかかる。（5分程度）
```sh
aws cloudformation validate-template --template-body file://cfn-documentdb.yaml
aws cloudformation create-stack --stack-name Demo-DocumentDB-Stack --template-body file://cfn-documentdb.yaml
```

## 9. EC2(Bastion)の作成
* psqlによるRDBのテーブル作成や、APIGatewayのPrivate APIにアクセスするための踏み台を作成
```sh
aws cloudformation validate-template --template-body file://cfn-bastion-ec2.yaml
aws cloudformation create-stack --stack-name Demo-Bastion-Stack --template-body file://cfn-bastion-ec2.yaml
```

* 必要に応じてキーペア名等のパラメータを指定
    * 「--parameters ParameterKey=KeyPairName,ParameterValue=myKeyPair」

## 10. RDBのテーブル作成
* マネージドコンソールからEC2にセッションマネージャで接続し、Bastionにログインする。psqlをインストールし、DB接続する。
    * 以下参考に、Bastionにpsqlをインストールするとよい
        * https://docs.aws.amazon.com/ja_jp/AmazonRDS/latest/UserGuide/CHAP_GettingStarted.CreatingConnecting.PostgreSQL.html#CHAP_GettingStarted.Connecting.PostgreSQL
* DB接続後、ユーザテーブルを作成する。        

```sh
sudo dnf update -y

sudo dnf install postgresql16 -y

#Auroraに直接接続
#CloudFormationのDemo-RDS-Stackスタックの出力「RDSClusterEndpointAddress」の値を参照
psql -h (Auroraのクラスタエンドポイント) -U postgres -d testdb    

#パスワードは、SecretsManagerの「Demo-RDS-Secrets」の「password」の値を参照し入力

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

## 11. DocumentDBのデータベース作成
* マネージドコンソールからEC2にセッションマネージャで接続し、Bastionにログインする。mongoシェルをインストールし、DB接続する。
    * 以下参考に、Bastionにmongoシェルをインストールするとよい
        * https://docs.aws.amazon.com/ja_jp/documentdb/latest/developerguide/get-started-guide.html#get-start-mongoshell

```sh
echo -e "[mongodb-org-5.0] \nname=MongoDB Repository\nbaseurl=https://repo.mongodb.org/yum/amazon/2023/mongodb-org/5.0/x86_64/\ngpgcheck=1 \nenabled=1 \ngpgkey=https://pgp.mongodb.com/server-5.0.asc" | sudo tee /etc/yum.repos.d/mongodb-org-5.0.repo

sudo yum install -y mongodb-org-shell
```

* サイトを参考に、DocumentDB 認証局 (CA) 証明書をダウンロードし、[mongo シェルでこのクラスターに接続する] で、提供された接続文字列をコピーし、DocumentDBに接続する。
    * https://docs.aws.amazon.com/ja_jp/documentdb/latest/developerguide/get-started-guide.html#get-start-connectcluster

```sh
sudo su ec2-user
```

```sh
cd ~
wget https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem

# mongo --ssl --host (DocumentDBのエンドポイント) --sslCAFile global-bundle.pem --username root --password (パスワード)
# パスワードは、SecretsManagerの「Demo-DocDB-Secrets」の「password」の値を参照し入力
# 例
mongo --ssl --host demo-documentdbcluster.cluster-crby0oqsjgv1.ap-northeast-1.docdb.amazonaws.com:27017 --sslCAFile global-bundle.pem --username root --password password

# sampledbという名前のデータベースを作成
> use sampledb

# 簡単な動作確認
# booksという名前のコレクションにサンプルデータ挿入
> db.books.insert({ "title": "こころ", "author": "夏目漱石", "publisher": "新潮社", "published_date": "2004-03-01", "isbn": "111-1-1111-1111-1" })
> db.books.insert({ "title": "坊っちゃん", "author": "夏目漱石", "publisher": "新潮社", "published_date": "2003-04-01", "isbn": "111-1-1111-1111-2" })

# booksコレクションのデータを表示
> db.books.find()

# 以下のように表示されればOK
{ "_id" : ObjectId("67352225289897f98f461d35"), "title" : "こころ", "author" : "夏目漱石", "publisher" : "新潮社", "published_date" : "2004-03-01", "isbn" : "111-1-1111-1111-1" }
{ "_id" : ObjectId("6735222c289897f98f461d36"), "title" : "坊っちゃん", "author" : "夏目漱石", "publisher" : "新潮社", "published_date" : "2003-04-01", "isbn" : "111-1-1111-1111-2" }

# booksコレクションへのクエリ（タイトルがこころを検索）
> db.books.find({"title": "こころ"}).pretty()

# 以下のように表示されればOK
{
        "_id" : ObjectId("67352225289897f98f461d35"),
        "title" : "こころ",
        "author" : "夏目漱石",
        "publisher" : "新潮社",
        "published_date" : "2004-03-01",
        "isbn" : "111-1-1111-1111-1"
}

# データベースの表示（sampledbがあることを確認）
show dbs
```

## 12. DynamoDBのテーブル作成
* DynamoDBに「todo」、「temp」、「queue_message」、「idempotency」の各テーブルを作成する。
```sh
aws cloudformation validate-template --template-body file://cfn-dynamodb.yaml
aws cloudformation create-stack --stack-name Demo-DynamoDB-Stack --template-body file://cfn-dynamodb.yaml
```

## 13. SQSの作成
```sh
aws cloudformation validate-template --template-body file://cfn-sqs.yaml
aws cloudformation create-stack --stack-name Demo-SQS-Stack --template-body file://cfn-sqs.yaml
```

## 14. S3の作成
```sh
aws cloudformation validate-template --template-body file://cfn-s3.yaml
aws cloudformation create-stack --stack-name Demo-S3-Stack --template-body file://cfn-s3.yaml
```

## 15. AppConfigの作成
* AppConfigの基本リソースを作成する。
```sh
aws cloudformation validate-template --template-body file://cfn-appconfig.yaml
aws cloudformation create-stack --stack-name Demo-AppConfig-Stack --template-body file://cfn-appconfig.yaml
```

## 16. AWS SAMでLambda/API Gatewayの手動でのAPのビルド・デプロイ     
* CI/CDを行う場合は、この手順は不要で [CI/CDの手順](#17-cicd)を実施してください。

* SAMビルド    
```sh
# トップのフォルダに戻る
cd ..

# ビルド
sam build
xcopy /I /S configs .aws-sam\build\BffFunction\configs
xcopy /I /S configs .aws-sam\build\UsersFunction\configs	
xcopy /I /S configs .aws-sam\build\TodoFunction\configs	
xcopy /I /S configs .aws-sam\build\TodoAsyncFunction\configs
xcopy /I /S configs .aws-sam\build\TodoAsyncFifoFunction\configs
xcopy /I /S configs .aws-sam\build\BooksFunction\configs	

# Windowsでもmakeをインストールすればmakeだけでいけます
make
```

* （参考）修正後再度ビルドするとき
```sh
# .aws-sam配下のビルド資材を削除
rmdir /s /q .aws-sam
# ビルド
sam build
xcopy /I /S configs .aws-sam\build\BffFunction\configs
xcopy /I /S configs .aws-sam\build\UsersFunction\configs	
xcopy /I /S configs .aws-sam\build\TodoFunction\configs	
xcopy /I /S configs .aws-sam\build\TodoAsyncFunction\configs
xcopy /I /S configs .aws-sam\build\TodoAsyncFifoFunction\configs
xcopy /I /S configs .aws-sam\build\BooksFunction\configs	

# Windowsでもmakeをインストールすればmakeだけでいけます
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

* （参考）samconfig.tomlで定義した特定の環境を指定する場合
```sh
sam deploy --config-env (環境名)
# 例
sam deploy --config-env prd

# Windowsでもmakeをインストールすればmakeでいけます
make deploy_env env=(環境名)
# 例
make deploy_env env=prd
```

## 17. CI/CD
* 作成中

* CIジョブ用のCodeBuildの作成

```sh
#cfnフォルダに移動
cd cfn
aws cloudformation validate-template --template-body file://cfn-codebuild-ci.yaml
aws cloudformation create-stack --stack-name Demo-CodeBuild-CI-Stack --template-body file://cfn-codebuild-ci.yaml
```

* CDジョブ用のCodeBuildの作成
```sh
TBD
```

* CodePipelineの作成
    * CIジョブとCDジョブを組み合わせたCodePipelineを作成
    * 実行すると、CIジョブが実行され、その後、SAMのアプリケーションがデプロイされる予定。
```sh
aws cloudformation validate-template --template-body file://cfn-codepipeline.yaml
aws cloudformation create-stack --stack-name Demo-CodePipeline-Stack --template-body file://cfn-codepipeline.yaml
```

## 18. AppConfigのデプロイ
* Hosted Configurationの設定バージョンの作成と初回デプロイする。
```sh
#cfnフォルダに移動
cd cfn
aws cloudformation validate-template --template-body file://cfn-appconfig-hosted-deploy.yaml
aws cloudformation create-stack --stack-name Demo-AppConfigHostedDeploy-Stack --template-body file://cfn-appconfig-hosted-deploy.yaml
```

* RDS用のSecretManagerの設定を初回デプロイする。
    * 同一のアプリケーション、環境に対してのデプロイは並列実行できないため、Hosted Configurationの設定のデプロイが完了後に実施すること
    * パラメータのSecretsManagerVersionのバージョンIDは、CLIまたはマネコンで確認してパラメータに設定する
```sh
# シークレットのバージョンIDを確認
aws secretsmanager list-secret-version-ids --secret-id Demo-RDS-Secrets --query Versions[?contains(VersionStages,`AWSCURRENT`)].VersionId

# CloudFormationの実行
aws cloudformation validate-template --template-body file://cfn-appconfig-rds-sm-deploy.yaml
aws cloudformation create-stack --stack-name Demo-AppConfigRDSSMDeploy-Stack --template-body file://cfn-appconfig-rds-sm-deploy.yaml --parameters ParameterKey=SecretsManagerVersion,ParameterValue=（SecretsManagerVersionのバージョンID）
```

* DocumentDB用のSecretManagerの設定を初回デプロイする。
    * 同一のアプリケーション、環境に対してのデプロイは並列実行できないため、RDS用のSecretMaangerの設定のデプロイが完了後に実施すること
    * パラメータのSecretsManagerVersionのバージョンIDは、CLIまたはマネコンで確認してパラメータに設定する
```sh
# シークレットのバージョンIDを確認
aws secretsmanager list-secret-version-ids --secret-id Demo-DocDB-Secrets --query Versions[?contains(VersionStages,`AWSCURRENT`)].VersionId

# CloudFormationの実行
aws cloudformation validate-template --template-body file://cfn-appconfig-docdb-sm-deploy.yaml
aws cloudformation create-stack --stack-name Demo-AppConfigDocDBSMDeploy-Stack --template-body file://cfn-appconfig-docdb-sm-deploy.yaml --parameters ParameterKey=SecretsManagerVersion,ParameterValue=（SecretsManagerVersionのバージョンID）
```


## 19. APの実行確認（バックエンド）
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
curl -X POST -H "Content-Type: application/json" -d '{ "user_name" : "Taro"}' https://42b4c7bk9g.execute-api.ap-northeast-1.amazonaws.com/Prod/users-api/v1/users

# 登録結果を返却
{"user_id":"99bf4d94-f6a4-11ed-85ec-be18af968bc1","user_name":"Taro"}
```

* Userサービスでユーザー情報を取得するGetのAPIの実行例（users/の後にPOSTのAPIで取得したユーザIDを指定）
```sh
curl https://42b4c7bk9g.execute-api.ap-northeast-1.amazonaws.com/Prod/users-api/v1/users/99bf4d94-f6a4-11ed-85ec-be18af968bc1

# 対象のユーザ情報をRDBから取得し返却
{"user_id":"99bf4d94-f6a4-11ed-85ec-be18af968bc1","user_name":"Taro"}
```

* Todoサービスでやることリストを登録するPOSTのAPI実行例
    * TodoサービスはDynamoDBアクセスするLambda/goのサンプルAP
```sh
# DynamoDBトランザクションなし
curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "ミルクを買う"}' https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/todo-api/v1/todo
# DynamoDBトランザクションあり
curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "ミルクを買う"}' https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/todo-api/v1/todo?tx=true

# 登録結果を返却
{"todo_id":"04a14ad3-f6a5-11ed-b40f-f2ead45b980a","todo_title":"ミルクを買う"}
```

* Todoサービスでやること（TODO）を取得するGetのAPI実行例（todo/の後にPOSTのAPIで取得したTodo IDを指定）
```sh
curl https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/todo-api/v1/todo/04a14ad3-f6a5-11ed-b40f-f2ead45b980a

# 対象のやることをDyanamoDBから取得し返却
{"todo_id":"04a14ad3-f6a5-11ed-b40f-f2ead45b980a","todo_title":"ミルクを買う"}
```

* BookサービスのAPI実行例

```sh
# 書籍の登録
curl -X POST -H "Content-Type: application/json" -d '{"title": "こころ", "author": "夏目漱石", "publisher": "新潮社", "published_date": "2004-03-01", "isbn": "111-1-1111-1111-1"}' https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/books-api/v1/books
# タイトルが「こころ」の書籍を取得(title=こころ)
# 以前の手順で、bastionからも「こころ」の書籍情報を登録していると、2件取得される
curl https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/books-api/v1/books?title=%E3%81%93%E3%81%93%E3%82%8D
```

## 20. APの実行確認（フロントエンド）
* 手元の端末のコンソールから、curlコマンドで動作確認
    * 以下の実行例のURLを、sam deployの結果出力される実際のURLをに置き換えること
* Windowsではgit bash等で実行できるが日本語が文字化けするので、PostmanやTalend API Tester等のツールを使ったほうがよい

* BFFサービスのAPI実行例
```sh
# Userサービスを利用し、ユーザー情報を登録
curl -X POST -H "Content-Type: application/json" -d '{ "user_name" : "Taro"}' https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/users

# 登録結果を返却
{"user_id":"416ad789-6fde-11ee-a3ec-0242ac110004","user_name":"Taro"}

# Todoサービスを利用し、やることを登録
curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "ミルクを買う"}' https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/todo

# 登録結果を返却
{"todo_id":"60d48f8f-6fde-11ee-a60c-0242ac110005","todo_title":"ミルクを買う"}

# TodoサービスとUseサービスを利用して、対象ユーザと対象のやることを取得し返却
# curlコマンドの場合は&をエスケープする
#
# curl https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/todo?user_id=（ユーザID）\&todo_id=(TODO ID)
curl https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/todo?user_id=416ad789-6fde-11ee-a3ec-0242ac110004\&todo_id=60d48f8f-6fde-11ee-a60c-0242ac110005

# 対象のユーザ情報とやることを一緒に取得
{"user":{"user_id":"416ad789-6fde-11ee-a3ec-0242ac110004","user_name":"Taro"},"todo":{"todo_id":"60d48f8f-6fde-11ee-a60c-0242ac110005","todo_title":"ミルクを買う"}}
```

```sh
# Bookサービスを利用して、書籍情報を登録
curl -X POST -H "Content-Type: application/json" -d @curl/books/post1.json https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books
curl -X POST -H "Content-Type: application/json" -d @curl/books/post2.json https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books
curl -X POST -H "Content-Type: application/json" -d @curl/books/post3.json https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books
curl -X POST -H "Content-Type: application/json" -d @curl/books/post4.json https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books
curl -X POST -H "Content-Type: application/json" -d @curl/books/post5.json https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books
# 著者が「村上春樹」の書籍を取得（author=村上春樹）
curl https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books?author=%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A8%B9
# 出版社が「新潮社」の書籍を取得（publisher=新潮社）
curl https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books?publisher=%E6%96%B0%E6%BD%AE%E7%A4%BE
# 著者が「村上春樹」で出版社が「講談社」の書籍を取得（author=村上春樹&publisher=講談社）
curl https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books?author=%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A8%B9\&publisher=%E8%AC%9B%E8%AB%87%E7%A4%BE
# Bookサービスを利用して、書籍情報を検索（author=村上春樹&publisher=講談社）
curl https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/books?author=%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A8%B9\&publisher=%E8%AC%9B%E8%AB%87%E7%A4%BE#
```

* ディレード処理の実行例
    * 実行結果はDynamoDBを確認するとよい
```sh
# BFFからの非同期実行依頼（標準キュー）
# 業務のDB更新を伴う場合
curl -X POST -H "Content-Type: application/json" -d '{ "todo_titles" : ["Buy Milk", "Study English"]}' https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/todo-async
# 業務のDB更新を伴わない場合
curl -X POST -H "Content-Type: application/json" -d '{ "todo_titles" : ["Buy Milk", "Study English"]}' https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/todo-async?dbtx=no

# BFFからの非同期実行依頼（FIFOキュー）
# 業務のDB更新を伴う場合
curl -X POST -H "Content-Type: application/json" -d '{ "todo_titles" : ["Buy Milk", "Study English"]}' https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/todo-async?fifo=true
# 業務のDB更新を伴わない場合
curl -X POST -H "Content-Type: application/json" -d '{ "todo_titles" : ["Buy Milk", "Study English"]}' https://adoscoxed14.execute-api.ap-northeast-1.amazonaws.com/Prod/bff-api/v1/todo-async?fifo=true\&dbtx=no

```

## 21. AppConfingの設定変更＆デプロイ
* cfn-appconfig-hosted-deploy.yaml内のホスト化された設定の内容を修正
```yaml
  AppConfigHostedConfigurationVersion:
    Type: AWS::AppConfig::HostedConfigurationVersion
    Properties:
      …
      # Contentを修正
      Content: |
        hoge_name: foo2
        fuga_name: gaa2
```

* 以下のコマンドを実行しAppConfigのスタック更新すると、新しいホスト化された設定が指定したデプロイ戦略に基づき再デプロイされる
```sh
cd cfn
aws cloudformation validate-template --template-body file://cfn-appconfig-hosted-deploy.yaml
aws cloudformation update-stack --stack-name Demo-AppConfigHostedDeploy-Stack --template-body file://cfn-appconfig-hosted-deploy.yaml
```

![AppConfigの設定再デプロイ](image/appconfig.png)

* Userサービスでユーザ情報を登録するPOSTのAPIを呼び出したときの、設定値を出力するCloudWatchのログの変化を確認するとよい。

* /aws/lambda/user-functionロググループのログを確認すると変化が分かる

```sh
#Before: APIを呼び出すと、hoge_nameが「foo」
{"level":"info","ts":1699780051.3576484,"caller":"service/user_service.go:39","msg":"hoge_name=foo"}

#途中でAgentによりAppConfigのキャッシュが更新される 
[appconfig agent] 2023/11/12 13:24:04 INFO updated cache data for 'todo-app:Prod:config' in 135.032ms

#After: APIを呼び出すと、hoge_nameが「foo2」に変化
{"level":"info","ts":1699780051.3576484,"caller":"service/user_service.go:39","msg":"hoge_name=foo2"}
```

## 22. AWSリソースの削除
* AppConfig Deploymentリソースの削除
```sh
aws cloudformation delete-stack --stack-name Demo-AppConfigDocDBSMDeploy-Stack
aws cloudformation delete-stack --stack-name Demo-AppConfigRDSSMDeploy-Stack
aws cloudformation delete-stack --stack-name Demo-AppConfigHostedDeploy-Stack
```

* SAMのCloudFormationスタック削除
    * VPC内Lambdaが参照するHyperplane ENIの削除に最大20分かかるため、スタックの削除に時間がかかる。

```sh
sam delete
# Windowsでもmakeをインストールすればmakeでいけます
make delete
```

* その他のリソース削除

```sh
aws cloudformation delete-stack --stack-name Demo-AppConfig-Stack
aws cloudformation delete-stack --stack-name Demo-Bastion-Stack
#aws s3 rm s3://(バケット名) --recursive
aws s3 rm s3://mysd33bucket123demo --recursive
aws cloudformation delete-stack --stack-name Demo-S3-Stack
aws cloudformation delete-stack --stack-name Demo-SQS-Stack
aws cloudformation delete-stack --stack-name Demo-DynamoDB-Stack
aws cloudformation delete-stack --stack-name Demo-DocumentDB-Stack
aws cloudformation delete-stack --stack-name Demo-RDS-Stack
aws cloudformation delete-stack --stack-name Demo-SM-Stack
aws cloudformation delete-stack --stack-name Demo-NATGW-Stack
aws cloudformation delete-stack --stack-name Demo-VPE-Stack
aws cloudformation delete-stack --stack-name Demo-SG-Stack
aws cloudformation delete-stack --stack-name Demo-VPC-Stack 
aws cloudformation delete-stack --stack-name Demo-IAM-Stack 
```

## 23. CloudWatch Logsのロググループ削除

```
aws logs delete-log-group --log-group-name /aws/apigateway/welcome
aws logs delete-log-group --log-group-name /aws/lambda-insights
aws logs delete-log-group --log-group-name /aws/rds/proxy/demo-rds-proxy
aws logs delete-log-group --log-group-name /aws/rds/cluster/aurora-postgresql-cluster/postgresql

# API Gatewayの実行ログの削除
aws logs describe-log-groups --log-group-name-prefix API-Gateway-Execution-Logs --query logGroups[*].logGroupName
aws logs delete-log-group --log-group-name（返却された各ロググループ名）

# 以下は、Lambdaのロググループが残ってしまっていた場合は実行し削除
# AppConfig Agentの終了のタイミングによっては残ってしまうことがある
aws logs delete-log-group --log-group-name /aws/lambda/user-function
aws logs delete-log-group --log-group-name /aws/lambda/book-function
aws logs delete-log-group --log-group-name /aws/lambda/todo-function
aws logs delete-log-group --log-group-name /aws/lambda/bff-function
aws logs delete-log-group --log-group-name /aws/lambda/todo-async-function
aws logs delete-log-group --log-group-name /aws/lambda/todo-async-fifo-function
```

## 24. ローカルでの実行確認
* 前述の手順の通り、AWS上でLambda等をデプロイしなくてもsam localコマンドを使ってローカル実行確認も可能である

* Postgres SQLのDockerコンテナを起動
```sh
cd postgres-local
docker compose up -d

#Postgresのコンテナにシェルで入って、psqlコマンドで接続
docker exec -i -t postgres-local /bin/bash
> psql -U postgres

# psqlで、testdbデータベースを作成
postgres> CREATE DATABASE testdb;
# testdbに切替
postgres> \c testdb
#ユーザテーブル作成
testdb> CREATE TABLE IF NOT EXISTS m_user (user_id VARCHAR(50) PRIMARY KEY, user_name VARCHAR(50));
#ユーザテーブルの作成を確認
testdb> \dt
#切断
testdb> \q
```

* DynamoDB Local実行の起動
    * AWSのDynamoDB Localの場合
        * [DynamoDB Local](https://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/developerguide/DynamoDBLocal.html)と[dynamodb-admin](https://github.com/aaronshaf/dynamodb-admin)のDockerコンテナを起動
            * Windowsでうまく起動しない場合は、いったん権限の問題で引っかかるdocker/dynamodbフォルダを削除してみて再作成してみてください。

        ```sh
        cd ..
        cd dynamodb-local
        docker compose up -d
        ```

        * dynamodb-adminでtodoテーブルを作成        
            * ブラウザで[http://localhost:8001/](http://localhost:8001/)にアクセスし「Create Table」ボタンをクリック    
            * 以下のテーブルを作成
                * 「Table Name」…「todo」、「Hash Attribute Name」…「todo_id」、「Hash Attribute Type」…「String」で作成
                * 「Table Name」…「temp」、「Hash Attribute Name」…「id」、「Hash Attribute Type」…「String」で作成
                * 「Table Name」…「queue_message」、「Hash Attribute Name」…「message_id」、「Hash Attribute Type」…「String」で作成        
                * 「Table Name」…「idempotency」、「Hash Attribute Name」…「idempotency_key」、「Hash Attribute Type」…「String」で作成

        * TODO: NoSQL WorkbenchでDynamoDB Localにアクセスする場合のテーブル作成手順も記載
      

    * [LocalStack](https://github.com/localstack/localstack)の場合
        * TODO: DockerComposeでの起動方法の整備


* SQL Local実行の起動
    * [Elastic MQ](https://github.com/softwaremill/elasticmq)の場合
        * Elastic MQのコンテナを起動

        ```sh
        cd ..
        cd elasticmq
        docker compose up -d
        ```

        * キューの確認
            * ブラウザで、[http://localhost:9325](http://localhost:9325)にアクセスするとキューの状態が確認できる
            * custom.confの設定に基づき、SampleQueueと、SampleQueue-DLQという標準キュー、SampleFIFOQueue.fifoというFIFOキューが作成されていることが分かる

    * [LocalStack](https://github.com/localstack/localstack)の場合
        * TODO: DockerComposeでの起動方法の整備

> [!NOTE]
> MinIOは、GNU AGPL v3によるOSSライセンスと商用ライセンスのデュアルライセンスで提供されており、MinIOを同梱しての配布、利用等には注意すること。
> また、現在、MinIOは、[Dockerイメージの配布を停止してしまった](https://github.com/minio/minio/issues/21647)ようなので、代替案としてLocalStackを利用する方法を整理した。これにより、他のAWSサービスのローカル実行もLocalStackで一元的に実行する手順も今後整備する。

* S3 Local実行の起動
    * [LocalStack](https://github.com/localstack/localstack)の場合        

        ```sh
        cd ..
        cd localstack
        docker compose up -d
        ```            

        ```sh
        $ aws configure --profile localstack
        AWS Access Key ID [None]: DUMMY
        AWS Secret Access Key [None]: DUMMY
        Default region name [None]: ap-northeast-1
        Default output format [None]: json
        ```

        * バケットの作成

            ```sh
            # バケットの作成
            aws s3 mb --endpoint-url=http://localhost:4566 --profile localstack s3://samplebucket123 
            # バケットの確認
            aws s3 ls --endpoint-url=http://localhost:4566 --profile localstack
            ```

        * バケット内のファイルの確認（サンプルAP実行後に使用）

            ```sh            
            aws s3 ls --endpoint-url=http://localhost:4566 --profile localstack s3://samplebucket123 --recursive
            ```        

    * [MinIO](https://min.io/)の場合
        * MinIOのDockerコンテナを起動

        ```sh
        cd ..
        cd minio
        docker compose up -d
        ```

        * バケットの作成
            * ブラウザで、[http://localhost:9001](http://localhost:9001)にアクセスするとMinIOのコンソールが表示される
            * ログイン画面で、UsernameとPasswordに、それぞれ「minioadmin」を入力しログインする
            * 「Object Browser」メニューから、「Create bucket」をクリックし、以下のバケットを作成する
                * 「Bucket Name」…「samplebucket123」

* DocumentDB Local実行の起動
    * MongoDB(DocumentDB互換)の場合
        * MongoDBのDockerコンテナを起動

            ```sh
            cd mongodb
            docker compose up -d
            ```

        * Mongo Expressでデータベースの作成
            * ブラウザで、[http://localhost:8081](http://localhost:8081)にアクセスするとMongo Expressのコンソールが表示される
            * 「+Create Database」ボタンをクリックし、以下のデータベースを作成する
                * 「Database Name」…「sampledb」

    * [LocalStack](https://github.com/localstack/localstack)の場合
        * TODO: DockerComposeでの起動方法の整備

* sam local start-apiコマンドを実行
    * local-env.jsonファイルに記載されてた、環境変数で上書きして実行

```sh
cd ..
sam local start-api --env-vars local-env.json

# Windowsでもmakeをインストールすればmakeでいけます

make local_startapi 
```

* APの動作確認
```sh
# hello-world
curl http://127.0.0.1:3000/hello

# Userサービス
curl -X POST -H "Content-Type: application/json" -d '{ "user_name" : "Taro"}' http://127.0.0.1:3000/users-api/v1/users

curl http://127.0.0.1:3000/users-api/v1/users/(ユーザID)

# Todoサービス
curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "Buy Milk"}' http://127.0.0.1:3000/todo-api/v1/todo
# トランザクション実施による登録の場合
curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "Buy Milk"}' http://127.0.0.1:3000/todo-api/v1/todo?tx=true

curl http://127.0.0.1:3000/todo-api/v1/todo/(TODO ID)

# Bookサービス
# 書籍の登録（WindowsでGit Bashだとcurlで日本語データのPostがうまくいかないのでファイルにしている）
curl -X POST -H "Content-Type: application/json" -d @curl/books/post1.json http://127.0.0.1:3000/books-api/v1/books
curl -X POST -H "Content-Type: application/json" -d @curl/books/post2.json http://127.0.0.1:3000/books-api/v1/books
curl -X POST -H "Content-Type: application/json" -d @curl/books/post3.json http://127.0.0.1:3000/books-api/v1/books
curl -X POST -H "Content-Type: application/json" -d @curl/books/post4.json http://127.0.0.1:3000/books-api/v1/books
curl -X POST -H "Content-Type: application/json" -d @curl/books/post5.json http://127.0.0.1:3000/books-api/v1/books
# タイトルが「こころ」の書籍を取得(title=こころ)
curl http://127.0.0.1:3000/books-api/v1/books?title=%E3%81%93%E3%81%93%E3%82%8D
# 著者が「村上春樹」の書籍を取得（author=村上春樹）
curl http://127.0.0.1:3000/books-api/v1/books?author=%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A8%B9
# 出版社が「新潮社」の書籍を取得（publisher=新潮社）
curl http://127.0.0.1:3000/books-api/v1/books?publisher=%E6%96%B0%E6%BD%AE%E7%A4%BE
# 著者が「村上春樹」で出版社が「講談社」の書籍を取得（author=村上春樹&publisher=講談社）
curl http://127.0.0.1:3000/books-api/v1/books?author=%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A8%B9\&publisher=%E8%AC%9B%E8%AB%87%E7%A4%BE

# BFF
curl -X POST -H "Content-Type: application/json" -d '{ "user_name" : "Taro"}' http://127.0.0.1:3000/bff-api/v1/users
curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "Buy Milk"}' http://127.0.0.1:3000/bff-api/v1/todo

curl -X POST -H "Content-Type: application/json" -d @curl/books/post1.json http://127.0.0.1:3000/bff-api/v1/books
curl http://127.0.0.1:3000/bff-api/v1/books?author=%E6%9D%91%E4%B8%8A%E6%98%A5%E6%A8%B9\&publisher=%E8%AC%9B%E8%AB%87%E7%A4%BE

#curlコマンドの場合は&をエスケープする
curl http://127.0.0.1:3000/bff-api/v1/todo?user_id=（ユーザID）\&todo_id=(TODO ID)

# BFF (エラー電文動作確認)
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/validation
curl -X POST -H "Content-Type: application/json" -d '{}' http://127.0.0.1:3000/bff-api/v1/error/validation2
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/validation3
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/business
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/business2
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/business3
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/business4
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/system
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/system2
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/hogehoge
curl -X POST http://127.0.0.1:3000/bff-api/v1/error/panic
# BFF（存在しないパス動作確認）
curl -X POST http://127.0.0.1:3000/bff-api/v1/xxx
# BFF（存在しないメソッド動作確認）
curl -X PUT http://127.0.0.1:3000/bff-api/v1/users
```

* ディレード実行依頼
    * ディレード実行依頼受信後の後方の非同期処理部分はsam local start-apiでは動作確認できないため、別途sam local invokeコマンドを実行し確認
        * local-env.jsonファイルに記載されてた、環境変数で上書きして実行

    * 標準キューの場合
        * ディレード処理実行依頼

        ```sh
        # 業務のDB更新を伴う場合
        curl -X POST -H "Content-Type: application/json" -d '{ "todo_titles" : ["Buy Milk", "Study English"]}' http://127.0.0.1:3000/bff-api/v1/todo-async
        # 業務のDB更新を伴わない場合
        curl -X POST -H "Content-Type: application/json" -d '{ "todo_titles" : ["Buy Milk", "Study English"]}' http://127.0.0.1:3000/bff-api/v1/todo-async?dbtx=no

        # Elastic MQから実行依頼したメッセージを取得し確認
        aws sqs receive-message --queue-url http://localhost:9324/000000000000/SampleQueue --endpoint-url http://localhost:9324 --attribute-names All --message-attribute-names All --no-cli-pager --max-number-of-messages 10
        ```

        * events/event-TodoAsyncFunction.jsonをメッセージ内容に合わせて編集

        ```json
            "Records": [
                {
                    # メッセージIDを修正       
                    "messageId": "99a809d5-0e25-42ad-89df-c466385dad58",
                    "receiptHandle": "99a809d5-0e25-42ad-89df-c466385dad58#d753a0fb-c153-495f-81f3-d849a3705116",
                    # Bodyを送信したメッセージに修正
                    "body": "{\"tempId\":\"45c95bba-a933-11ee-bc9f-0242ac110005\"}",
                    "attributes": {
                        "ApproximateFirstReceiveTimestamp": "1704094384254",
                        "SentTimestamp": "1704091021068",
                        "SenderId": "127.0.0.1",
                        # メッセージの受信回数（リトライ回数）を変更したい場合に適宜修正
                        "ApproximateReceiveCount": "1",
                        "AWSTraceHeader": "Parent=0000000000000000;Sampled=0"                        
                    },
                    "messageAttributes": {                        
                        # delete_timeの値の追加
                        "delete_time": {
                            "StringValue": "1704413117",
                            "DataType": "String"
                        }
                    },
                    …
                }
            ]
        ```

        * 後方の非同期処理の実行
        
        ```sh
        # ディレード実行依頼後の後方の非同期処理の実行
        sam local invoke --env-vars local-env.json --event events\event-TodoAsyncFunction.json TodoAsyncFunction

        # Windowsでもmakeをインストールすればmakeでいけます
        make local_invoke_TodoAsyncFunction
        ```

        * メッセージの削除

        ```sh
        aws sqs delete-message --queue-url http://localhost:9324/000000000000/SampleQueue --endpoint-url http://localhost:9324 --receipt-handle (ReceiptHandleの値)
        ```

    * FIFOキューの場合

        * ディレード実行依頼

        ```sh        
        # 業務のDB更新を伴う場合
        curl -X POST -H "Content-Type: application/json" -d '{ "todo_titles" : ["Buy Milk", "Study English"]}' http://127.0.0.1:3000/bff-api/v1/todo-async?fifo=true
        # 業務のDB更新を伴わない場合
        curl -X POST -H "Content-Type: application/json" -d '{ "todo_titles" : ["Buy Milk", "Study English"]}' http://127.0.0.1:3000/bff-api/v1/todo-async?fifo=true\&dbtx=no
        # Elastic MQから実行依頼したメッセージを取得し確認
        aws sqs receive-message --queue-url http://localhost:9324/000000000000/SampleFIFOQueue.fifo --endpoint-url http://localhost:9324 --attribute-names All --message-attribute-names All --no-cli-pager --max-number-of-messages 10
        ```

        * events/event-TodoAsyncFifoFunction.jsonをメッセージ内容に合わせて編集

        ```json
            "Records": [
                {
                    # メッセージIDを修正       
                    "messageId": "99a809d5-0e25-42ad-89df-c466385dad58",
                    "receiptHandle": "99a809d5-0e25-42ad-89df-c466385dad58#d753a0fb-c153-495f-81f3-d849a3705116",
                    # Bodyを送信したメッセージに修正
                    "body": "{\"tempId\":\"45c95bba-a933-11ee-bc9f-0242ac110005\"}",
                    "attributes": {
                        "ApproximateFirstReceiveTimestamp": "1545082649185",    
                        "SentTimestamp": "1545082649183",           
                        "MessageDeduplicationId": "049a0b4e-a882-11ee-80df-0242ac110005",
                        "SenderId": "AIDAIENQZJOLO23YVJ4VO",
                        # メッセージの受信回数（リトライ回数）を変更したい場合に適宜修正    
                        "ApproximateReceiveCount": "1",   
                        "AWSTraceHeader": "Parent=0000000000000000;Sampled=0",                        
                        # メッセージグループIDを修正 "049a0ab8-a882-11ee-80df-0242ac110005",                                                        
                        "SequenceNumber": "0"                                      
                    },                    
                    "messageAttributes": {
                        # delete_timeの値の追加
                        "delete_time": {
                            "StringValue": "1704413117",
                            "DataType": "String"
                        }                                              
                    },
                    …
                }
            ]        
        ```

        * 後方の非同期処理の実行

        ```sh
        # ディレード実行依頼後の後方の非同期処理の実行
        sam local invoke --env-vars local-env.json --event events\event-TodoAsyncFifoFunction.json TodoAsyncFifoFunction

        # Windowsでもmakeをインストールすればmakeでいけます
        make local_invoke_TodoAsyncFifoFunction
        ```

        * メッセージ削除

        ```sh
        aws sqs delete-message --queue-url http://localhost:9324/000000000000/SampleFIFOQueue.fifo --endpoint-url http://localhost:9324 --receipt-handle (ReceiptHandleの値)
        ```

## AWS SDKのClientLogMode

* [AWS SDKのClientLogMode](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/logging/#clientlogmode)に対応している
* ソフトウェアフレームワーク機能によるプロパティファイル（configs/config-（プロファイル名）.yml）または、AppConfigのホスト化された設定によりが可能
    * AWSSDK_CLIENT_LOG_MODEに、カンマ区切りでログモードを指定することで、ログの出力を制御できる

* config-（プロファイル名）.yamlでの設定例
```yaml
AWSSDK_CLIENT_LOG_MODE: "LogSigning,LogRetries,LogRequestWithBody,LogResponseWithBody,LogDeprecatedUsage,LogRequestEventMessage,LogResponseEventMessage"
```

* 設定後のDynamoDBアクセスでのログの例
    * リクエストの署名、リクエスト、レスポンスのデバッグログが出力されている

```sh
SDK 2024/03/09 14:14:13 DEBUG Request Signature:
---[ CANONICAL STRING  ]-----------------------------
POST
/
accept-encoding:identity
amz-sdk-invocation-id:45ebb81c-ca25-4d6c-af46-e5d3631be68f
amz-sdk-request:attempt=1; max=3
content-length:264
content-type:application/x-amz-json-1.0
host:host.docker.internal:8000
x-amz-date:20240309T051413Z
x-amz-target:DynamoDB_20120810.PutItem

accept-encoding;amz-sdk-invocation-id;amz-sdk-request;content-length;content-type;host;x-amz-date;x-amz-target
3632327c19662e8f924a4f107cb7f798d77b1d3d091a6f4ca48d2fc811bec37a
---[ STRING TO SIGN ]--------------------------------
AWS4-HMAC-SHA256
20240309T051413Z
20240309/ap-northeast-1/dynamodb/aws4_request
93d5240153452cb92468a6b764bc6bcca32547c5a9075322da0db219e81f154a
-----------------------------------------------------
```

```sh
SDK 2024/03/09 14:14:13 DEBUG Request
POST / HTTP/1.1
Host: host.docker.internal:8000
User-Agent: aws-sdk-go-v2/1.24.1 os/linux lang/go#1.21.4 md/GOOS#linux md/GOARCH#amd64 api/dynamodb#1.26.7
Content-Length: 264
Accept-Encoding: identity
Amz-Sdk-Invocation-Id: 45ebb81c-ca25-4d6c-af46-e5d3631be68f
Amz-Sdk-Request: attempt=1; max=3
Authorization: AWS4-HMAC-SHA256 Credential=AKIAJYAKKINDAXTEZWWA/20240309/ap-northeast-1/dynamodb/aws4_request, SignedHeaders=accept-encoding;amz-sdk-invocation-id;amz-sdk-request;content-length;content-type;host;x-amz-date;x-amz-target, Signature=236532a84d1508b8176dec783c18e87594bd0f7ea443ead187449930ff36263e
Content-Type: application/x-amz-json-1.0
X-Amz-Date: 20240309T051413Z
X-Amz-Target: DynamoDB_20120810.PutItem
X-Amzn-Trace-Id: Parent=0000000000000000;Sampled=0
{"ConditionExpression":"attribute_not_exists(#partition_key)","ExpressionAttributeNames":{"#partition_key":"todo_id"},"Item":{"todo_id":{"S":"de67c546-ddd3-11ee-a41c-0242ac110004"},"todo_title":{"S":"Buy Milk"}},"ReturnConsumedCapacity":"TOTAL","TableName":"todo"}
```

```sh
SDK 2024/03/09 14:14:13 DEBUG Response
HTTP/1.1 200 OK
Content-Length: 61
Content-Type: application/x-amz-json-1.0
Date: Sat, 09 Mar 2024 05:14:13 GMT
Server: Jetty(11.0.11)
X-Amz-Crc32: 539703220
X-Amzn-Requestid: 1c53023f-b61b-424e-b665-62ca6cdf3f2a
{"ConsumedCapacity":{"TableName":"todo","CapacityUnits":1.0}}
```

## sam localでのリモートデバッグ実行

> [!WARNING]  
> [aws-sam-cliのissue](https://github.com/aws/aws-sam-cli/issues/3718)によると、当該サンプルAPが使用する「provided.al2」、「provided.al2023」（カスタムランタイム）でのsam localのデバッグ実行は現状サポートされていない。  
> 回避策として、template.yaml中を「go1.x」ランタイムに一時的に修正する対応する手順を参考に記載する。 

* [AWSの開発者ガイド](https://docs.aws.amazon.com/ja_jp/serverless-application-model/latest/developerguide/serverless-sam-cli-using-debugging.html#serverless-sam-cli-running-locally)の記載にある通り、[delve](https://github.com/go-delve/delve)といったサードパーティのデバッガを使用することで、VSCodeでの sam localのリモートデバッグ実行が可能である。

> [!WARNING]  
> Linuxの場合はGLIBCのバージョンの問題が出ることがほとんどである。今分かっている簡易的な回避策としてあらかじめ、`CGO_ENABLED=0`を指定してから、delveのインストール、sam buildをビルドする方法がある。
> 
> ```sh
> go env -w CGO_ENABLED=0
> ```

* delveのインストール
    * Lambda関数及びdelveが実行されるのはLambdaコンテナ内(Amazon Linux)なので、
    GOOSに`linux`を指定し、インストール

        ```sh
        # Windows
        set GOARCH=amd64
        set GOOS=linux
        go install github.com/go-delve/delve/cmd/dlv@latest

        # Linux
        GOARCH=amd64 GOOS=linux go install github.com/go-delve/delve/cmd/dlv@latest
        ```

* サンプルAPを一時的にgo1.xのランタイムに切り替えてビルド
    * template.yamlのRuntimeを「go1.x」に変更、Handlerの値を「bootstrap」から任意のAP名に変更する。また、Lambda Insightの設定を無効化する。その後、クリーンして再ビルド(makeコマンドでOK)する。
        * Runtimeの修正例:  
            ![go1.xへのテンプレート変更](image/template_go1x_1.png)

        * Handlerの修正例（HelloWorldFunctionの場合）：  
            ![go1.xへのテンプレート変更2](image/template_go1x_2.png)

        * Lambda Insightの修正例:
            ![go1.xへのテンプレート変更2](image/template_go1x_3.png)

    * 本サンプルAPでは、すでに上記の修正をしたデバッグ実行用の一時的なtemplate.yamlとして[template-dbg.yaml](./template-dbg.yaml)を使ってビルドすればよい。makeコマンドも用意している。    

    ```sh
	sam build --template-file template-dbg.yaml
	xcopy /I /S configs .aws-sam\build\BffFunction\configs
	xcopy /I /S configs .aws-sam\build\UsersFunction\configs	
	xcopy /I /S configs .aws-sam\build\TodoFunction\configs	
	xcopy /I /S configs .aws-sam\build\TodoAsyncFunction\configs
	xcopy /I /S configs .aws-sam\build\TodoAsyncFifoFunction\configs

    # Windowsでもmakeをインストールすればmakeでいけます
    make build_dbg
    ```

> [!Note]
> 上記に記載のとおり、al2、al2023のランタイムではリモートデバッグができないことへの回避策として、ランタイムgo1.xのコンテナイメージを使ってsam localを実行する際、Lambda Insightを有効化すると、cloudwatch_lambda_agentが利用される際、ランタイムgo1.xに含まれるGLIBCが、GLIBC_2.18、2.25といった新しいバージョンに対応していないためと考えられる。  
> このため、デバッグ実行時はコメントアウトして無効化している。
>
> ```
> /opt/extensions/cloudwatch_lambda_agent: /lib64/libc.so.6: version `GLIBC_2.18' not found (required by /opt/extensions/cloudwatch_lambda_agent)
> /opt/extensions/cloudwatch_lambda_agent: /lib64/libc.so.6: version `GLIBC_2.25' not found (required by /opt/extensions/cloudwatch_lambda_agent)
> ```



* sam local start-apiの実行
    * デバッガパス、デバッグポート(この例では8099番)を指定して、sam local start-apiを実行する。
        * [AWSのデベロッパーガイド](https://docs.aws.amazon.com/ja_jp/serverless-application-model/latest/developerguide/serverless-sam-cli-using-debugging.html#serverless-sam-cli-running-locally)を参考のこと。

    * 本サンプルAPのように、SAMテンプレート内にFunctionが複数ある場合は、デバッグできるのは一度に1つの関数のみのためか、--debug-functionオプションでデバッグしたい関数を指定しないとデバッガが動かない

        ```sh
        # Windows
        sam local start-api -d 8099 --debugger-path=$GOPATH/bin/linux_amd64 --debug-args="-delveAPI=2" --debug-function (template.yamlのLambda関数の論理ID) --env-vars local-env.json

        # Linux
        sam local start-api -d 8099 --debugger-path=$HOME/go/bin --debug-args="-delveAPI=2" --debug-function (template.yamlのLambda関数の論理ID) --env-vars local-env.json

        # Windowsの例
        sam local start-api -d 8099 --debugger-path=$GOPATH/bin/linux_amd64 --debug-args="-delveAPI=2" --debug-function HelloWorldFunction --env-vars local-env.json
        sam local start-api -d 8099 --debugger-path=$GOPATH/bin/linux_amd64 --debug-args="-delveAPI=2" --debug-function UsersFunction --env-vars local-env.json
        sam local start-api -d 8099 --debugger-path=$GOPATH/bin/linux_amd64 --debug-args="-delveAPI=2" --debug-function TodoFunction --env-vars local-env.json
        sam local start-api -d 8099 --debugger-path=$GOPATH/bin/linux_amd64 --debug-args="-delveAPI=2" --debug-function BffFunction --env-vars local-env.json


        # Windowsでもmakeをインストールすればmakeでいけます
        make local_startapi_dbg_HelloWorldFunction
        make local_startapi_dbg_UsersFunction
        make local_startapi_dbg_TodoFunction
        make local_startapi_dbg_BffFunction
        ```

    * SAMテンプレート内にFunctionが1つの場合は--debug-functionオプションなくてもうまくいく

        ```sh        
        # Windows
        sam local start-api -d 8099 --debugger-path=$GOPATH/bin/linux_amd64 --debug-args="-delveAPI=2" --env-vars local-env.json

        # Linux
        sam local start-api -d 8099 --debugger-path=$HOME/go/bin --debug-args="-delveAPI=2" --env-vars local-env.json
        ```

* このサンプルAPでは作成済のものがあるのでそのまま使えばよいが、VSCodeからアタッチするため、`.vscode/launch.json`を作成

    ```json
    {
        "version": "0.2.0",
        "configurations": [
            {
                "name": "delve",
                "type": "go",
                "request": "attach",
                "mode": "remote",
                "port": 8099,
                "host": "127.0.0.1"
            }
        ]
    }
    ```

* curlコマンドで、確認対象のAPIを呼び出すと、処理が待ち状態で止まった状態になる

* VSCodeでブレイクポイントを設定、「実行とデバッグ」の▷ボタンを押すと、ブレイクポイントで止まる。
    * 以下のようなログが出力されれば、▷ボタンを押下して、デバッグ可能

    ```
    2024-09-16T08:57:15+09:00 info layer=debugger launching process with args: [/var/task/hello-world]
    2024-09-16T08:57:16+09:00 debug layer=debugger Adding target 36 "/var/task/hello-world"    
    ```

![SAM Localのデバッグ画面](image/sam-local-debug.png)

## sam localでの直接デバッグ実行

* [AWSの開発者ガイド](https://docs.aws.amazon.com/ja_jp/toolkit-for-vscode/latest/userguide/debug-apigateway.html)の記載にある通り、AWS Toolkitの機能で、VSCodeでの sam localの直接デバッグ実行することも可能である。

* TODO: 作者のWindows端末環境ではエラーになってしまい動作しない。今後、実施できたら手順を記載する。

## godocの表示
* godocをインストール
```sh
go install golang.org/x/tools/cmd/godoc@latest     
```

* 使い方は、[godoc](https://pkg.go.dev/golang.org/x/tools/cmd/godoc)を参照のこと

* appbaseフォルダでgodocコマンドを実行    
```sh
cd appbase
godoc

# Windowsでもmakeをインストールすればmakeでいけます
make doc_appbase
```

* appフォルダでgodocコマンドを実行
```sh
cd app
godoc

# Windowsでもmakeをインストールすればmakeでいけます
make doc_app
```

* godoc起動中の状態で、[http://localhost:6060](http://localhost:6060)へアクセス
    * 「example.com/」の「appbase」パッケージは、[http://localhost:6060/pkg/example.com/appbase/](http://localhost:6060/pkg/example.com/appbase/)に表示される。
    * 「app」パッケージは、ほぼ全てが「internal」パッケージに配置しているため、デフォルトでは表示されない。m=allをクエリパラメータに指定して、[http://localhost:6060/pkg/app/?m=all](http://localhost:6060/pkg/app/?m=all)にアクセスするとよい。

## ソフトウェアフレームワーク
* 本サンプルアプリケーションでは、ソフトウェアフレームワーク実装例も同梱している。簡単のため、アプリケーションと同じプロジェクトでソース管理している。
* ソースコードはappbaseフォルダ配下にexample.com/appbaseパッケージとして格納されている。    
    * 本格的な開発を実施する場合には、業務アプリケーションと別のGitリポジトリとして管理し、参照するようにすべきであるが、ここでは、あえて同じプロジェクトに格納してノウハウを簡単に参考にしてもらいやすいようにしている。
* 各機能と実現方式は、以下の通り。

| 機能 | 機能概要と実現方式 | 拡張実装 | 拡張実装の格納パッケージ |
| ---- | ---- | ---- | ---- |
| DI | ソフトウェアフレームワーク機能の各コンポーネントの依存関係の注入、インスタンス管理を実施する。 | ○ | com.example/appbase/pkg/component |
| オンラインAP実行制御 | APIの要求受信、ビジネスロジック実行、応答返却まで一連の定型的な処理を実行を制御する共通機能を提供する。AWS Lambda Go API Proxyを利用して、Lambda SDKとginを統合し実現する。 | ○ | com.example/appbase/pkg/handler<br/>com.example/appbase/pkg/api |
| 非同期AP実行制御 | SQSからの要求受信、ビジネスロジック実行、応答返却まで一連の定型的な処理を実行を制御する共通機能を提供する。Lambda SDKを利用して実現する。また、非同期実行依頼側の業務APでDynamoDBアクセスを伴う場合、DynamoDBトランザクション管理機能を用いてDB更新とメッセージ送達のデータ整合性を担保する。 | ○ | com.example/appbase/pkg/handler<br/>com.example/appbase/pkg/transaction |
| 二重実行防止（冪等性） | SQSの標準キューの場合には複数回同一メッセージが配信されるケースがある。また、FIFOキューでもLambdaのイベントソースマッピングの場合、EventBridgeをトリガとするLambdaやStepFunctionsでも同一イベントが重複して発生するため、いずれの場合にもLambdaが二重実行される恐れがあるため、DynamoDBのテーブルの条件付き更新を使って、PutItem時に、ConditionExpressionですでに同一のメッセージIDやイベントIDを主キーとするアイテムがないかチェックすることで、二重実行防止し冪等性を担保する機能を提供する。  | ○ | com.example/appbase/pkg/idempotency |
| 入力チェック| APIのリクエストデータの入力チェックを実施する、ginのバインディング機能でgo-playground/validator/v10を使ったバリデーションを実現する。バリデーションエラーメッセージの日本語化に対応する。 | ○ | com.example/appbase/pkg/validator |
| エラー（例外） | エラーコード（メッセージID）やメッセージを管理可能な共通的な入力エラー、ビジネスエラー、システムエラー用のGoのErrorオブジェクトを提供する。cockroachdb/errorsによりスタックトレースがログ出力できるようにする。 | ○ | com.example/appbase/pkg/errors |
 集約例外ハンドリング | オンラインAP制御機能、トランザクション管理機能と連携し、エラー（例外）発生時、エラーログの出力、DBのロールバック、エラー画面やエラー電文の返却といった共通的なエラーハンドリングを実施する。 | ○ | com.example/appbase/pkg/handler<br/>com.example/appbase/pkg/transaction |
| RDBアクセス | go標準のdatabase/sqlパッケージを利用しRDBへアクセスする。DB接続等の共通処理を個別に実装しなくてもよい仕組みとする。 | ○ | com.example/appbase/pkg/rdb |
| RDBトランザクション管理機能 | サービス（ビジネスロジック）の実行前後にRDBのトランザクション開始・終了を機能を提供する。 | ○ | com.example/appbase/pkg/rdb |
| DynamoDBアクセス | AWS SDKを利用しDynamoDBへアクセスする汎化したAPIを提供する。 | ○ | com.example/appbase/pkg/dynamodb |
| DynamoDBトランザクション管理機能 | サービス（ビジネスロジック）の実行前後にDynamoDBのトランザクション開始・終了を機能を提供する。 | ○ | com.example/appbase/pkg/transaction<br/>com.example/appbase/pkg/domain |
| DocumentDB（Mongo）アクセス | MongoDB Goドライバー(go.mongodb.org/mongo-driver/mongo)を利用しDBへアクセスする。DB接続等の共通処理を個別に実装しなくてもよい仕組みとする。  | ○ | com.example/appbase/pkg/documentdb |
| 非同期実行依頼 | AWS SDKを利用してSQSへ非同期処理実行依頼メッセージを送信する汎化したAPIを提供する。また、業務APでDynamoDBアクセスを伴う場合、DynamoDBトランザクション管理機能を用いてDB更新とメッセージ送達のデータ整合性を担保する。 | ○ | com.example/appbase/pkg/async<br/>com.example/appbase/pkg/transaction |
| オブジェクトストレージアクセス| AWS SDKを利用し、S3にアクセスする汎化したAPIを提供する。 | ○ | com.example/appbase/pkg/objectstorage |
| HTTPクライアント| net/http、ctxhttp等を利用しREST APIの呼び出しを汎化したAPIを提供する。 | ○ | com.example/appbase/pkg/httpclient |
| 分散トレーシング（X-Ray） | AWS X-Rayを利用して、サービス間の分散トレーシング・可視化を実現する。実現には、AWS SAMのtemplate.ymlで設定でAPI GatewayやLambdaのトレースを有効化する。またAWS SDKが提供するメソッドに、Lambdaのハンドラメソッドの引数のContextを引き渡すようにする。Contextは業務AP側で引き継いでメソッドの引数に引き渡さなくてもソフトウェアフレームワーク側で取得できるようにグローバル変数で管理する。 | ○ | com.example/appbase/pkg/apcontext |
| ロギング | zap(go.uber.org/zap)の機能を利用し、プロファイル（環境区分）によって動作環境に応じたログレベル、出力形式（プレーンテキストやJSON形式）等を切替可能とする。また、メッセージ管理機能と連携し、メッセージIDをもとにログ出力可能な汎用的なAPIを提供する。 | ○ | com.example/appbase/pkg/logging |
| プロパティ管理 | APから環境依存のパラメータを切り出し、プロファイル（環境区分）によって動作環境に応じたパラメータ値に置き換え可能とする。AWS AppConfigおよびAppConfig Agent Lambdaエクステンションを利用してAPの再デプロイせずとも設定変更を反映できる。また、変更が少ない静的な設定値やローカルでのAP実行用に、spf13/viperの機能を利用して、OS環境変数、yamlによる設定ファイルを読み込み反映する。なお、AppConfigに同等のプロパティがある場合には優先的に反映する。 | ○ | com.example/appbase/pkg/env<br/>com.example/appbase/pkg/config |
| メッセージ管理 | go標準のembededで、ログ等に出力するメッセージを設定ファイルで一元管理する。 | ○ | com.example/appbase/pkg/message |
| ID生成 | google/uuidを使用したUUID等を生成する。 | ○ | com.example/appbase/pkg/id |
| システム日時取得 | go標準のtimeパッケージを使用してシステムの現在日時を取得する。テスト用にプロパティから取得した固定の日時を返却するように設定切り替え可能とする。 | ○ | com.example/appbase/pkg/date |

* 以下は、今後追加を検討中。

| 機能 | 機能概要と実現方式 | 拡張実装 | 拡張実装の格納パッケージ |
| ---- | ---- | ---- | ---- |
| API認証・認可| APIGatewayのCognitoオーサライザまたはLambdaオーサライザを利用し、APIの認証、認可を行う。 | ○ | 未定 |