package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCloseDialogModel(t *testing.T) {
	dialog := NewCloseDialogModel()

	assert.Equal(t, CloseDialogHidden, dialog.state)
	assert.False(t, dialog.requireReason)
	assert.False(t, dialog.showError)
	assert.Empty(t, dialog.errorMsg)
	assert.Equal(t, reasonCharLimit, dialog.reasonInput.CharLimit)
	assert.Equal(t, reasonInputWidth, dialog.reasonInput.Width)
	assert.NotEmpty(t, dialog.reasonInput.Placeholder)
}

func TestCloseDialogModel_ShowHide(t *testing.T) {
	dialog := NewCloseDialogModel()

	// Test Show with required reason
	dialog.Show(true)
	assert.Equal(t, CloseDialogInput, dialog.state)
	assert.True(t, dialog.requireReason)
	assert.False(t, dialog.showError)
	assert.True(t, dialog.reasonInput.Focused())

	// Test Hide
	dialog.Hide()
	assert.Equal(t, CloseDialogHidden, dialog.state)
	assert.False(t, dialog.reasonInput.Focused())
	assert.False(t, dialog.showError)
}

func TestCloseDialogModel_StateChecks(t *testing.T) {
	dialog := NewCloseDialogModel()

	// Initial state
	assert.False(t, dialog.IsVisible())
	assert.False(t, dialog.IsConfirmed())
	assert.False(t, dialog.IsCancelled())

	// When visible
	dialog.state = CloseDialogInput
	assert.True(t, dialog.IsVisible())
	assert.False(t, dialog.IsConfirmed())
	assert.False(t, dialog.IsCancelled())

	// When confirmed
	dialog.state = CloseDialogConfirmed
	assert.False(t, dialog.IsVisible())
	assert.True(t, dialog.IsConfirmed())
	assert.False(t, dialog.IsCancelled())

	// When cancelled
	dialog.state = CloseDialogCancelled
	assert.False(t, dialog.IsVisible())
	assert.False(t, dialog.IsConfirmed())
	assert.True(t, dialog.IsCancelled())
}

func TestCloseDialogModel_GetReason(t *testing.T) {
	dialog := NewCloseDialogModel()

	// Test empty reason
	assert.Empty(t, dialog.GetReason())

	// Test with whitespace
	dialog.reasonInput.SetValue("  \n\t  ")
	assert.Empty(t, dialog.GetReason())

	// Test with actual content
	dialog.reasonInput.SetValue("  Test reason  ")
	assert.Equal(t, "Test reason", dialog.GetReason())
}

func TestCloseDialogModel_SetSize(t *testing.T) {
	dialog := NewCloseDialogModel()

	dialog.SetSize(100, 50)
	assert.Equal(t, 100, dialog.width)
	assert.Equal(t, 50, dialog.height)
}

func TestCloseDialogModel_Update_EscapeKey(t *testing.T) {
	dialog := NewCloseDialogModel()
	dialog.Show(false)

	// Test escape key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	result, cmd := dialog.Update(msg)

	assert.Equal(t, CloseDialogHidden, result.state)
	assert.False(t, result.reasonInput.Focused())
	assert.Nil(t, cmd)
}

func TestCloseDialogModel_Update_EnterKey_OptionalReason(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectSuccess bool
	}{
		{
			name:          "empty reason when optional",
			input:         "",
			expectSuccess: true,
		},
		{
			name:          "whitespace reason when optional",
			input:         "  \n\t  ",
			expectSuccess: true,
		},
		{
			name:          "valid reason when optional",
			input:         "Test reason",
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewCloseDialogModel()
			dialog.Show(false) // Optional reason
			dialog.reasonInput.SetValue(tt.input)

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			result, _ := dialog.Update(msg)

			if tt.expectSuccess {
				assert.Equal(t, CloseDialogConfirmed, result.state)
				assert.False(t, result.showError)
			}
		})
	}
}

func TestCloseDialogModel_Update_EnterKey_RequiredReason(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectSuccess bool
		expectedError string
	}{
		{
			name:          "empty reason when required",
			input:         "",
			expectSuccess: false,
			expectedError: ErrReasonRequired,
		},
		{
			name:          "whitespace reason when required",
			input:         "  \n\t  ",
			expectSuccess: false,
			expectedError: ErrReasonWhitespace,
		},
		{
			name:          "valid reason when required",
			input:         "Test reason",
			expectSuccess: true,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewCloseDialogModel()
			dialog.Show(true) // Required reason
			dialog.reasonInput.SetValue(tt.input)

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			result, _ := dialog.Update(msg)

			if tt.expectSuccess {
				assert.Equal(t, CloseDialogConfirmed, result.state)
				assert.False(t, result.showError)
			} else {
				assert.Equal(t, CloseDialogInput, result.state)
				assert.True(t, result.showError)
				assert.Equal(t, tt.expectedError, result.errorMsg)
			}
		})
	}
}

func TestCloseDialogModel_Update_ErrorClearingOnTyping(t *testing.T) {
	dialog := NewCloseDialogModel()
	dialog.Show(true) // Required reason

	// First, trigger an error
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := dialog.Update(msg)
	assert.True(t, result.showError)
	assert.NotEmpty(t, result.errorMsg)

	// Type a character - error should clear
	dialog = result // Use the updated dialog
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	result2, _ := dialog.Update(msg)
	assert.False(t, result2.showError)
	assert.Empty(t, result2.errorMsg)
}

func TestCloseDialogModel_Update_NotVisibleState(t *testing.T) {
	dialog := NewCloseDialogModel()
	// Dialog is hidden by default

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, cmd := dialog.Update(msg)

	// Should return unchanged when not visible
	assert.Equal(t, dialog.state, result.state)
	assert.Nil(t, cmd)
}

func TestCloseDialogModel_View_HiddenState(t *testing.T) {
	dialog := NewCloseDialogModel()
	// Dialog is hidden by default

	view := dialog.View()
	assert.Empty(t, view)
}

func TestCloseDialogModel_View_VisibleState(t *testing.T) {
	tests := []struct {
		name          string
		requireReason bool
		showError     bool
		errorMsg      string
		width         int
	}{
		{
			name:          "optional reason dialog",
			requireReason: false,
			showError:     false,
			errorMsg:      "",
			width:         100,
		},
		{
			name:          "required reason dialog",
			requireReason: true,
			showError:     false,
			errorMsg:      "",
			width:         100,
		},
		{
			name:          "dialog with error",
			requireReason: true,
			showError:     true,
			errorMsg:      "Test error",
			width:         100,
		},
		{
			name:          "narrow screen dialog",
			requireReason: false,
			showError:     false,
			errorMsg:      "",
			width:         50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewCloseDialogModel()
			dialog.Show(tt.requireReason)
			dialog.showError = tt.showError
			dialog.errorMsg = tt.errorMsg
			dialog.width = tt.width

			view := dialog.View()
			require.NotEmpty(t, view)

			// Check for required elements
			assert.Contains(t, view, "Close Ticket")
			assert.Contains(t, view, "Enter: Confirm")
			assert.Contains(t, view, "ESC: Cancel")

			if tt.requireReason {
				assert.Contains(t, view, "(Reason Required)")
			}

			if tt.showError && tt.errorMsg != "" {
				assert.Contains(t, view, tt.errorMsg)
			}
		})
	}
}

func TestCloseDialogModel_Init(t *testing.T) {
	dialog := NewCloseDialogModel()
	cmd := dialog.Init()

	// Init should return textinput.Blink command
	assert.NotNil(t, cmd)
	// We can't directly compare function pointers, just ensure it's not nil
}

func TestCloseDialogModel_ResponsiveWidth(t *testing.T) {
	dialog := NewCloseDialogModel()
	dialog.Show(false)

	tests := []struct {
		name          string
		screenWidth   int
		expectedWidth int // Expected width used in rendering
	}{
		{
			name:          "wide screen",
			screenWidth:   100,
			expectedWidth: defaultDialogWidth,
		},
		{
			name:          "narrow screen below breakpoint",
			screenWidth:   50,
			expectedWidth: 50 - dialogMargin,
		},
		{
			name:          "at breakpoint",
			screenWidth:   dialogWidthBreakpoint,
			expectedWidth: defaultDialogWidth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog.width = tt.screenWidth
			view := dialog.View()

			// The view should be non-empty and contain the dialog
			assert.NotEmpty(t, view)
			// We can't directly test the width used, but we ensure the view is generated
			assert.Contains(t, view, "Close Ticket")
		})
	}
}

// TestCloseDialogModel_ValidationLogic tests the complete validation flow
func TestCloseDialogModel_ValidationLogic(t *testing.T) {
	t.Run("todo tickets require reason", func(t *testing.T) {
		dialog := NewCloseDialogModel()
		dialog.Show(true) // Simulating todo ticket (requires reason)

		// Test empty input
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := dialog.Update(msg)
		assert.Equal(t, CloseDialogInput, result.state)
		assert.True(t, result.showError)
		assert.Equal(t, ErrReasonRequired, result.errorMsg)

		// Test valid input
		dialog.reasonInput.SetValue("Abandoning due to priority change")
		result2, _ := dialog.Update(msg)
		assert.Equal(t, CloseDialogConfirmed, result2.state)
		assert.False(t, result2.showError)
	})

	t.Run("doing tickets have optional reason", func(t *testing.T) {
		dialog := NewCloseDialogModel()
		dialog.Show(false) // Simulating doing ticket (optional reason)

		// Test empty input - should succeed
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		result, _ := dialog.Update(msg)
		assert.Equal(t, CloseDialogConfirmed, result.state)
		assert.False(t, result.showError)
	})
}
