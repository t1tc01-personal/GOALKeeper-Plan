FROM python:3.10-slim

# RUN apt-get update && apt-get install -y \
#     gcc curl \
#     poppler-utils\
#     tesseract-ocr\
#     tesseract-ocr-vie \
#     libpq-dev \
#     && rm -rf /var/lib/apt/lists/*
# RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash
COPY requirements requirements.txt
RUN pip install --upgrade pip
RUN pip install -r requirements.txt

COPY . /app

ENV PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION=python
ENV PYTHONPATH=/app

WORKDIR /app

CMD [ "uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]