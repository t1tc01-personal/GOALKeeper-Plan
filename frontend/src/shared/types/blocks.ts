/**
 * Block type definitions and metadata
 * Based on backend migration: 20260127134056_add_blocks_table.sql
 */

import {
  Type,
  Heading1,
  Heading2,
  Heading3,
  List,
  ListOrdered,
  CheckSquare,
  ChevronRight,
  Info,
  Quote,
  Minus,
} from 'lucide-react';

export type BlockCategory = 'content' | 'framework' | 'media' | 'smart';

export interface BlockTypeConfig {
  name: string;
  displayName: string;
  description: string;
  icon: React.ComponentType<{ className?: string }>;
  category: BlockCategory;
  defaultMetadata?: Record<string, any>;
}

/**
 * Content block types (from lines 26-34 of migration)
 */
export const CONTENT_BLOCK_TYPES: BlockTypeConfig[] = [
  {
    name: 'text',
    displayName: 'Text',
    description: 'Plain text block with Markdown support',
    icon: Type,
    category: 'content',
    defaultMetadata: { format: 'markdown' },
  },
  {
    name: 'heading_1',
    displayName: 'Heading 1',
    description: 'Main heading',
    icon: Heading1,
    category: 'content',
    defaultMetadata: { level: 1 },
  },
  {
    name: 'heading_2',
    displayName: 'Heading 2',
    description: 'Sub heading',
    icon: Heading2,
    category: 'content',
    defaultMetadata: { level: 2 },
  },
  {
    name: 'heading_3',
    displayName: 'Heading 3',
    description: 'Sub-sub heading',
    icon: Heading3,
    category: 'content',
    defaultMetadata: { level: 3 },
  },
  {
    name: 'bulleted_list',
    displayName: 'Bulleted List',
    description: 'Unordered list',
    icon: List,
    category: 'content',
    defaultMetadata: { items: [] },
  },
  {
    name: 'numbered_list',
    displayName: 'Numbered List',
    description: 'Ordered list',
    icon: ListOrdered,
    category: 'content',
    defaultMetadata: { items: [] },
  },
  {
    name: 'todo_list',
    displayName: 'Todo List',
    description: 'Basic checklist (different from advanced Task)',
    icon: CheckSquare,
    category: 'content',
    defaultMetadata: { items: [], checked: [] },
  },
  {
    name: 'toggle',
    displayName: 'Toggle',
    description: 'Collapsible block to hide/show content (important for Learning Roadmap)',
    icon: ChevronRight,
    category: 'content',
    defaultMetadata: { collapsed: false },
  },
  {
    name: 'callout',
    displayName: 'Callout',
    description: 'Highlighted block with icon for notices',
    icon: Info,
    category: 'content',
    defaultMetadata: { icon: 'info', color: 'blue' },
  },
];

/**
 * Get block type config by name
 */
export function getBlockTypeConfig(name: string): BlockTypeConfig | undefined {
  return CONTENT_BLOCK_TYPES.find(type => type.name === name);
}

/**
 * Get all block types for a category
 */
export function getBlockTypesByCategory(category: BlockCategory): BlockTypeConfig[] {
  return CONTENT_BLOCK_TYPES.filter(type => type.category === category);
}

/**
 * Search block types by query (for slash command filtering)
 */
export function searchBlockTypes(query: string): BlockTypeConfig[] {
  const lowerQuery = query.toLowerCase();
  return CONTENT_BLOCK_TYPES.filter(type =>
    type.name.toLowerCase().includes(lowerQuery) ||
    type.displayName.toLowerCase().includes(lowerQuery) ||
    type.description.toLowerCase().includes(lowerQuery)
  );
}

