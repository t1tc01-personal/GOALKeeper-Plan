from fastapi import HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
from services.document_processor import DocumentProcessor
from services.rag_service import RAGService
from services.llm_service import LLMService

class ProcessDocumentRequest(BaseModel):
    file_path: str
    file_type: str
    workspace_id: str
    is_private: bool = False

class RAGQueryRequest(BaseModel):
    query: str
    workspace_id: str

class InternalHandler:
    def __init__(self):
        self.doc_processor = DocumentProcessor()
        self.llm_service = LLMService()
        self.rag_service = RAGService(self.llm_service)

    async def process_document(self, request: ProcessDocumentRequest):
        try:
            chunks = self.doc_processor.process_document(request.file_path, request.file_type)
            # Here we would normally save chunks to pgvector
            # For now, we just return the count
            return {"status": "success", "chunks_processed": len(chunks)}
        except Exception as e:
            raise HTTPException(status_code=500, detail=str(e))

    async def rag_query(self, request: RAGQueryRequest):
        async def stream_generator():
            async for chunk in self.rag_service.query_with_context(request.query, request.workspace_id):
                yield chunk

        return StreamingResponse(stream_generator(), media_type="text/plain")
