package alitest_test

import (
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/toolzup/alitest"
)

//go:embed dataset/petstore.yaml
var petstoreYaml string

//go:embed dataset/petstore.json
var petstoreJson string

//go:embed dataset/wrong.json
var wrongJson string

//go:embed dataset/wrong.yaml
var wrongYaml string

//go:embed dataset/simple_get_specification.yaml
var simpleGet string

func TestParse(t *testing.T) {
	type testCase struct {
		description string
		fnToTest    func() (alitest.IntegrationTestSuite, error)
		err         error
	}

	testCases := []testCase{
		{
			description: "nominal yaml content",
			fnToTest:    func() (alitest.IntegrationTestSuite, error) { return alitest.ParseString(petstoreYaml) },
		},
		{
			description: "nominal yaml file",
			fnToTest:    func() (alitest.IntegrationTestSuite, error) { return alitest.ParseFile("./dataset/petstore.yaml") },
		},
		{
			description: "nominal json content",
			fnToTest:    func() (alitest.IntegrationTestSuite, error) { return alitest.ParseString(petstoreJson) },
		},
		{
			description: "wrong yaml content",
			fnToTest:    func() (alitest.IntegrationTestSuite, error) { return alitest.ParseString(wrongYaml) },
			err:         errors.New("cannot unmarshal into an open api document. Please check the input."),
		},
		{
			description: "wrong yaml file",
			fnToTest:    func() (alitest.IntegrationTestSuite, error) { return alitest.ParseFile("./dataset/wrong.yaml") },
			err:         errors.New("cannot unmarshal into an open api document. Please check the input."),
		},
		{
			description: "wrong json content",
			fnToTest:    func() (alitest.IntegrationTestSuite, error) { return alitest.ParseString(wrongJson) },
			err:         errors.New("cannot unmarshal into an open api document. Please check the input."),
		},
		{
			description: "wrong json file",
			fnToTest:    func() (alitest.IntegrationTestSuite, error) { return alitest.ParseFile("./dataset/wrong.json") },
			err:         errors.New("cannot unmarshal into an open api document. Please check the input."),
		},
		{
			description: "unexisting file",
			fnToTest:    func() (alitest.IntegrationTestSuite, error) { return alitest.ParseFile("./dataset/do_not_exist.yaml") },
			err:         errors.New("open ./dataset/do_not_exist.yaml: no such file or directory"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			integrationSuite, err := tc.fnToTest()

			if tc.err != nil && err.Error() != tc.err.Error() {
				t.Fatalf("expect error %s, but got %s", tc.err.Error(), err.Error())
			} else if err != nil {
				// successful test case with error. Stop it now.
				return
			}

			if tc.err == nil && err != nil {
				t.Fatalf("expect nil error, but got %v", err)
			}

			if integrationSuite.String() != "Swagger Petstore integration test suite" {
				t.Fatalf("expect %s, but got %s", "Swagger Petstore integration test suite", integrationSuite.String())
			}

			if integrationSuite.EndpointCount() != 20 {
				t.Fatalf("expect %d endpoint in integration test suite, but got %d", 20, integrationSuite.EndpointCount())
			}
		})
	}
}

// TestRun is a nominal global test for a simple specification file.
// There is one endpoint, three status code (200, 400, 404)
func TestRun(t *testing.T) {
	type PetResult struct {
		Id   int64  `json:"id"`
		Name string `json:"name"`
	}

	type ErrResult struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	var nominalCalled bool
	var badFormatCalled bool
	var notFoundCalled bool

	integrationSuite, err := alitest.ParseString(simpleGet)

	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		if r.Method != http.MethodGet {
			t.Fatalf("expect %s, but got %s", http.MethodGet, r.Method)
			w.WriteHeader(http.StatusInternalServerError)
		}

		encoder := json.NewEncoder(w)
		t.Log("---------------------------------------------- : ", r.URL.Path)

		if r.URL.Path == "/pet/0a62b985-17b5-48ee-ae04-ae0c99cb1109" {
			// Sucess case
			err = encoder.Encode(PetResult{
				Id:   321654,
				Name: "Medor",
			})
			nominalCalled = true
		} else if r.URL.Path == "/pet/bad-format" {
			w.WriteHeader(http.StatusBadRequest)
			err = encoder.Encode(ErrResult{
				Type:        "badIDFormat",
				Description: "The given id, 'bad-format', is invalid. Wait for a UUID V4.",
			})
			badFormatCalled = true
		} else if r.URL.Path == "/pet/9051be9a-5aa2-4912-9786-01ffe22401d7" {
			w.WriteHeader(http.StatusNotFound)
			err = encoder.Encode(ErrResult{
				Type:        "PetNotFound",
				Description: "No pet found for the given id : '9051be9a-5aa2-4912-9786-01ffe22401d7'",
			})
			notFoundCalled = true
		}

		if err != nil {
			t.Fatalf("expect nil error, but got %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	integrationSuite.Run(t, alitest.RunParameters{URL: srv.URL})

	if !nominalCalled {
		t.Fatal("nominal case not covered")
	}

	if !badFormatCalled {
		t.Fatal("bad format case not covered")
	}

	if !notFoundCalled {
		t.Fatal("not found case not covered")
	}
}

// Je veux pouvoir lancer tous les tests avec des valeurs par défaut
// Je veux un moyen pratique, facile et maintenable d'injecter des valeurs d'input
// Je veux pouvoir faire une vérification simple des valeurs de retour
// Je veux, optionnellement, faire une vérification + avancé des valeurs de retour.
// Je veux avoir une structure hierarchiques des tests suivant les path des endpoint
// Je veux que chaque test non implémenté plante d'une façon ou une autre par défaut
// Je veux pouvoir exposer un setup optionnel de l'env de test d'intégration / API via une fonction de cleanup / seed (avec les paramètres)
// Je veux pouvoir "skip" un test non encore implémenté (custom attribute ou via appel de fonction)
// Je veux pouvoir gérer les code d'authorizations
// Je veux pouvoir gérer du XML plus tard...
// Je veux vérifier les attributs obligatoires/optionnels
// Je veux vérifier le format des types des attributs
// Je veux pouvoir vérifier le schéma de la spec OpenApi
