package alitest

type operationRunContext struct {
	url        string
	verb       string
	parameters []OpenApiParameter
}
