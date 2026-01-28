'use client';

import React, { useEffect, useState, useRef } from 'react';
import { Card } from './ui/card';
import { BlockEditor } from './BlockEditor';
import { SlashCommandMenu } from './SlashCommandMenu';
import { blockApi, type Block as ApiBlock } from '@/services/blockApi';
import { CONTENT_BLOCK_TYPES, type BlockTypeConfig } from '@/shared/types/blocks';

interface BlockListProps {
  pageId?: string;
}

export function BlockList({ pageId }: BlockListProps) {
  const [blocks, setBlocks] = useState<ApiBlock[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newBlockContent, setNewBlockContent] = useState('');
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [showSlashMenu, setShowSlashMenu] = useState(false);
  const [slashMenuPosition, setSlashMenuPosition] = useState<{ top: number; left: number } | undefined>();
  const [slashQuery, setSlashQuery] = useState('');
  const newBlockInputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (pageId) {
      loadBlocks(pageId);
    } else {
      setBlocks([]);
    }
  }, [pageId]);

  const loadBlocks = async (id: string) => {
    try {
      setIsLoading(true);
      setError(null);
      const loadedBlocks = await blockApi.listBlocks(id);
      setBlocks(loadedBlocks);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load blocks');
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddBlock = async (blockType: BlockTypeConfig, content: string = '') => {
    if (!pageId) return;

    try {
      setIsLoading(true);
      setError(null);
      const newBlock = await blockApi.createBlock({
        pageId: pageId,
        type: blockType.name,
        content: content,
        position: blocks.length,
        blockConfig: blockType.defaultMetadata,
      });
      setBlocks([...blocks, newBlock]);
      setNewBlockContent('');
      setShowSlashMenu(false);
      setSlashQuery('');
      
      // Focus the new block to enter edit mode
      setTimeout(() => {
        const blockElement = document.querySelector(`[data-block-id="${newBlock.id}"]`);
        const editableElement = blockElement?.querySelector<HTMLElement>('[role="button"], textarea, input');
        editableElement?.click();
      }, 100);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to add block';
      setError(errorMsg);
    } finally {
      setIsLoading(false);
    }
  };

  const handleNewBlockInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setNewBlockContent(value);

    // Check if user typed "/" to trigger slash menu
    if (value.endsWith('/') && !showSlashMenu) {
      const input = e.target;
      const rect = input.getBoundingClientRect();
      setSlashMenuPosition({
        top: rect.bottom + 4,
        left: rect.left,
      });
      setShowSlashMenu(true);
      setSlashQuery('');
    } else if (showSlashMenu && value.includes('/')) {
      // Update query if menu is open
      const lastSlashIndex = value.lastIndexOf('/');
      const query = value.slice(lastSlashIndex + 1);
      setSlashQuery(query);
    } else if (!value.includes('/')) {
      // Close menu if "/" is removed
      setShowSlashMenu(false);
      setSlashQuery('');
    }
  };

  const handleNewBlockKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // If slash menu is open and user presses Escape, close it
    if (e.key === 'Escape' && showSlashMenu) {
      e.preventDefault();
      setShowSlashMenu(false);
      setSlashQuery('');
      // Remove the "/" from input
      const newValue = newBlockContent.slice(0, -1);
      setNewBlockContent(newValue);
      return;
    }

    // If user presses Enter without slash menu, create text block
    if (e.key === 'Enter' && !e.shiftKey && !showSlashMenu) {
      e.preventDefault();
      if (newBlockContent.trim()) {
        const textBlockType = CONTENT_BLOCK_TYPES.find(t => t.name === 'text')!;
        handleAddBlock(textBlockType, newBlockContent.trim());
      }
    }
  };

  const handleSlashMenuSelect = (blockType: BlockTypeConfig) => {
    // Remove the "/" and query from input
    const contentWithoutSlash = newBlockContent.replace(/\/.*$/, '').trim();
    handleAddBlock(blockType, contentWithoutSlash);
  };

  const handleUpdateBlock = (updatedBlock: ApiBlock) => {
    setBlocks(blocks.map(b => (b.id === updatedBlock.id ? updatedBlock : b)));
  };

  const handleDeleteBlock = (blockId: string) => {
    setBlocks(blocks.filter(b => b.id !== blockId));
  };

  const handleBlockError = (error: string) => {
    setError(error);
  };

  // Drag and drop handlers
  const handleDragStart = (index: number) => {
    setDraggedIndex(index);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDrop = async (targetIndex: number) => {
    if (draggedIndex === null || draggedIndex === targetIndex) {
      setDraggedIndex(null);
      return;
    }

    try {
      // Optimistic update
      const newBlocks = [...blocks];
      const [draggedBlock] = newBlocks.splice(draggedIndex, 1);
      newBlocks.splice(targetIndex, 0, draggedBlock);
      setBlocks(newBlocks);

      // Reorder on backend
      const blockIds = newBlocks.map(b => b.id);
      await blockApi.reorderBlocks({ pageId: pageId!, blockIds });
    } catch (err) {
      // Rollback on error
      await loadBlocks(pageId!);
      const errorMsg = err instanceof Error ? err.message : 'Failed to reorder blocks';
      setError(errorMsg);
    } finally {
      setDraggedIndex(null);
    }
  };

  if (!pageId) {
    return (
      <section className="space-y-2">
        <Card className="p-4 text-sm text-muted-foreground">
          Select a page to view and edit content blocks.
        </Card>
      </section>
    );
  }

  // Create refs map for block navigation
  const blockRefsMap = useRef<Map<string, HTMLDivElement>>(new Map());

  return (
    <section 
      className="space-y-4"
      role="region"
      aria-label="Page content blocks"
      aria-describedby="block-list-help"
    >
      <div id="block-list-help" className="sr-only">
        Content blocks for this page. Use arrow keys to navigate between blocks when editing.
        Drag blocks to reorder them.
      </div>
      {error && (
        <div className="p-3 bg-red-50 text-red-700 text-sm rounded flex justify-between items-center">
          <span>{error}</span>
          <button
            onClick={() => setError(null)}
            className="text-xs px-2 py-1 hover:bg-red-100 rounded"
          >
            ✕
          </button>
        </div>
      )}

      {isLoading && blocks.length === 0 && (
        <Card className="p-4 text-sm text-muted-foreground">
          Loading blocks...
        </Card>
      )}

      {blocks.length === 0 && !isLoading && (
        <Card className="p-4 text-sm text-muted-foreground">
          No content blocks yet. Start typing to create your first block.
        </Card>
      )}

      {blocks.map((block, index) => {
        const handleFocusPrevious = () => {
          if (index > 0) {
            const prevBlockId = blocks[index - 1].id;
            const prevBlockElement = blockRefsMap.current.get(prevBlockId);
            if (prevBlockElement) {
              const editableElement = prevBlockElement.querySelector<HTMLElement>('[role="button"], textarea, input');
              editableElement?.focus();
            }
          }
        };

        const handleFocusNext = () => {
          if (index < blocks.length - 1) {
            const nextBlockId = blocks[index + 1].id;
            const nextBlockElement = blockRefsMap.current.get(nextBlockId);
            if (nextBlockElement) {
              const editableElement = nextBlockElement.querySelector<HTMLElement>('[role="button"], textarea, input');
              editableElement?.focus();
            }
          }
        };

        return (
          <div
            key={block.id}
            ref={(el) => {
              if (el) blockRefsMap.current.set(block.id, el);
            }}
            draggable
            onDragStart={() => handleDragStart(index)}
            onDragOver={handleDragOver}
            onDrop={() => handleDrop(index)}
            className={`transition-opacity ${draggedIndex === index ? 'opacity-50' : 'opacity-100'}`}
          >
            <div className="flex gap-2 group">
              <div 
                className="cursor-grab active:cursor-grabbing pt-3 text-gray-400"
                aria-label={`Drag to reorder block ${index + 1}`}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    // Trigger drag (simplified - full implementation would use drag API)
                  }
                }}
              >
                ⋮⋮
              </div>
              <div className="flex-1">
                <BlockEditor
                  block={block}
                  onUpdate={handleUpdateBlock}
                  onDelete={handleDeleteBlock}
                  onSaveError={handleBlockError}
                  blockIndex={index}
                  totalBlocks={blocks.length}
                  onFocusPrevious={index > 0 ? handleFocusPrevious : undefined}
                  onFocusNext={index < blocks.length - 1 ? handleFocusNext : undefined}
                />
              </div>
            </div>
          </div>
        );
      })}

      <div className="mt-4 pt-4 border-t relative">
        <textarea
          ref={newBlockInputRef}
          value={newBlockContent}
          onChange={handleNewBlockInputChange}
          onKeyDown={handleNewBlockKeyDown}
          placeholder="Type '/' to choose a block type, or start typing..."
          className="w-full p-3 border rounded resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
          rows={2}
          disabled={isLoading}
        />
        {showSlashMenu && slashMenuPosition && (
          <SlashCommandMenu
            open={showSlashMenu}
            onClose={() => {
              setShowSlashMenu(false);
              setSlashQuery('');
            }}
            onSelect={handleSlashMenuSelect}
            query={slashQuery}
            position={slashMenuPosition}
          />
        )}
      </div>
    </section>
  );
}

