package main

import (
	"context"

	"github.com/toversus/env-injector/pkg/webhook/injector"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
)

func NewMutatingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	return injector.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"env-injector.toversus.dev",

		// The path on which to serve the webhook.
		"/inject",
	)
}

func main() {
	ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
		ServiceName: "env-injector",
		Port:        10443,
		SecretName:  "env-injector-certs",
	})

	sharedmain.WebhookMainWithContext(ctx, "env-injector",
		certificates.NewController,
		NewMutatingAdmissionController,
	)
}
