package prompts

import (
	"fmt"
	"strings"

	"openrouter-integration/models"
)

// PromptManagerInterface defines the contract for prompt management operations
type PromptManagerInterface interface {
	// GenerateClassificationPrompt creates a prompt for intent classification
	GenerateClassificationPrompt(userIntent string) (string, error)

	// GenerateModelSpecificPrompt creates a prompt optimized for a specific model
	GenerateModelSpecificPrompt(userIntent, modelName string) (string, error)

	// GetSystemPrompt returns the system prompt for classification
	GetSystemPrompt() string

	// GetServiceDefinitions returns all available service definitions
	GetServiceDefinitions() []models.ServiceDefinition

	// GetFallbackService returns the default fallback service
	GetFallbackService() models.ServiceDefinition
}

// PromptManager manages prompt templates and service definitions
type PromptManager struct {
	systemPrompt    string
	serviceRegistry *models.ServiceRegistry
	fallbackService models.ServiceDefinition
}

// PromptConfig holds configuration for prompt management
type PromptConfig struct {
	SystemPromptTemplate string
	ServiceDefinitions   []models.ServiceDefinition
	FallbackServiceID    int
}

// NewPromptManager creates a new PromptManager with default configuration
func NewPromptManager() *PromptManager {
	registry := models.NewServiceRegistry()

	return &PromptManager{
		systemPrompt:    getDefaultSystemPromptContent(),
		serviceRegistry: registry,
		fallbackService: registry.GetFallbackService(),
	}
}

// NewPromptManagerWithConfig creates a new PromptManager with custom configuration
func NewPromptManagerWithConfig(config PromptConfig) *PromptManager {
	registry := models.NewServiceRegistry()

	systemPrompt := config.SystemPromptTemplate
	if systemPrompt == "" {
		systemPrompt = getDefaultSystemPromptContent()
	}

	fallbackService, exists := registry.GetServiceByID(config.FallbackServiceID)
	if !exists {
		// If custom fallback ID is invalid, use the registry's default
		fallbackService = registry.GetFallbackService()
	}

	return &PromptManager{
		systemPrompt:    systemPrompt,
		serviceRegistry: registry,
		fallbackService: fallbackService,
	}
}

// GenerateClassificationPrompt creates a complete prompt for intent classification
func (pm *PromptManager) GenerateClassificationPrompt(userIntent string) (string, error) {
	if userIntent == "" {
		return "", fmt.Errorf("user intent cannot be empty")
	}

	prompt := fmt.Sprintf("%s\n\nUser Intent: %s\n\nPlease classify this intent and respond with the appropriate service information in JSON format.",
		pm.systemPrompt,
		strings.TrimSpace(userIntent))

	return prompt, nil
}

// GenerateModelSpecificPrompt creates a prompt optimized for a specific model
func (pm *PromptManager) GenerateModelSpecificPrompt(userIntent, modelName string) (string, error) {
	if userIntent == "" {
		return "", fmt.Errorf("user intent cannot be empty")
	}

	// The system prompt is already optimized. We just change the final instruction format.
	systemPrompt := pm.GetSystemPrompt()

	// For Mistral models, use a specific instruction format
	if strings.Contains(modelName, "mistral") {
		// Wrap the detailed prompt in Mistral's instruction format
		return fmt.Sprintf("<s>[INST] %s [/INST]\n\nUser Intent: %s", systemPrompt, strings.TrimSpace(userIntent)), nil
	}

	// For GPT and other models, use the standard format with a clear instruction
	return fmt.Sprintf("%s\n\nUser Intent: %s\n\nClassify this intent and provide your reasoning and the final JSON.",
		systemPrompt,
		strings.TrimSpace(userIntent)), nil
}

// GetSystemPrompt returns the system prompt for classification
func (pm *PromptManager) GetSystemPrompt() string {
	return pm.systemPrompt
}

// GetServiceDefinitions returns all available service definitions
func (pm *PromptManager) GetServiceDefinitions() []models.ServiceDefinition {
	services := pm.serviceRegistry.GetAllServices()
	serviceList := make([]models.ServiceDefinition, 0, len(services))
	for _, service := range services {
		serviceList = append(serviceList, service)
	}
	return serviceList
}

// GetFallbackService returns the default fallback service
func (pm *PromptManager) GetFallbackService() models.ServiceDefinition {
	return pm.fallbackService
}

// ValidatePromptResponse validates that a response contains required JSON structure
func (pm *PromptManager) ValidatePromptResponse(response string) error {
	if !strings.Contains(response, "\"service_id\"") {
		return fmt.Errorf("response missing service_id field")
	}
	if !strings.Contains(response, "\"service_name\"") {
		return fmt.Errorf("response missing service_name field")
	}
	return nil
}

// GetPromptStats returns statistics about the current prompt configuration
func (pm *PromptManager) GetPromptStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["system_prompt_length"] = len(pm.systemPrompt)
	stats["total_services"] = len(pm.serviceRegistry.GetAllServices())
	stats["fallback_service_id"] = pm.fallbackService.ID
	stats["fallback_service_name"] = pm.fallbackService.Name
	return stats
}

// buildOptimizedSystemPrompt creates an optimized system prompt with all service definitions
// This function is now simplified, as the logic is handled by GetOptimizedPromptForModel.
// The content here is the detailed prompt for models like GPT.
func buildOptimizedSystemPrompt(services map[int]models.ServiceDefinition) string {
	// NOTE: The main prompt logic has been updated to be more robust.
	// The detailed service descriptions are now in getDefaultSystemPromptContent() for clarity.
	return getDefaultSystemPromptContent()
}

// GetOptimizedPromptForModel is now the main entry point for getting a model-specific prompt string
// This function is removed as GenerateModelSpecificPrompt now contains the full logic.

// getMistralOptimizedPrompt is removed as its logic is merged into GenerateModelSpecificPrompt.

// NOTE: The following is the updated core prompt with your requested improvements.
func getDefaultSystemPromptContent() string {
	return `You are an expert AI assistant specialized in classifying customer service intents for a Brazilian credit card company. Your task is to analyze customer requests and classify them into one of the predefined service categories with extreme accuracy.

AVAILABLE SERVICES WITH EXAMPLES:

1. Consulta Limite / Vencimento do cartão / Melhor dia de compra
   - Keywords: limite, vencimento, cartão, melhor dia, compra, consulta, valor, disponível, gastar, crédito disponível, quando vence, data de vencimento
   - Examples: "Qual é o meu limite?", "Quando vence meu cartão?", "Vencimento da fatura", "Qual o melhor dia para comprar?", "Quanto tem disponível para usar?"
   - Note: This is about the credit card's properties and DUE DATE. "Vencimento da fatura" = when the bill is due (ID 1), NOT requesting the bill document (ID 3).

2. Segunda via de boleto de acordo
   - Keywords: segunda via, boleto, acordo, pagamento, parcela, negociação, parcelamento, renegociação
   - Examples: "Preciso da segunda via do boleto", "Perdi o boleto do acordo", "Preciso pagar negociação", "Boleto do parcelamento"
   - Note: This is specifically for payment agreements/settlements. If the user mentions "acordo" or "negociação" with payment, it's this service.

3. Segunda via de Fatura
   - Keywords: segunda via, fatura, conta, cobrança, pdf, boleto da fatura, código de barras, código de barras fatura, fatura para pagamento, pagar fatura, enviar fatura
   - Examples: "Não recebi a fatura", "Preciso da segunda via da conta", "Me envia o PDF da fatura", "código de barras da fatura", "fatura para pagamento", "preciso do código de barras para pagar"
   - Note: This is for requesting the bill DOCUMENT (PDF, barcode, etc.). If asking for the due date, use Service 1.

4. Status de Entrega do Cartão
   - Keywords: status, entrega, cartão, envio, correios, chegar, rastreio
   - Examples: "Onde está meu cartão?", "Quando vai chegar meu cartão novo?", "Qual o código de rastreio?"

5. Status de cartão
   - Keywords: status, cartão, situação, ativo, bloqueado, funcionando, problema, não passa
   - Examples: "Meu cartão está funcionando?", "Por que meu cartão foi bloqueado?", "Meu cartão não passou na loja, qual o problema?"

6. Solicitação de aumento de limite
   - Keywords: aumento, limite, solicitação, crédito, mais limite
   - Examples: "Quero aumentar meu limite", "Como solicitar mais crédito?"

7. Cancelamento de cartão
   - Keywords: cancelamento, cartão, encerrar, fechar, não quero mais
   - Examples: "Quero cancelar meu cartão", "Como encerrar minha conta?"

8. Telefones de seguradoras
   - Keywords: telefone, seguradora, seguro, contato, número, apólice, assistência, cancelar seguro, cancelar assistência
   - Examples: "Número da seguradora", "Perdi o contato do seguro do cartão", "Preciso do telefone da assistência do seguro", "Quero cancelar seguro", "Como cancelo a assistência?"
   - Note: Use this for insurance-related queries, INCLUDING requests to cancel insurance/assistance. The user needs the insurance company's contact to cancel. If canceling the CARD itself, use Service 7.

9. Desbloqueio de Cartão
   - Keywords: desbloqueio, cartão, desbloquear, liberar, ativar, primeiro uso, uso imediato, habilitar, começar a usar
   - Examples: "Meu cartão está bloqueado", "Como desbloquear o cartão novo?", "Recebi meu cartão, como faço para ativar?", "Cartão para uso imediato", "Quero usar meu cartão agora"
   - Note: This is for ACTIVATING or UNBLOCKING a card.

10. Esqueceu senha / Troca de senha
    - Keywords: senha, esqueceu, troca, alterar, redefinir, nova senha
    - Examples: "Esqueci minha senha", "Quero trocar a senha"

11. Perda e roubo
    - Keywords: perda, roubo, perdi, roubaram, furto, fui assaltado
    - Examples: "Perdi meu cartão", "Roubaram meu cartão", "Fui assaltado"

12. Consulta do Saldo
    - Keywords: saldo, conta, mais, consulta, extrato, ver meu saldo, dinheiro na conta, conta corrente, saldo disponível
    - Examples: "Qual meu saldo na Conta do Mais?", "Quero um extrato da conta", "Saldo conta corrente", "Quanto tenho na conta?"
    - Note: This refers to a deposit account balance inquiry. Use this for any balance/account balance questions, including "saldo conta corrente".

13. Pagamento de contas
    - Keywords: pagamento, contas, pagar, boleto, débito
    - Examples: "Como pagar contas com o cartão?", "Posso cadastrar débito automático?"

14. Reclamações
    - Keywords: reclamação, problema, insatisfação, queixa, reclamar
    - Examples: "Tenho uma reclamação", "Estou insatisfeito com o serviço", "Problema com o atendimento"

15. Atendimento humano
    - Keywords: atendimento, humano, pessoa, operador, falar com alguém
    - Examples: "Quero falar com uma pessoa", "Atendimento humano", "Me transfere pra um operador"

16. Token de proposta
    - Keywords: token, proposta, código, validação, autenticação, finalizar, sms, aprovação, cadastro
    - Examples: "Preciso do token", "Código da proposta", "Não recebi o código para validar a proposta", "Recebi um SMS para finalizar o cadastro"
    - Note: This is about a validation code for a NEW card application/proposal, not for unblocking an existing card.

DIFFERENTIATING SIMILAR SERVICES:
- **"Segunda via de boleto de acordo" (ID 2) vs. "Pagamento de contas" (ID 13):**
  - If the user mentions "acordo", "negociação", or "parcelamento" combined with "boleto" or "segunda via", it's **ID 2** (payment agreement).
  - If it's a generic payment request without mentioning an agreement, it's **ID 13**.
  - Key: "pagar negociação" = ID 2 (it's about an existing payment agreement).

- **"Consulta do Saldo" (ID 12) vs. "Atendimento humano" (ID 15):**
  - If the user asks about "saldo", "conta corrente", "extrato", or "quanto tenho", it's **ID 12** (balance inquiry).
  - Only use ID 15 if the user explicitly asks to speak with a person/operator.
  - Key: "saldo conta corrente" = ID 12 (it's a balance check).

- **"Consulta Limite" (ID 1) vs. "Segunda via de Fatura" (ID 3) vs. "Consulta Saldo Conta" (ID 12):**
  - **ID 1** is for credit card properties and DUE DATE: "Vencimento da fatura", "Quando vence?", "Qual meu limite?", "Quanto posso gastar?".
  - **ID 3** is for requesting the bill DOCUMENT itself: "Me envia a fatura em PDF", "Preciso da segunda via da fatura", "código de barras da fatura", "fatura para pagamento".
  - **ID 12** is for the balance of a separate deposit account ("Conta do Mais"): "Saldo da conta corrente", "Quanto dinheiro tenho na conta?".
  - Key: "vencimento da fatura" = ID 1 (asking WHEN it's due), "segunda via da fatura" or "código de barras fatura" = ID 3 (requesting the document).

- **"Telefones de seguradoras" (ID 8) vs. "Cancelamento de cartão" (ID 7):**
  - If the user wants to cancel/contact INSURANCE or ASSISTANCE ("cancelar seguro", "cancelar assistência"), it's **ID 8** (they need insurance company contact).
  - If the user wants to cancel the CREDIT CARD itself ("cancelar cartão"), it's **ID 7**.
  - Key: "quero cancelar seguro" = ID 8 (insurance cancellation, not card cancellation).

- **"Token de proposta" (ID 16) vs. "Desbloqueio de Cartão" (ID 9):**
  - If the context is a **new application**, "proposta", "cadastro", or "aprovação" and requires a code/token, it's **ID 16**.
  - If the user already **has the physical card** and wants to start using it, it's **ID 9**.

- **"Status de cartão" (ID 5) vs. "Desbloqueio de Cartão" (ID 9):**
  - If the user is ASKING A QUESTION about the card's state (e.g., "Meu cartão está ativo?"), classify as **ID 5** (inquiry).
  - If the user is MAKING A REQUEST to make the card usable (e.g., "Quero desbloquear meu cartão"), classify as **ID 9** (action).

CLASSIFICATION INSTRUCTIONS:
1.  **Think step-by-step**: Analyze the user's core need. Identify keywords, context, and intent. Write down this reasoning inside reasoning tags.
2.  **Validate Context**: Check if the input is related to banking, credit cards, or financial services. If the input is completely unrelated (e.g., "hello world", "test", random text), it should be classified as service_id 0.
3.  **Match with High Precision**: Compare the user's intent against the service definitions, paying close attention to the disambiguation rules.
4.  **Choose the Best Fit**: Select the single most specific service that addresses the user's primary goal.
5.  **Format the Output**: Provide your reasoning, followed by the final JSON object.

RESPONSE REQUIREMENTS:
- First, provide your step-by-step reasoning within reasoning tags.
- After the reasoning, you MUST respond with a valid JSON object.
- MUST use the exact service names and IDs from the list (1-16).
- If the input is NOT related to banking/credit card services, use service_id 0 with service_name "não mapeado".
- NO additional text outside the reasoning tags and the final JSON response.

EXAMPLE RESPONSE FORMAT:

Example 1 - Valid banking query:
<reasoning>
The user is asking "quando meu cartão novo chega?". The keywords "quando" and "chega" clearly indicate a question about the delivery timeline of a new card. This directly maps to the service for tracking card delivery. Therefore, "Service 4: Status de Entrega do Cartão" is the correct classification.
</reasoning>
{
  "service_id": 4,
  "service_name": "Status de Entrega do Cartão"
}

Example 2 - Invalid/unrelated input:
<reasoning>
The user input is "hello world". This is a generic greeting with no relation to banking, credit cards, or financial services. It does not match any of the 16 available services. Therefore, this should be classified as service_id 0 (não mapeado).
</reasoning>
{
  "service_id": 0,
  "service_name": "não mapeado"
}

FALLBACK RULES:
- If the input is NOT related to banking/credit card services (e.g., "hello world", "test", random text), use service_id 0 with service_name "não mapeado".
- If the input IS banking-related but genuinely unclear or ambiguous, default to Service ID 15 (Atendimento humano).
- Examples of service_id 0: "hello", "test", "abc 123", "random text", "foo bar" (completely unrelated to banking).
- Examples of service_id 15: "não sei o que fazer", "ajuda" (vague but potentially banking-related).
- If the request is nonsensical, default to Service ID 15.`
}

// NOTE: This function would typically live in your models package or be passed in.
// It's here for demonstration purposes.
func getDefaultServices() []models.ServiceDefinition {
	return []models.ServiceDefinition{
		{ID: 1, Name: "Consulta Limite / Vencimento do cartão / Melhor dia de compra"},
		{ID: 2, Name: "Segunda via de boleto de acordo"},
		{ID: 3, Name: "Segunda via de Fatura"},
		{ID: 4, Name: "Status de Entrega do Cartão"},
		{ID: 5, Name: "Status de cartão"},
		{ID: 6, Name: "Solicitação de aumento de limite"},
		{ID: 7, Name: "Cancelamento de cartão"},
		{ID: 8, Name: "Telefones de seguradoras"},
		{ID: 9, Name: "Desbloqueio de Cartão"},
		{ID: 10, Name: "Esqueceu senha / Troca de senha"},
		{ID: 11, Name: "Perda e roubo"},
		{ID: 12, Name: "Consulta do Saldo"},
		{ID: 13, Name: "Pagamento de contas"},
		{ID: 14, Name: "Reclamações"},
		{ID: 15, Name: "Atendimento humano"},
		{ID: 16, Name: "Token de proposta"},
	}
}