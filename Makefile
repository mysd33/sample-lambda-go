STACK_NAME := todo-app-stack

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
# sam build
# for windows
	sam.cmd build
	xcopy /I config .aws-sam\build\GetUsersFunction\config
	xcopy /I config .aws-sam\build\PostUsersFunction\config	
	xcopy /I config .aws-sam\build\GetTodoFunction\config
	xcopy /I config .aws-sam\build\PostTodoFunction\config	

unit_test:
	cd app & go test -v ./internal/...

validate:
	sam.cmd validate

deploy_guided:
# for windows
	sam.cmd deploy --guided

deploy:
# for windows
	sam.cmd deploy

delete:
	sam.cmd delete --stack-name $(STACK_NAME)