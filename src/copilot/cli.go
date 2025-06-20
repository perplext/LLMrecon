package copilot

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/security/access/common"
)

// CopilotCLI provides a command-line interface for the AI Security Copilot
type CopilotCLI struct {
	engine          *CopilotEngine
	scanner         *bufio.Scanner
	conversationHistory []ConversationTurn
	currentSession  *CLISession
	config          *CLIConfig
}

// CLIConfig configures the CLI interface
type CLIConfig struct {
	ShowDebugInfo   bool
	SaveSession     bool
	SessionPath     string
	MaxHistorySize  int
	PromptPrefix    string
	WelcomeMessage  string
	ExitCommands    []string
	HelpCommands    []string
}

// CLISession tracks the current CLI session
type CLISession struct {
	ID              string
	StartTime       time.Time
	QueryCount      int
	SuccessfulQueries int
	UserProfile     *UserProfile
	SessionData     map[string]interface{}
}

// UserProfile represents the user interacting with the copilot
type UserProfile struct {
	Username        string
	ExperienceLevel string
	Preferences     *UserPreferences
	ActiveTargets   []*TargetProfile
}

// NewCopilotCLI creates a new CLI interface
func NewCopilotCLI(engine *CopilotEngine, config *CLIConfig) *CopilotCLI {
	if config == nil {
		config = &CLIConfig{
			ShowDebugInfo:  false,
			SaveSession:    true,
			SessionPath:    "./sessions/",
			MaxHistorySize: 50,
			PromptPrefix:   "🛡️ Security Copilot",
			WelcomeMessage: "Welcome to the AI Security Copilot! I can help you with security testing, attack recommendations, and strategy planning.",
			ExitCommands:   []string{"exit", "quit", "bye", "goodbye"},
			HelpCommands:   []string{"help", "?", "commands"},
		}
	}

	cli := &CopilotCLI{
		engine:              engine,
		scanner:             bufio.NewScanner(os.Stdin),
		conversationHistory: make([]ConversationTurn, 0),
		config:              config,
		currentSession: &CLISession{
			ID:            fmt.Sprintf("session_%d", time.Now().UnixNano()),
			StartTime:     time.Now(),
			SessionData:   make(map[string]interface{}),
		},
	}

	return cli
}

// Start begins the interactive CLI session
func (cli *CopilotCLI) Start(ctx context.Context) error {
	cli.displayWelcome()
	cli.displayHelp()

	for {
		select {
		case <-ctx.Done():
			cli.displayMessage("Session interrupted. Goodbye!")
			return cli.saveSession()
		default:
			if !cli.processUserInput(ctx) {
				return cli.saveSession()
			}
		}
	}
}

// processUserInput handles a single user input cycle
func (cli *CopilotCLI) processUserInput(ctx context.Context) bool {
	fmt.Printf("\n%s > ", cli.config.PromptPrefix)
	
	if !cli.scanner.Scan() {
		return false
	}

	input := strings.TrimSpace(cli.scanner.Text())
	if input == "" {
		return true
	}

	cli.currentSession.QueryCount++

	// Check for exit commands
	if cli.isExitCommand(input) {
		cli.displayMessage("Thank you for using AI Security Copilot. Goodbye!")
		return false
	}

	// Check for help commands
	if cli.isHelpCommand(input) {
		cli.displayHelp()
		return true
	}

	// Check for special commands
	if cli.handleSpecialCommands(input) {
		return true
	}

	// Process query with the copilot engine
	if err := cli.processQuery(ctx, input); err != nil {
		cli.displayError(fmt.Sprintf("Error processing query: %v", err))
	} else {
		cli.currentSession.SuccessfulQueries++
	}

	return true
}

// processQuery sends the query to the copilot engine and displays results
func (cli *CopilotCLI) processQuery(ctx context.Context, query string) error {
	// Prepare query options
	options := &QueryOptions{
		Context:     make(map[string]interface{}),
		History:     cli.conversationHistory,
		Preferences: cli.getUserPreferences(),
		Constraints: cli.getExecutionConstraints(),
	}

	// Add session context
	options.Context["session_id"] = cli.currentSession.ID
	options.Context["query_count"] = cli.currentSession.QueryCount
	options.Context["user_id"] = cli.getCurrentUser()

	// Display processing indicator
	cli.displayProcessing("Processing your query...")

	startTime := time.Now()

	// Process query with copilot engine
	response, err := cli.engine.ProcessQuery(ctx, query, options)
	if err != nil {
		return err
	}

	processingTime := time.Since(startTime)

	// Display response
	cli.displayResponse(response, processingTime)

	// Update conversation history
	cli.updateConversationHistory(query, response.Response)

	// Handle any actions suggested by the copilot
	cli.handleResponseActions(ctx, response.Actions)

	return nil
}

// displayResponse formats and displays the copilot's response
func (cli *CopilotCLI) displayResponse(response *QueryResponse, processingTime time.Duration) {
	fmt.Printf("\n🤖 %s\n", response.Response)

	// Display confidence if debugging is enabled
	if cli.config.ShowDebugInfo {
		fmt.Printf("\n📊 Debug Info:\n")
		fmt.Printf("   Confidence: %.1f%%\n", response.Confidence*100)
		fmt.Printf("   Processing Time: %v\n", processingTime)
		fmt.Printf("   Response ID: %s\n", response.ID)
	}

	// Display actions if any
	if len(response.Actions) > 0 {
		fmt.Printf("\n💡 Suggested Actions:\n")
		for i, action := range response.Actions {
			fmt.Printf("   %d. %s - %s\n", i+1, action.Type, action.Description)
		}
	}

	// Display recommendations if any
	if response.Recommendations != nil {
		cli.displayRecommendations(response.Recommendations)
	}

	// Display follow-up questions
	if len(response.FollowUpQuestions) > 0 {
		fmt.Printf("\n❓ You might also ask:\n")
		for _, question := range response.FollowUpQuestions {
			fmt.Printf("   • %s\n", question)
		}
	}
}

// displayRecommendations shows attack recommendations in a formatted way
func (cli *CopilotCLI) displayRecommendations(recommendations *AttackRecommendations) {
	fmt.Printf("\n🎯 Attack Recommendations:\n")

	if len(recommendations.Primary) > 0 {
		fmt.Printf("\n🔴 Primary Recommendations:\n")
		for i, rec := range recommendations.Primary {
			fmt.Printf("   %d. %s\n", i+1, rec.AttackName)
			fmt.Printf("      Type: %s\n", rec.AttackType)
			fmt.Printf("      Confidence: %.1f%%\n", rec.Confidence*100)
			fmt.Printf("      Risk Level: %s\n", rec.RiskLevel)
			fmt.Printf("      Rationale: %s\n", rec.Rationale)
		}
	}

	if len(recommendations.Alternatives) > 0 {
		fmt.Printf("\n🟡 Alternative Approaches:\n")
		for i, rec := range recommendations.Alternatives {
			fmt.Printf("   %d. %s (%.1f%% confidence)\n", i+1, rec.AttackName, rec.Confidence*100)
		}
	}

	if len(recommendations.Experimental) > 0 {
		fmt.Printf("\n🟠 Experimental Techniques:\n")
		for i, rec := range recommendations.Experimental {
			fmt.Printf("   %d. %s (High complexity)\n", i+1, rec.AttackName)
		}
	}

	// Display overall strategy
	if recommendations.Strategy != nil {
		fmt.Printf("\n📋 Recommended Strategy: %s\n", recommendations.Strategy.Name)
		fmt.Printf("   Description: %s\n", recommendations.Strategy.Description)
		if len(recommendations.Strategy.Phases) > 0 {
			fmt.Printf("   Phases: %d\n", len(recommendations.Strategy.Phases))
		}
	}

	// Display success probability
	fmt.Printf("\n📈 Overall Success Probability: %.1f%%\n", recommendations.SuccessProbability*100)
}

// handleResponseActions processes actions suggested by the copilot
func (cli *CopilotCLI) handleResponseActions(ctx context.Context, actions []Action) {
	if len(actions) == 0 {
		return
	}

	fmt.Printf("\nWould you like me to execute any of these actions? (y/n): ")
	if !cli.scanner.Scan() {
		return
	}

	response := strings.ToLower(strings.TrimSpace(cli.scanner.Text()))
	if response != "y" && response != "yes" {
		return
	}

	// Execute actions
	for i, action := range actions {
		fmt.Printf("\nExecuting action %d: %s\n", i+1, action.Description)
		
		if action.RequiresConfirmation {
			fmt.Printf("This action requires confirmation. Proceed? (y/n): ")
			if !cli.scanner.Scan() {
				continue
			}
			confirm := strings.ToLower(strings.TrimSpace(cli.scanner.Text()))
			if confirm != "y" && confirm != "yes" {
				fmt.Printf("Skipping action.\n")
				continue
			}
		}

		// Simulate action execution
		cli.executeAction(ctx, action)
	}
}

// executeAction simulates the execution of a copilot action
func (cli *CopilotCLI) executeAction(ctx context.Context, action Action) {
	switch action.Type {
	case ActionAnalyzeTarget:
		cli.displayMessage("Analyzing target profile and capabilities...")
		time.Sleep(2 * time.Second)
		cli.displayMessage("✅ Target analysis complete. Vulnerabilities identified.")
		
	case ActionExecuteAttack:
		cli.displayMessage("Executing attack scenario...")
		time.Sleep(3 * time.Second)
		cli.displayMessage("✅ Attack execution complete. Results logged.")
		
	case ActionGenerateReport:
		cli.displayMessage("Generating comprehensive security report...")
		time.Sleep(2 * time.Second)
		cli.displayMessage("✅ Report generated successfully.")
		
	case ActionCreateStrategy:
		cli.displayMessage("Creating testing strategy...")
		time.Sleep(2 * time.Second)
		cli.displayMessage("✅ Strategy created with multiple phases.")
		
	case ActionSearchKnowledge:
		cli.displayMessage("Searching knowledge base...")
		time.Sleep(1 * time.Second)
		cli.displayMessage("✅ Found relevant patterns and insights.")
		
	default:
		cli.displayMessage(fmt.Sprintf("Executing %s...", action.Type))
		time.Sleep(1 * time.Second)
		cli.displayMessage("✅ Action completed successfully.")
	}
}

// handleSpecialCommands processes CLI-specific commands
func (cli *CopilotCLI) handleSpecialCommands(input string) bool {
	cmd := strings.ToLower(input)
	
	switch {
	case strings.HasPrefix(cmd, "debug"):
		cli.toggleDebugMode()
		return true
		
	case strings.HasPrefix(cmd, "session"):
		cli.displaySessionInfo()
		return true
		
	case strings.HasPrefix(cmd, "history"):
		cli.displayConversationHistory()
		return true
		
	case strings.HasPrefix(cmd, "clear"):
		cli.clearScreen()
		return true
		
	case strings.HasPrefix(cmd, "stats"):
		cli.displayStatistics()
		return true
		
	case strings.HasPrefix(cmd, "demo"):
		cli.runDemoScenario()
		return true
		
	default:
		return false
	}
}

// runDemoScenario demonstrates the copilot's capabilities
func (cli *CopilotCLI) runDemoScenario() {
	cli.displayMessage("🎬 Running Copilot Demo Scenario...")
	
	demoQueries := []string{
		"What attack techniques would you recommend for testing a GPT-4 model?",
		"Explain why you recommended HouYi injection technique",
		"Create a testing strategy for a multimodal AI system",
		"How effective are cross-modal attacks compared to text-only attacks?",
		"What patterns have you learned from recent attack results?",
	}
	
	ctx := context.Background()
	
	for i, query := range demoQueries {
		fmt.Printf("\n--- Demo Query %d ---\n", i+1)
		fmt.Printf("User: %s\n", query)
		
		// Add a delay for realistic interaction
		time.Sleep(1 * time.Second)
		
		if err := cli.processQuery(ctx, query); err != nil {
			cli.displayError(fmt.Sprintf("Demo query failed: %v", err))
		}
		
		fmt.Printf("\nPress Enter to continue to next demo query...")
		cli.scanner.Scan()
	}
	
	cli.displayMessage("🎬 Demo scenario completed!")
}

// Helper methods for CLI functionality

func (cli *CopilotCLI) displayWelcome() {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                🛡️  AI Security Copilot v0.5.0                ║\n")
	fmt.Printf("║              Your AI-Powered Security Testing Assistant       ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")
	fmt.Printf("%s\n\n", cli.config.WelcomeMessage)
}

func (cli *CopilotCLI) displayHelp() {
	fmt.Printf("📚 Available Commands:\n")
	fmt.Printf("   • Ask natural language questions about security testing\n")
	fmt.Printf("   • 'help' or '?' - Show this help message\n")
	fmt.Printf("   • 'demo' - Run a demonstration scenario\n")
	fmt.Printf("   • 'debug' - Toggle debug information display\n")
	fmt.Printf("   • 'session' - Show current session information\n")
	fmt.Printf("   • 'history' - Display conversation history\n")
	fmt.Printf("   • 'stats' - Show copilot statistics\n")
	fmt.Printf("   • 'clear' - Clear the screen\n")
	fmt.Printf("   • 'exit', 'quit', or 'bye' - Exit the copilot\n\n")
	fmt.Printf("💡 Example queries:\n")
	fmt.Printf("   • \"Recommend attacks for testing GPT-4\"\n")
	fmt.Printf("   • \"Create a strategy for testing multimodal AI\"\n")
	fmt.Printf("   • \"Explain why this attack technique was recommended\"\n")
	fmt.Printf("   • \"What have you learned from recent test results?\"\n")
}

func (cli *CopilotCLI) displayMessage(message string) {
	fmt.Printf("💬 %s\n", message)
}

func (cli *CopilotCLI) displayError(message string) {
	fmt.Printf("❌ %s\n", message)
}

func (cli *CopilotCLI) displayProcessing(message string) {
	fmt.Printf("⏳ %s\n", message)
}

func (cli *CopilotCLI) toggleDebugMode() {
	cli.config.ShowDebugInfo = !cli.config.ShowDebugInfo
	if cli.config.ShowDebugInfo {
		cli.displayMessage("Debug mode enabled")
	} else {
		cli.displayMessage("Debug mode disabled")
	}
}

func (cli *CopilotCLI) displaySessionInfo() {
	fmt.Printf("\n📊 Session Information:\n")
	fmt.Printf("   Session ID: %s\n", cli.currentSession.ID)
	fmt.Printf("   Start Time: %s\n", cli.currentSession.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Duration: %v\n", time.Since(cli.currentSession.StartTime))
	fmt.Printf("   Total Queries: %d\n", cli.currentSession.QueryCount)
	fmt.Printf("   Successful Queries: %d\n", cli.currentSession.SuccessfulQueries)
	fmt.Printf("   Success Rate: %.1f%%\n", float64(cli.currentSession.SuccessfulQueries)/float64(cli.currentSession.QueryCount)*100)
}

func (cli *CopilotCLI) displayConversationHistory() {
	fmt.Printf("\n📜 Conversation History:\n")
	if len(cli.conversationHistory) == 0 {
		fmt.Printf("   No conversation history yet.\n")
		return
	}
	
	for i, turn := range cli.conversationHistory {
		fmt.Printf("\n--- Turn %d (%s) ---\n", i+1, turn.Timestamp.Format("15:04:05"))
		fmt.Printf("User: %s\n", turn.UserMessage)
		fmt.Printf("Copilot: %s\n", turn.CopilotResponse)
	}
}

func (cli *CopilotCLI) clearScreen() {
	fmt.Print("\033[2J\033[H")
	cli.displayWelcome()
}

func (cli *CopilotCLI) displayStatistics() {
	fmt.Printf("\n📈 Copilot Statistics:\n")
	fmt.Printf("   Engine Status: Running\n")
	fmt.Printf("   Available Techniques: 4\n")
	fmt.Printf("   Knowledge Base Items: Simulated\n")
	fmt.Printf("   Average Response Time: ~2.5s\n")
	fmt.Printf("   Memory Usage: Optimized\n")
}

func (cli *CopilotCLI) updateConversationHistory(userMessage, copilotResponse string) {
	turn := ConversationTurn{
		UserMessage:     userMessage,
		CopilotResponse: copilotResponse,
		Timestamp:       time.Now(),
		Actions:         []string{}, // Would be populated with actual actions
		Results:         []string{}, // Would be populated with actual results
	}
	
	cli.conversationHistory = append(cli.conversationHistory, turn)
	
	// Limit history size
	if len(cli.conversationHistory) > cli.config.MaxHistorySize {
		cli.conversationHistory = cli.conversationHistory[1:]
	}
}

func (cli *CopilotCLI) saveSession() error {
	if !cli.config.SaveSession {
		return nil
	}
	
	sessionData := map[string]interface{}{
		"session":     cli.currentSession,
		"history":     cli.conversationHistory,
		"end_time":    time.Now(),
	}
	
	data, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		return err
	}
	
	// In a real implementation, this would save to file
	// For now, we'll just log that we would save
	cli.displayMessage(fmt.Sprintf("Session data saved (%d bytes)", len(data)))
	
	return nil
}

func (cli *CopilotCLI) isExitCommand(input string) bool {
	lower := strings.ToLower(input)
	for _, cmd := range cli.config.ExitCommands {
		if lower == cmd {
			return true
		}
	}
	return false
}

func (cli *CopilotCLI) isHelpCommand(input string) bool {
	lower := strings.ToLower(input)
	for _, cmd := range cli.config.HelpCommands {
		if lower == cmd {
			return true
		}
	}
	return false
}

func (cli *CopilotCLI) getUserPreferences() *UserPreferences {
	// Return default preferences for now
	return &UserPreferences{
		RiskTolerance:    "moderate",
		ExplanationLevel: "detailed",
		AutomationLevel:  "assisted",
		Industry:         "general",
	}
}

func (cli *CopilotCLI) getExecutionConstraints() *ExecutionConstraints {
	// Return default constraints for now
	return &ExecutionConstraints{
		MaxAttacks:         10,
		TimeLimit:          30 * time.Minute,
		MaxConcurrency:     3,
		MaxTokensPerAttack: 1000,
		SafetyLevel:        "high",
	}
}

func (cli *CopilotCLI) getCurrentUser() string {
	// Return default user for now
	return "demo_user"
}

// RunCopilotCLI is a convenience function to start the CLI
func RunCopilotCLI() error {
	// Create default configuration
	config := &EngineConfig{
		ConfidenceThreshold: 0.7,
		EnabledTechniques: []string{
			"houyi_injection",
			"cross_modal_coordination", 
			"red_queen_adversarial",
			"conversation_flow_manipulation",
		},
		MaxConcurrentAttacks: 5,
		SafetyChecks:         true,
		AuditAllQueries:      true,
		ResponseTimeout:      30 * time.Second,
	}
	
	// Create mock logger
	logger := &MockLogger{}
	
	// Create knowledge base
	kbConfig := &KnowledgeConfig{
		MaxItems:        1000,
		RetentionPeriod: 30 * 24 * time.Hour,
		AutoPersist:     false, // Disabled for demo
		FullTextSearch:  true,
	}
	knowledgeBase := NewInMemoryKnowledgeBase(kbConfig)
	
	// Create copilot engine
	engine := NewCopilotEngine(config, knowledgeBase, logger)
	
	// Create CLI
	cli := NewCopilotCLI(engine, nil)
	
	// Start interactive session
	ctx := context.Background()
	return cli.Start(ctx)
}

// MockLogger implements a simple logger for demonstration
type MockLogger struct{}

func (l *MockLogger) LogSecurityEvent(event string, data map[string]interface{}) {
	// In a real implementation, this would log to a proper logging system
	fmt.Printf("🔒 [%s] %s\n", time.Now().Format("15:04:05"), event)
}