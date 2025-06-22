package telegram

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"github.com/giovannigabriele/go-todo-bot/internal/llm"
	"github.com/giovannigabriele/go-todo-bot/internal/sheets"
)

// ValidateTelegramToken checks if a Telegram bot token is valid
func ValidateTelegramToken(token string) error {
	// Check if token is empty
	if token == "" {
		return fmt.Errorf("token is empty")
	}

	// Check if token matches expected format: bot123456:ABC-DEF...
	// Telegram bot tokens are typically in format: <bot_id>:<auth_token>
	tokenPattern := regexp.MustCompile(`^\d+:[A-Za-z0-9_-]+$`)
	if !tokenPattern.MatchString(token) {
		return fmt.Errorf("token format is invalid (expected format: 123456:ABC-DEF...)")
	}

	// Check for obvious placeholder values
	lowerToken := strings.ToLower(token)
	placeholders := []string{"placeholder", "your_token", "bot_token", "token_here", "example", "test"}
	for _, placeholder := range placeholders {
		if strings.Contains(lowerToken, placeholder) {
			return fmt.Errorf("token appears to be a placeholder value")
		}
	}

	// Try to create a bot instance to validate the token
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("failed to validate token with Telegram API: %w", err)
	}

	// If we can get the bot info, the token is valid
	log.Debug().Str("bot_username", bot.Self.UserName).Msg("Telegram token validated successfully")
	return nil
}

// Handler handles Telegram bot interactions
type Handler struct {
	bot          *tgbotapi.BotAPI
	llmClient    *llm.Client
	sheetsClient *sheets.Client
}

// NewHandler creates a new Telegram handler
func NewHandler(token string, llmClient *llm.Client, sheetsClient *sheets.Client) (*Handler, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Info().Str("username", bot.Self.UserName).Msg("Telegram bot authorized")

	return &Handler{
		bot:          bot,
		llmClient:    llmClient,
		sheetsClient: sheetsClient,
	}, nil
}

// Start starts the bot with long polling
func (h *Handler) Start(ctx context.Context) error {
	log.Info().Msg("Starting Telegram bot with long polling")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stopping Telegram bot")
			h.bot.StopReceivingUpdates()
			return ctx.Err()
		case update := <-updates:
			if update.Message != nil {
				go h.handleMessage(ctx, update.Message)
			}
		}
	}
}

// handleMessage processes incoming messages
func (h *Handler) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	log.Debug().
		Int64("chat_id", message.Chat.ID).
		Str("text", message.Text).
		Str("username", message.From.UserName).
		Msg("Received message")

	// Handle commands
	if message.IsCommand() {
		h.handleCommand(ctx, message)
		return
	}

	// Process regular messages as tasks
	if message.Text != "" {
		h.processTaskMessage(ctx, message)
	}
}

// handleCommand processes bot commands
func (h *Handler) handleCommand(ctx context.Context, message *tgbotapi.Message) {
	command := message.Command()

	log.Debug().
		Str("command", command).
		Int64("chat_id", message.Chat.ID).
		Msg("Processing command")

	var response string

	switch command {
	case "start":
		response = h.getStartMessage()
	case "help":
		response = h.getHelpMessage()
	case "status":
		response = h.getStatusMessage(ctx)
	default:
		response = "Unknown command. Use /help to see available commands."
	}

	h.sendMessage(message.Chat.ID, response)
}

// processTaskMessage processes a message as a potential task
func (h *Handler) processTaskMessage(ctx context.Context, message *tgbotapi.Message) {
	// Send immediate acknowledgment
	h.sendMessage(message.Chat.ID, "🔖 Processing your message...")

	// Parse message with LLM
	parseResp, err := h.llmClient.ParseMessage(ctx, message.Text)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse message with LLM")
		h.handleParseError(message, err)
		return
	}

	// Convert to sheet rows
	var taskRows []sheets.TaskRow
	botNotes := ""

	for i, task := range parseResp.Tasks {
		summary := task.Summary
		if len(parseResp.Tasks) > 1 {
			summary = fmt.Sprintf("%s (%d/%d)", task.Summary, i+1, len(parseResp.Tasks))
		}

		if task.Confidence < 0.7 {
			botNotes = fmt.Sprintf("Low confidence parse (%.2f)", task.Confidence)
		}

		taskRow := sheets.CreateTaskRow(
			task.People,
			task.Client,
			summary,
			message.Text,
			task.DueDate,
			botNotes,
		)
		taskRows = append(taskRows, taskRow)
	}

	// Save to Google Sheets
	if err := h.sheetsClient.AddTasks(ctx, taskRows); err != nil {
		log.Error().Err(err).Msg("Failed to save tasks to Google Sheets")
		h.handleSaveError(message, err)
		return
	}

	// Send success response
	h.sendSuccessResponse(message, taskRows)
}

// handleParseError handles LLM parsing errors
func (h *Handler) handleParseError(message *tgbotapi.Message, err error) {
	log.Warn().Err(err).Str("message", message.Text).Msg("Parse error, creating fallback task")

	// Create fallback task assigned to "team"
	taskRow := sheets.CreateTaskRow(
		[]string{"team"},
		"Internal",
		h.extractSimpleSummary(message.Text),
		message.Text,
		"unclear",
		fmt.Sprintf("Parse error: %s", err.Error()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.sheetsClient.AddTasks(ctx, []sheets.TaskRow{taskRow}); err != nil {
		log.Error().Err(err).Msg("Failed to save fallback task")
		h.sendMessage(message.Chat.ID, "❌ Sorry, I couldn't process your message. Please try again later.")
		return
	}

	response := "⚠️ I had trouble parsing your message, but I've saved it as a team task. " +
		"You can update the assignment in the Google Sheet if needed."
	h.sendMessage(message.Chat.ID, response)
}

// handleSaveError handles Google Sheets save errors
func (h *Handler) handleSaveError(message *tgbotapi.Message, err error) {
	response := "❌ I understood your message but couldn't save it to the sheet. " +
		"Please try again in a few minutes."
	h.sendMessage(message.Chat.ID, response)
}

// sendSuccessResponse sends a success message after saving tasks
func (h *Handler) sendSuccessResponse(message *tgbotapi.Message, taskRows []sheets.TaskRow) {
	var response strings.Builder
	response.WriteString("✅ Saved ")

	if len(taskRows) == 1 {
		row := taskRows[0]
		response.WriteString(fmt.Sprintf("task for %s: \"%s\"", row.People, row.Summary))
	} else {
		response.WriteString(fmt.Sprintf("%d tasks:\n", len(taskRows)))
		for i, row := range taskRows {
			response.WriteString(fmt.Sprintf("%d. %s: \"%s\"\n", i+1, row.People, row.Summary))
		}
	}

	h.sendMessage(message.Chat.ID, response.String())
}

// sendMessage sends a message to a chat
func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := h.bot.Send(msg); err != nil {
		log.Error().Err(err).Int64("chat_id", chatID).Msg("Failed to send message")
	}
}

// getStartMessage returns the start message
func (h *Handler) getStartMessage() string {
	return `👋 Welcome to the TODO Bot!

I help you capture tasks and organize them in Google Sheets.

Just send me a message like:
• "Lilly to call Johnny about the project"
• "Team meeting tomorrow at 2pm"
• "Review the budget AND send it to Alice"

I'll automatically:
✅ Extract who's responsible
✅ Create a task summary
✅ Save it to your Google Sheet
✅ Split multiple tasks when needed

Use /help for more commands.`
}

// getHelpMessage returns the help message
func (h *Handler) getHelpMessage() string {
	return `🤖 TODO Bot Commands:

/start - Show welcome message
/help - Show this help
/status - Check bot status

📝 How to use:
Just send me any message describing a task or reminder. I'll automatically parse it and save it to your Google Sheet.

Examples:
• "Lilly to reach out to Johnny for a quote"
• "Jemma and Lexi to set up the meeting"
• "Call the vendor AND review the contract"
• "Team standup at 9am tomorrow"

The bot will:
• Identify who's responsible
• Extract the task description
• Save it with timestamp and status
• Split multiple tasks automatically`
}

// getStatusMessage returns the status message
func (h *Handler) getStatusMessage(ctx context.Context) string {
	// Try to get team data to test connectivity
	_, err := h.sheetsClient.GetTeam(ctx)
	if err != nil {
		return fmt.Sprintf("⚠️ Bot is running but Google Sheets connection failed: %s", err.Error())
	}

	return "✅ Bot is running and connected to Google Sheets!"
}

// extractSimpleSummary creates a simple summary from text
func (h *Handler) extractSimpleSummary(text string) string {
	const maxLength = 80

	// Take first sentence or truncate
	sentences := strings.Split(text, ".")
	if len(sentences) > 0 && len(sentences[0]) > 0 {
		summary := strings.TrimSpace(sentences[0])
		if len(summary) <= maxLength {
			return summary
		}
	}

	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}
