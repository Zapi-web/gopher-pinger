KEY_PATH ?= ~/.ssh/aws-key.pem
TF_DIR = infrastructure/terraform
AN_DIR = infrastructure/ansible

export ANSIBLE_HOST_KEY_CHECKING=False

init:
	terraform -chdir=$(TF_DIR) init
	go mod download
deploy:
	@echo "Make an infrastructure with terraform"
	terraform -chdir=$(TF_DIR) apply -auto-approve

	@echo "Ping all servers"
	cd $(AN_DIR) && ansible all -m ping --private-key=$(KEY_PATH)

	@echo "Configure servers, and deploy"
	cd $(AN_DIR) && ansible-playbook main.yml --private-key=$(KEY_PATH)
test:
	terraform -chdir=$(TF_DIR) validate
	terraform -chdir=$(TF_DIR) fmt -recursive
	go mod tidy
	go fmt
	go test -v -race ./...
destroy:
	terraform -chdir=$(TF_DIR) destroy -auto-approve