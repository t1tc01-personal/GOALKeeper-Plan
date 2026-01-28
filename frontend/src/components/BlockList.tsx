'use client';

import React, { useEffect, useState, useRef } from 'react';
import { Card } from './ui/card';
import { BlockEditor } from './BlockEditor';
import { blockApi, type Block as ApiBlock } from '@/services/blockApi';

interface BlockListProps {
  pageId?: string;
}

export function BlockList({ pageId }: BlockListProps) {
  const [blocks, setBlocks] = useState<ApiBlock[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newBlockContent, setNewBlockContent] = useState('');
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);

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

  const handleAddBlock = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newBlockContent.trim() || !pageId) return;

    try {
      setIsLoading(true);
      setError(null);
      const newBlock = await blockApi.createBlock({
        pageId: pageId,
        type: 'text', // Use 'text' block type from backend
        content: newBlockContent,
        position: blocks.length,
      });
      setBlocks([...blocks, newBlock]);
      setNewBlockContent('');
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to add block';
      setError(errorMsg);
    } finally {
      setIsLoading(false);
    }
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

      <form onSubmit={handleAddBlock} className="mt-4 pt-4 border-t">
        <textarea
          value={newBlockContent}
          onChange={e => setNewBlockContent(e.target.value)}
          placeholder="Type to add a new block..."
          className="w-full p-3 border rounded resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
          rows={2}
          disabled={isLoading}
        />
        <button
          type="submit"
          className="mt-2 px-4 py-2 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors disabled:opacity-50"
          disabled={isLoading || !newBlockContent.trim()}
        >
          {isLoading ? 'Adding...' : 'Add Block'}
        </button>
      </form>
    </section>
  );
}

