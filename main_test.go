package main

import (
	"os"
	"strings"
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

// Helper function for Go versions before 1.21 which don't have min in standard library
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
