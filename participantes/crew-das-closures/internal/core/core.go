package core

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/dyammarcano/crew-das-closures/internal/client/openrouter"
	"github.com/dyammarcano/crew-das-closures/internal/model"
	"github.com/dyammarcano/crew-das-closures/internal/prompt"
)

type Core struct {
	*openrouter.Client
	*prompt.PromptManager
}

func NewCore(urlStr string, opts openrouter.Option) (*Core, error) {
	return &Core{
		Client:        openrouter.NewClient(urlStr, opts),
		PromptManager: prompt.NewPromptManager(),
	}, nil
}

// AskQuestion decodes the request, prepares a (mock) service response, and analyzes
// coherence issues between input and output. Diagnostics are returned to help
// detect problems early when integrating with external APIs.
func (c *Core) AskQuestion(question []byte) (*model.FindServiceResponse, error) {
	obj := &model.FindServiceRequest{}
	if err := json.Unmarshal(question, obj); err != nil {
		return nil, err
	}

	// montar o prompt para o OpenRouter com base no obj.Intent
	result, err := c.PromptManager.GenerateModelSpecificPrompt(obj.Intent)
	if err != nil {
		return nil, err
	}

	var msgs []openrouter.Message

	msg := openrouter.Message{
		Role:    "user",
		Content: result,
	}

	msgs = append(msgs, msg)

	oRequest := &openrouter.OpenRouterRequest{
		Model:    c.PromptManager.GetModelName(),
		Messages: msgs,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// chamar o OpenRouter e obter a resposta
	response, err := c.Client.ChatCompletion(ctx, oRequest)
	if err != nil {
		return nil, err
	}

	sData := &model.ServiceData{
		ServiceID:   response.ServiceID,
		ServiceName: response.ServiceName,
	}

	diagnostics := analyzeCoherence(obj, sData)

	if len(diagnostics) > 0 {
		// logar os diagnósticos para análise futura
		for _, diag := range diagnostics {
			log.Printf("Coherence issue detected: %s", diag)
		}
	}

	return &model.FindServiceResponse{
		Success:     len(diagnostics) == 0,
		Data:        sData,
		Diagnostics: diagnostics,
	}, nil
}

// analyzeCoherence performs lightweight checks to surface coherence issues
// between the incoming request and the produced service data. This is useful
// when consuming web/API outputs where format/content can drift.
func analyzeCoherence(req *model.FindServiceRequest, data *model.ServiceData) []string {
	issues := make([]string, 0, 4)
	if req == nil {
		issues = append(issues, "request is nil")
		return issues
	}

	if req.Intent == "" {
		issues = append(issues, "request.intent is empty")
	}

	if data == nil {
		issues = append(issues, "response data is nil")
		return issues
	}

	if data.ServiceID <= 0 {
		issues = append(issues, "response.service_id must be > 0")
	}

	if data.ServiceName == "" {
		issues = append(issues, "response.service_name is empty")
	}

	// Basic semantic check: ensure the intent appears related (very naive heuristic)
	// This is intentionally simple to avoid heavy dependencies.
	if req.Intent != "" && data.ServiceName != "" {
		lowerIntent := strings.ToLower(req.Intent)
		lowerName := strings.ToLower(data.ServiceName)
		if !strings.Contains(lowerIntent, "segur") && strings.Contains(lowerName, "segur") {
			issues = append(issues, "potential mismatch: intent may not relate to returned service name")
		}
	}

	return issues
}
