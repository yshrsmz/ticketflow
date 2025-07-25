package ticket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketProgress(t *testing.T) {
	ticket := New("test-progress", "Test progress tracking")

	// Test initial progress
	assert.Equal(t, 0, ticket.Progress)
	assert.Equal(t, 0, ticket.CalculateProgress())

	// Test updating progress
	err := ticket.UpdateProgress(50)
	require.NoError(t, err)
	assert.Equal(t, 50, ticket.Progress)

	// Test invalid progress values
	err = ticket.UpdateProgress(-1)
	assert.Error(t, err)
	err = ticket.UpdateProgress(101)
	assert.Error(t, err)
}

func TestTicketTasks(t *testing.T) {
	ticket := New("test-tasks", "Test task management")

	// Add tasks
	ticket.AddTask("First task")
	ticket.AddTask("Second task")
	ticket.AddTask("Third task")

	assert.Len(t, ticket.Tasks, 3)
	assert.Equal(t, "First task", ticket.Tasks[0].Description)
	assert.False(t, ticket.Tasks[0].Completed)

	// Complete a task
	err := ticket.CompleteTask(0)
	require.NoError(t, err)
	assert.True(t, ticket.Tasks[0].Completed)
	assert.NotNil(t, ticket.Tasks[0].CompletedAt)
	assert.Equal(t, 1, ticket.GetCompletedTasksCount())

	// Test task index out of range
	err = ticket.CompleteTask(-1)
	assert.Error(t, err)
	err = ticket.CompleteTask(10)
	assert.Error(t, err)

	// Complete another task
	err = ticket.CompleteTask(1)
	require.NoError(t, err)
	assert.Equal(t, 2, ticket.GetCompletedTasksCount())

	// Test progress calculation based on tasks
	progress := ticket.CalculateProgress()
	assert.Equal(t, 66, progress) // 2 out of 3 tasks = 66%
	assert.Equal(t, 66, ticket.Progress) // Should be auto-updated

	// Complete all tasks
	err = ticket.CompleteTask(2)
	require.NoError(t, err)
	assert.Equal(t, 100, ticket.CalculateProgress())
	assert.Equal(t, 100, ticket.Progress)
}

func TestTicketProgressWithoutTasks(t *testing.T) {
	ticket := New("test-manual", "Test manual progress")

	// Set manual progress without tasks
	err := ticket.UpdateProgress(75)
	require.NoError(t, err)
	assert.Equal(t, 75, ticket.Progress)
	assert.Equal(t, 75, ticket.CalculateProgress()) // Should return manual progress when no tasks
}

func TestTicketProgressPersistence(t *testing.T) {
	// Create ticket with progress
	ticket := New("test-persist", "Test progress persistence")
	ticket.UpdateProgress(50)
	ticket.AddTask("Task 1")
	ticket.AddTask("Task 2")
	ticket.CompleteTask(0)

	// Serialize
	data, err := ticket.ToBytes()
	require.NoError(t, err)

	// Parse back
	parsed, err := Parse(data)
	require.NoError(t, err)

	// Check progress data is preserved
	assert.Equal(t, 50, parsed.Progress) // Manual progress
	assert.Len(t, parsed.Tasks, 2)
	assert.True(t, parsed.Tasks[0].Completed)
	assert.NotNil(t, parsed.Tasks[0].CompletedAt)
	assert.False(t, parsed.Tasks[1].Completed)
}