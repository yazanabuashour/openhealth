package openapi_test

import (
	"os"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPIContractIsValid(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("openapi.yaml")
	if err != nil {
		t.Fatalf("load spec: %v", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		t.Fatalf("validate spec: %v", err)
	}
}

func TestOpenAPIContractIsSingleUserAndStable(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec := string(content)

	if strings.Contains(spec, "UnauthorizedErrorResponse") {
		t.Fatalf("spec should not expose auth responses in single-user mode")
	}
	if strings.Contains(spec, "userId:") {
		t.Fatalf("spec should not expose userId fields")
	}
	if !strings.Contains(spec, `$ref: "#/components/schemas/HealthWeightEntryList"`) {
		t.Fatalf("weight list should keep the items envelope")
	}
	if !strings.Contains(spec, "enum: [30d, 90d, 1y, all]") {
		t.Fatalf("weight trend enum drifted")
	}
	if !strings.Contains(spec, "correlationId") {
		t.Fatalf("error envelope should expose a correlationId")
	}
	for _, operationID := range []string{
		"createHealthBloodPressure",
		"replaceHealthBloodPressure",
		"deleteHealthBloodPressure",
		"createHealthMedication",
		"replaceHealthMedication",
		"deleteHealthMedication",
		"createHealthLabCollection",
		"replaceHealthLabCollection",
		"deleteHealthLabCollection",
	} {
		if !strings.Contains(spec, "operationId: "+operationID) {
			t.Fatalf("spec missing operationId %s", operationID)
		}
	}
	if !strings.Contains(spec, "CreateHealthLabCollectionRequest") {
		t.Fatalf("spec should expose lab collection write request")
	}
	if !strings.Contains(spec, "updatedAt") || !strings.Contains(spec, "deletedAt") {
		t.Fatalf("lab collection lifecycle fields should be exposed")
	}
}
