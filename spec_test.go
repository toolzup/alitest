package alitest_test

import (
	"testing"

	"github.com/toolzup/alitest"
)

func TestMatchBodyResponse(t *testing.T) {

	testCases := []struct {
		description  string
		response     alitest.AliResponse
		bodyResponse []byte
		identical    bool
	}{
		{
			description: "simple object match",
			response: alitest.AliResponse{
				Expected: struct {
					Name string `json:"name"`
				}{Name: "Medor"},
			},
			bodyResponse: []byte(`{"name": "Medor"}`),
			identical:    true,
		},
		{
			description: "object don't match with more attributes",
			response: alitest.AliResponse{
				Expected: struct {
					Name string `json:"name"`
				}{Name: "Medor"},
			},
			bodyResponse: []byte(`{"name": "Medor", "age": 5}`),
			identical:    false,
		},
		{
			description: "object match with more attributes",
			response: alitest.AliResponse{
				Expected: struct {
					Name string `json:"name"`
				}{Name: "Medor"},
				AcceptAdditionalProps: true,
			},
			bodyResponse: []byte(`{"name": "Medor", "age": 5}`),
			identical:    true,
		},
		{
			description: "simple object mismatch. Expected is 'rex'",
			response: alitest.AliResponse{
				Expected: struct {
					Name string `json:"name"`
				}{Name: "Rex"},
			},
			bodyResponse: []byte(`{"name": "Medor"}`),
			identical:    false,
		},
		{
			description: "simple object mismatch. Expected is 'medor'",
			response: alitest.AliResponse{
				Expected: struct {
					Name string `json:"name"`
				}{Name: "Medor"},
			},
			bodyResponse: []byte(`{"name": "Rex"}`),
			identical:    false,
		},
		{
			description: "simple array match",
			response: alitest.AliResponse{
				Expected: []struct {
					Name string `json:"name"`
				}{{Name: "Medor"}},
			},
			bodyResponse: []byte(`[{"name": "Medor"}]`),
			identical:    true,
		},
		{
			description: "array match with additionals props",
			response: alitest.AliResponse{
				Expected: []struct {
					Name string `json:"name"`
				}{{Name: "Medor"}},
				AcceptAdditionalProps: true,
			},
			bodyResponse: []byte(`[{"name": "Medor", "age": 5}]`),
			identical:    true,
		},
		{
			description: "simple array mismatch. Expected is 'rex'",
			response: alitest.AliResponse{
				Expected: []struct {
					Name string `json:"name"`
				}{{Name: "Rex"}},
			},
			bodyResponse: []byte(`[{"name": "Medor"}]`),
			identical:    false,
		},
		{
			description: "simple array mismatch. Expected is 'medor'",
			response: alitest.AliResponse{
				Expected: []struct {
					Name string `json:"name"`
				}{{Name: "Medor"}},
			},
			bodyResponse: []byte(`[{"name": "Rex"}]`),
			identical:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			identical, details := testCase.response.Compare(testCase.bodyResponse)

			if identical != testCase.identical {
				t.Errorf("Expect identifical to be %t, but id %t", testCase.identical, identical)
			}

			t.Log(details)

		})
	}

}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name        string
		response    alitest.OpenApiResponse
		url         string
		params      []alitest.OpenApiParameter
		expectedURL string
	}{
		{
			name:        "no parameters",
			response:    alitest.OpenApiResponse{},
			params:      []alitest.OpenApiParameter{},
			url:         "some/path/without/params",
			expectedURL: "some/path/without/params",
		},
		{
			name: "some query parameters",
			response: alitest.OpenApiResponse{
				AliParameters: map[string]alitest.AliParameter{
					"eventID": {
						Value: "some-event-id",
					},
					"raceID": {
						Value: "some-race-id",
					},
				},
			},
			params: []alitest.OpenApiParameter{
				{
					Name: "eventID",
					In:   alitest.Path,
				},
				{
					Name: "raceID",
					In:   alitest.Path,
				},
			},
			url:         "some/path/{eventID}/races/{raceID}",
			expectedURL: "some/path/some-event-id/races/some-race-id",
		},
		{
			name: "some query parameters",
			response: alitest.OpenApiResponse{
				AliParameters: map[string]alitest.AliParameter{
					"search": {
						Value: "ert+titi",
					},
					"from": {
						Value: "f;gihzdfgkjhfdgj",
					},
				},
			},
			params: []alitest.OpenApiParameter{
				{
					Name: "search",
					In:   alitest.Query,
				},
				{
					Name: "from",
					In:   alitest.Query,
				},
			},
			url:         "some/path/some-event/races/some-race",
			expectedURL: "some/path/some-event/races/some-race?search=ert%2Btiti&from=f%3Bgihzdfgkjhfdgj",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			actualURL := test.response.ResolveURL(test.url, test.params)

			if test.expectedURL != actualURL {
				t.Fatalf("expect %v, got %v", test.expectedURL, actualURL)
			}
		})
	}
}
