TARGET_NAMESPACE := $(shell kubectl config current-context | cut -d/ -f1)
MANIFESTS_DIR ?= manifests

undeploy:
	kubectl delete -f $(MANIFESTS_DIR) 2> /dev/null || echo "undeployed only available resources"

deploy: undeploy
	sed "s/TARGET_NAMESPACE/$(TARGET_NAMESPACE)/" $(MANIFESTS_DIR)/$(TASK_NAME)-cluster-rbac.yaml | kubectl apply -f -
	set -e; $(foreach SUBTASK_NAME, $(SUBTASK_NAMES), kubectl apply -f $(MANIFESTS_DIR)/$(SUBTASK_NAME);)

deploy-namespace: undeploy
	kubectl apply -f manifests/$(TASK_NAME)-namespace-rbac.yaml
	set -e; $(foreach SUBTASK_NAME, $(SUBTASK_NAMES), kubectl apply -f $(MANIFESTS_DIR)/$(SUBTASK_NAME);)


.PHONY: \
	undeploy \
	deploy \
	deploy-namespace