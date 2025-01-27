package vcfa

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func convertToStringMap(param map[string]interface{}) map[string]string {
	temp := make(map[string]string)
	for k, v := range param {
		temp[k] = v.(string)
	}
	return temp
}

// convertSchemaSetToSliceOfStrings accepts Terraform's *schema.Set object and converts it to slice
// of strings.
// This is useful for extracting values from a set of strings
func convertSchemaSetToSliceOfStrings(param *schema.Set) []string {
	paramList := param.List()
	result := make([]string, len(paramList))
	for index, value := range paramList {
		result[index] = fmt.Sprint(value)
	}

	return result
}

// convertTypeListToSliceOfStrings accepts Terraform's TypeList structure `[]interface{}` and
// converts it to slice of strings.
func convertTypeListToSliceOfStrings(param []interface{}) []string {
	result := make([]string, len(param))
	for i, v := range param {
		result[i] = v.(string)
	}
	return result
}

// extractIdsFromOpenApiReferences extracts []string with IDs from []types.OpenApiReference which contains ID and Names
func extractIdsFromOpenApiReferences(refs []types.OpenApiReference) []string {
	resultStrings := make([]string, len(refs))
	for index := range refs {
		resultStrings[index] = refs[index].ID
	}

	return resultStrings
}

// convertSliceOfStringsToOpenApiReferenceIds converts []string to []types.OpenApiReference by filling
// types.OpenApiReference.ID fields
func convertSliceOfStringsToOpenApiReferenceIds(ids []string) []types.OpenApiReference {
	resultReferences := make([]types.OpenApiReference, len(ids))
	for i, v := range ids {
		resultReferences[i].ID = v
	}

	return resultReferences
}

// contains returns true if `sliceToSearch` contains `searched`. Returns false otherwise.
func contains(sliceToSearch []string, searched string) bool {
	found := false
	for _, idInSlice := range sliceToSearch {
		if searched == idInSlice {
			found = true
			break
		}
	}
	return found
}

// Checks if a file exists
func fileExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	fileMode := f.Mode()
	return fileMode.IsRegular()
}

// mustStrToInt will convert string to int and panic if an error while convert occurs
// Note. It is convenient to use for inline type conversions, but the string _must be_ validated before
// e.g. field validation using `ValidateFunc: IsIntAndAtLeast(1), `
func mustStrToInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("failed converting '%s' to int: %s", s, err))
	}
	return v
}
