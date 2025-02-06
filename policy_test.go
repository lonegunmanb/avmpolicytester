package avmpolicytester

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var skippedRegoKeyWords = []string{
	"test",
	"utils",
}
var planFields = []string{
	"resource_changes",
	"configuration",
	"terraform_version",
	"planned_values",
	"output_changes",
	"format_version",
}

func TestPolicy(t *testing.T) {
	rootDir := os.Getenv("POLICY_DIR")
	if rootDir == "" {
		t.Skip("environment varialbe POLICY_DIR not set")
	}
	var err error
	rootDir, err = filepath.Abs(filepath.Clean(rootDir))
	require.NoError(t, err, "failed to get absolute path of POLICY_DIR")
	var regoFiles []string
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, skipWord := range skippedRegoKeyWords {
			if strings.Contains(filepath.Base(path), skipWord) {
				return nil
			}
		}
		if !info.IsDir() && filepath.Ext(path) == ".rego" {
			regoFiles = append(regoFiles, path)
		}
		return nil
	})
	require.NoError(t, err, "failed to walk through POLICY_DIR")
	var utilsRego []string
	utilsRegoStr := os.Getenv("UTILS_REGO")
	if utilsRegoStr != "" {
		utilsRego = strings.Split(utilsRegoStr, ",")
	}
	for _, filePath := range regoFiles {
		testPolicyFile(t, filePath, utilsRego)
	}
}

func testPolicyFile(t *testing.T, filePath string, utilsRego []string) {
	t.Run(filePath, func(t *testing.T) {
		validCases, invalidCases := parseJsonCases(t, filePath)

		for caseName, payload := range validCases {
			t.Run(fmt.Sprintf("VALID CASE_%s", caseName), func(t *testing.T) {
				result, err := testCase(t, caseName, payload, append(utilsRego, filePath))
				require.NoError(t, err, "failed to run test for %s", caseName)
				assert.True(t, result)
			})
		}
		for caseName, payload := range invalidCases {
			t.Run(fmt.Sprintf("INVALID CASE_%s", caseName), func(t *testing.T) {
				result, err := testCase(t, caseName, payload, append(utilsRego, filePath))
				require.NoError(t, err, "failed to run test for %s", caseName)
				assert.False(t, result)
			})
		}
	})
}

func parseJsonCases(t *testing.T, filePath string) (map[string]any, map[string]any) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("file %s does not exist", filePath)
	}
	if filepath.Ext(filePath) != ".rego" {
		t.Fatalf("file %s is not a .rego file", filePath)
	}
	mockFilePath := strings.TrimSuffix(filePath, ".rego") + ".mock.json"
	if _, err := os.Stat(mockFilePath); os.IsNotExist(err) {
		t.Fatalf("mock file %s does not exist", mockFilePath)
	}
	mockFileContent, err := os.ReadFile(mockFilePath)
	require.NoError(t, err, "failed to read mock file %s", mockFilePath)
	var mockData map[string]any
	err = json.Unmarshal(mockFileContent, &mockData)
	require.NoError(t, err, "failed to unmarshal mock file %s", mockFilePath)
	mockField, ok := mockData["mock"]
	require.True(t, ok, "mock field not found in %s", mockFilePath)
	mockMap := mockField.(map[string]any)
	validCases := readCases(t, true, mockMap, mockFilePath)
	invalidCases := readCases(t, false, mockMap, mockFilePath)
	return validCases, invalidCases
}

func readCases(t *testing.T, valid bool, mockMap map[string]any, filePath string) map[string]any {
	result := make(map[string]any)
	fieldKey := "valid"
	if !valid {
		fieldKey = "invalid"
	}
	field, ok := mockMap[fieldKey]
	if ok {
		m, ok := field.(map[string]any)
		require.True(t, ok, "fieldKey field in mock of %s is not a map", fieldKey, filePath)

		for _, pf := range planFields {
			if containsKey(m, pf) {
				result["default"] = field
				return result
			}
		}
		return m
	}
	for name, value := range mockMap {
		if (strings.HasPrefix(name, fieldKey)) || (valid && !strings.HasPrefix(name, "invalid")) {
			result[name] = value
		}
	}
	return result
}

func testCase(t *testing.T, caseName string, payload any, policyFiles []string) (bool, error) {
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err, "failed to marshal payload for %s", caseName)
	payloadStr := string(payloadBytes)
	tmpRegoFile, err := os.CreateTemp("", "mock_json*.json")
	require.NoError(t, err, "failed to create temp mock json file")
	defer func() {
		_ = os.RemoveAll(tmpRegoFile.Name())
	}()
	_, err = tmpRegoFile.WriteString(payloadStr)
	require.NoError(t, err, "failed to write payload to temp mock json file")
	cmdArgs := []string{"test", "--all-namespaces"}
	for _, policyFile := range policyFiles {
		policyFile, err = filepath.Abs(policyFile)
		require.NoError(t, err, "failed to get absolute path of %s", policyFile)
		cmdArgs = append(cmdArgs, "-p", policyFile)
	}
	cmdArgs = append(cmdArgs, tmpRegoFile.Name())
	cmd := exec.Command("conftest", cmdArgs...)
	return cmd.Run() == nil, nil
}

func containsKey[T any](m map[string]T, key string) bool {
	_, ok := m[key]
	return ok
}
