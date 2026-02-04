from enum import Enum
from typing import Any, Dict, List, Optional
from pydantic import Field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class AppEnvTypes(str, Enum):
    """Application environment types."""
    PROD = "prod"
    DEV = "dev"
    TEST = "test"


class BaseAppSettings(BaseSettings):
    """Base application settings."""
    app_env: AppEnvTypes = AppEnvTypes.DEV

    model_config = SettingsConfigDict(
        env_file=".env",
        extra="ignore",
        case_sensitive=False,  # Allow case-insensitive env vars
    )


class AppSettings(BaseAppSettings):
    """Application settings with multi-provider LLM configuration."""

    # FastAPI settings
    debug: bool = False
    docs_url: str = "/docs"
    openapi_prefix: str = ""
    openapi_url: str = "/openapi.json"
    redoc_url: str = "/redoc"
    title: str = "AI Agent Service Demo"
    version: str = "0.0.0"

    # CORS & Database
    allowed_hosts: Any = Field(
        default=["*"],
        description="Allowed CORS hosts (comma-separated string or list)"
    )
    pool_size: int = Field(default=40, ge=1, description="Database pool size")
    max_overflow: int = Field(
        default=5, ge=0, description="Database max overflow connections"
    )
    pool_recycle: int = Field(
        default=600, ge=1, description="Database pool recycle time (seconds)"
    )

    # Redis Settings
    redis_host: str = Field(default="localhost", description="Redis host")
    redis_port: int = Field(default=6379, description="Redis port")

    # LiteLLM Proxy Settings
    litellm_proxy_url: str = Field(
        default="http://localhost:8002",
        alias="LITELLM_PROXY_URL",
        description="LiteLLM Proxy base URL"
    )
    litellm_master_key: str = Field(
        default="sk-1234",
        alias="LITELLM_MASTER_KEY",
        description="LiteLLM Proxy master key"
    )

    # LiteLLM Provider-agnostic settings
    llm_provider: Optional[str] = Field(
        default=None,
        description="LLM provider (azure, openai, anthropic, etc.)"
    )
    llm_model: Optional[str] = Field(
        default=None,
        description="LLM model identifier (e.g., 'gpt-4', 'claude-3-5-sonnet', 'azure/gpt-4')"
    )
    llm_api_key: Optional[str] = Field(
        default=None,
        description="Provider API key"
    )
    llm_api_base: Optional[str] = Field(
        default=None,
        description="API base URL (for Azure or custom endpoints)"
    )
    llm_api_version: Optional[str] = Field(
        default=None,
        description="API version (for Azure OpenAI)"
    )

    # Azure OpenAI settings (kept for backward compatibility)
    azure_tenant_id: Optional[str] = Field(
        default=None,
        description="Azure AD tenant ID"
    )
    azure_client_id: Optional[str] = Field(
        default=None,
        description="Azure AD client ID (required for Azure AD auth)"
    )
    azure_client_secret: Optional[str] = Field(
        default=None,
        description="Azure AD client secret (required for Azure AD auth)"
    )
    openai_use_key: bool = Field(
        default=False,
        description="Use API key authentication (False = use Azure AD)"
    )
    openai_api_version: Optional[str] = Field(
        default=None,
        description="Azure OpenAI API version (e.g., '2024-02-15-preview')"
    )
    openai_azure_endpoint: Optional[str] = Field(
        default=None,
        description="Azure OpenAI endpoint URL"
    )
    openai_azure_endpoint_3: Optional[str] = Field(
        default=None,
        description="Azure OpenAI endpoint for embeddings model 3"
    )
    openai_api_key: Optional[str] = Field(
        default=None,
        description="Azure OpenAI API key (required if openai_use_key=True)"
    )
    openai_api_key_3: Optional[str] = Field(
        default=None,
        description="Azure OpenAI API key for embeddings model 3"
    )
    openai_engine: Optional[str] = Field(
        default=None,
        description="Azure OpenAI deployment name"
    )
    openai_model_name: Optional[str] = Field(
        default=None,
        description="Azure OpenAI model name"
    )
    openai_embedding_3_deployment: Optional[str] = Field(
        default=None,
        description="Azure OpenAI embedding deployment name for model 3"
    )

    # Message settings
    number_message: Optional[int] = Field(
        default=2,
        description=(
            "Number of messages to process "
            "(purpose unclear - needs documentation)"
        ),
    )

    # Pagination
    max_page_size: int = Field(
        default=100,
        ge=1,
        description="Maximum page size for pagination"
    )

    model_config = SettingsConfigDict(
        validate_assignment=True,
        env_file=".env",
        extra="ignore",
        case_sensitive=False,
    )

    # core/settings.py (thêm vào)
    # Multi-provider configuration
    llm_providers: Optional[List[Dict[str, Any]]] = Field(
        default=None,
        description="List of provider configurations. Format: [{'name': 'azure', 'model': 'azure/gpt-4', 'api_key': '...', 'priority': 1}, ...]"
    )

    @field_validator("allowed_hosts", mode="before")
    @classmethod
    def parse_allowed_hosts(cls, v: Any) -> List[str]:
        """Parse allowed_hosts from string or list."""
        if isinstance(v, str):
            v = v.strip()
            # Handle possible bracketed string from .env
            if v.startswith("[") and v.endswith("]"):
                v = v[1:-1]
            if not v:
                return ["*"]
            return [host.strip().replace('"', '').replace("'", "") for host in v.split(",") if host.strip()]
        if isinstance(v, list):
            return v
        return ["*"]

    @field_validator("max_page_size")
    @classmethod
    def validate_max_page_size(cls, v: int) -> int:
        if v < 1:
            raise ValueError("max_page_size must be >= 1")
        return v

    @field_validator("openai_azure_endpoint", "openai_azure_endpoint_3", "llm_api_base")
    @classmethod
    def validate_endpoint_url(cls, v: Optional[str]) -> Optional[str]:
        """Validate endpoint URL format."""
        if v and not v.startswith(("http://", "https://")):
            raise ValueError(
                "Endpoint must start with http:// or https://"
            )
        return v

    @field_validator("openai_api_key")
    @classmethod
    def validate_api_key_if_needed(
        cls, v: Optional[str], info
    ) -> Optional[str]:
        """Validate API key is provided when openai_use_key=True."""
        if info.data.get("openai_use_key") and not v:
            raise ValueError(
                "openai_api_key is required when openai_use_key=True"
            )
        return v

    def get_llm_model_string(self) -> str:
        """
        Get the LiteLLM model string from settings.
        Supports both new provider-agnostic settings and legacy Azure settings.
        """
        # Use new settings if provided
        if self.llm_model:
            return self.llm_model
        
        # Fallback to Azure settings for backward compatibility
        if self.openai_engine:
            return f"azure/{self.openai_engine}"
        if self.openai_model_name:
            return f"azure/{self.openai_model_name}"
        
        raise ValueError(
            "LLM model not configured. Set LLM_MODEL or use legacy Azure settings."
        )

    def get_llm_api_key(self) -> Optional[str]:
        """Get API key from new or legacy settings."""
        if self.llm_api_key:
            return self.llm_api_key
        return self.openai_api_key

    def get_llm_api_base(self) -> Optional[str]:
        """Get API base URL from new or legacy settings."""
        if self.llm_api_base:
            return self.llm_api_base
        return self.openai_azure_endpoint

    def get_llm_api_version(self) -> Optional[str]:
        """Get API version from new or legacy settings."""
        if self.llm_api_version:
            return self.llm_api_version
        return self.openai_api_version

    @property
    def fastapi_kwargs(self) -> Dict[str, Any]:
        """Get FastAPI initialization kwargs."""
        return {
            "debug": self.debug,
            "docs_url": self.docs_url,
            "openapi_prefix": self.openapi_prefix,
            "openapi_url": self.openapi_url,
            "redoc_url": self.redoc_url,
            "title": self.title,
            "version": self.version,
        }


settings = AppSettings()
