package alitest

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

type (
	// IntegrationTestSuite .... TODO, complete me
	IntegrationTestSuite struct {
		doc OpenApiDocument
	}

	// TODO: doc me
	RunParameters struct {
		URL string
	}
)

func ParseFile(fileName string) (IntegrationTestSuite, error) {
	var testSuite IntegrationTestSuite
	var doc OpenApiDocument
	f, err := os.Open(fileName)

	if err != nil {
		return testSuite, err
	}

	yamlDecoder := yaml.NewDecoder(f)
	err = yamlDecoder.Decode(&doc)

	if err != nil {
		return testSuite, errors.New("cannot unmarshal into an open api document. Please check the input.")
	}

	testSuite.doc = doc

	return testSuite, nil
}

func ParseString(specContent string) (IntegrationTestSuite, error) {
	var testSuite IntegrationTestSuite
	var doc OpenApiDocument

	err := yaml.Unmarshal([]byte(specContent), &doc)
	if err != nil {
		return testSuite, errors.New("cannot unmarshal into an open api document. Please check the input.")
	}

	testSuite.doc = doc

	return testSuite, nil
}

func (s IntegrationTestSuite) EndpointCount() int {
	var count int
	for _, path := range s.doc.Paths {
		count += path.CountOperations()
	}
	return count
}

func (s *IntegrationTestSuite) Run(t *testing.T, parameters RunParameters) {
	t.Run(fmt.Sprintf("api test for %s", s.doc.Info.Title), func(t *testing.T) {
		for path, pathObject := range s.doc.Paths {
			// TODO improve that: ensure there is only one "/"
			pathObject.runTests(t, fmt.Sprintf("%s%s", parameters.URL, path))
		}
	})
}

func (s IntegrationTestSuite) String() string {
	return fmt.Sprintf("%s integration test suite", s.doc.Info.Title)
}
