.PHONY: clean
.PHONY: fmt
.PHONY: lint
.PHONY: vet
.PHONY: validate
.PHONY: build
.PHONY: unit_test
.PHONY: integration_test
.PHONY: local_invoke_%
.PHONY: local_startapi
.PHONY: local_startapi_dg_%
.PHONY: deploy
.PHONY: deploy_guided
.PHONY: delete

.DEFAULT_GOAL := build

clean:
# for windows
	if exist ".aws-sam" (	\
		rmdir /s /q .aws-sam	\
	)
# for Linux
#	rm -rf .aws-sam

fmt:
	cd app & go fmt ./...
	cd appbase & go fmt ./...

lint:
	cd app & staticcheck ./...
	cd appbase & staticcheck ./...
	
vet:
	cd app & go vet ./...	
	cd appbase & go vet ./...
	cd app & shadow ./...
	cd appbase & shadow ./...

validate:
	sam validate

build: clean
# for windows	
	sam build
	xcopy /I /S configs .aws-sam\build\BffFunction\configs
	xcopy /I /S configs .aws-sam\build\UsersFunction\configs	
	xcopy /I /S configs .aws-sam\build\TodoFunction\configs	
	xcopy /I /S configs .aws-sam\build\TodoAsyncFunction\configs
	xcopy /I /S configs .aws-sam\build\TodoAsyncFifoFunction\configs
# for linux
# TODO	

unit_test:
	cd app & go test -v ./internal/...
	cd appbase & go test -v ./pkg/...

integration_test:
	cd dynamodb-local & docker-compose up -d
	cd minio & docker-compose up -d
	cd app & go test -v ./cmd/...
	cd minio & docker-compose stop
	cd dynamodb-local & docker-compose stop	

local_invoke_%:
	sam local invoke --env-vars local-env.json --event events\event-${@:local_invoke_%=%}.json ${@:local_invoke_%=%}

local_startapi:
	sam local start-api --env-vars local-env.json	

# support only go1.x runtime 
local_startapi_dg_%:
# for windows
	sam local start-api -d 8099 --debugger-path=%GOPATH%/bin/linux_amd64 --debug-args="-delveAPI=2" --debug-function ${@:local_startapi_dg_%=%} --env-vars local-env.json
# for Linux
#	sam local start-api -d 8099 --debugger-path=$(HOME)/go/bin --debug-args="-delveAPI=2" --debug-function ${@:local_startapi_dg_%=%} --env-vars local-env.json 

deploy_guided:
	sam deploy --guided

deploy:
	sam deploy

delete:
	sam delete