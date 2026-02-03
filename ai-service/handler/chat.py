from typing import Any, AsyncIterator, Dict, List, Literal, Optional

from fastapi import HTTPException
from fastapi.responses import StreamingResponse

from loguru import logger
from pydantic import BaseModel, Field, model_validator


class ChatMessage(BaseModel):
    """Single chat message."""

    role: Literal["system", "user", "assistant"]
    content: str = Field(min_length=1, description="Message content")


class ChatRequest(BaseModel):
    """Chat completion request payload."""

    messages: List[ChatMessage]
    temperature: float = Field(
        default=1.0,
        ge=0.0,
        le=2.0,
        description="Sampling temperature (0.0 to 2.0)"
    )
    max_completion_tokens: Optional[int] = Field(
        default=8192,
        gt=0,
        le=16384,
        description=(
            "Max completion tokens for the response. "
            "Note: Reasoning models may use tokens for internal reasoning, "
            "so higher values (2048+) are recommended."
        ),
    )
    model: Optional[str] = Field(
        default="azure/gpt-5-nano",
        description="Model to use for chat completion (e.g., 'azure/gpt-5-nano', 'gemini/gemini-3-flash-preview')"
    )

    @model_validator(mode="after")
    def validate_messages(self) -> "ChatRequest":
        if not self.messages:
            raise ValueError("messages cannot be empty")
        if not any(msg.role == "user" for msg in self.messages):
            raise ValueError("messages must contain at least one user message")
        return self


class ChatResponse(BaseModel):
    """Chat completion response body."""

    reply: str
    model: str
    reasoning: Optional[str] = Field(default=None, description="Model reasoning/thought process")


class ModelInfo(BaseModel):
    """Information about an available model."""
    name: str # The alias used in API (smart, fast, etc.)
    full_name: str # The real model ID (azure/gpt-5-nano, etc.)
    provider: str
    supports_thinking: bool = False


class ModelListResponse(BaseModel):
    """List of available models."""
    models: List[ModelInfo]


class ChatHandler:
    """Handle chat completion requests."""

    _default_system_prompt = "You are a helpful assistant."
    _max_history_messages = 6  # Keep last N messages (reduced for safety)
    _max_message_length = 1500  # Max chars per message (reduced)

    def __init__(self) -> None:
        """Initialize ChatHandler with LiteLLM configuration."""
        import litellm
        import os
        
        # Configure LiteLLM to send requests to our Proxy (LiteLLM Proxy)
        # Defaults to Docker internal name, but falls back to .env/localhost
        self.api_base = os.getenv("LITELLM_PROXY_URL", "http://litellm-proxy:4000")
        self.api_key = os.getenv("LITELLM_MASTER_KEY", "sk-1234")
        
        # Globally set litellm params to avoid per-call overhead
        litellm.api_base = self.api_base
        litellm.api_key = self.api_key
        
        logger.info(f"ChatHandler initialized using LiteLLM Proxy at: {self.api_base}")

    @staticmethod
    def _truncate_conversation_history(
        messages: List[ChatMessage],
        max_messages: int = None,
        max_message_length: int = None
    ) -> List[ChatMessage]:
        """
        Truncate conversation history to prevent token limit issues.

        Strategy:
        1. Keep all system messages
        2. Keep the most recent N messages (user + assistant pairs)
        3. Truncate individual messages if too long

        Args:
            messages: List of chat messages
            max_messages: Maximum number of non-system messages to keep
            max_message_length: Maximum characters per message

        Returns:
            Truncated list of messages
        """
        if max_messages is None:
            max_messages = ChatHandler._max_history_messages
        if max_message_length is None:
            max_message_length = ChatHandler._max_message_length

        if len(messages) <= max_messages:
            # No truncation needed, but still truncate long messages
            truncated = []
            for msg in messages:
                content = msg.content
                if len(content) > max_message_length:
                    content = content[:max_message_length] + "..."
                truncated.append(
                    ChatMessage(role=msg.role, content=content)
                )
            return truncated

        # Separate system messages from conversation
        system_messages = [msg for msg in messages if msg.role == "system"]
        conversation_messages = [
            msg for msg in messages if msg.role != "system"
        ]

        # Keep most recent N conversation messages
        recent_messages = conversation_messages[-max_messages:]

        # Truncate long messages
        truncated = []
        for msg in recent_messages:
            content = msg.content
            if len(content) > max_message_length:
                content = content[:max_message_length] + "..."
            truncated.append(
                ChatMessage(role=msg.role, content=content)
            )

        # Prepend system messages
        result = system_messages + truncated

        logger.debug(
            f"Truncated conversation: {len(messages)} -> {len(result)} "
            f"messages (kept {len(system_messages)} system, "
            f"{len(recent_messages)} recent)"
        )

        return result

    @staticmethod
    def _to_openai_messages(
        messages: List[ChatMessage]
    ) -> List[Dict[str, str]]:
        """Convert ChatMessage list to OpenAI-compatible message format."""
        has_system = any(msg.role == "system" for msg in messages)
        converted: List[Dict[str, str]] = []
        
        if not has_system:
            converted.append({
                "role": "system",
                "content": ChatHandler._default_system_prompt
            })

        for msg in messages:
            converted.append({
                "role": msg.role,
                "content": msg.content
            })
        return converted

    @staticmethod
    def _extract_response_metadata(response: Any) -> Dict[str, Any]:
        """Extract finish_reason and usage_metadata from LiteLLM response."""
        finish_reason = None
        usage_metadata: Dict[str, Any] = {}
        
        if hasattr(response, "choices") and response.choices:
            choice = response.choices[0]
            if hasattr(choice, "finish_reason"):
                finish_reason = choice.finish_reason
        
        if hasattr(response, "usage"):
            usage = response.usage
            usage_metadata = {
                "prompt_tokens": getattr(usage, "prompt_tokens", 0),
                "completion_tokens": getattr(usage, "completion_tokens", 0),
                "total_tokens": getattr(usage, "total_tokens", 0),
                "output_token_details": getattr(usage, "output_token_details", {})
            }
        
        return {
            "finish_reason": finish_reason,
            "usage_metadata": usage_metadata,
        }

    async def get_models(self) -> ModelListResponse:
        """Get list of available models with provider info from litellm_config.yaml."""
        import yaml
        import os
        
        config_path = os.getenv("LITELLM_CONFIG_PATH", "litellm_config.yaml")
        model_infos = []
        seen_names = set()
        
        try:
            if os.path.exists(config_path):
                with open(config_path, "r") as f:
                    config = yaml.safe_load(f)
                    model_list = config.get("model_list", [])
                    for m in model_list:
                        name = m.get("model_name")
                        if name and name not in seen_names:
                            # Try to extract provider from 'model' param (e.g., 'azure/gpt-4' -> 'Azure')
                            litellm_params = m.get("litellm_params", {})
                            model_str = litellm_params.get("model", "")
                            provider = "unknown"
                            
                            if "/" in model_str:
                                raw_provider = model_str.split("/")[0].lower()
                                # Prettify provider name
                                provider_map = {
                                    "azure": "Azure OpenAI (GPT)",
                                    "gemini": "Google Gemini",
                                    "openai": "OpenAI",
                                    "anthropic": "Anthropic Claude",
                                    "vertex_ai": "Google Vertex AI"
                                }
                                provider = provider_map.get(raw_provider, raw_provider.capitalize())
                            
                            # Extract metadata
                            metadata = m.get("metadata", {})
                            supports_thinking = metadata.get("supports_thinking", False)
                            
                            model_infos.append(ModelInfo(
                                name=name,
                                full_name=model_str,
                                provider=provider,
                                supports_thinking=supports_thinking
                            ))
                            seen_names.add(name)
            
            # Fallback if config is empty or failed
            if not model_infos:
                model_infos = [
                    ModelInfo(name="azure/gpt-5-nano", full_name="azure/gpt-5-nano", provider="Azure OpenAI (GPT)", supports_thinking=False),
                    ModelInfo(name="gemini/gemini-3-flash-preview", full_name="gemini/gemini-3-flash-preview", provider="Google Gemini", supports_thinking=False)
                ]
                
        except Exception as e:
            logger.error(f"Failed to load model list: {e}")
            model_infos = [ModelInfo(name="error", full_name="error", provider=str(e), supports_thinking=False)]
            
        return ModelListResponse(models=model_infos)

    def _adapt_payload(self, model: str, payload: ChatRequest) -> Dict[str, Any]:
        """Adapt payload based on model capabilities (Thinking, O1 roles, etc.)"""
        is_thinking = "reasoner" in model or "thinking" in model or "o1" in model or "o3" in model
        
        # 1. Manage Max Tokens
        # Reasoning models need much more tokens as it includes thinking process
        if is_thinking and (payload.max_completion_tokens is None or payload.max_completion_tokens < 16384):
             # Don't override if user explicitly set a very high value
             logger.info(f"Boosting max_completion_tokens for thinking model: {model}")
             payload.max_completion_tokens = 16384

        # 2. Extract OpenAI messages
        openai_messages = self._to_openai_messages(payload.messages)

        # 3. Handle O1/O3 role limitations
        # OpenAI O1/O3 models don't support 'system' role in the usual way
        if "o1-" in model or "o3-" in model:
            logger.debug(f"Adapting roles for OpenAI reasoning model: {model}")
            new_messages = []
            for msg in openai_messages:
                if msg["role"] == "system":
                    # Convert system to 'developer' if supported or 'user' with prefix
                    # LiteLLM handles 'developer' role for O1 usually, 
                    # but some environments still prefer converting to user.
                    new_messages.append({"role": "user", "content": f"[SYSTEM INSTRUCTION]\n{msg['content']}"})
                else:
                    new_messages.append(msg)
            openai_messages = new_messages

        return {
            "model": f"openai/{model}", # Always route via proxy
            "messages": openai_messages,
            "temperature": payload.temperature,
            "max_tokens": payload.max_completion_tokens,
            "api_base": self.api_base,
            "api_key": self.api_key
        }

    def _extract_content_from_chunk(self, chunk: Any) -> tuple[str, Optional[str]]:
        """Extract content and reasoning from a LiteLLM stream chunk."""
        content = ""
        reasoning = None
        
        # LiteLLM returns chunks in OpenAI format
        if hasattr(chunk, "choices") and chunk.choices:
            delta = getattr(chunk.choices[0], "delta", None)
            if delta:
                if hasattr(delta, "reasoning_content") and delta.reasoning_content:
                    reasoning = delta.reasoning_content
                
                if hasattr(delta, "content") and delta.content:
                    content = delta.content
                elif hasattr(delta, "text") and delta.text:
                    content = delta.text
        elif hasattr(chunk, "reasoning_content") and chunk.reasoning_content:
            reasoning = chunk.reasoning_content
        elif hasattr(chunk, "content"):
            content = chunk.content or ""
        elif hasattr(chunk, "text"):
            content = chunk.text or ""
        elif isinstance(chunk, str):
            content = chunk
        elif isinstance(chunk, dict):
            # Handle dict format chunks
            if "choices" in chunk and chunk["choices"]:
                delta = chunk["choices"][0].get("delta", {})
                content = delta.get("content", "") or delta.get("text", "")
                reasoning = delta.get("reasoning_content")
            elif "content" in chunk:
                content = chunk["content"] or ""
        
        return content, reasoning

    async def chat(self, payload: ChatRequest) -> ChatResponse:
        """Process chat completion request."""
        # Truncate conversation history if needed
        original_count = len(payload.messages)
        truncated_messages = self._truncate_conversation_history(
            payload.messages
        )

        if len(truncated_messages) < original_count:
            logger.info(
                f"Truncated conversation history: "
                f"{original_count} -> {len(truncated_messages)} messages"
            )
            # Create new payload with truncated messages
            payload = ChatRequest(
                messages=truncated_messages,
                temperature=payload.temperature,
                max_completion_tokens=payload.max_completion_tokens,
            )

        # Convert messages to OpenAI format
        openai_messages = self._to_openai_messages(payload.messages)

        response: Any = None
        reply = ""
        reasoning = None

        try:
            import litellm
            
            # Use adaptive logic to prepare request
            requested_model = payload.model or "azure/gpt-5-nano"
            completion_kwargs = self._adapt_payload(requested_model, payload)

            response = await litellm.acompletion(**completion_kwargs)
            provider_name = response.model if hasattr(response, "model") else "unknown"
            logger.debug(f"Response from {provider_name}, type: {type(response)}")

            # Extract content from LiteLLM response
            if hasattr(response, "choices") and response.choices:
                choice = response.choices[0]
                if hasattr(choice, "message"):
                    message = choice.message
                    if hasattr(message, "content"):
                        reply = message.content.strip() if message.content else ""
                    elif hasattr(message, "text"):
                        reply = message.text.strip() if message.text else ""
                    
                    if hasattr(message, "reasoning_content"):
                        reasoning = message.reasoning_content
                elif hasattr(choice, "text"):
                    reply = choice.text.strip() if choice.text else ""
            
            # Fallback: try to get content directly
            if not reply and hasattr(response, "content"):
                reply = str(response.content).strip()
            
            if not reasoning and hasattr(response, "reasoning_content"):
                reasoning = response.reasoning_content

            # Extract metadata for logging
            metadata = self._extract_response_metadata(response)
            finish_reason = metadata["finish_reason"]
            usage_metadata = metadata["usage_metadata"]

            # If content is empty but we have token usage, log it
            if not reply and usage_metadata:
                output_tokens = usage_metadata.get("completion_tokens", 0)
                if output_tokens > 0:
                    logger.warning(
                        f"Response empty but {output_tokens} completion "
                        f"tokens used. Consider increasing max_completion_tokens."
                    )

        except HTTPException:
            # Re-raise HTTP exceptions as-is
            raise
        except Exception as exc:  # pragma: no cover - runtime protection
            logger.exception(
                f"Chat completion failed: {type(exc).__name__}: {exc}"
            )
            raise HTTPException(
                status_code=500,
                detail=f"Chat completion failed: {str(exc)}"
            ) from exc

        logger.debug(f"Extracted reply length: {len(reply) if reply else 0}")

        # Handle empty response (but allow reasoning-only responses)
        if not reply and not reasoning:
            # Extract metadata if response exists
            if response is not None:
                metadata = self._extract_response_metadata(response)
                finish_reason = metadata["finish_reason"]
                usage_metadata = metadata["usage_metadata"]
            else:
                finish_reason = None
                usage_metadata = {}

            error_detail = "Empty response from model."
            status_code = 502
            if finish_reason == "length":
                reasoning_tokens = usage_metadata.get(
                    "output_token_details", {}
                ).get("reasoning", 0)
                max_tokens = payload.max_completion_tokens or 8192
                status_code = 400
                error_detail = (
                    f"Response was truncated (hit token limit). "
                    f"Used {reasoning_tokens} tokens for reasoning without generating output. "
                    f"Try increasing max_completion_tokens "
                    f"(current: {max_tokens})."
                )

            logger.error(
                f"{error_detail} "
                f"Response type: {type(response) if response else 'None'}, "
                f"Finish reason: {finish_reason}, "
                f"Usage: {usage_metadata}"
            )
            raise HTTPException(
                status_code=status_code,
                detail=error_detail
            )

        # Extract model name from response or client
        model_name = "unknown"
        if hasattr(response, "model"):
            model_name = response.model
        else:
            # MultiProviderClient knows which provider it used
            model_name = provider_name

        return ChatResponse(reply=reply, model=model_name, reasoning=reasoning)

    async def chat_stream(self, payload: ChatRequest) -> StreamingResponse:
        """Process chat completion with Server-Sent Events streaming."""
        # Truncate conversation history if needed
        original_count = len(payload.messages)
        truncated_messages = self._truncate_conversation_history(
            payload.messages
        )

        if len(truncated_messages) < original_count:
            logger.info(
                f"Truncated conversation history: "
                f"{original_count} -> {len(truncated_messages)} messages"
            )
            # Create new payload with truncated messages
            payload = ChatRequest(
                messages=truncated_messages,
                temperature=payload.temperature,
                max_completion_tokens=payload.max_completion_tokens,
            )

        # Estimate input tokens (more conservative: ~3 chars per token)
        # Also account for message overhead (role, formatting, etc.)
        total_chars = sum(len(msg.content) for msg in payload.messages)
        # More conservative estimate: ~3 chars per token + overhead
        # Each message adds ~10 tokens overhead (role, formatting)
        message_overhead = len(payload.messages) * 10
        estimated_input_tokens = (total_chars // 3) + message_overhead

        logger.info(
            f"Starting stream request: messages={len(payload.messages)}, "
            f"temperature={payload.temperature}, "
            f"max_tokens={payload.max_completion_tokens}, "
            f"estimated_input_tokens={estimated_input_tokens}"
        )

        # Warn if input might be too large
        if estimated_input_tokens > 2000:
            logger.warning(
                f"Large input detected: ~{estimated_input_tokens} tokens. "
                f"May leave little room for output with "
                f"max_completion_tokens={payload.max_completion_tokens}"
            )

        # Convert messages to OpenAI format
        openai_messages = self._to_openai_messages(payload.messages)

        async def generate_stream() -> AsyncIterator[str]:
            """Generate SSE stream with optimized batching."""
            chunk_count = 0
            total_chars_sent = 0
            chunks_received = 0
            empty_chunks = 0

            try:
                chunk_buffer = ""
                buffer_size = 0
                max_buffer_size = 30  # Characters before flushing
                sentence_endings = ['.', '!', '?', '\n']

                logger.debug("Starting to stream from LiteLLM...")

                # Stream chunks from LiteLLM Proxy
            
                import litellm
                requested_model = payload.model or "azure/gpt-5-nano"
                completion_kwargs = self._adapt_payload(requested_model, payload)
                completion_kwargs["stream"] = True

                async for chunk in await litellm.acompletion(**completion_kwargs):
                    provider_name = "proxy-model" # Proxy hides actual provider name often, or returns it in model field
                    chunks_received += 1

                    # Extract content using our refined parser
                    content, reasoning = self._extract_content_from_chunk(chunk)
                    
                    if reasoning:
                        yield f"data: [THOUGHT] {reasoning}\n\n"
                        chunks_received += 1
                        continue
                    
                    chunk_type = type(chunk).__name__

                    # Check for finish reason
                    if hasattr(chunk, "choices") and chunk.choices:
                        choice = chunk.choices[0]
                        if hasattr(choice, "finish_reason") and choice.finish_reason:
                            finish_reason = choice.finish_reason
                            if finish_reason == "length":
                                max_tokens = payload.max_completion_tokens
                                logger.warning(
                                    f"Chunk #{chunks_received} hit token limit! "
                                    f"max_tokens={max_tokens}, "
                                    f"input_messages={len(payload.messages)}, "
                                    f"estimated_input_tokens={estimated_input_tokens}, "
                                    f"provider={provider_name}. "
                                    "Consider increasing max_completion_tokens if this is unintended."
                                )

                        logger.debug(
                            f"Chunk #{chunks_received} structure: "
                            f"type={chunk_type}, "
                            f"choice_count={len(chunk.choices) if hasattr(chunk, 'choices') else 0}"
                        )
                        if hasattr(chunk, 'choices') and chunk.choices:
                            choice = chunk.choices[0]
                            logger.debug(
                                f"Choice 0: delta={getattr(choice, 'delta', 'N/A')}, "
                                f"finish_reason={getattr(choice, 'finish_reason', 'N/A')}"
                            )

                    if content:
                        chunk_buffer += content
                        buffer_size += len(content)

                        logger.debug(
                            f"Chunk #{chunks_received} received: "
                            f"type={chunk_type}, size={len(content)}, "
                            f"buffer_size={buffer_size}, "
                            f"preview={repr(content[:50])}"
                        )

                        # Flush buffer if threshold reached or sentence end
                        should_flush = (
                            buffer_size >= max_buffer_size or
                            any(
                                content.endswith(ending)
                                for ending in sentence_endings
                            )
                        )

                        if should_flush and chunk_buffer:
                            chunk_count += 1
                            total_chars_sent += len(chunk_buffer)

                            logger.debug(
                                f"Flushing buffer #{chunk_count}: "
                                f"size={len(chunk_buffer)}, "
                                f"total_sent={total_chars_sent}"
                            )

                            # Format as SSE: data: <content>\n\n
                            yield f"data: {chunk_buffer}\n\n"
                            chunk_buffer = ""
                            buffer_size = 0
                    else:
                        empty_chunks += 1
                        logger.debug(
                            f"Empty chunk #{chunks_received}: type={chunk_type}"
                        )

                # Check if we hit token limit early
                early_token_limit = (
                    chunks_received <= 5
                    and chunk_count == 0
                    and empty_chunks > 0
                )

                logger.info(
                    f"Stream completed: received={chunks_received}, "
                    f"empty={empty_chunks}, sent={chunk_count}, "
                    f"chars={total_chars_sent}, buffer={buffer_size}"
                )

                if early_token_limit:
                    # Suggest more aggressive truncation
                    suggested_max = max(2, len(payload.messages) // 2)
                    logger.error(
                        f"Stream failed: Hit token limit immediately. "
                        f"Input messages={len(payload.messages)}, "
                        f"estimated_input_tokens={estimated_input_tokens}, "
                        f"max_completion_tokens="
                        f"{payload.max_completion_tokens}. "
                        f"Model may have small context window. "
                        f"Suggested: Reduce to {suggested_max} messages "
                        f"or increase max_completion_tokens."
                    )

                # Flush any remaining buffer
                if chunk_buffer:
                    chunk_count += 1
                    total_chars_sent += len(chunk_buffer)
                    logger.debug(
                        f"Flushing final buffer: size={len(chunk_buffer)}"
                    )
                    yield f"data: {chunk_buffer}\n\n"
                elif chunks_received > 0 and chunk_count == 0:
                    logger.warning(
                        f"No data sent despite {chunks_received} chunks. "
                        f"All chunks may be empty or buffered."
                    )

                # Send completion signal
                logger.debug("Sending [DONE] signal")
                yield "data: [DONE]\n\n"

            except HTTPException:
                # Re-raise HTTP exceptions as-is
                logger.error(
                    f"HTTPException: received={chunks_received}, "
                    f"sent={chunk_count}"
                )
                raise
            except Exception as exc:  # pragma: no cover - runtime protection
                logger.exception(
                    f"Streaming failed: type={type(exc).__name__}, "
                    f"error={exc}, received={chunks_received}, "
                    f"sent={chunk_count}, empty={empty_chunks}"
                )
                # Send error as SSE event
                error_msg = f"Error: {str(exc)}"
                yield f"data: {error_msg}\n\n"
                yield "data: [DONE]\n\n"

        return StreamingResponse(
            generate_stream(),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "X-Accel-Buffering": "no",  # Disable nginx buffering
                "Transfer-Encoding": "chunked",  # Explicit chunked encoding
                "X-Content-Type-Options": "nosniff",
            },
        )
