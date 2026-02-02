import os
from opik import track, start_as_current_span
from typing import List, Dict, Any
import pypdf
import docx

class DocumentProcessor:
    @track(project_name="document-processing")
    def process_document(self, file_path: str, file_type: str) -> List[Dict[str, Any]]:
        text_content = ""
        with start_as_current_span(name="extract_text", input={"file_path": file_path}) as span:
            if file_type == "pdf":
                text_content = self._extract_from_pdf(file_path)
            elif file_type == "docx":
                text_content = self._extract_from_docx(file_path)
            elif file_type == "txt" or file_type == "md":
                text_content = self._extract_from_text(file_path)
            
            span.update(output={"text_length": len(text_content)})
            
        chunks = []
        with start_as_current_span(name="chunk_text", input={"length": len(text_content)}) as span:
            chunks = self._chunk_text(text_content)
            span.update(output={"num_chunks": len(chunks)})
            
        return chunks

    def _extract_from_pdf(self, path: str) -> str:
        text = ""
        with open(path, 'rb') as f:
            reader = pypdf.PdfReader(f)
            for page in reader.pages:
                text += page.extract_text() + "\n"
        return text

    def _extract_from_docx(self, path: str) -> str:
        doc = docx.Document(path)
        return "\n".join([para.text for para in doc.paragraphs])

    def _extract_from_text(self, path: str) -> str:
        with open(path, 'r', encoding='utf-8') as f:
            return f.read()

    def _chunk_text(self, text: str, chunk_size: int = 1000, overlap: int = 100) -> List[Dict[str, Any]]:
        # Simple character splitting for now. Use RecursiveCharacterTextSplitter for better results.
        chunks = []
        for i in range(0, len(text), chunk_size - overlap):
            chunk_text = text[i:i + chunk_size]
            chunks.append({"text": chunk_text, "index": len(chunks)})
        return chunks
