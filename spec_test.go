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
