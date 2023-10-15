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

deploy_guided:
	sam deploy --guided

deploy:
	sam deploy

delete:
	sam delete