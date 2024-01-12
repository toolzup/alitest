package alitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/wI2L/jsondiff"
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

		if o.Post != nil {
			o.Post.runTests(t, url, http.MethodPost)
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
		if o.Responses.Created != nil {
			t.Run("Run tests for status: 201", func(t *testing.T) {
				o.Responses.Created.runTest(t, ctx, http.StatusCreated)
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
		if o.Responses.Expired != nil {
			t.Run("Run tests for status: 419", func(t *testing.T) {
				o.Responses.Expired.runTest(t, ctx, 419)
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
	Created    *OpenApiResponse `json:"201" yaml:"201"`
	BadRequest *OpenApiResponse `json:"400" yaml:"400"`
	NotFound   *OpenApiResponse `json:"404" yaml:"404"`
	Expired    *OpenApiResponse `json:"419" yaml:"419"`
}

type OpenApiResponse struct {
	Description   string                  `json:"description" yaml:"description"`
	Json          OpenApiResponseContent  `json:"application/json" yaml:"application/json"`
	AliParameters map[string]AliParameter `json:"x-ali-parameters" yaml:"x-ali-parameters"`
	AliBody       interface{}             `json:"x-ali-body" yaml:"x-ali-body"`
	AliResponse   *aliResponse            `json:"x-ali-response" yaml:"x-ali-response"`
}

type aliResponse struct {
	// Ignore id array of json pointer string to exclude from check
	Ignore   []string    `json:"ignore" yaml:"ignore"`
	Expected interface{} `json:"expected" yaml:"expected"`
}

func (o OpenApiResponse) runTest(t *testing.T, ctx operationRunContext, status int) {
	var reader io.Reader
	var err error
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
	if o.AliBody != nil {
		reader, err = ioReader(o.AliBody)
	}

	if err != nil {
		t.Fatalf("Got unexpected marshalling error (%v) when performing a %s on %s", err, ctx.verb, resolvedURL)
	}

	// TODO handle the error
	request, _ := http.NewRequest(ctx.verb, resolvedURL, reader)
	request.Header.Add("Accept", "application/json")

	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	response, err := netClient.Do(request)

	if err != nil {
		t.Fatalf("Got unexpected error (%v) when performing a %s on %s", err, ctx.verb, resolvedURL)
	}

	if response.StatusCode != status {
		t.Fatalf("Expect status %d but got status %d", status, response.StatusCode)
	}

	// Stop the process now, no returned data to verify
	if o.AliResponse == nil {
		return
	}

	expectedPayload, err := json.Marshal(o.AliResponse.Expected)

	if err != nil {
		t.Fatalf("Got unexpected marshalling error (%v) when reading expected response from spec for a %s on %s", err, ctx.verb, resolvedURL)
	}

	actualPayload, err := io.ReadAll(response.Body)

	if err != nil {
		t.Fatalf("Got unexpected error (%v) when reading response from %s on %s", err, ctx.verb, resolvedURL)
	}

	opts := []jsondiff.Option{
		jsondiff.Equivalent(),
	}

	if len(o.AliResponse.Ignore) > 0 {
		opts = append(opts, jsondiff.Ignores(o.AliResponse.Ignore...))
	}

	diff, err := jsondiff.CompareJSON(
		actualPayload,
		expectedPayload,
		opts...,
	)

	if len(diff) > 0 {
		t.Fatalf("Got differences on expectedPayload : %s, actualPayload: %s. => %v", string(expectedPayload), string(actualPayload), diff)
	} else {
		t.Logf("Expected payload: %s and actual one : %s do match. Ignored fields : %v", string(expectedPayload), string(actualPayload), o.AliResponse.Ignore)
	}

}

func ioReader(data interface{}) (io.Reader, error) {
	jsonEncoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(jsonEncoded), nil
}

type OpenApiResponseContent struct {
	Schema interface{} `json:"schema" yaml:"schema"`
}

type AliParameter struct {
	Value any `json:"value" yaml:"value"`
}
