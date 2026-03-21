package tests

import (
	"encoding/json"
	"testing"
)

// Benchmark tests for Claude Pipeline

// Benchmark JSON marshaling for tasks
func BenchmarkTaskMarshal(b *testing.B) {
	task := map[string]interface{}{
		"id":        "task-abc123",
		"skill_id":  "code-review",
		"status":    "completed",
		"duration":  5000,
		"parameters": map[string]interface{}{
			"target": "src/",
			"depth":  "deep",
		},
		"result": map[string]interface{}{
			"files_analyzed": 42,
			"issues_found":   15,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(task)
	}
}

// Benchmark JSON unmarshaling for tasks
func BenchmarkTaskUnmarshal(b *testing.B) {
	data := []byte(`{
		"id": "task-abc123",
		"skill_id": "code-review",
		"status": "completed",
		"duration": 5000,
		"parameters": {"target": "src/", "depth": "deep"},
		"result": {"files_analyzed": 42, "issues_found": 15}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var task map[string]interface{}
		_ = json.Unmarshal(data, &task)
	}
}

// Benchmark skill list creation
func BenchmarkCreateSkillList(b *testing.B) {
	skills := make([]map[string]interface{}, 10)
	for i := range skills {
		skills[i] = map[string]interface{}{
			"id":          string(rune('0' + i)),
			"name":        "Skill",
			"version":     "1.0.0",
			"category":    "quality",
			"enabled":     true,
			"parameters":  []interface{}{},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(map[string]interface{}{"skills": skills})
	}
}

// Benchmark pipeline validation
func BenchmarkPipelineValidation(b *testing.B) {
	pipeline := map[string]interface{}{
		"name": "benchmark-pipeline",
		"mode": "serial",
		"steps": []interface{}{
			map[string]interface{}{"id": "step1", "cli": "claude"},
			map[string]interface{}{"id": "step2", "cli": "npm"},
			map[string]interface{}{"id": "step3", "cli": "git"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Validate pipeline structure
		_ = pipeline["name"] != nil
		_ = pipeline["mode"] != nil
		_ = pipeline["steps"] != nil
	}
}

// Benchmark event creation
func BenchmarkCreateEvent(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = map[string]interface{}{
			"id":      "event-123",
			"type":    "task.completed",
			"source":  "api",
			"subject": "task-123",
			"data": map[string]interface{}{
				"status":   "completed",
				"duration": 5000,
			},
		}
	}
}

// Benchmark concurrent map access
func BenchmarkConcurrentMapAccess(b *testing.B) {
	m := make(map[string]string)
	for i := 0; i < 1000; i++ {
		m[string(rune(i))] = "value"
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = m[string(rune(i%1000))]
			i++
		}
	})
}

// Benchmark JSON stream encoding
func BenchmarkJSONStreamEncode(b *testing.B) {
	items := make([]map[string]interface{}, 100)
	for i := range items {
		items[i] = map[string]interface{}{
			"id":    i,
			"name":  "Item",
			"value": i * 100,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(items)
	}
}

// Benchmark response construction
func BenchmarkResponseConstruction(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = map[string]interface{}{
			"status": "ok",
			"data": map[string]interface{}{
				"items":      make([]interface{}, 0),
				"total":      0,
				"page":       1,
				"page_size":  20,
				"total_page": 0,
			},
		}
	}
}

// Benchmark parameter validation
func BenchmarkParameterValidation(b *testing.B) {
	params := map[string]interface{}{
		"target":  "src/",
		"depth":   "deep",
		"timeout": 600,
		"enabled": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = params["target"] != nil
		_ = params["depth"] != nil
		_ = params["timeout"] != nil
	}
}