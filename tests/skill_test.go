package tests

import (
	"testing"

	"github.com/company/claude-pipeline/internal/service"
)

// Skill Service Tests

func TestSkillService_New(t *testing.T) {
	ss := service.NewSkillService()
	if ss == nil {
		t.Fatal("Expected non-nil skill service")
	}
}

func TestSkillService_Create(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID:          "skill-1",
		Name:        "Code Review",
		Description: "Review code quality",
		Version:     "1.0.0",
		Category:    "quality",
	}

	err := ss.Create(skill)
	if err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}
}

func TestSkillService_Create_MissingName(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID: "no-name",
	}

	err := ss.Create(skill)
	if err == nil {
		t.Error("Expected error for missing name")
	}
}

func TestSkillService_Get(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID:       "get-skill",
		Name:     "Get Skill",
		Category: "testing",
	}
	ss.Create(skill)

	retrieved, err := ss.Get("get-skill")
	if err != nil {
		t.Fatalf("Failed to get skill: %v", err)
	}

	if retrieved.Name != "Get Skill" {
		t.Error("Skill name mismatch")
	}
}

func TestSkillService_Get_NotFound(t *testing.T) {
	ss := service.NewSkillService()

	_, err := ss.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent skill")
	}
}

func TestSkillService_List(t *testing.T) {
	ss := service.NewSkillService()

	ss.Create(&service.Skill{
		ID:       "list-1",
		Name:     "List 1",
		Category: "quality",
	})

	ss.Create(&service.Skill{
		ID:       "list-2",
		Name:     "List 2",
		Category: "devops",
	})

	skills := ss.List("")
	if len(skills) < 2 {
		t.Errorf("Expected at least 2 skills, got %d", len(skills))
	}
}

func TestSkillService_ListByCategory(t *testing.T) {
	ss := service.NewSkillService()

	ss.Create(&service.Skill{
		ID:       "cat-1",
		Name:     "Category Test 1",
		Category: "quality",
	})

	ss.Create(&service.Skill{
		ID:       "cat-2",
		Name:     "Category Test 2",
		Category: "devops",
	})

	skills := ss.List("quality")
	for _, s := range skills {
		if s.Category != "quality" {
			t.Error("Expected only quality skills")
		}
	}
}

func TestSkillService_Update(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID:          "update-skill",
		Name:        "Update Skill",
		Description: "Original",
	}
	ss.Create(skill)

	err := ss.Update("update-skill", map[string]interface{}{
		"description": "Updated description",
		"version":     "2.0.0",
	})

	if err != nil {
		t.Fatalf("Failed to update skill: %v", err)
	}

	retrieved, _ := ss.Get("update-skill")
	if retrieved.Description != "Updated description" {
		t.Error("Description not updated")
	}
}

func TestSkillService_Delete(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID:   "delete-skill",
		Name: "Delete Skill",
	}
	ss.Create(skill)

	err := ss.Delete("delete-skill")
	if err != nil {
		t.Fatalf("Failed to delete skill: %v", err)
	}

	_, err = ss.Get("delete-skill")
	if err == nil {
		t.Error("Expected error for deleted skill")
	}
}

func TestSkillService_Enable(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID:      "enable-skill",
		Name:    "Enable Skill",
		Enabled: false,
	}
	ss.Create(skill)

	err := ss.Enable("enable-skill")
	if err != nil {
		t.Fatalf("Failed to enable skill: %v", err)
	}

	retrieved, _ := ss.Get("enable-skill")
	if !retrieved.Enabled {
		t.Error("Skill should be enabled")
	}
}

func TestSkillService_Disable(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID:      "disable-skill",
		Name:    "Disable Skill",
		Enabled: true,
	}
	ss.Create(skill)

	err := ss.Disable("disable-skill")
	if err != nil {
		t.Fatalf("Failed to disable skill: %v", err)
	}

	retrieved, _ := ss.Get("disable-skill")
	if retrieved.Enabled {
		t.Error("Skill should be disabled")
	}
}

func TestSkillService_SetParameter(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID:   "param-skill",
		Name: "Param Skill",
	}
	ss.Create(skill)

	param := service.SkillParameter{
		Name:        "target",
		Type:        "string",
		Required:    true,
		Description: "Target path",
	}

	err := ss.SetParameter("param-skill", param)
	if err != nil {
		t.Fatalf("Failed to set parameter: %v", err)
	}
}

func TestSkillService_GetParameters(t *testing.T) {
	ss := service.NewSkillService()

	skill := &service.Skill{
		ID:   "get-param-skill",
		Name: "Get Param Skill",
	}
	ss.Create(skill)
	ss.SetParameter("get-param-skill", service.SkillParameter{
		Name: "param1",
		Type: "string",
	})
	ss.SetParameter("get-param-skill", service.SkillParameter{
		Name: "param2",
		Type: "number",
	})

	params := ss.GetParameters("get-param-skill")
	if len(params) < 2 {
		t.Errorf("Expected at least 2 parameters, got %d", len(params))
	}
}

func TestSkillService_GetStats(t *testing.T) {
	ss := service.NewSkillService()

	ss.Create(&service.Skill{ID: "stats-1", Name: "Stats 1", Category: "quality"})
	ss.Create(&service.Skill{ID: "stats-2", Name: "Stats 2", Category: "devops"})

	stats := ss.GetStats()

	if stats.TotalSkills < 2 {
		t.Errorf("Expected at least 2 skills, got %d", stats.TotalSkills)
	}
}

func TestSkillService_SkillToJSON(t *testing.T) {
	skill := &service.Skill{
		ID:          "json-skill",
		Name:        "JSON Skill",
		Description: "Test",
		Version:     "1.0.0",
		Category:    "testing",
		Enabled:     true,
	}

	data, err := skill.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}
}