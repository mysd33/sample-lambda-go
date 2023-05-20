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

build: clean
	sam build
	xcopy /I config .aws-sam\build\GetUsersFunction\config
	xcopy /I config .aws-sam\build\PostUsersFunction\config	
	xcopy /I config .aws-sam\build\GetTodoFunction\config
	xcopy /I config .aws-sam\build\PostTodoFunction\config	

unit_test:
	cd app & go test -v ./internal/...

validate:
	sam validate

deploy_guided:
	sam deploy --guided

deploy:
	sam deploy --parameter-overrides DBUsername=%DB_USER_NAME% DBPassword=%DB_PASSWORD%

delete:
	sam delete