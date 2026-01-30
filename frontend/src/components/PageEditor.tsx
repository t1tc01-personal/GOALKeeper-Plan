'use client';

import React, { useEffect, useState, useRef, useCallback } from 'react';
import { Card } from './ui/card';
import { InlineBlockEditor } from './InlineBlockEditor';
import { blockApi, type Block } from '@/services/blockApi';
import { BlockSyncQueue, type BatchSyncResponse } from '@/services/blockSyncQueue';
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
  const pendingCreatesRef = useRef<Map<string, Block>>(new Map()); // Ref to access latest pending blocks in callbacks
  const blockRefsMap = useRef<Map<string, HTMLDivElement>>(new Map());
  const [cursorPositions, setCursorPositions] = useState<Map<string, number>>(new Map());
  const [draggedBlockId, setDraggedBlockId] = useState<string | null>(null);
  const [dragOverBlockId, setDragOverBlockId] = useState<string | null>(null);
  const syncQueueRef = useRef<BlockSyncQueue | null>(null);
  const tempIdMapRef = useRef<Map<string, string>>(new Map()); // Maps temp IDs to real block IDs

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

  // Helper: Calculate list number for numbered_list by counting consecutive numbered_list blocks before current
  const calculateListNumber = useCallback((blockId: string): number => {
    const allBlocks = getAllBlocks();
    const blockIndex = allBlocks.findIndex(b => b.id === blockId);
    if (blockIndex === -1) return 1;

    const currentBlock = allBlocks[blockIndex];
    const currentType = currentBlock.type_id || currentBlock.type;

    // Chỉ tính nếu là numbered_list
    if (currentType !== 'numbered_list') return 1;

    let count = 1;
    // Count consecutive numbered_list blocks before current
    for (let i = blockIndex - 1; i >= 0; i--) {
      const prevType = allBlocks[i].type_id || allBlocks[i].type;
      if (prevType === 'numbered_list') {
        count++;
      } else {
        break; // Stop at first non-numbered-list block
      }
    }
    return count;
  }, [getAllBlocks]);

  // Handle batch sync success - update state with real block IDs
  const handleSyncSuccess = useCallback((response: BatchSyncResponse) => {
    // Process creates - map temp IDs to real IDs
    if (response.creates && response.creates.length > 0) {
      const tempToRealMap = new Map<string, string>();

      response.creates.forEach((createResult) => {
        tempToRealMap.set(createResult.tempId, createResult.block.id);
        tempIdMapRef.current.set(createResult.tempId, createResult.block.id);
      });

      // GIẢI PHÁP: Cập nhật tất cả state liên quan đến ID transition trong cùng một "batch" (React tự động làm điều này)
      // Nhưng quan trọng hơn là thứ tự và tính toàn vẹn của dữ liệu

      // 1. Add created blocks to blocks list with type_id
      setBlocks((prevBlocks) => {
        const newBlocks = [...prevBlocks];
        response.creates!.forEach((createResult) => {
          // Tránh thêm trùng lặp nếu block đã tồn tại
          if (!newBlocks.some(b => b.id === createResult.block.id)) {
            // Lấy thông tin block gốc từ ref để dự phòng trường hợp server trả về thiếu data (chưa restart backend)
            const originalBlock = pendingCreatesRef.current.get(createResult.tempId);

            newBlocks.push({
              id: createResult.block.id,
              pageId: createResult.block.pageId,
              // FALLBACK: Ưu tiên type từ server, nhưng nếu rỗng thì dùng type từ block gốc
              type: createResult.block.type || originalBlock?.type || 'text',
              type_id: createResult.block.type || originalBlock?.type || 'text',
              content: createResult.block.content,
              position: createResult.block.position,
              blockConfig: createResult.block.blockConfig || originalBlock?.blockConfig,
              created_at: createResult.block.created_at,
              updated_at: createResult.block.updated_at,
            });
          }
        });
        return newBlocks.sort((a, b) => a.position - b.position);
      });

      // 2. Remove temp blocks
      setPendingCreates((prev) => {
        const next = new Map(prev);
        response.creates!.forEach((createResult) => {
          next.delete(createResult.tempId);
        });
        pendingCreatesRef.current = next; // Sync ref
        return next;
      });

      // 3. Update cursor position map with real IDs
      setCursorPositions((prev) => {
        const next = new Map(prev);
        tempToRealMap.forEach((realId, tempId) => {
          if (next.has(tempId)) {
            next.set(realId, next.get(tempId)!);
            next.delete(tempId);
          }
        });
        return next;
      });

      // 4. Update focused block if it was temp
      setFocusedBlockId((prev) => {
        if (prev && tempToRealMap.has(prev)) {
          return tempToRealMap.get(prev)!;
        }
        return prev;
      });
    }

    // Process updates - state should already be updated optimistically
    if (response.updates && response.updates.length > 0) {
      // Verify updates succeeded, state is already optimistically updated
      response.updates.forEach((updateResult) => {
        console.log('Block updated:', updateResult.id);
      });
    }

    // Process deletes - state should already be updated optimistically
    if (response.deletes && response.deletes.length > 0) {
      // Verify deletes succeeded, state is already optimistically updated
      response.deletes.forEach((blockId) => {
        console.log('Block deleted:', blockId);
      });
    }

    // Handle errors if any
    if (response.errors && response.errors.length > 0) {
      const errorMessages = response.errors.map((e) => `${e.type} (${e.operationId}): ${e.error}`).join('; ');
      setError(`Some operations failed: ${errorMessages}`);
    }
  }, []);

  // Initialize sync queue with callbacks
  useEffect(() => {
    if (!syncQueueRef.current) {
      syncQueueRef.current = new BlockSyncQueue({
        syncInterval: 100, // Debounce for 100ms - almost immediate persistence
        maxBatchSize: 50, // Max 50 operations per batch
        maxRetries: 5, // Retry failed ops up to 5 times
        baseRetryDelay: 1000,
        maxRetryDelay: 30000,
      });

      // Set sync callbacks
      syncQueueRef.current.setSyncCallbacks({
        onSyncSuccess: handleSyncSuccess,
        onSyncError: (error: Error) => {
          console.error('Sync queue error:', error);
          setError(error.message || 'Failed to sync blocks');
        },
      });
    }

    return () => {
      // Cleanup on unmount
      if (syncQueueRef.current) {
        syncQueueRef.current.destroy();
        syncQueueRef.current = null;
      }
    };
  }, [handleSyncSuccess]);

  // Force sync on page unload to prevent data loss
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (syncQueueRef.current && !syncQueueRef.current.isEmpty()) {
        // Try to force sync (may not complete in time, but attempt it)
        syncQueueRef.current.forceSync();

        // Show warning if there are pending changes
        e.preventDefault();
        e.returnValue = '';
      }
    };

    const handleVisibilityChange = () => {
      if (document.hidden && syncQueueRef.current && !syncQueueRef.current.isEmpty()) {
        // Force sync when page becomes hidden
        syncQueueRef.current.forceSync();
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);

  // Cleanup block refs when blocks are deleted to prevent memory leaks
  useEffect(() => {
    const currentBlockIds = new Set(getAllBlocks().map(b => b.id));
    for (const [blockId] of blockRefsMap.current.entries()) {
      if (!currentBlockIds.has(blockId)) {
        blockRefsMap.current.delete(blockId);
      }
    }
  }, [blocks, pendingCreates, getAllBlocks]);

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

      // Add to pending creates AND update positions of existing blocks in one go
      // Note: React 18 batches these, but consolidating logic ensures consistency
      setPendingCreates((prev) => {
        const next = new Map(prev);
        next.set(tempId, newBlock);
        pendingCreatesRef.current = next; // Sync ref
        return next;
      });

      setBlocks((prevBlocks) => {
        return prevBlocks.map((b) => {
          if (b.position >= newPosition) {
            return { ...b, position: b.position + 1 };
          }
          return b;
        });
      });

      // Focus on new block immediately to prevent placeholder flicker
      const isListType = type === 'numbered_list' || type === 'bulleted_list';

      // Batch setFocusedBlockId with other state updates if possible
      // but creation and focus usually happen together
      setFocusedBlockId(tempId);
      setCursorPositions((prev) => {
        const next = new Map(prev);
        next.set(tempId, 0); // Cursor at start of new block
        return next;
      });

      // Ensure editor is focused and cursor is visible
      requestAnimationFrame(() => {
        // Try multiple times for list blocks (structure is more complex)
        const tryFocus = (attempts = 0) => {
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
            } else if (attempts < 3) {
              // Retry if editor not found (DOM might not be ready yet)
              setTimeout(() => tryFocus(attempts + 1), 10);
            }
          } else if (attempts < 5) { // More attempts for list items
            // Retry if block element not found
            setTimeout(() => tryFocus(attempts + 1), 10);
          }
        };

        tryFocus();
      });

      // Enqueue create operation to sync queue
      if (syncQueueRef.current) {
        syncQueueRef.current.enqueue({
          id: tempId,
          type: 'create',
          blockId: tempId,
          data: {
            pageId: pageId,
            type: blockType.name,
            content: content,
            position: newPosition,
            blockConfig: blockType.defaultMetadata || {},
          },
          timestamp: Date.now(),
          priority: 'normal',
        });
      }
    },
    [pageId, getAllBlocks]
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
    (blockId: string, data: { content?: string; type?: string; blockConfig?: Record<string, any> }) => {
      // GIẢI PHÁP: Luôn sử dụng prevBlocks để đảm bảo lấy dữ liệu mới nhất trong hàng đợi state
      setBlocks((prevBlocks) => {
        const existingBlock = prevBlocks.find((b) => b.id === blockId);

        // Nếu không thấy trong blocks, có thể nó nằm trong pendingCreates (temp block)
        if (!existingBlock) {
          if (blockId.startsWith('temp-')) {
            setPendingCreates((prevPending) => {
              const next = new Map(prevPending);
              const block = next.get(blockId);
              if (block) {
                next.set(blockId, {
                  ...block,
                  ...(data.content !== undefined && { content: data.content }),
                  ...(data.type !== undefined && { type: data.type, type_id: data.type }),
                  ...(data.blockConfig !== undefined && { blockConfig: data.blockConfig, metadata: data.blockConfig }),
                });
              }
              return next;
            });
          }
          return prevBlocks;
        }

        // Cập nhật block thật với dữ liệu mới nhất từ prevBlocks
        return prevBlocks.map((b) =>
          b.id === blockId
            ? {
              ...b,
              ...(data.content !== undefined && { content: data.content }),
              ...(data.type !== undefined && { type: data.type, type_id: data.type }),
              ...(data.blockConfig !== undefined && { blockConfig: data.blockConfig, metadata: data.blockConfig }),
            }
            : b
        );
      });

      // Enqueue update operation
      if (syncQueueRef.current) {
        syncQueueRef.current.enqueue({
          id: `update-${blockId}-${Date.now()}`,
          type: 'update',
          blockId: blockId,
          data: {
            content: data.content,
            type: data.type,
            blockConfig: data.blockConfig,
          },
          timestamp: Date.now(),
          priority: 'normal',
        });
      }
    },
    []
  );

  // Handle Enter key
  const handleEnter = useCallback(
    (blockId: string) => {
      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      if (!block) return;

      // Lấy nội dung thực tế trực tiếp từ DOM trước khi tạo dòng mới
      const blockElement = blockRefsMap.current.get(blockId);
      const editor = blockElement?.querySelector<HTMLElement>('[contenteditable]');
      const currentActualContent = editor?.textContent || '';

      // CẬP NHẬT NGAY LẬP TỨC: Đảm bảo nội dung dòng hiện tại không bị mất
      handleUpdateBlock(blockId, { content: currentActualContent });

      if (editor) {
        const selection = globalThis.getSelection();
        if (selection && selection.rangeCount > 0) {
          const range = selection.getRangeAt(0);
          let cursorPos = 0;
          if (range.startContainer.nodeType === Node.TEXT_NODE) {
            cursorPos = range.startOffset;
          } else {
            cursorPos = currentActualContent.length;
          }
          setCursorPositions((prev) => {
            const next = new Map(prev);
            next.set(blockId, cursorPos);
            return next;
          });
        }
      }

      const currentType = block.type_id || block.type || 'text';

      // Notion-like behavior for lists:
      // - If empty list item + Enter → create text block (break out of list)
      // - If list item with content + Enter → create new list item
      let newBlockType = 'text'; // Default to text

      if (currentType === 'numbered_list' || currentType === 'bulleted_list') {
        // Check if current list item has content
        if (currentActualContent && currentActualContent.trim().length > 0) {
          // Has content → create new list item of same type
          newBlockType = currentType;
        } else {
          // Empty list item → break out to text
          newBlockType = 'text';
        }
      }

      createNewBlock(blockId, newBlockType, '');
    },
    [createNewBlock, getAllBlocks, handleUpdateBlock] // Phải có handleUpdateBlock ở đây
  );

  // Handle Backspace on empty block
  const handleBackspace = useCallback(
    (blockId: string) => {
      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      if (!block) return;

      const blockIndex = allBlocks.findIndex((b) => b.id === blockId);
      const prevBlock = blockIndex > 0 ? allBlocks[blockIndex - 1] : null;

      if (block.id.startsWith('temp-')) {
        // Remove pending block - single state update
        setPendingCreates((prev) => {
          const next = new Map(prev);
          next.delete(blockId);
          return next;
        });

        // CRITICAL: Cancel the create operation in sync queue
        if (syncQueueRef.current) {
          syncQueueRef.current.dequeue(blockId);
        }
      } else {
        // Optimistic delete - single state update combining all changes
        setBlocks((prevBlocks) => prevBlocks.filter((b) => b.id !== blockId));

        // Enqueue delete operation - NO forceSync - let queue debounce naturally
        if (syncQueueRef.current) {
          syncQueueRef.current.enqueue({
            id: `delete-${blockId}-${Date.now()}`,
            type: 'delete',
            blockId: blockId,
            timestamp: Date.now(),
            priority: 'high', // Use high priority but let queue debounce
          });
          // REMOVED: forceSync() - let natural debouncing handle it
        }
      }

      // Focus on previous block - combine state updates
      if (prevBlock) {
        const prevContent = prevBlock.content || '';
        setFocusedBlockId(prevBlock.id);
        setCursorPositions((prev) => {
          const next = new Map(prev);
          next.set(prevBlock.id, prevContent.length);
          return next;
        });

        // Use queueMicrotask instead of setTimeout(0)
        queueMicrotask(() => {
          const blockElement = blockRefsMap.current.get(prevBlock.id);
          if (blockElement) {
            const editor = blockElement.querySelector<HTMLElement>('[contenteditable]');
            editor?.focus();
          }
        });
      }
    },
    [getAllBlocks]
  );

  // Handle Merge: Merge current block with previous block (Notion-style)
  const handleMerge = useCallback(
    (blockId: string) => {
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

        // Cancel the create operation in sync queue
        if (syncQueueRef.current) {
          syncQueueRef.current.dequeue(blockId);
        }

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
          // Previous is real block - optimistic update + enqueue
          setBlocks((prevBlocks) =>
            prevBlocks.map((b) =>
              b.id === prevBlock.id ? { ...b, content: mergedContent } : b
            )
          );

          // Enqueue update operation
          if (syncQueueRef.current) {
            syncQueueRef.current.enqueue({
              id: `update-${prevBlock.id}-${Date.now()}`,
              type: 'update',
              blockId: prevBlock.id,
              data: {
                content: mergedContent,
              },
              timestamp: Date.now(),
              priority: 'normal',
            });
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

          // Update the create operation in sync queue with new content
          if (syncQueueRef.current) {
            syncQueueRef.current.enqueue({
              id: prevBlock.id,
              type: 'create',
              blockId: prevBlock.id,
              data: {
                pageId: prevBlock.pageId,
                type: prevBlock.type,
                content: mergedContent,
                position: prevBlock.position,
                blockConfig: prevBlock.blockConfig || {},
              },
              timestamp: Date.now(),
              priority: 'normal',
            });
          }

          // Optimistic delete for current real block
          setBlocks((prevBlocks) => prevBlocks.filter((b) => b.id !== blockId));

          // Enqueue delete operation
          if (syncQueueRef.current) {
            syncQueueRef.current.enqueue({
              id: `delete-${blockId}-${Date.now()}`,
              type: 'delete',
              blockId: blockId,
              timestamp: Date.now(),
              priority: 'normal',
            });
          }
        } else {
          // Both blocks are real - optimistic update & delete + enqueue both
          setBlocks((prevBlocks) => {
            const updated = prevBlocks
              .map((b) =>
                b.id === prevBlock.id ? { ...b, content: mergedContent } : b
              )
              .filter((b) => b.id !== blockId);
            return updated;
          });

          // Enqueue both operations
          if (syncQueueRef.current) {
            syncQueueRef.current.enqueue({
              id: `update-${prevBlock.id}-${Date.now()}`,
              type: 'update',
              blockId: prevBlock.id,
              data: {
                content: mergedContent,
              },
              timestamp: Date.now(),
              priority: 'normal',
            });

            syncQueueRef.current.enqueue({
              id: `delete-${blockId}-${Date.now()}`,
              type: 'delete',
              blockId: blockId,
              timestamp: Date.now(),
              priority: 'normal',
            });
          }
        }
      }

      // Focus on previous block at merge point (end of previous content)
      setFocusedBlockId(prevBlock.id);
      setCursorPositions((prev) => {
        const next = new Map(prev);
        next.set(prevBlock.id, cursorPos);
        return next;
      });

      // Use queueMicrotask instead of setTimeout(0)
      queueMicrotask(() => {
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
      });
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

        // Move to previous block - use queueMicrotask
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

        // Move to next block - use queueMicrotask
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

  // Handle drag start
  const handleDragStart = useCallback(
    (blockId: string) => {
      setDraggedBlockId(blockId);
    },
    []
  );

  // Handle drag over
  const handleDragOver = useCallback(
    (e: React.DragEvent<HTMLDivElement>, blockId: string) => {
      e.preventDefault();
      e.stopPropagation();
      setDragOverBlockId(blockId);
    },
    []
  );

  // Handle drag leave
  const handleDragLeave = useCallback(
    (e: React.DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      e.stopPropagation();
      setDragOverBlockId(null);
    },
    []
  );

  // Handle drop
  const handleDrop = useCallback(
    (e: React.DragEvent<HTMLDivElement>, dropBlockId: string) => {
      e.preventDefault();
      e.stopPropagation();
      setDragOverBlockId(null);

      if (!draggedBlockId || draggedBlockId === dropBlockId) {
        setDraggedBlockId(null);
        return;
      }

      const allBlocks = getAllBlocks();
      const draggedBlock = allBlocks.find((b) => b.id === draggedBlockId);
      const dropBlock = allBlocks.find((b) => b.id === dropBlockId);

      if (!draggedBlock || !dropBlock) {
        setDraggedBlockId(null);
        return;
      }

      // Calculate new positions
      const draggedIndex = allBlocks.findIndex((b) => b.id === draggedBlockId);
      const dropIndex = allBlocks.findIndex((b) => b.id === dropBlockId);

      if (draggedIndex === dropIndex) {
        setDraggedBlockId(null);
        return;
      }

      // Reorder blocks
      const newBlocks = [...allBlocks];
      const [movedBlock] = newBlocks.splice(draggedIndex, 1);
      newBlocks.splice(dropIndex, 0, movedBlock);

      // Recalculate positions (100, 200, 300, etc.)
      const reorderedBlocks = newBlocks.map((block, index) => ({
        ...block,
        position: (index + 1) * 100,
      }));

      // Separate real and pending blocks
      const realBlocks = reorderedBlocks.filter((b) => !b.id.startsWith('temp-'));
      const pendingBlocks = reorderedBlocks.filter((b) => b.id.startsWith('temp-'));

      // Update UI optimistically
      setBlocks(realBlocks);
      if (pendingBlocks.length > 0) {
        setPendingCreates(
          new Map(pendingBlocks.map((b) => [b.id, b]))
        );
      }

      // Enqueue position updates for real blocks
      if (syncQueueRef.current) {
        realBlocks.forEach((block) => {
          syncQueueRef.current!.enqueue({
            id: `update-position-${block.id}-${Date.now()}`,
            type: 'update',
            blockId: block.id,
            data: {
              position: block.position,
            },
            timestamp: Date.now(),
            priority: 'high', // Higher priority for drag operations
          });
        });
      }

      setDraggedBlockId(null);
    },
    [draggedBlockId, getAllBlocks]
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
          draggable={true}
          onDragStart={() => handleDragStart(block.id)}
          onDragOver={(e) => handleDragOver(e, block.id)}
          onDragLeave={(e) => handleDragLeave(e)}
          onDrop={(e) => handleDrop(e, block.id)}
          className={`block-editor-item transition-all duration-150 ${draggedBlockId === block.id ? 'opacity-50 cursor-grabbing' : ''
            } ${dragOverBlockId === block.id ? 'border-t-2 border-blue-400 bg-blue-50' : ''
            }`}
          style={{
            border: dragOverBlockId === block.id ? 'none' : 'none',
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
            listNumber={calculateListNumber(block.id)}
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

