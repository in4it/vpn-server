
all: build-webapp build-arm64 build-amd64

install-qa: build-webapp build-arm64 build-amd64 install-qa-aws install-qa-azure
install: build-webapp build-arm64 build-amd64 install-aws install-azure

build-webapp:
	cd webapp && npm run build && cd ..

build-arm64:
	go generate ./...
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o configmanager-linux-arm64 cmd/configmanager/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o restserver-linux-arm64 cmd/rest-server/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o reset-admin-password-linux-arm64 cmd/reset-admin-password/main.go
	shasum -a 256 configmanager-linux-arm64 > configmanager-linux-arm64.sha256
	shasum -a 256 restserver-linux-arm64 > restserver-linux-arm64.sha256
	shasum -a 256 reset-admin-password-linux-arm64 > reset-admin-password-linux-arm64.sha256

build-amd64:
	go generate ./...
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o configmanager-linux-amd64 cmd/configmanager/main.go
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o restserver-linux-amd64 cmd/rest-server/main.go
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o reset-admin-password-linux-amd64 cmd/reset-admin-password/main.go
	shasum -a 256 configmanager-linux-amd64 > configmanager-linux-amd64.sha256
	shasum -a 256 restserver-linux-amd64 > restserver-linux-amd64.sha256
	shasum -a 256 reset-admin-password-linux-amd64 > reset-admin-password-linux-amd64.sha256

install-qa-aws:
	cd provisioning && AWS_PROFILE=in4it-compute packer build -var-file=whitelist.pkr.hcl packer-amd64.pkr.hcl

install-qa-azure:
	cd provisioning && packer build -var image_version=$(shell cat latest) packer-azure-amd64.pkr.hcl

test:
	go test ./...

install-aws:
	cd provisioning && AWS_PROFILE=in4it-vpn-server AWS_REGION=us-east-1 packer build -var-file=whitelist.pkr.hcl packer-amd64.pkr.hcl

install-gcp:
	cd provisioning && packer build packer-gcp-amd64.pkr.hcl

install-azure:
	cd provisioning && packer build -var image_version=$(shell cat latest) packer-azure-amd64.pkr.hcl

install-digitalocean:
	cd provisioning && packer build packer-digitalocean-amd64.pkr.hcl

install-s3:
	cd provisioning && AWS_PROFILE=in4it-vpn-server scripts/install_s3.sh --release

install-docs:
	provisioning/scripts/deploy_documentation.sh
