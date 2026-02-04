"""
Internal API routes for the AI service.

This module defines the InternalRoute class which sets up endpoints for
internal operations like document processing and RAG queries.
"""
from fastapi import APIRouter
from handler.internal import InternalHandler


class InternalRoute:
    """
    Internal API routes for the AI service.

    This class defines the InternalRoute class which sets up endpoints for
    internal operations like document processing and RAG queries.
    """
    router: APIRouter
    handler: InternalHandler

    def __init__(self, handler: InternalHandler):
        self.router = APIRouter()
        self.handler = handler

        self.router.add_api_route(
            path="/process-document",
            endpoint=self.handler.process_document,
            methods=["POST"],
            summary="Process uploaded document",
        )

        self.router.add_api_route(
            path="/rag-query",
            endpoint=self.handler.rag_query,
            methods=["POST"],
            summary="RAG Query with streaming response",
        )
