STACK_NAME := todo-app-stack
STACK_BUCKET := mysd33bucket123sam

.PHONY: build
.PHONY: deploy
.PHONY: deploy_guided
.PHONY: delete

build:
# sam build
# for windows
	sam.cmd build

deploy_guided:
# for windows
	sam.cmd deploy --guided

deploy:
# for windows
	sam.cmd deploy

delete:
	aws cloudformation delete-stack --stack-name $(STACK_NAME)
	aws s3 rm "s3://$(STACK_BUCKET)" --recursive