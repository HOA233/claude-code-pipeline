package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// CLI client for Claude Pipeline API

var (
	baseURL = flag.String("url", "http://localhost:8080", "API base URL")
	apiKey  = flag.String("key", "", "API key")
	output  = flag.String("output", "text", "Output format: text, json")
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	client := &Client{
		BaseURL: *baseURL,
		APIKey:  *apiKey,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}

	switch cmd {
	case "skills":
		handleSkills(client, args)
	case "tasks":
		handleTasks(client, args)
	case "pipelines":
		handlePipelines(client, args)
	case "runs":
		handleRuns(client, args)
	case "status":
		handleStatus(client, args)
	case "watch":
		handleWatch(client, args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Claude Pipeline CLI

Usage:
  pipeline <command> [arguments]

Commands:
  skills list              List all available skills
  skills get <id>          Get skill details

  tasks list               List all tasks
  tasks create <skill>     Create a new task
  tasks get <id>           Get task details
  tasks result <id>        Get task result
  tasks cancel <id>        Cancel a task

  pipelines list           List all pipelines
  pipelines create <file>  Create pipeline from JSON file
  pipelines run <id>       Execute a pipeline

  runs list                List all runs
  runs get <id>            Get run details

  status                   Get service status
  watch <task-id>          Watch task output in real-time

Options:
  -url string     API base URL (default "http://localhost:8080")
  -key string     API key
  -output string  Output format: text, json (default "text")

Examples:
  pipeline skills list
  pipeline tasks create code-review -p target=src/ -p depth=deep
  pipeline tasks watch task-abc123
  pipeline pipelines create pipeline.json
  pipeline pipelines run pipeline-001`)
}

// Client is the API client
type Client struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	return c.Client.Do(req)
}

func (c *Client) Get(path string) (*http.Response, error) {
	return c.Do("GET", path, nil)
}

func (c *Client) Post(path string, body io.Reader) (*http.Response, error) {
	return c.Do("POST", path, body)
}

func (c *Client) Delete(path string) (*http.Response, error) {
	return c.Do("DELETE", path, nil)
}

func handleSkills(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: pipeline skills [list|get <id>]")
		os.Exit(1)
	}

	subCmd := args[0]
	switch subCmd {
	case "list":
		resp, err := client.Get("/api/skills")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if *output == "json" {
			json.NewEncoder(os.Stdout).Encode(result)
			return
		}

		fmt.Println("\nAvailable Skills:")
		fmt.Println("=================")
		for _, skill := range result["skills"].([]interface{}) {
			s := skill.(map[string]interface{})
			fmt.Printf("\n  %-20s %s\n", s["id"], s["name"])
			fmt.Printf("  %-20s %s\n", "", s["description"])
			fmt.Printf("  %-20s v%s | %s\n", "", s["version"], s["category"])
		}
		fmt.Println()

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: pipeline skills get <id>")
			os.Exit(1)
		}

		resp, err := client.Get("/api/skills/" + args[1])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var skill map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&skill)

		if *output == "json" {
			json.NewEncoder(os.Stdout).Encode(skill)
			return
		}

		fmt.Printf("\nSkill: %s\n", skill["name"])
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("ID:          %s\n", skill["id"])
		fmt.Printf("Version:     %s\n", skill["version"])
		fmt.Printf("Category:    %s\n", skill["category"])
		fmt.Printf("Description: %s\n", skill["description"])

		if params, ok := skill["parameters"].([]interface{}); ok && len(params) > 0 {
			fmt.Println("\nParameters:")
			for _, p := range params {
				param := p.(map[string]interface{})
				req := ""
				if param["required"].(bool) {
					req = " (required)"
				}
				fmt.Printf("  - %-15s %s%s\n", param["name"], param["type"], req)
				if param["description"] != nil {
					fmt.Printf("    %s\n", param["description"])
				}
			}
		}
		fmt.Println()
	}
}

func handleTasks(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: pipeline tasks [list|create|get|result|cancel]")
		os.Exit(1)
	}

	subCmd := args[0]
	switch subCmd {
	case "list":
		resp, err := client.Get("/api/tasks")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if *output == "json" {
			json.NewEncoder(os.Stdout).Encode(result)
			return
		}

		fmt.Println("\nTasks:")
		fmt.Println("======")
		for _, task := range result["tasks"].([]interface{}) {
			t := task.(map[string]interface{})
			status := getStatusIcon(t["status"].(string))
			fmt.Printf("  %s %-15s %-20s %s\n", status, t["id"], t["skill_id"], t["status"])
		}
		fmt.Println()

	case "create":
		if len(args) < 2 {
			fmt.Println("Usage: pipeline tasks create <skill-id> [-p key=value]")
			os.Exit(1)
		}

		skillID := args[1]
		params := make(map[string]interface{})

		// Parse -p flags
		for i := 2; i < len(args); i++ {
			if args[i] == "-p" && i+1 < len(args) {
				kv := strings.SplitN(args[i+1], "=", 2)
				if len(kv) == 2 {
					params[kv[0]] = kv[1]
				}
				i++
			}
		}

		reqBody := map[string]interface{}{
			"skill_id":   skillID,
			"parameters": params,
		}

		body, _ := json.Marshal(reqBody)
		resp, err := client.Post("/api/tasks", strings.NewReader(string(body)))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var task map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&task)

		fmt.Printf("\nTask created: %s\n", task["id"])
		fmt.Printf("Status: %s\n", task["status"])
		fmt.Printf("\nUse 'pipeline tasks get %s' to check status\n", task["id"])

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: pipeline tasks get <id>")
			os.Exit(1)
		}

		resp, err := client.Get("/api/tasks/" + args[1])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var task map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&task)

		if *output == "json" {
			json.NewEncoder(os.Stdout).Encode(task)
			return
		}

		printTaskDetails(task)

	case "result":
		if len(args) < 2 {
			fmt.Println("Usage: pipeline tasks result <id>")
			os.Exit(1)
		}

		resp, err := client.Get("/api/tasks/" + args[1] + "/result")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if *output == "json" {
			json.NewEncoder(os.Stdout).Encode(result)
			return
		}

		fmt.Println("\nTask Result:")
		fmt.Println("============")
		json.NewEncoder(os.Stdout).Encode(result["result"])

	case "cancel":
		if len(args) < 2 {
			fmt.Println("Usage: pipeline tasks cancel <id>")
			os.Exit(1)
		}

		resp, err := client.Delete("/api/tasks/" + args[1])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		fmt.Printf("Task %s cancelled\n", args[1])
	}
}

func handlePipelines(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: pipeline pipelines [list|create|run]")
		os.Exit(1)
	}

	subCmd := args[0]
	switch subCmd {
	case "list":
		resp, err := client.Get("/api/pipelines")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		fmt.Println("\nPipelines:")
		fmt.Println("==========")
		for _, p := range result["pipelines"].([]interface{}) {
			pipeline := p.(map[string]interface{})
			fmt.Printf("  %-20s %-20s %s\n", pipeline["id"], pipeline["name"], pipeline["mode"])
		}
		fmt.Println()

	case "create":
		if len(args) < 2 {
			fmt.Println("Usage: pipeline pipelines create <file.json>")
			os.Exit(1)
		}

		file, err := os.Open(args[1])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		resp, err := client.Post("/api/pipelines", file)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var pipeline map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&pipeline)

		fmt.Printf("Pipeline created: %s\n", pipeline["id"])

	case "run":
		if len(args) < 2 {
			fmt.Println("Usage: pipeline pipelines run <id>")
			os.Exit(1)
		}

		resp, err := client.Post("/api/pipelines/"+args[1]+"/run", nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var run map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&run)

		fmt.Printf("Pipeline execution started: %s\n", run["id"])
	}
}

func handleRuns(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: pipeline runs [list|get]")
		os.Exit(1)
	}

	subCmd := args[0]
	switch subCmd {
	case "list":
		resp, err := client.Get("/api/runs")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		fmt.Println("\nRuns:")
		fmt.Println("=====")
		for _, r := range result["runs"].([]interface{}) {
			run := r.(map[string]interface{})
			status := getStatusIcon(run["status"].(string))
			fmt.Printf("  %s %-15s %-20s %s\n", status, run["id"], run["pipeline_id"], run["status"])
		}
		fmt.Println()

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: pipeline runs get <id>")
			os.Exit(1)
		}

		resp, err := client.Get("/api/runs/" + args[1])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var run map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&run)

		if *output == "json" {
			json.NewEncoder(os.Stdout).Encode(run)
			return
		}

		fmt.Printf("\nRun: %s\n", run["id"])
		fmt.Printf("Pipeline: %s\n", run["pipeline_id"])
		fmt.Printf("Status: %s\n", run["status"])

		if steps, ok := run["step_results"].([]interface{}); ok {
			fmt.Println("\nSteps:")
			for _, s := range steps {
				step := s.(map[string]interface{})
				status := getStatusIcon(step["status"].(string))
				fmt.Printf("  %s %s: %s\n", status, step["step_id"], step["status"])
			}
		}
	}
}

func handleStatus(client *Client, args []string) {
	resp, err := client.Get("/api/status")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var status map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&status)

	fmt.Println("\nService Status:")
	fmt.Println("===============")
	fmt.Printf("Status: %s\n", status["status"])

	if cli, ok := status["cli"].(map[string]interface{}); ok {
		fmt.Printf("Active Tasks: %v\n", cli["active_count"])
		fmt.Printf("Max Concurrency: %v\n", cli["max_concurrency"])
	}
	fmt.Println()
}

func handleWatch(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: pipeline watch <task-id>")
		os.Exit(1)
	}

	taskID := args[0]
	fmt.Printf("Watching task %s...\n\n", taskID)

	// Poll for updates
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastStatus string

	for range ticker.C {
		resp, err := client.Get("/api/tasks/" + taskID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		var task map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&task)
		resp.Body.Close()

		status := task["status"].(string)
		if status != lastStatus {
			fmt.Printf("[%s] Status: %s\n", time.Now().Format("15:04:05"), status)
			lastStatus = status
		}

		if status == "completed" || status == "failed" || status == "cancelled" {
			fmt.Println()
			printTaskDetails(task)

			if status == "completed" {
				// Get result
				resp, _ := client.Get("/api/tasks/" + taskID + "/result")
				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				resp.Body.Close()

				fmt.Println("\nResult:")
				json.NewEncoder(os.Stdout).Encode(result["result"])
			}
			break
		}
	}
}

func printTaskDetails(task map[string]interface{}) {
	fmt.Printf("\nTask: %s\n", task["id"])
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Skill:    %s\n", task["skill_id"])
	fmt.Printf("Status:   %s\n", task["status"])

	if task["duration"] != nil {
		fmt.Printf("Duration: %v ms\n", task["duration"])
	}

	if task["created_at"] != nil {
		fmt.Printf("Created:  %v\n", task["created_at"])
	}

	if task["error"] != nil && task["error"] != "" {
		fmt.Printf("Error:    %s\n", task["error"])
	}
}

func getStatusIcon(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "running":
		return "▶️"
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "cancelled":
		return "⏹️"
	default:
		return "❓"
	}
}

// Interactive mode
func interactiveMode(client *Client) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Claude Pipeline Interactive Mode")
	fmt.Println("Type 'help' for commands, 'exit' to quit")
	fmt.Println()

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			break
		}

		// Parse and execute command
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "skills", "tasks", "pipelines", "runs", "status":
			handleSkills(client, parts)
		case "help":
			printUsage()
		case "clear":
			fmt.Print("\033[H\033[2J")
		default:
			fmt.Printf("Unknown command: %s\n", parts[0])
		}
	}
}