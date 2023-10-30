package alitest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

//go:generate stringer -type ParameterLocation

type ParameterLocation int

const (
	Query ParameterLocation = iota
	Header
	Path
	Cookie
)

type OpenApiDocument struct {
	Info       ApiInfo                `json:"info" yaml:"info"`
	Paths      map[string]OpenApiPath `json:"paths" yaml:"paths"`
	Components ApiComponents          `json:"components" yaml:"components"`
}

type ApiInfo struct {
	Title string `json:"title" yaml:"title"`

	// TODO: determine if I should use it ?
	Summary     string `json:"summary" yaml:"summary"`
	Description string `json:"description" yaml:"description"`
	Version     string `json:"version" yaml:"version"`
}

type ApiComponents struct {
	Schemas map[string]interface{} `json:"schemas" yaml:"schemas"`
	// TODO implements responses, parameters, examples, requestBodies, headers, securitySchemes, links, callbacks, pathItems
}

type OpenApiPath struct {
	Summary     string            `json:"summary" yaml:"summary"`
	Description string            `json:"description" yaml:"description"`
	Get         *OpenApiOperation `json:"get" yaml:"get"`
	Put         *OpenApiOperation `json:"put" yaml:"put"`
	Post        *OpenApiOperation `json:"post" yaml:"post"`
	Delete      *OpenApiOperation `json:"delete" yaml:"delete"`
	Options     *OpenApiOperation `json:"options" yaml:"options"`
	Head        *OpenApiOperation `json:"head" yaml:"head"`
	Patch       *OpenApiOperation `json:"patch" yaml:"patch"`
	Trace       *OpenApiOperation `json:"trace" yaml:"trace"`
	// TODO test put, post, delete, options, head, patch, trace, servers, $ref
}

func (p OpenApiPath) CountOperations() int {
	var count int

	if p.Get != nil {
		count++
	}

	if p.Put != nil {
		count++
	}

	if p.Post != nil {
		count++
	}

	if p.Delete != nil {
		count++
	}

	if p.Options != nil {
		count++
	}

	if p.Head != nil {
		count++
	}

	if p.Patch != nil {
		count++
	}

	if p.Trace != nil {
		count++
	}
	return count
}

func (o OpenApiPath) runTests(t *testing.T, url string) {
	// TODO check the operations + the path
	t.Run(fmt.Sprintf("run tests for path %s", url), func(t *testing.T) {

		if o.Get != nil {
			o.Get.runTests(t, url, http.MethodGet)
		}

		// TODO check the response schema if any

	})
}

type OpenApiOperation struct {
	Summary     string             `json:"summary" yaml:"summary"`
	Description string             `json:"description" yaml:"description"`
	Parameters  []OpenApiParameter `json:"parameters" yaml:"parameters"`
	Responses   OpenApiResponses   `json:"responses" yaml:"responses"`
}

func (o OpenApiOperation) runTests(t *testing.T, url, verb string) {
	t.Run(fmt.Sprintf("found an operation: %s", o.Description), func(t *testing.T) {
		ctx := operationRunContext{url: url, verb: verb, parameters: o.Parameters}
		if o.Responses.Ok != nil {
			t.Run("Run tests for status: 200", func(t *testing.T) {
				o.Responses.Ok.runTest(t, ctx, http.StatusOK)
			})
		}
		if o.Responses.BadRequest != nil {
			t.Run("Run tests for status: 400", func(t *testing.T) {
				o.Responses.BadRequest.runTest(t, ctx, http.StatusBadRequest)
			})
		}
		if o.Responses.NotFound != nil {
			t.Run("Run tests for status: 404", func(t *testing.T) {
				o.Responses.NotFound.runTest(t, ctx, http.StatusNotFound)
			})
		}
	})
}

type OpenApiParameter struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	In          ParameterLocation `json:"in" yaml:"in"`
	Required    bool              `json:"required" yaml:"required"`
}

func (i ParameterLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToLower(i.String()))
}

func (i *ParameterLocation) UnmarshalYAML(data *yaml.Node) (err error) {
	var inVal string
	var result ParameterLocation
	if err := data.Decode(&inVal); err != nil {
		return err
	}
	switch strings.ToLower(inVal) {
	case strings.ToLower(Query.String()):
		result = Query
	case strings.ToLower(Header.String()):
		result = Header
	case strings.ToLower(Path.String()):
		result = Path
	case strings.ToLower(Cookie.String()):
		result = Cookie
	default:
		// TODO test me
		return fmt.Errorf("unknown parameter location : %s", inVal)
	}
	*i = result
	return nil
}

type OpenApiResponses struct {
	Ok         *OpenApiResponse `json:"200" yaml:"200"`
	BadRequest *OpenApiResponse `json:"400" yaml:"400"`
	NotFound   *OpenApiResponse `json:"404" yaml:"404"`
}

type OpenApiResponse struct {
	Description   string                  `json:"description" yaml:"description"`
	Json          OpenApiResponseContent  `json:"application/json" yaml:"application/json"`
	AliParameters map[string]AliParameter `json:"x-ali-parameters" yaml:"x-ali-parameters"`
}

func (o OpenApiResponse) runTest(t *testing.T, ctx operationRunContext, status int) {
	// TODO compute the length ?
	resolvedURL := ctx.url

	for _, params := range ctx.parameters {
		if params.In != Path {
			continue
		}
		// TODO handle not provided parameter => failed
		paramValue, _ := o.AliParameters[params.Name]
		resolvedURL = strings.ReplaceAll(resolvedURL, fmt.Sprintf("{%s}", params.Name), fmt.Sprintf("%s", paramValue.Value))
	}

	// TODO handle the error
	request, _ := http.NewRequest(ctx.verb, resolvedURL, nil)
	request.Header.Add("Accept", "application/json")

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	response, _ := netClient.Do(request)

	if response.StatusCode != status {
		t.Fatalf("Expect status %d but got status %d", status, response.StatusCode)
	}
}

type OpenApiResponseContent struct {
	Schema interface{} `json:"schema" yaml:"schema"`
}

type AliParameter struct {
	Value any `json:"value" yaml:"value"`
}
