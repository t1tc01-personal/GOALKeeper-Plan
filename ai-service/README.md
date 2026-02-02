# ai-agent-service-demo

AI Agent Service with multi-provider LLM support via LiteLLM. Supports Azure OpenAI, OpenAI, Anthropic, and 100+ other providers.

## Quickstart

1) Install dependencies:
```bash
pip install -r requirements.txt
```

2) Configure your LLM provider:

### Option A: Azure OpenAI (New Provider-Agnostic Config)
```bash
export LLM_MODEL="azure/<your-deployment-name>"
export LLM_API_KEY="<your-azure-api-key>"
export LLM_API_BASE="https://<your-resource>.cognitiveservices.azure.com/"
export LLM_API_VERSION="2024-12-01-preview"
```

### Option B: Azure OpenAI (Legacy Config - Still Supported)
```bash
export AZURE_CLIENT_ID="<AZURE_CLIENT_ID>"
export AZURE_TENANT_ID="<AZURE_TENANT_ID>"
export AZURE_CLIENT_SECRET="<AZURE_CLIENT_SECRET>"

export OPENAI_API_VERSION="2024-12-01-preview"
export OPENAI_AZURE_ENDPOINT="https://<your-resource>.cognitiveservices.azure.com/"
export OPENAI_ENGINE="<your-deployment-name>"
export OPENAI_MODEL_NAME="<your-model-name>"
```

For API key authentication instead of Azure AD:
```bash
export OPENAI_USE_KEY=true
export OPENAI_API_KEY="<api-key>"
```

### Option C: OpenAI
```bash
export LLM_MODEL="gpt-4"
export LLM_API_KEY="<your-openai-api-key>"
```

### Option D: Anthropic Claude
```bash
export LLM_MODEL="anthropic/claude-3-5-sonnet-20241022"
export LLM_API_KEY="<your-anthropic-api-key>"
```

### Option E: Other Providers
LiteLLM supports 100+ providers. See [LiteLLM documentation](https://docs.litellm.ai/docs/providers) for model string formats.

3) Run the API:
```bash
uvicorn app:app --host 0.0.0.0 --port 8000 --reload
```

4) Call the chat endpoint:
```bash
curl -X POST http://localhost:8000/api/v1/chat/ \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "I am going to Paris, what should I see?"}
    ],
    "temperature": 0.7,
    "max_completion_tokens": 512
  }'
```

5) Stream chat responses:
```bash
curl -X POST http://localhost:8000/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "Tell me a story"}
    ],
    "temperature": 0.7,
    "max_completion_tokens": 1024
  }'
```

## API Documentation

Swagger docs: `http://localhost:8000/docs`

## Features

- Multi-provider LLM support via LiteLLM (Azure OpenAI, OpenAI, Anthropic, and 100+ more)
- Streaming and non-streaming chat completions
- Provider-agnostic configuration
- Backward compatible with existing Azure OpenAI configurations
- Server-Sent Events (SSE) for streaming responses