package main

import (
	"log"
	"net/http"

	genericfilters "cloudcome.net/k8s-handler/filter"
	k8shandler "cloudcome.net/k8s-handler/handler"
	"cloudcome.net/k8s-handler/webservice"
)

var (

	BuildHandlerChainFunc =  DefaultBuildHandlerChain
	DefaultLegacyAPIPrefix = "/apis"

)

//GetHandler get http.Handlers used by the API server
func GetHandler() *k8shandler.APIServerHandler {
	name:= "k8s-handler"
	handlerChainBuilder := func(handler http.Handler) http.Handler {
		return BuildHandlerChainFunc(handler)
	}

	 apiServerHandler := k8shandler.NewAPIServerHandler(name, handlerChainBuilder, nil)
	 return apiServerHandler
}


// InstallAPIGroups adds a service to GoRestfulContainer
func InstallAPIGroups() *k8shandler.APIServerHandler {

	//InstallLegacyAPIGroup(DefaultLegacyAPIPrefix, &apis[0])
	handler:= GetHandler()
	handler.GoRestfulContainer.Add(webservice.NewLegacyRootAPIHandler(DefaultLegacyAPIPrefix).WebService())
	return handler

}

//DefaultBuildHandlerChain build custom handler chains by decorating the apiHandler.
func DefaultBuildHandlerChain(apiHandler http.Handler) http.Handler {

	//handler := genericapifilters.WithAuthorization(apiHandler, c.Authorization.Authorizer, c.Serializer)
	//handler = genericfilters.WithMaxInFlightLimit(handler, c.MaxRequestsInFlight, c.MaxMutatingRequestsInFlight, c.LongRunningFunc)
	//handler = genericapifilters.WithImpersonation(handler, c.Authorization.Authorizer, c.Serializer)
	//handler = genericapifilters.WithAudit(handler, c.AuditBackend, c.AuditPolicyChecker, c.LongRunningFunc)
	//failedHandler := genericapifilters.Unauthorized(c.Serializer, c.Authentication.SupportsBasicAuth)
	//failedHandler = genericapifilters.WithFailedAuthenticationAudit(failedHandler, c.AuditBackend, c.AuditPolicyChecker)
	//handler = genericapifilters.WithAuthentication(handler, c.Authentication.Authenticator, failedHandler, c.Authentication.APIAudiences)
	// handler = genericfilters.WithCORS(handler, c.CorsAllowedOriginList, nil, nil, nil, "true")
	//handler = genericfilters.WithTimeoutForNonLongRunningRequests(handler, c.LongRunningFunc, c.RequestTimeout)
	//handler = genericfilters.WithWaitGroup(handler, c.LongRunningFunc, c.HandlerChainWaitGroup)
	//handler = genericapifilters.WithRequestInfo(handler, c.RequestInfoResolver)

	handler := genericfilters.WithPanicRecovery(apiHandler)
	return handler
}



func main() {

	http.Handle("/apis", InstallAPIGroups())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
