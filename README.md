# Private APIでのAPIGatewayを使ったLambda/GoのAWS SAMサンプルAP

## 構成イメージ
![構成イメージ](image/demo.drawio.png)

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
* hello-worldのサンプルAPでは、"https://checkip.amazonaws.com"へアクセスしに行くので、これを試す場合には作成が必要となる。

```sh
aws cloudformation validate-template --template-body file://cfn-ngw.yaml
aws cloudformation create-stack --stack-name Demo-NATGW-Stack --template-body file://cfn-ngw.yaml
```

## 6. Aurora for PostgreSQLのクラスタ作成
* TODO:作成中


## 7. RDS Proxy作成
* TODO:作成中

## 8. EC2(Bastion)の作成
* psqlによるRDBのテーブル作成や、APIGatewayのPrivate APIにアクセスするための踏み台を作成
```sh
aws cloudformation validate-template --template-body file://cfn-bastion-ec2.yaml
aws cloudformation create-stack --stack-name Demo-Bastion-Stack --template-body file://cfn-bastion-ec2.yaml
```

* 必要に応じてキーペア名等のパラメータを指定
    * 「--parameters ParameterKey=KeyPairName,ParameterValue=myKeyPair」

## 9. RDBのテーブル作成
* Bastionにログインし、psqlをインストールし、DB接続する。
    * 以下参考に、Bastionにpsqlをインストールするとよい
        * https://techviewleo.com/how-to-install-postgresql-database-on-amazon-linux/
* DB接続後、ユーザテーブルを作成する。        
```sh
sudo amazon-linux-extras install epel

sudo tee /etc/yum.repos.d/pgdg.repo<<EOF
[pgdg14]
name=PostgreSQL 14 for RHEL/CentOS 7 - x86_64
baseurl=http://download.postgresql.org/pub/repos/yum/14/redhat/rhel-7-x86_64
enabled=1
gpgcheck=0
EOF

sudo yum makecache
sudo yum install postgresql14

#DBに接続    
psql -h (Auroraのクラスタエンドポイント) -U postgres -d testdb    
#TODO: ユーザテーブル作成

```


## 10. AWS SAMでLambda/API Gateway等のデプロイ       
* SAMビルド    
```sh
# トップのフォルダに戻る
cd ..
sam build
# Windowsにmakeをインストールすればmakeでもいけます
make
```

* 必要に応じてローカル実行可能(hello-worldのみ)
```sh
sam local invoke
sam local start-api
curl http://127.0.0.1:3000/hello
```

* SAMデプロイ
```sh
# 1回目は
sam deploy --guided
# Windowsにmakeをインストールすればmakeでもいけます
make deploy_guided

# 2回目以降は
sam deploy
# Windowsにmakeをインストールすればmakeでもいけます
make deploy
```



## 11. APの実行確認

* マネージドコンソールから、EC2(Bation)へSystems Manager Session Managerで接続して、動作確認
```sh
# hello-worldの例
curl https://5h5zxybd3c.execute-api.ap-northeast-1.amazonaws.com/Prod/hello
```

```sh
# User APIサービスのPOSTコマンドの例
curl -X POST -H "Content-Type: application/json" -d '{ "name" : "Taro"}' https://42b4c7bk9g.execute-api.ap-northeast-1.amazonaws.com/Prod/users

# User APIサービスのGetコマンドの例（users/の後にPOSTコマンドで取得したユーザIDを指定）
curl https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/users/d4d6cb7f-7691-11ec-9520-1ee887dd490e
```

```sh
# Todo APIサービスのPOSTコマンドの例
curl -X POST -H "Content-Type: application/json" -d '{ "todo_title" : "ミルクを買う"}' https://42b4c7bk9g.execute-api.ap-northeast-1.amazonaws.com/Prod/todo

# Todo APIサービスのGetコマンドの例（todo/の後にPOSTコマンドで取得したTodo IDを指定）
curl https://civuzxdd14.execute-api.ap-northeast-1.amazonaws.com/Prod/tod/d4d6cb7f-7691-11ec-9520-1ee887dd490e
```
## SAMのCloudFormationスタック削除
```sh
sam delete
# Windowsにmakeをインストールすればmakeでもいけます
make delete
```

## その他のCloudFormationスタック削除
```sh
aws cloudformation delete-stack --stack-name Demo-Bastion-Stack
aws cloudformation delete-stack --stack-name Demo-NATGW-Stack
aws cloudformation delete-stack --stack-name Demo-SG-Stack
aws cloudformation delete-stack --stack-name Demo-VPC-Stack 
aws cloudformation delete-stack --stack-name Demo-IAM-Stack 
```