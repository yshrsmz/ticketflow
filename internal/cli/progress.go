package cli

import (
	"fmt"
	"strings"

	"github.com/yshrsmz/ticketflow/internal/ticket"
)

// UpdateProgress updates the progress of a ticket
func (app *App) UpdateProgress(ticketID string, progress int) error {
	t, err := app.Manager.Get(ticketID)
	if err != nil {
		return err
	}

	if err := t.UpdateProgress(progress); err != nil {
		return err
	}

	return app.Manager.Update(t)
}

// AddTask adds a task to a ticket
func (app *App) AddTask(ticketID, description string) error {
	t, err := app.Manager.Get(ticketID)
	if err != nil {
		return err
	}

	t.AddTask(description)
	return app.Manager.Update(t)
}

// CompleteTask marks a task as completed
func (app *App) CompleteTask(ticketID string, taskIndex int) error {
	t, err := app.Manager.Get(ticketID)
	if err != nil {
		return err
	}

	if err := t.CompleteTask(taskIndex); err != nil {
		return err
	}

	return app.Manager.Update(t)
}

// ShowProgress shows the progress of a ticket
func (app *App) ShowProgress(ticketID string, format OutputFormat) error {
	t, err := app.Manager.Get(ticketID)
	if err != nil {
		return err
	}

	if format == FormatJSON {
		return outputJSON(map[string]interface{}{
			"ticket_id":   t.ID,
			"progress":    t.Progress,
			"tasks_total": len(t.Tasks),
			"tasks_completed": t.GetCompletedTasksCount(),
			"tasks": t.Tasks,
		})
	}

	// Text format
	fmt.Printf("Ticket: %s\n", t.ID)
	fmt.Printf("Progress: %d%%\n", t.Progress)
	
	if len(t.Tasks) > 0 {
		fmt.Printf("\nTasks (%d/%d completed):\n", t.GetCompletedTasksCount(), len(t.Tasks))
		for i, task := range t.Tasks {
			status := "[ ]"
			if task.Completed {
				status = "[✓]"
			}
			fmt.Printf("%d. %s %s\n", i+1, status, task.Description)
		}
		
		// Show progress bar
		fmt.Printf("\nProgress: ")
		progressBar := generateProgressBar(t.CalculateProgress(), 30)
		fmt.Println(progressBar)
	}

	return nil
}

// ReportProgress generates a progress report for all active tickets
func (app *App) ReportProgress(format OutputFormat) error {
	activeTickets, err := app.Manager.List(string(ticket.StatusDoing))
	if err != nil {
		return err
	}

	if format == FormatJSON {
		type progressReport struct {
			TicketID   string `json:"ticket_id"`
			Progress   int    `json:"progress"`
			TasksTotal int    `json:"tasks_total"`
			TasksDone  int    `json:"tasks_done"`
		}

		reports := make([]progressReport, len(activeTickets))
		for i, t := range activeTickets {
			reports[i] = progressReport{
				TicketID:   t.ID,
				Progress:   t.Progress,
				TasksTotal: len(t.Tasks),
				TasksDone:  t.GetCompletedTasksCount(),
			}
		}

		return outputJSON(map[string]interface{}{
			"active_tickets": len(activeTickets),
			"reports":        reports,
		})
	}

	// Text format
	if len(activeTickets) == 0 {
		fmt.Println("No active tickets.")
		return nil
	}

	fmt.Printf("Progress Report - %d active ticket(s)\n", len(activeTickets))
	fmt.Println(strings.Repeat("=", 60))

	for _, t := range activeTickets {
		fmt.Printf("\n%s: %s\n", t.ID, t.Description)
		
		if len(t.Tasks) > 0 {
			fmt.Printf("Tasks: %d/%d completed\n", t.GetCompletedTasksCount(), len(t.Tasks))
		}
		
		progressBar := generateProgressBar(t.Progress, 40)
		fmt.Printf("Progress: %s %d%%\n", progressBar, t.Progress)
	}

	return nil
}

// generateProgressBar creates a visual progress bar
func generateProgressBar(progress, width int) string {
	filled := (progress * width) / 100
	empty := width - filled
	
	bar := "["
	bar += strings.Repeat("=", filled)
	if filled < width {
		bar += ">"
		bar += strings.Repeat(" ", empty-1)
	}
	bar += "]"
	
	return bar
}

// ParseTasksFromArgs parses task descriptions from command line arguments
func ParseTasksFromArgs(args []string) []string {
	var tasks []string
	for _, arg := range args {
		// Handle comma-separated tasks
		parts := strings.Split(arg, ",")
		for _, part := range parts {
			task := strings.TrimSpace(part)
			if task != "" {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}