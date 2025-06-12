package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	// Test with existing env var
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")
	
	if val := getEnvOrDefault("TEST_VAR", "default"); val != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", val)
	}
	
	// Test with non-existing env var
	if val := getEnvOrDefault("NON_EXISTING_VAR", "default"); val != "default" {
		t.Errorf("Expected 'default', got '%s'", val)
	}
}

func TestGenerateSignature(t *testing.T) {
	timestamp := "1622222222"
	secret := "test_secret"
	signature := generateSignature(timestamp, secret)
	
	// We know the expected output for this specific input
	if signature == "" {
		t.Error("Expected non-empty signature")
	}
}

func TestCreateLarkCard_StatusOverride(t *testing.T) {
	tests := []struct {
		name           string
		droneStatus    string
		pluginStatus   string
		expectedStatus string
	}{
		{
			name:           "Default Success",
			droneStatus:    "success",
			pluginStatus:   "",
			expectedStatus: "green", // Header color for success
		},
		{
			name:           "Default Failure",
			droneStatus:    "failure",
			pluginStatus:   "",
			expectedStatus: "red", // Header color for failure
		},
		{
			name:           "Override Success to Failure",
			droneStatus:    "success",
			pluginStatus:   "failure",
			expectedStatus: "red", // Should be overridden to failure (red)
		},
		{
			name:           "Override Failure to Success",
			droneStatus:    "failure",
			pluginStatus:   "success",
			expectedStatus: "green", // Should be overridden to success (green)
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			if tc.droneStatus != "" {
				os.Setenv("DRONE_BUILD_STATUS", tc.droneStatus)
			} else {
				os.Unsetenv("DRONE_BUILD_STATUS")
			}
			
			if tc.pluginStatus != "" {
				os.Setenv("PLUGIN_STATUS", tc.pluginStatus)
			} else {
				os.Unsetenv("PLUGIN_STATUS")
			}
			
			// Reset environment after test
			defer func() {
				os.Unsetenv("DRONE_BUILD_STATUS")
				os.Unsetenv("PLUGIN_STATUS")
			}()
			
			// Call the function
			card := createLarkCard("v1.0.0")
			
			// Extract and verify the header color
			cardObj, ok := card["card"].(map[string]any)
			if !ok {
				t.Fatal("Expected card object")
			}
			
			header, ok := cardObj["header"].(map[string]any)
			if !ok {
				t.Fatal("Expected header object")
			}
			
			template, ok := header["template"].(string)
			if !ok {
				t.Fatal("Expected template string")
			}
			
			if template != tc.expectedStatus {
				t.Errorf("Expected header color '%s', got '%s'", tc.expectedStatus, template)
			}
		})
	}
}

func TestCreateLarkTextMessage_StatusOverride(t *testing.T) {
	tests := []struct {
		name           string
		droneStatus    string
		pluginStatus   string
		expectedStatus string
	}{
		{
			name:           "Default Success",
			droneStatus:    "success",
			pluginStatus:   "",
			expectedStatus: "✅ PIPELINE SUCCEEDED",
		},
		{
			name:           "Default Failure",
			droneStatus:    "failure",
			pluginStatus:   "",
			expectedStatus: "🚨 PIPELINE FAILED",
		},
		{
			name:           "Override Success to Failure",
			droneStatus:    "success",
			pluginStatus:   "failure",
			expectedStatus: "🚨 PIPELINE FAILED",
		},
		{
			name:           "Override Failure to Success",
			droneStatus:    "failure",
			pluginStatus:   "success",
			expectedStatus: "✅ PIPELINE SUCCEEDED",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			if tc.droneStatus != "" {
				os.Setenv("DRONE_BUILD_STATUS", tc.droneStatus)
			} else {
				os.Unsetenv("DRONE_BUILD_STATUS")
			}
			
			if tc.pluginStatus != "" {
				os.Setenv("PLUGIN_STATUS", tc.pluginStatus)
			} else {
				os.Unsetenv("PLUGIN_STATUS")
			}
			
			// Reset environment after test
			defer func() {
				os.Unsetenv("DRONE_BUILD_STATUS")
				os.Unsetenv("PLUGIN_STATUS")
			}()
			
			// Call the function
			message := createLarkTextMessage("v1.0.0")
			
			// Extract and verify the message content
			contentObj, ok := message["content"].(map[string]any)
			if !ok {
				t.Fatal("Expected content object")
			}
			
			text, ok := contentObj["text"].(string)
			if !ok {
				t.Fatal("Expected text string")
			}
			
			// Check if the text starts with the expected status message
			if len(text) < len(tc.expectedStatus) || text[:len(tc.expectedStatus)] != tc.expectedStatus {
				t.Errorf("Expected message to start with '%s', got '%s'", tc.expectedStatus, text[:min(len(text), len(tc.expectedStatus))])
			}
		})
	}
}

func TestGetProjectVersion(t *testing.T) {
	// Test with tag
	os.Setenv("CI_COMMIT_TAG", "v1.0.0")
	os.Setenv("CI_COMMIT_SHA", "abcdef1234567890")
	defer func() {
		os.Unsetenv("CI_COMMIT_TAG")
		os.Unsetenv("CI_COMMIT_SHA")
	}()
	
	if version := getProjectVersion(); version != "v1.0.0" {
		t.Errorf("Expected 'v1.0.0', got '%s'", version)
	}
	
	// Test with SHA only
	os.Unsetenv("CI_COMMIT_TAG")
	if version := getProjectVersion(); version != "abcdef1" {
		t.Errorf("Expected 'abcdef1', got '%s'", version)
	}
	
	// Test with no env vars
	os.Unsetenv("CI_COMMIT_SHA")
	if version := getProjectVersion(); version != "" {
		t.Errorf("Expected empty string, got '%s'", version)
	}
}

func TestCreateActionButtons(t *testing.T) {
	// Setup test environment
	os.Setenv("CI_PIPELINE_URL", "https://example.com/pipeline")
	os.Setenv("CI_COMMIT_TAG", "v1.0.0")
	os.Setenv("CI_REPO_URL", "https://github.com/user/repo")
	defer func() {
		os.Unsetenv("CI_PIPELINE_URL")
		os.Unsetenv("CI_COMMIT_TAG")
		os.Unsetenv("CI_REPO_URL")
		os.Unsetenv("PLUGIN_BUTTONS")
	}()
	
	// Test with all buttons
	actions := createActionButtons()
	if len(actions) != 2 {
		t.Errorf("Expected 2 buttons, got %d", len(actions))
	}
	
	// Test with filtered buttons
	os.Setenv("PLUGIN_BUTTONS", "pipeline")
	actions = createActionButtons()
	if len(actions) != 1 {
		t.Errorf("Expected 1 button, got %d", len(actions))
	}
	
	// Test with commit instead of tag
	os.Unsetenv("CI_COMMIT_TAG")
	os.Unsetenv("PLUGIN_BUTTONS")
	os.Setenv("CI_PIPELINE_FORGE_URL", "https://github.com/user/repo/commit/abc123")
	defer os.Unsetenv("CI_PIPELINE_FORGE_URL")
	
	actions = createActionButtons()
	if len(actions) != 2 {
		t.Errorf("Expected 2 buttons, got %d", len(actions))
	}
}

func TestPrintBuildInfo(t *testing.T) {
	// This is mostly a visual test, but we can at least ensure it doesn't crash
	os.Setenv("CI_REPO_NAME", "test-repo")
	os.Setenv("CI_COMMIT_BRANCH", "main")
	os.Setenv("DRONE_BUILD_STATUS", "success")
	defer func() {
		os.Unsetenv("CI_REPO_NAME")
		os.Unsetenv("CI_COMMIT_BRANCH")
		os.Unsetenv("DRONE_BUILD_STATUS")
	}()
	
	// Just make sure it doesn't panic
	printBuildInfo("v1.0.0")
}

func TestPrintDebugInfo(t *testing.T) {
	// This is mostly a visual test, but we can at least ensure it doesn't crash
	messageBytes := []byte(`{"msg_type":"text","content":{"text":"Test message"}}`)
	
	// Just make sure it doesn't panic
	printDebugInfo(messageBytes)
}

func TestMain_MissingWebhookURL(t *testing.T) {
	// Save original osExit and restore it after the test
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	// Clear any existing environment variables
	os.Unsetenv("PLUGIN_WEBHOOK_URL")

	// Mock os.Exit
	exitCalled := false
	exitCode := 0
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	// Call main
	main()

	// Verify that os.Exit was called with the expected code
	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestMain_WithWebhook(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": 0, "message": "success"}`)) // Lark API success response
	}))
	defer testServer.Close()

	// Save original osExit and restore it after the test
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	// Set up environment variables
	os.Setenv("PLUGIN_WEBHOOK_URL", testServer.URL)
	os.Setenv("CI_REPO_NAME", "test-repo")
	os.Setenv("CI_COMMIT_BRANCH", "main")
	os.Setenv("CI_COMMIT_SHA", "abcdef1234567890")
	os.Setenv("DRONE_BUILD_STATUS", "success")
	defer func() {
		os.Unsetenv("PLUGIN_WEBHOOK_URL")
		os.Unsetenv("CI_REPO_NAME")
		os.Unsetenv("CI_COMMIT_BRANCH")
		os.Unsetenv("CI_COMMIT_SHA")
		os.Unsetenv("DRONE_BUILD_STATUS")
	}()

	// Mock os.Exit to prevent actual exit
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
	}

	// Call main
	main()

	// Verify that os.Exit was not called
	if exitCalled {
		t.Error("os.Exit should not have been called")
	}
}

func TestSendMessage(t *testing.T) {
	// Create a test server that mimics Lark API
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the request is a POST
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify that the content type is application/json
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" && contentType != "" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Verify that the body is valid JSON
		var requestData map[string]interface{}
		if err := json.Unmarshal(body, &requestData); err != nil {
			t.Fatalf("Request body is not valid JSON: %v", err)
		}

		// Return a successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": 0, "message": "success"}`)) // Lark API success response
	}))
	defer testServer.Close()

	// Test with success response
	messageBytes := []byte(`{"msg_type":"text","content":{"text":"Test message"}}`)
	sendMessage(testServer.URL, messageBytes)

	// Test with error response
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code": 1, "message": "error"}`)) // Lark API error response
	}))
	defer errorServer.Close()

	// We need to use a custom exit function to prevent the test from exiting
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()
	
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		if code != 1 {
			t.Errorf("Expected exit code 1, got %d", code)
		}
	}

	// This should call osExit(1) due to the error response
	sendMessage(errorServer.URL, messageBytes)

	if !exitCalled {
		t.Error("Expected os.Exit to be called")
	}
}

// Helper function for Go versions before 1.21 which don't have min in standard library
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
