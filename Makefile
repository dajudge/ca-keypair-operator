RELEASE_VERSION ?= 0.0.0-dev

CHARTS_DIRECTORY			:= charts
CHART_CAKEYPAIR_OPERATOR	:= $(CHARTS_DIRECTORY)/cakeypair-operator
KUSTOMIZE					:= $(BIN)/kustomize

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif


#############################################
# Helm
#############################################
CHART_TEMPLATE_PATH := $(CHART_CAKEYPAIR_OPERATOR)/templates
DO_NOT_EDIT := THIS IS GENERATED CODE. DO NOT EDIT!

.PHONY:helm-clean
helm-clean:
	rm $(CHART_TEMPLATE_PATH)/*.yaml -f
	rm $(CHART_CAKEYPAIR_OPERATOR)/Chart.yaml -f

.PHONY:helm-regenerate
helm-regenerate: helm-clean helm-generate

.PHONY:helm-generate
helm-generate: $(CHARTS_DIRECTORY)/index.yaml

$(CHARTS_DIRECTORY)/index.yaml: $(CHARTS_DIRECTORY)/cakeypair-operator-$(RELEASE_VERSION).tgz
	helm repo index --url ./ $(CHARTS_DIRECTORY)

$(CHARTS_DIRECTORY)/cakeypair-operator-$(RELEASE_VERSION).tgz: \
	$(CHART_TEMPLATE_PATH)/clusterRole.yaml \
	$(CHART_TEMPLATE_PATH)/clusterRoleBinding.yaml \
	$(CHART_TEMPLATE_PATH)/role.yaml \
	$(CHART_TEMPLATE_PATH)/roleBinding.yaml \
	$(CHART_TEMPLATE_PATH)/deployment.yaml \
	$(CHART_TEMPLATE_PATH)/serviceAccount.yaml \
	$(CHART_CAKEYPAIR_OPERATOR)/Chart.yaml
	helm package $(CHART_CAKEYPAIR_OPERATOR) \
		--version $(RELEASE_VERSION) \
		--app-version $(RELEASE_VERSION) \
		--destination $(CHARTS_DIRECTORY)

$(CHART_TEMPLATE_PATH)/serviceAccount.yaml: config/helm/rbac/serviceAccount.template.yaml
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/serviceAccount.yaml
	cat config/helm/rbac/serviceAccount.template.yaml >> $(CHART_TEMPLATE_PATH)/serviceAccount.yaml

$(CHART_TEMPLATE_PATH)/clusterRole.yaml: $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/clusterRole.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/clusterRole.yaml
	kustomize build --reorder legacy config/helm/rbac | \
	kustomize cfg grep --annotate=false 'kind=ClusterRole' | \
	kustomize cfg grep --annotate=false --invert-match 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/clusterRole.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/clusterRole.yaml

$(CHART_TEMPLATE_PATH)/role.yaml: $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/role.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/role.yaml
	kustomize build --reorder legacy config/helm/rbac | \
	kustomize cfg grep --annotate=false 'kind=Role' | \
	kustomize cfg grep --annotate=false --invert-match 'kind=RoleBinding' | \
	kustomize cfg grep --annotate=false --invert-match 'kind=ClusterRole' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/role.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/role.yaml

$(CHART_TEMPLATE_PATH)/clusterRoleBinding.yaml: $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/clusterRoleBinding.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/clusterRoleBinding.yaml
	kustomize build --reorder legacy config/helm/rbac | \
	kustomize cfg grep --annotate=false 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/clusterRoleBinding.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/clusterRoleBinding.yaml

$(CHART_TEMPLATE_PATH)/roleBinding.yaml: $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/roleBinding.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/roleBinding.yaml
	kustomize build --reorder legacy config/helm/rbac | \
	kustomize cfg grep --annotate=false 'kind=RoleBinding' | \
	kustomize cfg grep --annotate=false --invert-match 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/roleBinding.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/roleBinding.yaml

$(CHART_TEMPLATE_PATH)/deployment.yaml: $(wildcard config/helm/deployment/*) $(wildcard config/manager/*) $(wildcard config/config/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/deployment.yaml
	kustomize build --reorder legacy config/helm/deployment | \
	kustomize cfg grep --annotate=false 'kind=Deployment' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/deployment.yaml

$(CHART_CAKEYPAIR_OPERATOR)/Chart.yaml: $(CHART_CAKEYPAIR_OPERATOR)/Chart.template.yaml
	echo '# $(DO_NOT_EDIT)' > $(CHART_CAKEYPAIR_OPERATOR)/Chart.yaml
	cat $(CHART_CAKEYPAIR_OPERATOR)/Chart.template.yaml \
		| sed "s/{{ VERSION }}/$(RELEASE_VERSION)/g" >> $(CHART_CAKEYPAIR_OPERATOR)/Chart.yaml