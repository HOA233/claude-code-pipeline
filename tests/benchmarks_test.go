package tests

import (
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

// BenchmarkSkillService tests skill service performance
func BenchmarkSkillService_Create(b *testing.B) {
	ss := service.NewSkillService()
	skill := &service.Skill{
		ID:          "bench-skill",
		Name:        "Benchmark Skill",
		Description: "For benchmarking",
		Version:     "1.0.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		skill.ID = "bench-skill-" + string(rune(i))
		ss.Create(skill)
	}
}

func BenchmarkSkillService_Get(b *testing.B) {
	ss := service.NewSkillService()
	skill := &service.Skill{
		ID:   "bench-get-skill",
		Name: "Benchmark Get",
	}
	ss.Create(skill)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss.Get("bench-get-skill")
	}
}

func BenchmarkSkillService_List(b *testing.B) {
	ss := service.NewSkillService()
	for i := 0; i < 100; i++ {
		ss.Create(&service.Skill{
			ID:   "bench-list-skill-" + string(rune(i)),
			Name: "Benchmark List",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss.List()
	}
}

// BenchmarkTaskService tests task service performance
func BenchmarkTaskService_Create(b *testing.B) {
	ts := service.NewTaskService()
	task := &service.Task{
		ID:      "bench-task",
		SkillID: "skill-1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task.ID = "bench-task-" + string(rune(i))
		ts.Create(task)
	}
}

func BenchmarkTaskService_Get(b *testing.B) {
	ts := service.NewTaskService()
	task := &service.Task{
		ID:      "bench-get-task",
		SkillID: "skill-1",
	}
	ts.Create(task)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.Get("bench-get-task")
	}
}

func BenchmarkTaskService_UpdateStatus(b *testing.B) {
	ts := service.NewTaskService()
	task := &service.Task{
		ID:      "bench-update-task",
		SkillID: "skill-1",
	}
	ts.Create(task)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.UpdateStatus("bench-update-task", service.TaskStatusRunning)
	}
}

// BenchmarkCacheService tests cache service performance
func BenchmarkCacheService_Set(b *testing.B) {
	cs := service.NewCacheService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.Set("bench-key", "bench-value", 0)
	}
}

func BenchmarkCacheService_Get(b *testing.B) {
	cs := service.NewCacheService()
	cs.Set("bench-get-key", "value", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.Get("bench-get-key")
	}
}

// BenchmarkPriorityQueue tests priority queue performance
func BenchmarkPriorityQueue_Enqueue(b *testing.B) {
	pq := service.NewPriorityQueue()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Enqueue("item", i%10)
	}
}

func BenchmarkPriorityQueue_Dequeue(b *testing.B) {
	pq := service.NewPriorityQueue()
	for i := 0; i < 10000; i++ {
		pq.Enqueue("item-"+string(rune(i)), i%10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if pq.Size() > 0 {
			pq.Dequeue()
		}
	}
}

// BenchmarkRateLimiter tests rate limiter performance
func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := service.NewRateLimiter(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("bench-key")
	}
}

// BenchmarkEventBus tests event bus performance
func BenchmarkEventBus_Publish(b *testing.B) {
	eb := service.NewEventBus()
	eb.Subscribe("bench-event", func(e service.Event) {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.Publish("bench-event", "data")
	}
}

// BenchmarkWebhookService tests webhook service performance
func BenchmarkWebhookService_Create(b *testing.B) {
	ws := service.NewWebhookService()
	webhook := &service.Webhook{
		ID:      "bench-webhook",
		URL:     "https://example.com/webhook",
		Events:  []string{"task.completed"},
		Enabled: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		webhook.ID = "bench-webhook-" + string(rune(i))
		ws.Create(webhook)
	}
}

// BenchmarkTenantService tests tenant service performance
func BenchmarkTenantService_Create(b *testing.B) {
	ts := service.NewTenantService()
	tenant := &service.Tenant{
		ID:   "bench-tenant",
		Name: "Benchmark Tenant",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tenant.ID = "bench-tenant-" + string(rune(i))
		ts.Create(tenant)
	}
}

// BenchmarkSecretService tests secret service performance
func BenchmarkSecretService_Create(b *testing.B) {
	ss := service.NewSecretService()
	secret := &service.Secret{
		ID:    "bench-secret",
		Name:  "Benchmark Secret",
		Value: "secret-value",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		secret.ID = "bench-secret-" + string(rune(i))
		ss.Create(secret)
	}
}

func BenchmarkSecretService_GetValue(b *testing.B) {
	ss := service.NewSecretService()
	secret := &service.Secret{
		ID:    "bench-get-secret",
		Name:  "Get Secret",
		Value: "secret-value",
	}
	ss.Create(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss.GetValue("bench-get-secret")
	}
}