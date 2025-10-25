// Package execution implements Layer 4 (Main Agent) of the 7-layer architecture.
// This is the core execution layer that uses tools to complete tasks.
package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/internal/logging"
	"github.com/abrksh22/bplus/models"
	"github.com/abrksh22/bplus/security"
	"github.com/abrksh22/bplus/tools"
)

// Agent represents the main execution agent (Layer 4).
type Agent struct {
	provider    models.Provider
	config      *AgentConfig
	toolReg     *tools.Registry
	permMgr     *security.PermissionManager
	logger      *logging.Logger
	costTracker *CostTracker
}

// AgentConfig holds configuration for the agent.
type AgentConfig struct {
	ModelName     string  // Model to use (e.g., "anthropic/claude-sonnet-4-5")
	SystemPrompt  string  // System prompt for the agent
	MaxIterations int     // Maximum agent loop iterations
	Temperature   float64 // Temperature for generation
	MaxTokens     int     // Maximum tokens per generation
	Streaming     bool    // Enable streaming responses
}

// NewAgent creates a new agent with the given configuration.
func NewAgent(provider models.Provider, config *AgentConfig, toolReg *tools.Registry, permMgr *security.PermissionManager) (*Agent, error) {
	if provider == nil {
		return nil, errors.New(errors.ErrCodeValidation, "provider cannot be nil")
	}
	if config == nil {
		return nil, errors.New(errors.ErrCodeValidation, "config cannot be nil")
	}
	if toolReg == nil {
		return nil, errors.New(errors.ErrCodeValidation, "tool registry cannot be nil")
	}
	if permMgr == nil {
		return nil, errors.New(errors.ErrCodeValidation, "permission manager cannot be nil")
	}

	return &Agent{
		provider:    provider,
		config:      config,
		toolReg:     toolReg,
		permMgr:     permMgr,
		logger:      logging.NewDefaultLogger().WithComponent("execution"),
		costTracker: NewCostTracker(),
	}, nil
}

// AgentRequest represents a request to the agent.
type AgentRequest struct {
	// User's message
	UserMessage string

	// Conversation history (excluding current message)
	History []models.Message

	// Session ID for context
	SessionID string

	// Context from Layer 6 (optional)
	Context string
}

// AgentResponse represents the agent's response.
type AgentResponse struct {
	// Response content
	Content string

	// Tool calls made during execution
	ToolCalls []ToolExecution

	// Token usage
	Usage models.Usage

	// Total iterations
	Iterations int

	// Whether the task is complete
	Complete bool

	// Any errors encountered
	Error error
}

// ToolExecution represents a single tool execution in the agent loop.
type ToolExecution struct {
	ToolName   string
	Arguments  map[string]interface{}
	Result     *tools.Result
	Timestamp  time.Time
	Permission bool // Whether permission was granted
}

// Execute runs the agent loop to complete the given request.
func (a *Agent) Execute(ctx context.Context, req *AgentRequest) (*AgentResponse, error) {
	if req == nil {
		return nil, errors.New(errors.ErrCodeValidation, "request cannot be nil")
	}

	a.logger.Info("Starting agent execution", "session_id", req.SessionID, "model", a.config.ModelName)

	response := &AgentResponse{
		ToolCalls: make([]ToolExecution, 0),
	}

	// Build conversation messages
	messages := append(req.History, models.Message{
		Role:    "user",
		Content: req.UserMessage,
	})

	// Get available tools
	availableTools := a.getAvailableTools()

	// Agent loop
	for iteration := 0; iteration < a.config.MaxIterations; iteration++ {
		response.Iterations = iteration + 1

		a.logger.Debug("Agent iteration", "iteration", iteration+1, "max", a.config.MaxIterations)

		// Call LLM
		completionReq := &models.CompletionRequest{
			Model:     a.config.ModelName,
			Messages:  messages,
			System:    a.config.SystemPrompt,
			Tools:     availableTools,
			MaxTokens: a.config.MaxTokens,
		}

		if a.config.Temperature > 0 {
			completionReq.Temperature = &a.config.Temperature
		}

		var completionResp *models.CompletionResponse
		var err error

		if a.config.Streaming && a.provider.SupportsStreaming() {
			completionResp, err = a.executeStreaming(ctx, completionReq)
		} else {
			completionResp, err = a.provider.CreateCompletion(ctx, completionReq)
		}

		if err != nil {
			a.logger.Error("LLM call failed", err, "iteration", iteration+1)
			return nil, errors.Wrap(err, errors.ErrCodeProvider, "LLM call failed")
		}

		// Track usage and cost
		a.costTracker.AddUsage(completionResp.Usage)
		response.Usage.InputTokens += completionResp.Usage.InputTokens
		response.Usage.OutputTokens += completionResp.Usage.OutputTokens
		response.Usage.TotalTokens += completionResp.Usage.TotalTokens
		response.Usage.Cost += completionResp.Usage.Cost

		// Check stop reason
		if completionResp.StopReason == "end_turn" || completionResp.StopReason == "stop_sequence" {
			// Task complete
			response.Content = completionResp.Content
			response.Complete = true
			a.logger.Info("Agent execution complete", "iterations", iteration+1, "cost", response.Usage.Cost)
			return response, nil
		}

		if completionResp.StopReason == "tool_use" && len(completionResp.ToolCalls) > 0 {
			// Execute tool calls
			toolResults := make([]models.Message, 0, len(completionResp.ToolCalls))

			for _, toolCall := range completionResp.ToolCalls {
				execution := ToolExecution{
					ToolName:  toolCall.Name,
					Arguments: toolCall.Arguments,
					Timestamp: time.Now(),
				}

				// Execute tool with permission check
				result, err := a.executeTool(ctx, toolCall.Name, toolCall.Arguments)
				execution.Result = result
				execution.Permission = (err == nil) // Permission was granted if no error

				response.ToolCalls = append(response.ToolCalls, execution)

				// Format tool result as message
				var resultContent string
				if err != nil {
					resultContent = fmt.Sprintf("Error: %v", err)
				} else if result.Success {
					resultContent = fmt.Sprintf("%v", result.Output)
				} else {
					resultContent = fmt.Sprintf("Tool failed: %v", result.Error)
				}

				toolResults = append(toolResults, models.Message{
					Role:    "tool",
					Content: resultContent,
					Name:    toolCall.Name,
				})
			}

			// Add assistant message with tool calls
			if completionResp.Content != "" {
				messages = append(messages, models.Message{
					Role:    "assistant",
					Content: completionResp.Content,
				})
			}

			// Add tool results to conversation
			messages = append(messages, toolResults...)

			// Continue loop to let agent process tool results
			continue
		}

		// If we get here, something unexpected happened
		response.Content = completionResp.Content
		response.Complete = false
		a.logger.Warn("Agent stopped with unexpected reason", "stop_reason", completionResp.StopReason)
		return response, nil
	}

	// Max iterations reached
	a.logger.Warn("Agent reached max iterations", "max", a.config.MaxIterations)
	response.Complete = false
	return response, errors.New(errors.ErrCodeInternal, "agent reached maximum iterations without completing task")
}

// executeTool executes a single tool with permission checking.
func (a *Agent) executeTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*tools.Result, error) {
	a.logger.Debug("Executing tool", "tool", toolName, "args", arguments)

	// Get tool from registry
	tool, err := a.toolReg.Get(toolName)
	if err != nil {
		return nil, errors.Newf(errors.ErrCodeToolNotFound, "tool %s not found", toolName)
	}

	// Check permissions
	if tool.RequiresPermission() {
		permission := determinePermission(tool)
		resource := determineResource(arguments)

		req := &security.PermissionRequest{
			Permission: permission,
			Resource:   resource,
			Operation:  fmt.Sprintf("execute %s", toolName),
			Reason:     "Tool execution requested by agent",
			ToolName:   toolName,
		}

		granted, err := a.permMgr.Check(ctx, req)
		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrCodeToolPermission, "failed to request permission for tool %s", toolName)
		}
		if !granted {
			return nil, errors.Newf(errors.ErrCodeToolPermission, "permission denied for tool %s", toolName)
		}
	}

	// Execute tool
	result, err := tool.Execute(ctx, arguments)
	if err != nil {
		a.logger.Error("Tool execution failed", err, "tool", toolName)
		return nil, errors.Wrapf(err, errors.ErrCodeToolExecution, "tool %s execution failed", toolName)
	}

	return result, nil
}

// executeStreaming handles streaming completions.
func (a *Agent) executeStreaming(ctx context.Context, req *models.CompletionRequest) (*models.CompletionResponse, error) {
	tokenChan, err := a.provider.StreamCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	var content string
	var toolCalls []models.ToolCall
	var stopReason string
	var usage models.Usage

	for token := range tokenChan {
		if token.Error != nil {
			return nil, token.Error
		}

		if token.Content != "" {
			content += token.Content
			// TODO: Stream to UI in Phase 6.2
		}

		if token.ToolCall != nil {
			toolCalls = append(toolCalls, *token.ToolCall)
		}

		if token.Done {
			if token.Usage != nil {
				usage = *token.Usage
			}
			stopReason = "end_turn"
			if len(toolCalls) > 0 {
				stopReason = "tool_use"
			}
			break
		}
	}

	return &models.CompletionResponse{
		Content:    content,
		ToolCalls:  toolCalls,
		StopReason: stopReason,
		Usage:      usage,
		Model:      req.Model,
	}, nil
}

// getAvailableTools returns tools in the format expected by the LLM provider.
func (a *Agent) getAvailableTools() []models.Tool {
	registeredTools := a.toolReg.AllTools()
	llmTools := make([]models.Tool, 0, len(registeredTools))

	for _, tool := range registeredTools {
		params := make([]models.Parameter, 0, len(tool.Parameters()))
		required := make([]string, 0)

		for _, param := range tool.Parameters() {
			params = append(params, models.Parameter{
				Name:        param.Name,
				Type:        string(param.Type),
				Description: param.Description,
				Required:    param.Required,
			})
			if param.Required {
				required = append(required, param.Name)
			}
		}

		llmTools = append(llmTools, models.Tool{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  params,
			Required:    required,
		})
	}

	return llmTools
}

// determinePermission maps tool categories to permission types.
func determinePermission(tool tools.Tool) security.Permission {
	switch tool.Category() {
	case "file":
		// Determine read vs write based on tool name
		switch tool.Name() {
		case "core.read", "core.glob", "core.grep":
			return security.PermissionRead
		case "core.write", "core.edit":
			return security.PermissionWrite
		default:
			return security.PermissionWrite
		}
	case "exec":
		return security.PermissionExecute
	case "web":
		return security.PermissionNetwork
	case "mcp":
		return security.PermissionMCP
	default:
		return security.PermissionWrite // Default to write for safety
	}
}

// determineResource extracts the resource being accessed from tool arguments.
func determineResource(arguments map[string]interface{}) string {
	// Common resource parameter names
	for _, key := range []string{"file_path", "path", "pattern", "command", "url"} {
		if val, ok := arguments[key]; ok {
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

// GetCostTracker returns the cost tracker for this agent.
func (a *Agent) GetCostTracker() *CostTracker {
	return a.costTracker
}

// UpdateConfig updates the agent's configuration.
func (a *Agent) UpdateConfig(config *AgentConfig) {
	if config != nil {
		a.config = config
	}
}
