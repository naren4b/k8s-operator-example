# Image URL to use all building/pushing image targets
IMG ?= narenp/imagearray-operator:v0.1
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Define the Kubernetes namespace
NAMESPACE ?= default

# Set GO flags and environment
GO111MODULE=on

# All Targets
all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./main.go

# Install CRDs into the Kubernetes cluster
install: manifests
	kubectl apply -f config/crd/bases

# Uninstall CRDs from the Kubernetes cluster
uninstall:
	kubectl delete -f config/crd/bases

# Deploy controller to the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Undeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	kustomize build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code
generate:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build:
	docker build -t ${IMG} .

# Push the docker image
docker-push:
	docker push ${IMG}

# Install kustomize locally if necessary
kustomize:
	@if ! [ -x "$$(command -v kustomize)" ]; then \
		echo "Installing kustomize..."; \
		cd /tmp && wget --quiet https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v3.8.7/kustomize_v3.8.7_linux_amd64.tar.gz; \
		tar -xvf kustomize_v3.8.7_linux_amd64.tar.gz && mv kustomize /usr/local/bin/; \
		echo "kustomize installed"; \
	fi

# Run against kind cluster (for local dev)
kind-deploy: docker-build kind-load deploy

# Load docker image into kind
kind-load:
	kind load docker-image ${IMG}

# Verify your code
fmt:
	go fmt ./...

vet:
	go vet ./...
