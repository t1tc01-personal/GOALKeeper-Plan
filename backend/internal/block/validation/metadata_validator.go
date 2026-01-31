package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"goalkeeper-plan/internal/block/model"

	"github.com/xeipuuv/gojsonschema"
)

type MetadataValidator interface {
	Validate(ctx context.Context, blockType *model.BlockType, metadata model.JSONBMap) error
	ValidateJSON(ctx context.Context, schemaJSON, metadataJSON string) error
}

type metadataValidator struct{}

func NewMetadataValidator() MetadataValidator {
	return &metadataValidator{}
}

func (v *metadataValidator) Validate(ctx context.Context, blockType *model.BlockType, metadata model.JSONBMap) error {
	if blockType == nil {
		return fmt.Errorf("block type cannot be nil")
	}

	if blockType.MetadataSchema == nil || len(blockType.MetadataSchema) == 0 {
		return nil
	}

	schemaBytes, err := json.Marshal(blockType.MetadataSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata schema: %w", err)
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return v.ValidateJSON(ctx, string(schemaBytes), string(metadataBytes))
}

func (v *metadataValidator) ValidateJSON(ctx context.Context, schemaJSON, metadataJSON string) error {
	if schemaJSON == "" || schemaJSON == "{}" {
		return nil
	}

	schemaLoader := gojsonschema.NewStringLoader(schemaJSON)
	documentLoader := gojsonschema.NewStringLoader(metadataJSON)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		errMsg := "metadata validation failed:"
		for _, desc := range result.Errors() {
			errMsg += fmt.Sprintf("\n  - %s", desc)
		}
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}
