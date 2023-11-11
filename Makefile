.PHONY: clean
.PHONY: build
.PHONY: fmt
.PHONY: validate
.PHONY: local_startapi
.PHONY: deploy
.PHONY: deploy_guided
.PHONY: delete

.DEFAULT_GOAL := build

clean:
# for windows
	rmdir /s /q .aws-sam

build:
	sam build
	xcopy /I configs .aws-sam\build\UsersFunction\configs	
	xcopy /I configs .aws-sam\build\TodoFunction\configs
	xcopy /I configs .aws-sam\build\BffFunction\configs

fmt:
	cd app & go fmt ./...
	cd appbase & go fmt ./...

unit_test:
	cd app & go test -v ./internal/...
	cd appbase & go test -v ./pkg/...

validate:
	sam validate

local_startapi:
	sam local start-api --env-vars local-env.json	

# support only go1.x runtime 
local_startapi_dg_%:
#	sam local start-api -d 8099 --debugger-path=$GOPATH/bin/linux_amd64 --debug-args="-delveAPI=2" --debug-function ${@:local_startapi_dg_%=%} --env-vars local-env.json 
	sam local start-api -d 8099 --debugger-path=%GOPATH%/bin/linux_amd64 --debug-args="-delveAPI=2" --debug-function ${@:local_startapi_dg_%=%} --env-vars local-env.json

deploy_guided:
	sam deploy --guided

deploy:
	sam deploy

delete:
	sam delete