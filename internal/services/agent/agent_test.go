package agent

import (
	"testing"

	"github.com/pAran0k/calc_go/models"
)

func TestProcessTask(t *testing.T) {
	agent := NewAgent()
	// Устанавливаем значения конфигурации для теста
	agent.Config.TimeAdditionMS = 100
	agent.Config.TimeSubtractionMS = 150
	agent.Config.TimeMultiplicationMS = 200
	agent.Config.TimeDivisionMS = 250

	tests := []struct {
		task     *models.Task
		expected float64
		wantErr  bool
	}{
		{&models.Task{ID: "task-1", Arg1: "2", Arg2: "3", Operation: "+"}, 5, false},
		{&models.Task{ID: "task-2", Arg1: "4", Arg2: "0", Operation: "/"}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.task.ID, func(t *testing.T) {
			result, err := agent.processTask(tt.task, "http://fake-url")
			if tt.wantErr {
				if err == nil {
					t.Errorf("processTask(%+v) expected error, got nil", tt.task)
				}
			} else {
				if err != nil {
					t.Errorf("processTask(%+v) unexpected error: %v", tt.task, err)
				}
				if result.Value != tt.expected {
					t.Errorf("processTask(%+v) = %f, want %f", tt.task, result.Value, tt.expected)
				}
			}
		})
	}
}
