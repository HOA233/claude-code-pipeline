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
)

// CLI tool for Claude Pipeline API management

var (
	baseURL = "http://localhost:8080"
	apiKey  = ""
)

func main() {
	flag.StringVar(&baseURL, "url", "http://localhost:8080", "API base URL")
	flag.StringVar(&apiKey, "key", "", "API key")
	flag.Parse()

	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	command := flag.Arg(0)

	switch command {
	case "skills":
		handleSkills()
	case "tasks":
		handleTasks()
	case "pipelines":
		handlePipelines()
	case "runs":
		handleRuns()
	case "status":
		handleStatus()
	case "templates":
		handleTemplates()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Claude Pipeline CLI")
	fmt.Println()
	fmt.Println("Usage: pipeline-cli [flags] <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  skills [list|get <id>|sync]    Manage skills")
	fmt.Println("  tasks [list|create|get <id>]    Manage tasks")
	fmt.Println("  pipelines [list|create|get]     Manage pipelines")
	fmt.Println("  runs [list|get <id>]            View runs")
	fmt.Println("  templates [list|use <id>]       Manage templates")
	fmt.Println("  status                          Show service status")
	fmt.Println("  help                            Show this help")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -url string   API base URL (default: http://localhost:8080)")
	fmt.Println("  -key string   API key for authentication")
}

func handleSkills() {
	subCmd := "list"
	if flag.NArg() > 1 {
		subCmd = flag.Arg(1)
	}

	switch subCmd {
	case "list":
		resp := get("/api/skills")
		printJSON(resp)
	case "get":
		if flag.NArg() < 3 {
			fmt.Println("Usage: skills get <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := get("/api/skills/" + id)
		printJSON(resp)
	case "sync":
		resp := post("/api/skills/sync", nil)
		printJSON(resp)
	}
}

func handleTasks() {
	subCmd := "list"
	if flag.NArg() > 1 {
		subCmd = flag.Arg(1)
	}

	switch subCmd {
	case "list":
		resp := get("/api/tasks")
		printJSON(resp)
	case "get":
		if flag.NArg() < 3 {
			fmt.Println("Usage: tasks get <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := get("/api/tasks/" + id)
		printJSON(resp)
	case "create":
		createTask()
	case "result":
		if flag.NArg() < 3 {
			fmt.Println("Usage: tasks result <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := get("/api/tasks/" + id + "/result")
		printJSON(resp)
	case "cancel":
		if flag.NArg() < 3 {
			fmt.Println("Usage: tasks cancel <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := delete("/api/tasks/" + id)
		printJSON(resp)
	}
}

func createTask() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Skill ID: ")
	skillID, _ := reader.ReadString('\n')
	skillID = strings.TrimSpace(skillID)

	fmt.Print("Target: ")
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(target)

	fmt.Print("Depth (quick/standard/deep) [standard]: ")
	depth, _ := reader.ReadString('\n')
	depth = strings.TrimSpace(depth)
	if depth == "" {
		depth = "standard"
	}

	data := map[string]interface{}{
		"skill_id": skillID,
		"parameters": map[string]interface{}{
			"target": target,
			"depth":  depth,
		},
	}

	resp := post("/api/tasks", data)
	printJSON(resp)
}

func handlePipelines() {
	subCmd := "list"
	if flag.NArg() > 1 {
		subCmd = flag.Arg(1)
	}

	switch subCmd {
	case "list":
		resp := get("/api/pipelines")
		printJSON(resp)
	case "get":
		if flag.NArg() < 3 {
			fmt.Println("Usage: pipelines get <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := get("/api/pipelines/" + id)
		printJSON(resp)
	case "create":
		createPipeline()
	case "run":
		if flag.NArg() < 3 {
			fmt.Println("Usage: pipelines run <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := post("/api/pipelines/"+id+"/run", map[string]interface{}{})
		printJSON(resp)
	case "delete":
		if flag.NArg() < 3 {
			fmt.Println("Usage: pipelines delete <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := delete("/api/pipelines/" + id)
		printJSON(resp)
	}
}

func createPipeline() {
	// Read pipeline definition from stdin
	fmt.Println("Enter pipeline JSON (or Ctrl+D to finish):")
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	var pipeline map[string]interface{}
	if err := json.Unmarshal(data, &pipeline); err != nil {
		fmt.Printf("Invalid JSON: %v\n", err)
		os.Exit(1)
	}

	resp := post("/api/pipelines", pipeline)
	printJSON(resp)
}

func handleRuns() {
	subCmd := "list"
	if flag.NArg() > 1 {
		subCmd = flag.Arg(1)
	}

	switch subCmd {
	case "list":
		resp := get("/api/runs")
		printJSON(resp)
	case "get":
		if flag.NArg() < 3 {
			fmt.Println("Usage: runs get <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := get("/api/runs/" + id)
		printJSON(resp)
	case "cancel":
		if flag.NArg() < 3 {
			fmt.Println("Usage: runs cancel <id>")
			os.Exit(1)
		}
		id := flag.Arg(2)
		resp := delete("/api/runs/" + id)
		printJSON(resp)
	}
}

func handleTemplates() {
	subCmd := "list"
	if flag.NArg() > 1 {
		subCmd = flag.Arg(1)
	}

	switch subCmd {
	case "list":
		resp := get("/api/templates")
		printJSON(resp)
	}
}

func handleStatus() {
	resp := get("/api/status")
	printJSON(resp)
}

func get(path string) []byte {
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	return doRequest(req)
}

func post(path string, data interface{}) []byte {
	var body io.Reader
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = strings.NewReader(string(jsonData))
	}

	req, _ := http.NewRequest("POST", baseURL+path, body)
	req.Header.Set("Content-Type", "application/json")
	return doRequest(req)
}

func delete(path string) []byte {
	req, _ := http.NewRequest("DELETE", baseURL+path, nil)
	return doRequest(req)
}

func doRequest(req *http.Request) []byte {
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		fmt.Printf("Error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	return body
}

func printJSON(data []byte) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println(string(data))
		return
	}

	pretty, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(pretty))
}