CREATE TABLE IF NOT EXISTS framework_library (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    config_template JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Script to initialize sample data for the framework_library table
-- Assumes the table structure: id (uuid/varchar), name (varchar), config_template (jsonb)

INSERT INTO framework_library (id, name, config_template) VALUES
(
  'template-kanban', 
  'Kanban Board', 
  '{
    "framework_type": "kanban",
    "default_columns": [
      { "key": "todo", "label": "To Do", "color": "gray" },
      { "key": "doing", "label": "In Progress", "color": "blue" },
      { "key": "done", "label": "Completed", "color": "green" }
    ],
    "required_metadata": ["status", "priority"],
    "ai_instructions": "Categorize tasks based on their current progress status. Prioritize sorting by priority level."
  }'::jsonb
),
(
  'template-gantt', 
  'Gantt Chart', 
  '{
    "framework_type": "gantt",
    "timeline_axis": "time",
    "required_fields": ["start_date", "end_date", "duration"],
    "constraint_rules": {
      "allow_dependencies": true,
      "overlap_prevention": false
    },
    "ai_instructions": "Identify overlapping tasks or violations of dependency logic between blocks."
  }'::jsonb
),
(
  'template-habit', 
  'Habit Tracker', 
  '{
    "framework_type": "habit_tracker",
    "display_style": "grid_7_days",
    "tracking_logic": {
      "type": "boolean_check",
      "reset_period": "daily"
    },
    "default_metadata": {
      "frequency": "daily",
      "goal": 1,
      "streak": 0
    },
    "ai_instructions": "Analyze streaks and suggest optimal times for habit execution based on the user schedule."
  }'::jsonb
),
(
  'template-wbs', 
  'Work Breakdown Structure', 
  '{
    "framework_type": "wbs",
    "structure": "hierarchical",
    "calculation_rules": {
      "aggregate_progress": true,
      "sum_budget": true
    },
    "required_metadata": ["wbs_code", "cost", "owner"],
    "ai_instructions": "Assist in breaking down high-level goals into 3-5 actionable sub-tasks."
  }'::jsonb
),
(
  'template-action', 
  'Action Plan', 
  '{
    "framework_type": "action_plan",
    "view_mode": "matrix",
    "matrix_config": {
      "x_axis": "effort",
      "y_axis": "impact"
    },
    "required_metadata": ["impact", "effort", "metric"],
    "ai_instructions": "Position actions within an Impact/Effort or Eisenhower matrix to optimize resource allocation."
  }'::jsonb
),
(
  'template-roadmap', 
  'Learning Roadmap', 
  '{
    "framework_type": "learning_roadmap",
    "view_mode": "vertical_timeline",
    "node_types": ["theory", "practice", "quiz"],
    "required_metadata": ["estimated_time", "difficulty", "assets"],
    "ai_instructions": "Recommend relevant resource URLs based on the difficulty level of each learning step."
  }'::jsonb
);