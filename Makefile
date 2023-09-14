.PHONY: clean
.PHONY: build
.PHONY: validate
.PHONY: deploy
.PHONY: deploy_guided
.PHONY: delete

.DEFAULT_GOAL := build

clean:
# for windows
	rmdir /s /q .aws-sam

build:
	sam build
	xcopy /I config .aws-sam\build\UsersFunction\config	
	xcopy /I config .aws-sam\build\TodoFunction\config

unit_test:
	cd app & go test -v ./internal/...

validate:
	sam validate

deploy_guided:
	sam deploy --guided

deploy:
	sam deploy

delete:
	sam delete