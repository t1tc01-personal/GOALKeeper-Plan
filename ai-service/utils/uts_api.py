from pathlib import Path
from tempfile import NamedTemporaryFile

from common.exceptions.common import ExceptionFileTooLarge
from core.settings import settings as st
from fastapi import UploadFile

import os
from typing import Optional


def get_paging_params(
    page: int = 1,
    pagesize: int = 10
) -> tuple[int, int]:
    """Get pagination parameters (skip, limit).

    Args:
        page: Page number (must be >= 1)
        pagesize: Items per page (must be >= 1)

    Returns:
        Tuple of (skip, limit) for database query
    """
    if page < 1 or pagesize < 1:
        raise ValueError("page and pagesize must be >= 1")

    if pagesize > st.max_page_size:
        pagesize = st.max_page_size

    skip = (page - 1) * pagesize
    return skip, pagesize


async def read_file(file: UploadFile, max_size: int) -> bytes:
    """Read file content with size limit.

    Args:
        file: FastAPI UploadFile object
        max_size: Maximum file size in bytes

    Returns:
        File content as bytes

    Raises:
        ExceptionFileTooLarge: If file exceeds max_size
    """
    file_content = bytearray()
    file_size = 0

    try:
        while chunk := await file.read(1024):
            file_size += len(chunk)

            if file_size > max_size:
                raise ExceptionFileTooLarge(max_size)

            file_content.extend(chunk)

        return bytes(file_content)
    finally:
        # Reset file pointer for potential reuse
        await file.seek(0)


async def save_uploaded_file(file: UploadFile) -> str:
    """Save uploaded file to temporary location.

    Args:
        file: FastAPI UploadFile object

    Returns:
        Path to saved temporary file

    Raises:
        ValueError: If filename is invalid
        OSError: If file operations fail
    """
    if not file.filename:
        raise ValueError("Filename is required")

    # Sanitize filename to prevent path traversal
    filename = Path(file.filename).name  # Only get basename
    suffix = Path(filename).suffix

    # Validate suffix (optional: whitelist allowed extensions)
    # if suffix not in ALLOWED_EXTENSIONS:
    #     raise ValueError(f"File extension {suffix} not allowed")

    tmp_path: Optional[str] = None

    try:
        with NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
            # Use async read if available, or sync copy
            content = await file.read()
            tmp.write(content)
            tmp_path = Path(tmp.name).as_posix()

        return tmp_path
    except Exception:
        # Cleanup on error
        if tmp_path and os.path.exists(tmp_path):
            try:
                os.unlink(tmp_path)
            except OSError:
                pass  # Ignore cleanup errors
        raise
    finally:
        # Reset file pointer if needed
        try:
            await file.seek(0)
        except Exception:
            pass  # File may already be closed
