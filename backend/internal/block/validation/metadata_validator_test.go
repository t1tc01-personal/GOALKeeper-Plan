package validation

import (
	"context"
	"goalkeeper-plan/internal/block/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataValidator_Validate(t *testing.T) {
	validator := NewMetadataValidator()
	ctx := context.Background()

	t.Run("valid metadata passes validation", func(t *testing.T) {
		schema := model.JSONBMap{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type": "string",
					"enum": []interface{}{"todo", "in_progress", "done"},
				},
				"priority": map[string]interface{}{
					"type": "string",
					"enum": []interface{}{"low", "medium", "high"},
				},
			},
			"required": []interface{}{"status"},
		}

		blockType := &model.BlockType{
			MetadataSchema: schema,
		}

		metadata := model.JSONBMap{
			"status":   "todo",
			"priority": "high",
		}

		err := validator.Validate(ctx, blockType, metadata)
		assert.NoError(t, err)
	})

	t.Run("invalid metadata fails validation", func(t *testing.T) {
		schema := model.JSONBMap{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type": "string",
					"enum": []interface{}{"todo", "in_progress", "done"},
				},
			},
			"required": []interface{}{"status"},
		}

		blockType := &model.BlockType{
			MetadataSchema: schema,
		}

		metadata := model.JSONBMap{
			"status": "invalid_status",
		}

		err := validator.Validate(ctx, blockType, metadata)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("missing required field fails validation", func(t *testing.T) {
		schema := model.JSONBMap{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []interface{}{"status"},
		}

		blockType := &model.BlockType{
			MetadataSchema: schema,
		}

		metadata := model.JSONBMap{
			"priority": "high",
		}

		err := validator.Validate(ctx, blockType, metadata)
		assert.Error(t, err)
	})

	t.Run("nil or empty schema passes validation", func(t *testing.T) {
		blockType := &model.BlockType{
			MetadataSchema: nil,
		}

		metadata := model.JSONBMap{
			"anything": "goes",
		}

		err := validator.Validate(ctx, blockType, metadata)
		assert.NoError(t, err)

		blockType.MetadataSchema = model.JSONBMap{}
		err = validator.Validate(ctx, blockType, metadata)
		assert.NoError(t, err)
	})

	t.Run("nil block type returns error", func(t *testing.T) {
		metadata := model.JSONBMap{}
		err := validator.Validate(ctx, nil, metadata)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestMetadataValidator_ValidateJSON(t *testing.T) {
	validator := NewMetadataValidator()
	ctx := context.Background()

	t.Run("valid JSON passes validation", func(t *testing.T) {
		schemaJSON := `{
			"type": "object",
			"properties": {
				"framework_type": {
					"type": "string",
					"enum": ["kanban", "gantt", "habit"]
				}
			},
			"required": ["framework_type"]
		}`

		metadataJSON := `{
			"framework_type": "kanban",
			"columns": []
		}`

		err := validator.ValidateJSON(ctx, schemaJSON, metadataJSON)
		assert.NoError(t, err)
	})

	t.Run("empty schema passes validation", func(t *testing.T) {
		err := validator.ValidateJSON(ctx, "{}", `{"anything": "goes"}`)
		assert.NoError(t, err)

		err = validator.ValidateJSON(ctx, "", `{"anything": "goes"}`)
		assert.NoError(t, err)
	})

	t.Run("invalid JSON format returns error", func(t *testing.T) {
		schemaJSON := `{"type": "object"}`
		invalidJSON := `{invalid json}`

		err := validator.ValidateJSON(ctx, schemaJSON, invalidJSON)
		require.Error(t, err)
	})

	t.Run("kanban framework schema validation", func(t *testing.T) {
		schemaJSON := `{
			"type": "object",
			"properties": {
				"framework_type": {"type": "string"},
				"columns": {
					"type": "array",
					"items": {
						"type": "object",
						"properties": {
							"key": {"type": "string"},
							"label": {"type": "string"},
							"color": {"type": "string"}
						},
						"required": ["key", "label"]
					}
				}
			},
			"required": ["framework_type", "columns"]
		}`

		validMetadata := `{
			"framework_type": "kanban",
			"columns": [
				{"key": "todo", "label": "To Do", "color": "gray"},
				{"key": "done", "label": "Done", "color": "green"}
			]
		}`

		err := validator.ValidateJSON(ctx, schemaJSON, validMetadata)
		assert.NoError(t, err)

		invalidMetadata := `{
			"framework_type": "kanban",
			"columns": [
				{"key": "todo"}
			]
		}`

		err = validator.ValidateJSON(ctx, schemaJSON, invalidMetadata)
		assert.Error(t, err)
	})
}
