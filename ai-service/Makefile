run:
	uvicorn app:app --host 0.0.0.0 --port 8000

lite-llm-local:
	export $(grep -v '^#' .env | xargs) && env -u DATABASE_URL litellm --config litellm_config.yaml --port 8002