package injector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"gomodules.xyz/jsonpatch/v3"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/webhook"
)

const (
	EnvKey              = "TARGET"
	EnvValue            = "Sample Injector v1"
	TargetAppLabelKey   = "app"
	TargetAppLabelValue = "helloworld-go"
)

// Admit implements AdmissionController
func (ac *reconciler) Admit(ctx context.Context, request *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	logger := logging.FromContext(ctx)
	switch request.Operation {
	case admissionv1beta1.Create, admissionv1beta1.Update:
	default:
		logger.Infof("Unhandled webhook operation, letting it through %v", request.Operation)
		return &admissionv1beta1.AdmissionResponse{Allowed: true}
	}

	patch, err := injectEnvVar(ctx, request)
	if err != nil {
		return webhook.MakeErrorStatus("mutation failed: %v", err)
	}

	return &admissionv1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patch,
		PatchType: func() *admissionv1beta1.PatchType {
			pt := admissionv1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func injectEnvVar(ctx context.Context, req *admissionv1beta1.AdmissionRequest) ([]byte, error) {
	logger := logging.FromContext(ctx)
	kind := req.Kind
	newBytes := req.Object.Raw
	auid := req.UID

	// Why, oh why are these different types...
	gvk := schema.GroupVersionKind{
		Group:   kind.Group,
		Version: kind.Version,
		Kind:    kind.Kind,
	}

	resourceGVK := appsv1.SchemeGroupVersion.WithKind("Deployment")
	if gvk != resourceGVK {
		logger.Errorf("Unhandled kind: %v", gvk)
		return nil, fmt.Errorf("unhandled kind: %v", gvk)
	}

	var deploy appsv1.Deployment
	if len(newBytes) != 0 {
		newDecoder := json.NewDecoder(bytes.NewBuffer(newBytes))
		if err := newDecoder.Decode(&deploy); err != nil {
			return nil, fmt.Errorf("cannot decode incoming new object: %w", err)
		}
	}
	if !isTarget(ctx, deploy) {
		return nil, nil
	}

	// Got mutated object reference
	mutated := mutate(ctx, deploy)
	mutatedJSON, err := json.Marshal(mutated)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal mutated object: %w", err)
	}

	patch, err := jsonpatch.CreatePatch(newBytes, mutatedJSON)
	if err != nil {
		return nil, fmt.Errorf("cannot create JSON patch: %w", err)
	}

	marshalledPatch, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("cannot marshel JSON patch: %w", err)
	}
	logger.Debugf("json patch for request %s: %s", auid, string(marshalledPatch))

	return marshalledPatch, nil
}

// isTarget checks if the deployment has target label and meets the other requirements
func isTarget(ctx context.Context, deploy appsv1.Deployment) bool {
	logger := logging.FromContext(ctx)
	containers := deploy.Spec.Template.Spec.Containers
	val, ok := deploy.Spec.Template.ObjectMeta.Labels[TargetAppLabelKey]
	if !ok {
		logger.Infof("Target app label key (%s) is missing, skip injection to %s",
			TargetAppLabelKey, deploy.GetObjectMeta().GetName())
		return false
	}
	if val != TargetAppLabelValue {
		logger.Infof("Target app label (%s: %s) is missing, skip injection to %s",
			TargetAppLabelKey, TargetAppLabelValue, deploy.GetObjectMeta().GetName())
		return false
	}

	if len(containers) != 1 {
		logger.Infof("Unsupported number of containers (> 1) defined in spec, skip injection to %s",
			deploy.GetObjectMeta().GetName())
		return false
	}
	return true
}

// mutate is the main logic to inject EnvVar to the deployment
func mutate(ctx context.Context, deploy appsv1.Deployment) *appsv1.Deployment {
	logger := logging.FromContext(ctx)
	var injected bool
	containers := deploy.Spec.Template.Spec.Containers
	for _, env := range containers[0].Env {
		if env.Name != EnvKey {
			continue
		}
		logger.Infof("The env key (%s) in deployment %s is overwritten to %s",
			EnvKey, deploy.GetObjectMeta().GetName(), EnvValue)
		env.Value = EnvValue
		injected = true
	}
	if !injected {
		logger.Infof("The env (%s: %s) in deployment %s is added",
			EnvKey, EnvValue, deploy.GetObjectMeta().GetName())
		containers[0].Env = append(containers[0].Env, corev1.EnvVar{
			Name:  EnvKey,
			Value: EnvValue,
		})
	}
	return &deploy
}
