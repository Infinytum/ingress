bold := $(shell tput bold)
grey := $(shell tput setaf 8)
blue := $(shell tput setaf 6)
pink := $(shell tput setaf 5)
yellow := $(shell tput setaf 3)
reset := $(shell tput sgr0)

usage:
	@printf "$(blue)Infinytum Ingress $(grey)- Build Script$(reset)\n"
	@printf "$(pink)optional $(blue)command $(yellow)arguments\n\n"
	@printf "$(grey)Image Commands:\n"
	@printf "$(grey)╔ $(blue)build$(reset): Builds the docker image\n"
	@printf "$(grey)╠ $(blue)push$(reset): Pushes the docker image\n"
	@printf "$(grey)╠ $(blue)build-and-push$(reset): Builds the docker image and pushes it to the registry\n"
	@printf "$(grey)╚═══ $(pink)image$(grey): Optionally specify a custom image name (infinytum/ingress)\n"
	
	@printf "$(grey)\nGo Commands:\n"
	@printf "$(grey)╔ $(blue)run$(reset): Runs the ingress locally\n"
	@printf "$(grey)╠═══ $(pink)class-name$(grey): Configured Ingress Class\n"
	@printf "$(grey)╠═══ $(pink)namespace$(grey): Configured Namespace\n"
	@printf "$(grey)╠ $(blue)security$(reset): Scans the code for known vulnerabilites using govulncheck\n"
	@printf "$(grey)╚ $(blue)tidy$(reset): Cleans up the go module files\n"

build:
	@printf "$(blue)Building $(pink)$(if $(image),$(image),infinytum/ingress:dev)$(blue) image...$(reset)\n"
	@docker build -t $(if $(image),$(image),infinytum/ingress:dev) .

push:
	@printf "$(blue)Pushing $(pink)$(if $(image),$(image),infinytum/ingress:dev)$(blue) image...$(reset)\n"
	@docker push $(if $(image),$(image),infinytum/ingress:dev)

build-and-push: build push

run:
	@printf "$(blue)Running ingress locally$(reset)\n"
	@POD_NAME="localhost-dev" POD_NAMESPACE="infinytum-system" go run main.go --kube-config ~/.kube/config --config-map=infinytum-ingress-controller-configmap --nginx-annotations=true $(if $(class-name),--class-name=$(class-name),"") $(if $(namespace),--namespace=$(namespace),"")
 
security:
	@printf "$(blue)Scanning code for vulnerabilities...$(reset)\n"
	@govulncheck ./...

tidy:
	@printf "$(blue)Cleaning up Go Module files...$(reset)\n"
	@go mod tidy