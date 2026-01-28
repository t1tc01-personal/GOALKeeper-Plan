'use client';

import React, { useEffect, useState, useRef, useCallback } from 'react';
import { Card } from './ui/card';
import { InlineBlockEditor } from './InlineBlockEditor';
import { blockApi, type Block } from '@/services/blockApi';
import { CONTENT_BLOCK_TYPES, type BlockTypeConfig } from '@/shared/types/blocks';

interface PageEditorProps {
  pageId?: string;
}

// Temp ID generator
let tempIdCounter = 0;
function generateTempId(): string {
  return `temp-${Date.now()}-${++tempIdCounter}`;
}

export function PageEditor({ pageId }: PageEditorProps) {
  const [blocks, setBlocks] = useState<Block[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [focusedBlockId, setFocusedBlockId] = useState<string | null>(null);
  const [pendingCreates, setPendingCreates] = useState<Map<string, Block>>(new Map());
  const blockRefsMap = useRef<Map<string, HTMLDivElement>>(new Map());
  const [cursorPositions, setCursorPositions] = useState<Map<string, number>>(new Map());

  // Load blocks from server
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

  // Get all blocks including pending ones
  const getAllBlocks = useCallback((): Block[] => {
    const allBlocks = [...blocks];
    pendingCreates.forEach((block) => {
      allBlocks.push(block);
    });
    // Sort by position
    return allBlocks.sort((a, b) => a.position - b.position);
  }, [blocks, pendingCreates]);

  // Create block on server for a temp block (optimistic create with rollback)
  const createBlockOnServer = useCallback(
    async (tempBlock: Block, blockType: BlockTypeConfig) => {
      if (!pageId) return;

      try {
        const createdBlock = await blockApi.createBlock({
          pageId,
          type: blockType.name,
          content: tempBlock.content,
          position: tempBlock.position,
          blockConfig: blockType.defaultMetadata || {},
        });

        // Remove temp from pending
        setPendingCreates((prev) => {
          const next = new Map(prev);
          next.delete(tempBlock.id);
          return next;
        });

        // Insert created block
        setBlocks((prevBlocks) => {
          const next = [...prevBlocks, createdBlock].sort(
            (a, b) => a.position - b.position
          );
          return next;
        });

        // Preserve focus and cursor position
        setFocusedBlockId((prev) =>
          prev === tempBlock.id ? createdBlock.id : prev
        );
        setCursorPositions((prev) => {
          const next = new Map(prev);
          if (next.has(tempBlock.id)) {
            const pos = next.get(tempBlock.id)!;
            next.delete(tempBlock.id);
            next.set(createdBlock.id, pos);
          }
          return next;
        });
      } catch (err) {
        console.error('Failed to create block:', err);
        setError(err instanceof Error ? err.message : 'Failed to create block');

        // Rollback: remove temp block and restore positions
        setPendingCreates((prev) => {
          const next = new Map(prev);
          next.delete(tempBlock.id);
          return next;
        });

        setBlocks((prevBlocks) =>
          prevBlocks.map((b) =>
            b.position > tempBlock.position
              ? { ...b, position: b.position - 1 }
              : b
          )
        );
      }
    },
    [pageId]
  );

  // Create new block
  const createNewBlock = useCallback(
    (afterBlockId: string, type: string = 'text', content: string = '') => {
      if (!pageId) return;

      const allBlocks = getAllBlocks();
      const afterBlock = allBlocks.find((b) => b.id === afterBlockId);
      const afterIndex = afterBlock
        ? allBlocks.findIndex((b) => b.id === afterBlockId)
        : -1;

      const newPosition = afterIndex >= 0 ? afterBlock!.position + 1 : allBlocks.length;

      const tempId = generateTempId();
      const blockType = CONTENT_BLOCK_TYPES.find((t) => t.name === type) || CONTENT_BLOCK_TYPES[0];

      const newBlock: Block = {
        id: tempId,
        pageId: pageId,
        type: blockType.name,
        type_id: blockType.name,
        content: content,
        position: newPosition,
        metadata: blockType.defaultMetadata || {},
        blockConfig: blockType.defaultMetadata || {},
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      // Add to pending creates
      setPendingCreates((prev) => {
        const next = new Map(prev);
        next.set(tempId, newBlock);
        return next;
      });

      // Update positions of blocks after
      setBlocks((prevBlocks) => {
        return prevBlocks.map((b) => {
          if (b.position >= newPosition) {
            return { ...b, position: b.position + 1 };
          }
          return b;
        });
      });

      // Focus on new block and set cursor position
      setTimeout(() => {
        setFocusedBlockId(tempId);
        setCursorPositions((prev) => {
          const next = new Map(prev);
          next.set(tempId, 0); // Cursor at start of new block
          return next;
        });
        
        // Ensure editor is focused and cursor is visible
        requestAnimationFrame(() => {
          const blockElement = blockRefsMap.current.get(tempId);
          if (blockElement) {
            const editor = blockElement.querySelector<HTMLElement>('[contenteditable]');
            if (editor) {
              editor.focus();
              // Set cursor at start
              const range = document.createRange();
              const selection = globalThis.getSelection();
              if (editor.firstChild && editor.firstChild.nodeType === Node.TEXT_NODE) {
                range.setStart(editor.firstChild, 0);
                range.setEnd(editor.firstChild, 0);
              } else {
                range.setStart(editor, 0);
                range.setEnd(editor, 0);
              }
              selection?.removeAllRanges();
              selection?.addRange(range);
            }
          }
        });
      }, 0);
      
      // Optimistic create: persist to server for this single block
      void createBlockOnServer(newBlock, blockType);
    },
    [pageId, getAllBlocks, createBlockOnServer]
  );

  // Note: saveBlockToServer removed - all creates now go through batch sync queue
  // This ensures all operations are batched together for better performance

  // Handle content change
  const handleContentChange = useCallback((blockId: string, content: string) => {
    const allBlocks = getAllBlocks();
    const block = allBlocks.find((b) => b.id === blockId);
    if (!block) {
      // Block might have been deleted, ignore
      return;
    }

    if (block.id.startsWith('temp-')) {
      // Update pending block
      setPendingCreates((prev) => {
        const next = new Map(prev);
        const updated = { ...block, content };
        next.set(blockId, updated);
        return next;
      });
    } else {
      // Update existing block - check if it still exists
      const existingBlock = blocks.find((b) => b.id === blockId);
      if (!existingBlock) {
        // Block was deleted, ignore update
        return;
      }
      
      setBlocks((prevBlocks) =>
        prevBlocks.map((b) => (b.id === blockId ? { ...b, content } : b))
      );
    }
  }, [getAllBlocks, blocks]);

  // Handle block update - optimistic update + API call (debounced from InlineBlockEditor)
  const handleUpdateBlock = useCallback(
    async (blockId: string, data: { content?: string; type?: string; blockConfig?: Record<string, any> }) => {
      // Check if block exists
      const existingBlock = blocks.find((b) => b.id === blockId);
      if (!existingBlock || blockId.startsWith('temp-')) {
        // Block doesn't exist or is temp, ignore
        return;
      }

      // Keep snapshot for rollback
      const previousBlocks = blocks;

      // Optimistic update in local state
      setBlocks((prevBlocks) => 
        prevBlocks.map((b) =>
          b.id === blockId
            ? {
                ...b,
                ...(data.content !== undefined && { content: data.content }),
                ...(data.type !== undefined && { type: data.type, type_id: data.type }),
                ...(data.blockConfig !== undefined && { blockConfig: data.blockConfig, metadata: data.blockConfig }),
              }
            : b
        )
      );

      try {
        await blockApi.updateBlock(blockId, {
          content: data.content,
          type: data.type,
          blockConfig: data.blockConfig,
        });
      } catch (err) {
        // Rollback on failure
        console.error('Failed to update block:', err);
        setBlocks(previousBlocks);
        setError(err instanceof Error ? err.message : 'Failed to update block');
      }
    },
    [blocks]
  );

  // Handle Enter key
  const handleEnter = useCallback(
    (blockId: string) => {
      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      if (!block) return;

      // Save current cursor position in current block before creating new block
      const editor = blockRefsMap.current.get(blockId)?.querySelector<HTMLElement>('[contenteditable]');
      if (editor) {
        const selection = globalThis.getSelection();
        if (selection && selection.rangeCount > 0) {
          const range = selection.getRangeAt(0);
          let cursorPos = 0;
          if (range.startContainer.nodeType === Node.TEXT_NODE) {
            cursorPos = range.startOffset;
          } else {
            cursorPos = editor.textContent?.length || 0;
          }
          setCursorPositions((prev) => {
            const next = new Map(prev);
            next.set(blockId, cursorPos);
            return next;
          });
        }
      }

      // Create new block right after current block
      createNewBlock(blockId, 'text', '');
    },
    [createNewBlock, getAllBlocks]
  );

  // Handle Backspace on empty block
  const handleBackspace = useCallback(
    async (blockId: string) => {
      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      if (!block) return;

      const blockIndex = allBlocks.findIndex((b) => b.id === blockId);
      const prevBlock = blockIndex > 0 ? allBlocks[blockIndex - 1] : null;

      if (block.id.startsWith('temp-')) {
        // Remove pending block
        setPendingCreates((prev) => {
          const next = new Map(prev);
          next.delete(blockId);
          return next;
        });
      } else {
        // Optimistic delete - remove from UI immediately
        const previousBlocks = blocks;
        setBlocks((prevBlocks) => prevBlocks.filter((b) => b.id !== blockId));

        try {
          await blockApi.deleteBlock(blockId);
        } catch (err) {
          console.error('Failed to delete block:', err);
          setBlocks(previousBlocks);
        }
      }

      // Focus on previous block
      if (prevBlock) {
        setTimeout(() => {
          setFocusedBlockId(prevBlock.id);
          // Move cursor to end of previous block
          const prevContent = prevBlock.content || '';
          setCursorPositions((prev) => {
            const next = new Map(prev);
            next.set(prevBlock.id, prevContent.length);
            return next;
          });
        }, 0);
      }
    },
    [getAllBlocks, blocks]
  );

  // Handle Merge: Merge current block with previous block (Notion-style)
  const handleMerge = useCallback(
    async (blockId: string) => {
      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      if (!block) return;

      const blockIndex = allBlocks.findIndex((b) => b.id === blockId);
      const prevBlock = blockIndex > 0 ? allBlocks[blockIndex - 1] : null;

      if (!prevBlock) {
        // No previous block to merge with
        return;
      }

      // Merge content: previous block content + newline + current block content
      const prevContent = prevBlock.content || '';
      const currentContent = block.content || '';
      // Add newline between blocks if both have content
      const separator = prevContent && currentContent ? '\n' : '';
      const mergedContent = prevContent + separator + currentContent;

      // Calculate cursor position at merge point (end of previous content)
      const cursorPos = prevContent.length + separator.length;

      // Handle different block combinations
      if (block.id.startsWith('temp-')) {
        // Current block is temp - just remove it and update previous
        setPendingCreates((prev) => {
          const next = new Map(prev);
          next.delete(blockId);
          return next;
        });

        // Update previous block content
        if (prevBlock.id.startsWith('temp-')) {
          // Both are temp blocks
          setPendingCreates((prev) => {
            const next = new Map(prev);
            const updated = { ...prevBlock, content: mergedContent };
            next.set(prevBlock.id, updated);
            return next;
          });
        } else {
          // Previous is real block - optimistic update + API call
          const previousBlocks = blocks;
          setBlocks((prevBlocks) =>
            prevBlocks.map((b) =>
              b.id === prevBlock.id ? { ...b, content: mergedContent } : b
            )
          );

          try {
            await blockApi.updateBlock(prevBlock.id, { content: mergedContent });
          } catch (err) {
            console.error('Failed to merge blocks:', err);
            setBlocks(previousBlocks);
            return;
          }
        }
      } else {
        // Current block is real - need to update and delete
        if (prevBlock.id.startsWith('temp-')) {
          // Previous is temp, current is real
          // Update temp block with merged content
          setPendingCreates((prev) => {
            const next = new Map(prev);
            const updated = { ...prevBlock, content: mergedContent };
            next.set(prevBlock.id, updated);
            return next;
          });

          // Optimistic delete for current real block
          const previousBlocks = blocks;
          setBlocks((prevBlocks) => prevBlocks.filter((b) => b.id !== blockId));

          try {
            await blockApi.deleteBlock(blockId);
          } catch (err) {
            console.error('Failed to delete block during merge:', err);
            setBlocks(previousBlocks);
            return;
          }
        } else {
          // Both blocks are real - optimistic update & delete + API calls
          const previousBlocks = blocks;
          setBlocks((prevBlocks) => {
            const updated = prevBlocks
              .map((b) =>
                b.id === prevBlock.id ? { ...b, content: mergedContent } : b
              )
              .filter((b) => b.id !== blockId);
            return updated;
          });

          try {
            await blockApi.updateBlock(prevBlock.id, { content: mergedContent });
            await blockApi.deleteBlock(blockId);
          } catch (err) {
            console.error('Failed to merge blocks:', err);
            setBlocks(previousBlocks);
            return;
          }
        }
      }

      // Focus on previous block at merge point (end of previous content)
      setTimeout(() => {
        setFocusedBlockId(prevBlock.id);
        setCursorPositions((prev) => {
          const next = new Map(prev);
          next.set(prevBlock.id, cursorPos);
          return next;
        });
        
        // Ensure editor is focused and cursor is set
        requestAnimationFrame(() => {
          const blockElement = blockRefsMap.current.get(prevBlock.id);
          if (blockElement) {
            const editor = blockElement.querySelector<HTMLElement>('[contenteditable]');
            if (editor) {
              editor.focus();
              // Set cursor at merge point
              const range = document.createRange();
              const selection = globalThis.getSelection();
              const textContent = editor.textContent || '';
              const targetPos = Math.min(cursorPos, textContent.length);
              
              // Find text node and set cursor
              const walker = document.createTreeWalker(
                editor,
                NodeFilter.SHOW_TEXT,
                null
              );
              
              let currentPos = 0;
              let textNode: Text | null = null;
              let nodePos = 0;
              
              while (walker.nextNode()) {
                const node = walker.currentNode as Text;
                const nodeLength = node.length;
                
                if (currentPos + nodeLength >= targetPos) {
                  textNode = node;
                  nodePos = targetPos - currentPos;
                  break;
                }
                
                currentPos += nodeLength;
              }
              
              if (textNode) {
                range.setStart(textNode, nodePos);
                range.setEnd(textNode, nodePos);
              } else if (editor.firstChild && editor.firstChild.nodeType === Node.TEXT_NODE) {
                const firstNode = editor.firstChild as Text;
                range.setStart(firstNode, Math.min(targetPos, firstNode.length));
                range.setEnd(firstNode, Math.min(targetPos, firstNode.length));
              } else {
                range.setStart(editor, 0);
                range.setEnd(editor, 0);
              }
              
              selection?.removeAllRanges();
              selection?.addRange(range);
            }
          }
        });
      }, 0);
    },
    [getAllBlocks]
  );

  // Handle Arrow Up - move to previous block
  const handleArrowUp = useCallback(
    (blockId: string) => {
      const allBlocks = getAllBlocks();
      const blockIndex = allBlocks.findIndex((b) => b.id === blockId);
      if (blockIndex > 0) {
        const prevBlock = allBlocks[blockIndex - 1];
        // Save current cursor position
        const editor = blockRefsMap.current.get(blockId)?.querySelector<HTMLElement>('[contenteditable]');
        if (editor) {
          const selection = globalThis.getSelection();
          if (selection && selection.rangeCount > 0) {
            const range = selection.getRangeAt(0);
            let cursorPos = 0;
            if (range.startContainer.nodeType === Node.TEXT_NODE) {
              cursorPos = range.startOffset;
            } else {
              cursorPos = editor.textContent?.length || 0;
            }
            setCursorPositions((prev) => {
              const next = new Map(prev);
              next.set(blockId, cursorPos);
              return next;
            });
          }
        }
        
        // Move to previous block
        setTimeout(() => {
          setFocusedBlockId(prevBlock.id);
          // Restore or set cursor position in previous block
          const savedPos = cursorPositions.get(prevBlock.id);
          const prevContent = prevBlock.content || '';
          const targetPos = savedPos !== undefined ? savedPos : prevContent.length;
          setCursorPositions((prev) => {
            const next = new Map(prev);
            next.set(prevBlock.id, targetPos);
            return next;
          });
        }, 0);
      }
    },
    [getAllBlocks, cursorPositions]
  );

  // Handle Arrow Down - move to next block
  const handleArrowDown = useCallback(
    (blockId: string) => {
      const allBlocks = getAllBlocks();
      const blockIndex = allBlocks.findIndex((b) => b.id === blockId);
      if (blockIndex < allBlocks.length - 1) {
        const nextBlock = allBlocks[blockIndex + 1];
        // Save current cursor position
        const editor = blockRefsMap.current.get(blockId)?.querySelector<HTMLElement>('[contenteditable]');
        if (editor) {
          const selection = globalThis.getSelection();
          if (selection && selection.rangeCount > 0) {
            const range = selection.getRangeAt(0);
            let cursorPos = 0;
            if (range.startContainer.nodeType === Node.TEXT_NODE) {
              cursorPos = range.startOffset;
            } else {
              cursorPos = editor.textContent?.length || 0;
            }
            setCursorPositions((prev) => {
              const next = new Map(prev);
              next.set(blockId, cursorPos);
              return next;
            });
          }
        }
        
        // Move to next block
        setTimeout(() => {
          setFocusedBlockId(nextBlock.id);
          // Restore or set cursor position in next block
          const savedPos = cursorPositions.get(nextBlock.id);
          const nextContent = nextBlock.content || '';
          const targetPos = savedPos !== undefined ? savedPos : 0;
          setCursorPositions((prev) => {
            const next = new Map(prev);
            next.set(nextBlock.id, targetPos);
            return next;
          });
        }, 0);
      }
    },
    [getAllBlocks, cursorPositions]
  );

  // Handle type change
  const handleTypeChange = useCallback(
    async (blockId: string, newType: string) => {
      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      if (!block) {
        console.warn('Block not found for type change:', blockId);
        return;
      }

      if (block.id.startsWith('temp-')) {
        // Update pending block
        setPendingCreates((prev) => {
          const next = new Map(prev);
          const updated = { ...block, type: newType, type_id: newType };
          next.set(blockId, updated);
          return next;
        });
      } else {
        // Update on server - check if block still exists
        const existingBlock = blocks.find((b) => b.id === blockId);
        if (!existingBlock) {
          console.warn('Block not found in server blocks:', blockId);
          return;
        }

        // Optimistic update
        setBlocks((prevBlocks) =>
          prevBlocks.map((b) =>
            b.id === blockId ? { ...b, type: newType, type_id: newType } : b
          )
        );

        // Queue for batch sync
        if (syncQueueRef.current) {
          syncQueueRef.current.enqueue({
            id: `update-type-${blockId}-${Date.now()}`,
            type: 'update',
            blockId: blockId,
            data: {
              type: newType,
            },
            timestamp: Date.now(),
          });
        }
        // Note: No fallback - if queue is not available, optimistic update is enough
        // The queue will be initialized on mount, so this should rarely happen
      }
    },
    [getAllBlocks, blocks]
  );

  if (!pageId) {
    return (
      <section className="space-y-2">
        <Card className="p-4 text-sm text-muted-foreground">
          Select a page to view and edit content blocks.
        </Card>
      </section>
    );
  }

  const allBlocks = getAllBlocks();

  return (
    <section
      className="space-y-1"
      role="region"
      aria-label="Page content editor"
    >
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

      {isLoading && allBlocks.length === 0 && (
        <Card className="p-4 text-sm text-muted-foreground">
          Loading blocks...
        </Card>
      )}

      {allBlocks.length === 0 && !isLoading && (
        <div
          className="py-8 text-center text-gray-400 cursor-text"
          onClick={() => {
            createNewBlock('', 'text', '');
          }}
        >
          Click here to start typing...
        </div>
      )}

      {allBlocks.map((block, index) => (
        <div
          key={block.id}
          ref={(el) => {
            if (el) blockRefsMap.current.set(block.id, el);
          }}
          className="block-editor-item"
          style={{ 
            border: 'none', 
            outline: 'none',
            boxShadow: 'none',
          }}
        >
          <InlineBlockEditor
            block={block}
            isFocused={focusedBlockId === block.id}
            onContentChange={handleContentChange}
            onEnter={handleEnter}
            onBackspace={handleBackspace}
            onMerge={handleMerge}
            onTypeChange={handleTypeChange}
            onFocus={() => setFocusedBlockId(block.id)}
            onArrowUp={() => handleArrowUp(block.id)}
            onArrowDown={() => handleArrowDown(block.id)}
            autoFocus={index === allBlocks.length - 1 && block.id.startsWith('temp-')}
            restoreCursorPosition={cursorPositions.get(block.id) ?? null}
            onUpdateBlock={handleUpdateBlock}
          />
        </div>
      ))}

      {/* Empty state - create first block */}
      {allBlocks.length === 0 && !isLoading && (
        <div
          className="py-2 px-2 text-gray-400 cursor-text"
          onClick={() => {
            createNewBlock('', 'text', '');
          }}
        >
          <div className="opacity-0 hover:opacity-100 transition-opacity">
            Type to start...
          </div>
        </div>
      )}
    </section>
  );
}

