from fastapi import APIRouter
from handler.internal import InternalHandler, ProcessDocumentRequest, RAGQueryRequest

class InternalRoute:
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
