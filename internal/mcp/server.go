// Package mcp implements a JSON-RPC based MCP (Model Context Protocol) server
// that exposes Rune repository context to AI coding agents.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rune-context/rune/internal/config"
	"github.com/rune-context/rune/internal/context"
	"github.com/rune-context/rune/internal/doctor"
	"github.com/rune-context/rune/internal/graph"
	"github.com/rune-context/rune/internal/version"
)

// JSON-RPC types
type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP protocol types
type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type initializeResult struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      serverInfo        `json:"serverInfo"`
}

type serverCapabilities struct {
	Tools *toolsCapability `json:"tools,omitempty"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema inputSchema `json:"inputSchema"`
}

type inputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]schemaProperty `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

type schemaProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type toolsListResult struct {
	Tools []tool `json:"tools"`
}

type callToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type toolResult struct {
	Content []contentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Server is the MCP server instance.
type Server struct {
	root string
}

// NewServer creates a new MCP server.
func NewServer(root string) *Server {
	return &Server{root: root}
}

// Serve starts the MCP server on stdio.
func (s *Server) Serve() error {
	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeError(writer, nil, -32700, "Parse error")
			continue
		}

		resp := s.handleRequest(&req)
		if resp != nil {
			s.writeResponse(writer, resp)
		}
	}
}

func (s *Server) handleRequest(req *jsonrpcRequest) *jsonrpcResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "initialized":
		return nil // notification, no response
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "ping":
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{},
		}
	default:
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32601, Message: "Method not found: " + req.Method},
		}
	}
}

func (s *Server) handleInitialize(req *jsonrpcRequest) *jsonrpcResponse {
	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: initializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: serverCapabilities{
				Tools: &toolsCapability{ListChanged: false},
			},
			ServerInfo: serverInfo{
				Name:    "rune",
				Version: version.Version,
			},
		},
	}
}

func (s *Server) handleToolsList(req *jsonrpcRequest) *jsonrpcResponse {
	tools := []tool{
		{
			Name:        "rune_context",
			Description: "Get relevant repository context for a task or query. Returns file summaries, architecture info, and conventions.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProperty{
					"query": {Type: "string", Description: "The task or question to get context for"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "rune_files",
			Description: "List all files tracked in the repository dependency graph.",
			InputSchema: inputSchema{Type: "object"},
		},
		{
			Name:        "rune_graph",
			Description: "Get the dependency graph for a specific file, showing what it depends on and what depends on it.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProperty{
					"file": {Type: "string", Description: "The file path to get dependencies for"},
				},
				Required: []string{"file"},
			},
		},
		{
			Name:        "rune_summary",
			Description: "Get the summary for a specific file.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]schemaProperty{
					"file": {Type: "string", Description: "The file path to get summary for"},
				},
				Required: []string{"file"},
			},
		},
		{
			Name:        "rune_architecture",
			Description: "Get the repository architecture summary.",
			InputSchema: inputSchema{Type: "object"},
		},
		{
			Name:        "rune_conventions",
			Description: "Get the repository coding conventions.",
			InputSchema: inputSchema{Type: "object"},
		},
		{
			Name:        "rune_doctor",
			Description: "Check repository health and report any issues.",
			InputSchema: inputSchema{Type: "object"},
		},
	}

	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolsListResult{Tools: tools},
	}
}

func (s *Server) handleToolsCall(req *jsonrpcRequest) *jsonrpcResponse {
	var params callToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32602, Message: "Invalid params"},
		}
	}

	var result toolResult

	switch params.Name {
	case "rune_context":
		query, _ := params.Arguments["query"].(string)
		if query == "" {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: "Error: query is required"}},
				IsError: true,
			}
		} else {
			ctx, err := context.Query(s.root, query)
			if err != nil {
				result = toolResult{
					Content: []contentItem{{Type: "text", Text: "Error: " + err.Error()}},
					IsError: true,
				}
			} else {
				result = toolResult{
					Content: []contentItem{{Type: "text", Text: ctx.Format()}},
				}
			}
		}

	case "rune_files":
		g, err := graph.Load(config.SubPath(s.root, config.GraphFile))
		if err != nil {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: "Error: " + err.Error()}},
				IsError: true,
			}
		} else {
			files := g.Files()
			data, _ := json.MarshalIndent(files, "", "  ")
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: string(data)}},
			}
		}

	case "rune_graph":
		file, _ := params.Arguments["file"].(string)
		g, err := graph.Load(config.SubPath(s.root, config.GraphFile))
		if err != nil {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: "Error: " + err.Error()}},
				IsError: true,
			}
		} else {
			deps := g.Get(file)
			dependents := g.Dependents(file)
			info := map[string]interface{}{
				"file":       file,
				"depends_on": deps,
				"used_by":    dependents,
			}
			data, _ := json.MarshalIndent(info, "", "  ")
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: string(data)}},
			}
		}

	case "rune_summary":
		file, _ := params.Arguments["file"].(string)
		sum := loadSummary(s.root, file)
		if sum == "" {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: "No summary found for: " + file}},
				IsError: true,
			}
		} else {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: sum}},
			}
		}

	case "rune_architecture":
		data, err := os.ReadFile(config.SubPath(s.root, config.ArchitectureFile))
		if err != nil {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: "No architecture summary found. Run 'rune index'."}},
				IsError: true,
			}
		} else {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: string(data)}},
			}
		}

	case "rune_conventions":
		data, err := os.ReadFile(config.SubPath(s.root, config.ConventionsFile))
		if err != nil {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: "No conventions file found."}},
				IsError: true,
			}
		} else {
			result = toolResult{
				Content: []contentItem{{Type: "text", Text: string(data)}},
			}
		}

	case "rune_doctor":
		issues := doctor.Check(s.root)
		data, _ := json.MarshalIndent(issues, "", "  ")
		result = toolResult{
			Content: []contentItem{{Type: "text", Text: string(data)}},
		}

	default:
		result = toolResult{
			Content: []contentItem{{Type: "text", Text: "Unknown tool: " + params.Name}},
			IsError: true,
		}
	}

	return &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *Server) writeResponse(w io.Writer, resp *jsonrpcResponse) {
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "%s\n", data)
}

func (s *Server) writeError(w io.Writer, id interface{}, code int, msg string) {
	resp := &jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	}
	s.writeResponse(w, resp)
}

func loadSummary(root, file string) string {
	// Reuse the context package's summary loading
	safeName := file
	for _, c := range []string{"/"} {
		safeName = replaceAll(safeName, c, "__")
	}
	ext := ""
	for i := len(safeName) - 1; i >= 0; i-- {
		if safeName[i] == '.' {
			ext = safeName[i:]
			break
		}
	}
	if ext != "" {
		safeName = safeName[:len(safeName)-len(ext)]
	}
	safeName += ".md"

	path := config.SubPath(root, config.FilesDir) + "/" + safeName
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}
