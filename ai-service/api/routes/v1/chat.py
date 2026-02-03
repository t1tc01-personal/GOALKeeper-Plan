from fastapi import APIRouter
from handler.chat import ChatHandler, ChatResponse, ModelListResponse


class ChatRoute:
    """Routes for chat completion."""

    router: APIRouter
    handler: ChatHandler

    def __init__(self, handler: ChatHandler):
        self.router = APIRouter()
        self.handler = handler

        self.router.add_api_route(
            path="",
            endpoint=self.handler.chat,
            methods=["POST"],
            response_model=ChatResponse,
            summary="Chat completion",
            description=(
                "REST endpoint for chat completion using configured LLM provider "
                "(supports Azure OpenAI, OpenAI, Anthropic, and 100+ providers via LiteLLM)."
            ),
        )

        self.router.add_api_route(
            path="/models",
            endpoint=self.handler.get_models,
            methods=["GET"],
            response_model=ModelListResponse,
            summary="List available models",
            description="Returns a list of unique model names configured in LiteLLM Proxy.",
        )

        self.router.add_api_route(
            path="/stream",
            endpoint=self.handler.chat_stream,
            methods=["POST"],
            summary="Stream chat completion",
            description=(
                "Streaming endpoint that uses Server-Sent Events (SSE) to "
                "stream chat completion responses from the configured LLM provider."
            ),
        )
