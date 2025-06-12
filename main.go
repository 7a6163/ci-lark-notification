package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// osExit is a variable for os.Exit that can be overridden in tests
var osExit = os.Exit

func main() {
	webhookURL := getEnvOrDefault("PLUGIN_WEBHOOK_URL", "")
	if webhookURL == "" {
		fmt.Println("Need to set Lark Webhook URL")
		osExit(1)
	}

	projectVersion := getProjectVersion()

	// Check if using signature verification
	secret := getEnvOrDefault("PLUGIN_SECRET", "")
	useCard := getEnvOrDefault("PLUGIN_USE_CARD", "true") == "true"

	var message map[string]any
	if useCard {
		message = createLarkCard(projectVersion)
	} else {
		message = createLarkTextMessage(projectVersion)
	}

	// Add signature if secret is provided
	if secret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		sign := generateSignature(timestamp, secret)
		message["timestamp"] = timestamp
		message["sign"] = sign
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error creating message JSON: %v\n", err)
		osExit(1)
	}

	if getEnvOrDefault("PLUGIN_DEBUG", "false") == "true" {
		printDebugInfo(messageBytes)
	}

	printBuildInfo(projectVersion)
	
	// Only send message if webhook URL is provided
	if webhookURL != "" {
		sendMessage(webhookURL, messageBytes)
	}
}

func generateSignature(timestamp, secret string) string {
	stringToSign := fmt.Sprintf("%s\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func getProjectVersion() string {
	if tag := getEnvOrDefault("CI_COMMIT_TAG", ""); tag != "" {
		return tag
	}
	if sha := getEnvOrDefault("CI_COMMIT_SHA", ""); sha != "" {
		return sha[:7]
	}
	return ""
}

func createLarkCard(projectVersion string) map[string]any {
	// Allow overriding build status via plugin settings
	status := getEnvOrDefault("PLUGIN_STATUS", getEnvOrDefault("DRONE_BUILD_STATUS", ""))

	var headerColor, statusIcon, statusText string
	if status == "failure" {
		headerColor = "red"
		statusIcon = "🚨"
		statusText = "Pipeline Failed"
	} else {
		headerColor = "green"
		statusIcon = "✅"
		statusText = "Pipeline Succeeded"
	}

	elements := []map[string]any{
		{
			"tag": "div",
			"text": map[string]any{
				"content": fmt.Sprintf("**Project:** %s\n**Branch:** %s\n**Author:** %s\n**Version:** %s",
					getEnvOrDefault("CI_REPO", ""),
					getEnvOrDefault("CI_COMMIT_BRANCH", ""),
					getEnvOrDefault("CI_COMMIT_AUTHOR", ""),
					projectVersion),
				"tag": "lark_md",
			},
		},
		{
			"tag": "hr",
		},
		{
			"tag": "div",
			"text": map[string]any{
				"content": fmt.Sprintf("**Commit Message:**\n%s",
					strings.Split(getEnvOrDefault("CI_COMMIT_MESSAGE", ""), "\n")[0]),
				"tag": "lark_md",
			},
		},
	}

	// Add variables if specified
	if variables := getEnvOrDefault("PLUGIN_VARIABLES", ""); variables != "" {
		elements = append(elements, map[string]any{
			"tag": "hr",
		})

		varContent := "**Variables:**\n"
		for _, varName := range strings.Split(variables, ",") {
			varName = strings.TrimSpace(varName)
			varContent += fmt.Sprintf("• `%s`: %s\n", varName, getEnvOrDefault(varName, ""))
		}

		elements = append(elements, map[string]any{
			"tag": "div",
			"text": map[string]any{
				"content": varContent,
				"tag": "lark_md",
			},
		})
	}

	// Add action buttons
	actions := createActionButtons()
	if len(actions) > 0 {
		elements = append(elements, map[string]any{
			"tag": "action",
			"actions": actions,
		})
	}

	projectName := getEnvOrDefault("CI_REPO_NAME", "")
	headerTitle := fmt.Sprintf("%s - %s %s", projectName, statusIcon, statusText)

	return map[string]any{
		"msg_type": "interactive",
		"card": map[string]any{
			"header": map[string]any{
				"title": map[string]any{
					"content": headerTitle,
					"tag": "plain_text",
				},
				"template": headerColor,
			},
			"elements": elements,
		},
	}
}

func createLarkTextMessage(projectVersion string) map[string]any {
	// Allow overriding build status via plugin settings
	status := getEnvOrDefault("PLUGIN_STATUS", getEnvOrDefault("DRONE_BUILD_STATUS", ""))

	var statusIcon, statusText string
	if status == "failure" {
		statusIcon = "🚨"
		statusText = "PIPELINE FAILED"
	} else {
		statusIcon = "✅"
		statusText = "PIPELINE SUCCEEDED"
	}

	message := fmt.Sprintf("%s %s\n\n", statusIcon, statusText)
	message += fmt.Sprintf("📋 Project: %s\n", getEnvOrDefault("CI_REPO", ""))
	message += fmt.Sprintf("🌿 Branch: %s\n", getEnvOrDefault("CI_COMMIT_BRANCH", ""))
	message += fmt.Sprintf("👤 Author: %s\n", getEnvOrDefault("CI_COMMIT_AUTHOR", ""))
	message += fmt.Sprintf("🏷️ Version: %s\n", projectVersion)
	message += fmt.Sprintf("💬 Message: %s\n", strings.Split(getEnvOrDefault("CI_COMMIT_MESSAGE", ""), "\n")[0])

	// Add variables if specified
	if variables := getEnvOrDefault("PLUGIN_VARIABLES", ""); variables != "" {
		message += "\n📊 Variables:\n"
		for _, varName := range strings.Split(variables, ",") {
			varName = strings.TrimSpace(varName)
			message += fmt.Sprintf("• %s: %s\n", varName, getEnvOrDefault(varName, ""))
		}
	}

	// Add links
	if pipelineURL := getEnvOrDefault("CI_PIPELINE_URL", ""); pipelineURL != "" {
		message += fmt.Sprintf("\n🔗 Pipeline: %s", pipelineURL)
	}

	return map[string]any{
		"msg_type": "text",
		"content": map[string]any{
			"text": message,
		},
	}
}

func createActionButtons() []map[string]any {
	var actions []map[string]any

	// Pipeline button
	if pipelineURL := getEnvOrDefault("CI_PIPELINE_URL", ""); pipelineURL != "" {
		actions = append(actions, map[string]any{
			"tag": "button",
			"text": map[string]any{
				"content": "View Pipeline",
				"tag": "plain_text",
			},
			"type": "primary",
			"url": pipelineURL,
		})
	}

	// Commit/Release button
	if tag := getEnvOrDefault("CI_COMMIT_TAG", ""); tag != "" {
		// Release button
		if repoURL := getEnvOrDefault("CI_REPO_URL", ""); repoURL != "" {
			releaseURL := fmt.Sprintf("%s/releases/tag/%s", repoURL, tag)
			actions = append(actions, map[string]any{
				"tag": "button",
				"text": map[string]any{
					"content": "View Release",
					"tag": "plain_text",
				},
				"type": "default",
				"url": releaseURL,
			})
		}
	} else {
		// Commit button
		if commitURL := getEnvOrDefault("CI_PIPELINE_FORGE_URL", ""); commitURL != "" {
			actions = append(actions, map[string]any{
				"tag": "button",
				"text": map[string]any{
					"content": "View Commit",
					"tag": "plain_text",
				},
				"type": "default",
				"url": commitURL,
			})
		}
	}

	// Filter buttons based on PLUGIN_BUTTONS if specified
	requestedButtons := getEnvOrDefault("PLUGIN_BUTTONS", "")
	if requestedButtons != "" {
		var filteredActions []map[string]any
		buttonNames := strings.Split(requestedButtons, ",")

		for _, name := range buttonNames {
			name = strings.TrimSpace(name)
			for _, action := range actions {
				if text, ok := action["text"].(map[string]any); ok {
					if content, ok := text["content"].(string); ok {
						if (name == "pipeline" && strings.Contains(content, "Pipeline")) ||
						   (name == "commit" && strings.Contains(content, "Commit")) ||
						   (name == "release" && strings.Contains(content, "Release")) {
							filteredActions = append(filteredActions, action)
							break
						}
					}
				}
			}
		}
		return filteredActions
	}

	return actions
}

func printBuildInfo(projectVersion string) {
	fmt.Println("\nBuild Info:")
	fmt.Printf(" PROJECT: %s\n", getEnvOrDefault("CI_REPO", ""))
	fmt.Printf(" BRANCH:  %s\n", getEnvOrDefault("CI_COMMIT_BRANCH", ""))
	fmt.Printf(" VERSION: %s\n", projectVersion)
	fmt.Printf(" STATUS:  %s\n", getEnvOrDefault("DRONE_BUILD_STATUS", ""))
	fmt.Printf(" DATE:    %s\n", time.Now().UTC().Format(time.RFC3339))
}

func sendMessage(webhookURL string, messageBytes []byte) {
	fmt.Println("\nSending to Lark...")

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(messageBytes))
	if err != nil {
		fmt.Printf("Error sending to Lark: %v\n", err)
		osExit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error response from Lark: %s\n", string(body))
		osExit(1)
	}

	// Parse response to check if successful
	var response map[string]any
	if err := json.Unmarshal(body, &response); err == nil {
		if code, ok := response["code"].(float64); ok && code != 0 {
			fmt.Printf("Lark API error: %v\n", response)
			osExit(1)
		}
	}

	fmt.Println("Done!")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func printDebugInfo(messageBytes []byte) {
	fmt.Println("\n** DEBUG ENABLED **")
	fmt.Println("\nEnvironment Variables:")

	envVars := os.Environ()
	sort.Strings(envVars)

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			fmt.Printf(" %-30s = %s\n", parts[0], parts[1])
		}
	}

	fmt.Println("\nLark Message JSON:")
	fmt.Println(string(messageBytes))
}
