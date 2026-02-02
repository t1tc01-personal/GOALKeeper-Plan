import os
from typing import Callable
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()
import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from starlette.middleware.gzip import GZipMiddleware
from loguru import logger
from api.routes.v1.hello_world import HelloWorldRoute
from api.routes.v1.chat import ChatRoute
from handler.hello_world import HelloWorldHandler
from handler.chat import ChatHandler
from utils.uts_scheduler import scheduler
from fastapi import Response


class Settings:
    """Application settings configuration"""
    fastapi_kwargs = {
        "title": "AI Agent Service",
        "version": "1.0.0",
    }
    allowed_hosts = ["*"]  # Allow all origins in development


settings = Settings()


class App:
    application: FastAPI

    def on_init_app(self) -> Callable:
        async def start_app() -> None:

            hello_world_handle = HelloWorldHandler()
            chat_handler = ChatHandler()
            hello_world_router = HelloWorldRoute(hello_world_handle)
            chat_router = ChatRoute(chat_handler)
            
            # Internal routes (RAG, Doc Processing)
            from handler.internal import InternalHandler
            from api.routes.v1.internal import InternalRoute
            internal_handler = InternalHandler()
            internal_router = InternalRoute(internal_handler)

            prefix = "/api/v1"

            self.application.include_router(
                hello_world_router.router,
                prefix=prefix + "/hello-world",
                tags=["Hello-world"]
            )
            self.application.include_router(
                chat_router.router,
                prefix=prefix + "/chat",
                tags=["Chat"]
            )
            self.application.include_router(
                internal_router.router,
                prefix=prefix + "/internal",
                tags=["Internal"]
            )
            scheduler.start()

        return start_app

    def on_terminate_app(self) -> Callable:
        @logger.catch
        async def stop_app() -> None:
            pass

        return stop_app

    def __init__(self):
        self.application = FastAPI(**settings.fastapi_kwargs)

        self.application.add_middleware(
            GZipMiddleware,
            minimum_size=1000,  # Compress responses larger than 1KB
        )

        self.application.add_middleware(
            CORSMiddleware,
            allow_origins=settings.allowed_hosts,
            allow_credentials=True,
            allow_methods=["*"],
            allow_headers=["*"],
        )

        # Add timeout middleware to handle long-running requests
        @self.application.middleware("http")
        async def timeout_middleware(request, call_next):
            import asyncio
            try:
                # Set a longer timeout for large responses (10 minutes)
                response = await asyncio.wait_for(
                    call_next(request),
                    timeout=600.0
                )
                return response
            except asyncio.TimeoutError:
                logger.error(f"Request timeout: {request.url}")
                return Response(
                    content='{"detail": "Request timeout"}',
                    status_code=504,
                    media_type="application/json"
                )

        self.application.add_event_handler("startup", self.on_init_app())
        self.application.add_event_handler("shutdown", self.on_terminate_app())


app = App().application


if __name__ == "__main__":
    uvicorn.run("app:app", host="0.0.0.0", port=8000, reload=True)
