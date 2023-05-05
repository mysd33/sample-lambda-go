STACK_NAME := todo-app-stack
STACK_BUCKET := mysd33bucket123sam

.PHONY: clean
.PHONY: build
.PHONY: validate
.PHONY: deploy
.PHONY: deploy_guided
.PHONY: delete

clean:
# for windows
	rmdir /s /q .aws-sam

build:
# sam build
# for windows
	sam.cmd build
	xcopy /I config .aws-sam\build\GetUsersFunction\config
	xcopy /I config .aws-sam\build\PostUsersFunction\config	

validate:
	sam.cmd validate

deploy_guided:
# for windows
	sam.cmd deploy --guided

deploy:
# for windows
	sam.cmd deploy

delete:
	aws cloudformation delete-stack --stack-name $(STACK_NAME)
	aws s3 rm "s3://$(STACK_BUCKET)" --recursive