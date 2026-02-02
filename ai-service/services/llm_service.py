import os
from opik import track, Opik
from litellm import completion
from typing import AsyncGenerator, Dict, Any

class LLMService:
    def __init__(self):
        self.opik = Opik(
            api_key=os.getenv("OPIK_API_KEY"),
            workspace=os.getenv("OPIK_WORKSPACE", "goalkeeper")
        )
    
    @track(project_name="block-generation")
    async def generate_blocks(
        self,
        prompt: str,
        context: str = "",
        model_tier: str = "default",
        workspace_id: str = None
    ) -> AsyncGenerator[str, None]:
        
        system_prompt = self._build_system_prompt()
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": f"Context:\n{context}\n\nPrompt: {prompt}"}
        ]
        
        # Default = Proxy routing.
        import litellm
        litellm.api_base = os.getenv("LITELLM_PROXY_URL", "http://litellm-proxy:4000")
        litellm.api_key = os.getenv("LITELLM_MASTER_KEY", "sk-1234")

        response = await completion(
            model=self._get_model(model_tier),
            messages=messages,
            stream=True,
            metadata={
                "workspace_id": workspace_id,
                "context_length": len(context),
                "model_tier": model_tier
            }
        )
        
        full_response = ""
        async for chunk in response:
            if chunk and chunk.choices and chunk.choices[0].delta.content:
                content = chunk.choices[0].delta.content
                full_response += content
                yield content
    
    def _get_model(self, model_tier: str) -> str:
        # default = Google Gemini (free). Fallbacks: OpenAI, Azure via LiteLLM.
        # These aliases must match what is defined in litellm_config.yaml (router_settings)
        tier_map = {
            "default": "openai/default", # Maps to gemini-flash, gemini-pro via proxy
            "smart": "openai/smart",     # Maps to gemini-pro, gpt-4-turbo, azure-gpt-4 via proxy
            "fast": "openai/fast",       # Maps to gemini-flash, gpt-3.5-turbo, azure-gpt-35 via proxy
        }
        return tier_map.get(model_tier, "default")

    def _build_system_prompt(self) -> str:
        return """You are a helpful AI assistant that generates structured content blocks.
You must output clean, valid JSON when requested, or helpful text otherwise.
When generating blocks, follow the system's block schema (headings, paragraphs, kanban, etc.)."""
