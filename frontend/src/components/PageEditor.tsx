'use client';

import React, { useEffect, useState, useRef, useCallback } from 'react';
import { Card } from './ui/card';
import { InlineBlockEditor } from './InlineBlockEditor';
import { BlockWithChildren } from './BlockWithChildren';
import { blockApi, type Block } from '@/services/blockApi';
import { BlockSyncManager, type BatchSyncResponse } from '@/services/blockSyncQueue';
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
  const syncQueueRef = useRef<BlockSyncManager | null>(null);
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

        // Remap pending operations in queue (e.g. updates that happened while create was in flight)
        if (syncQueueRef.current) {
          syncQueueRef.current.remapBlockId(createResult.tempId, createResult.block.id);
        }
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
              // CRITICAL FIX: Ensure parent_block_id is preserved
              parent_block_id: createResult.block.parent_block_id || originalBlock?.parent_block_id,
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
      syncQueueRef.current = new BlockSyncManager({
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
      const syncManager = syncQueueRef.current;
      if (syncManager) {
        // Clear ref immediately so next mount creates a new instance
        syncQueueRef.current = null;

        // Try to sync pending changes before destroying
        // Note: This is fire-and-forget on unmount
        syncManager.forceSync();

        // Destroy instance after a small delay to allow sync to start
        setTimeout(() => {
          syncManager.destroy();
        }, 0);
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
    (afterBlockId: string, type: string = 'text', content: string = '', parentBlockId?: string | null) => {
      if (!pageId) return;

      const allBlocks = getAllBlocks();
      const afterBlock = allBlocks.find((b) => b.id === afterBlockId);

      // Determine parent ID: use explicit arg if provided, otherwise inherit from sibling
      // If parentBlockId is undefined (not passed), we use afterBlock's parent (sibling logic)
      // If it is null (explicitly root), we use null.
      const targetParentId = parentBlockId !== undefined ? parentBlockId : (afterBlock?.parent_block_id || undefined);

      console.log('[createNewBlock] Position calculation:', {
        afterBlockId,
        afterBlockParent: afterBlock?.parent_block_id,
        targetParentId,
        isToggleChild: targetParentId === afterBlockId
      });

      // Calculate new position within the target parent scope
      let newPosition = 0;

      // Filter siblings in the target scope
      const siblings = allBlocks.filter(b => b.parent_block_id === targetParentId);
      console.log('[createNewBlock] Siblings in target scope:', siblings.length, siblings.map(b => ({ id: b.id, pos: b.position })));

      if (afterBlock && afterBlock.parent_block_id === targetParentId) {
        // Inserting after a sibling
        newPosition = (afterBlock.position || 0) + 100; // Use gap of 100 to avoid excessive shifting
        // In a real implementation we might want to shift subsequent blocks if gap is too small,
        // but for now let's try to assume we can just +1 or +100.
        // BUT wait, existing logic used +1 and shifted everyone.
        // Let's stick to +1 shift logic but SCOPED to parent.
        newPosition = (afterBlock.position || 0) + 1;
        console.log('[createNewBlock] Inserting after sibling, position:', newPosition);
      } else {
        // Inserting as first child (or appending if no afterBlock matches)
        if (siblings.length > 0) {
          // Append to end
          const maxPos = Math.max(...siblings.map(b => b.position || 0));
          newPosition = maxPos + 1;
          console.log('[createNewBlock] Appending to end, position:', newPosition);
        } else {
          newPosition = 0;
          console.log('[createNewBlock] First child, position:', newPosition);
        }
      }

      const tempId = generateTempId();
      const blockType = CONTENT_BLOCK_TYPES.find((t) => t.name === type) || CONTENT_BLOCK_TYPES[0];

      // Handle framework types: map specific framework names (kanban, habit) to the generic 'framework_container'
      // type expected by the backend, while preserving the specific type in metadata.
      let backendType = blockType.name;
      if (['kanban', 'habit'].includes(blockType.name)) {
        backendType = 'framework_container';
      }

      const newBlock: Block = {
        id: tempId,
        pageId: pageId,
        type: backendType, // Use backend-compatible type
        type_id: backendType,
        content: content,
        position: newPosition,
        parent_block_id: targetParentId || undefined, // undefined in interface vs null in DB? Check Block interface.
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
          // Only shift blocks in the SAME parent scope
          if (b.parent_block_id === targetParentId && (b.position || 0) >= newPosition) {
            return { ...b, position: (b.position || 0) + 1 };
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
      console.log('[PageEditor] Enqueueing CREATE for:', tempId, { pageId, type: backendType });
      if (syncQueueRef.current) {
        syncQueueRef.current.enqueue({
          id: tempId,
          type: 'create',
          blockId: tempId,
          data: {
            pageId: pageId,
            type: backendType, // Send generic type to backend
            content: content,
            position: newPosition,
            parent_block_id: targetParentId || undefined,
            blockConfig: blockType.defaultMetadata || {},
          },
          timestamp: Date.now(),
          priority: 'normal',
        });
      }

      return newBlock as Block;
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
    (blockId: string, data: { content?: string; type?: string; blockConfig?: Record<string, any>; parent_block_id?: string | null }) => {
      console.log('[DEBUG handleUpdateBlock] blockId:', blockId, 'data:', data);

      if (data.content !== undefined && data.content === '') {
        console.warn('[DEBUG handleUpdateBlock] ⚠️ Content is empty string! BlockId:', blockId);
      }

      setBlocks((prevBlocks) => {
        const existingBlock = prevBlocks.find((b) => b.id === blockId);

        if (!existingBlock) {
          if (blockId.startsWith('temp-')) {
            setPendingCreates((prevPending) => {
              const next = new Map(prevPending);
              const block = next.get(blockId);
              if (block) {
                const isFrameworkBlock = block.type === 'framework_container' || block.type_id === 'framework_container';
                next.set(blockId, {
                  ...block,
                  ...(data.content !== undefined && { content: data.content }),
                  ...(data.type !== undefined && !isFrameworkBlock && { type: data.type, type_id: data.type }),
                  ...(data.blockConfig !== undefined && { blockConfig: data.blockConfig, metadata: data.blockConfig }),
                });
              }
              return next;
            });
          }
          return prevBlocks;
        }

        if (data.blockConfig !== undefined) {
          console.log('[handleUpdateBlock] Updating blockConfig for', blockId);
          console.log('[handleUpdateBlock] Old blockConfig:', existingBlock.blockConfig);
          console.log('[handleUpdateBlock] New blockConfig:', data.blockConfig);
        }

        const isFrameworkBlock = existingBlock.type === 'framework_container' || existingBlock.type_id === 'framework_container';

        if (isFrameworkBlock && data.type && data.type !== 'framework_container') {
          console.warn('[handleUpdateBlock] ⚠️ Preventing type change for framework_container block!', {
            blockId,
            currentType: existingBlock.type,
            attemptedType: data.type,
            blockConfig: data.blockConfig
          });
        }

        return prevBlocks.map((b) =>
          b.id === blockId
            ? {
              ...b,
              ...(data.content !== undefined && { content: data.content }),
              ...(data.type !== undefined && !isFrameworkBlock && { type: data.type, type_id: data.type }),
              ...(data.blockConfig !== undefined && { blockConfig: data.blockConfig, metadata: data.blockConfig }),
            }
            : b
        );
      });

      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      const isFrameworkBlock = block && (block.type === 'framework_container' || block.type_id === 'framework_container');

      if (syncQueueRef.current) {
        syncQueueRef.current.enqueue({
          id: `update-${blockId}-${Date.now()}`,
          type: 'update',
          blockId: blockId,
          data: {
            content: data.content,
            type: (!isFrameworkBlock && data.type) ? data.type : undefined,
            parent_block_id: data.parent_block_id,
            blockConfig: data.blockConfig,
          },
          timestamp: Date.now(),
          priority: 'normal',
        });
      }
    },
    [getAllBlocks]
  );

  // Handle Indent (Tab)
  const handleIndent = useCallback((blockId: string) => {
    const allBlocks = getAllBlocks();
    const block = allBlocks.find((b) => b.id === blockId);
    if (!block) return;

    // Find siblings (same parent) sorted by position
    const siblings = allBlocks
      .filter((b) => b.parent_block_id === block.parent_block_id)
      .sort((a, b) => (a.position || 0) - (b.position || 0));

    const currentIndex = siblings.findIndex((b) => b.id === blockId);

    // Cannot indent if first child
    if (currentIndex <= 0) return;

    const prevSibling = siblings[currentIndex - 1];

    console.log(`[handleIndent] Indenting block ${blockId} into new parent ${prevSibling.id}`);

    // Update parent_block_id
    handleUpdateBlock(blockId, {
      parent_block_id: prevSibling.id
    });

    // Expand the new parent if it's a collapsed toggle
    if ((prevSibling.type === 'toggle' || prevSibling.type_id === 'toggle') && prevSibling.blockConfig?.collapsed) {
      if (process.env.NODE_ENV === 'development') {
        console.log('[Indent Auto-Expand]', {
          toggleId: prevSibling.id,
          wasCollapsed: true,
          expandingForChild: blockId,
        });
      }
      handleUpdateBlock(prevSibling.id, {
        blockConfig: { ...prevSibling.blockConfig, collapsed: false }
      });
    }

  }, [getAllBlocks, handleUpdateBlock]);

  // Handle Outdent (Shift+Tab)
  const handleOutdent = useCallback((blockId: string) => {
    const allBlocks = getAllBlocks();
    const block = allBlocks.find((b) => b.id === blockId);
    if (!block) return;

    // Cannot outdent if at root level (no parent)
    if (!block.parent_block_id) return;

    // Find current parent to get its parent (grandparent)
    const currentParent = allBlocks.find(b => b.id === block.parent_block_id);
    const newParentId = currentParent ? currentParent.parent_block_id : null;

    console.log(`[handleOutdent] Outdenting block ${blockId} to new parent ${newParentId}`);

    handleUpdateBlock(blockId, {
      parent_block_id: newParentId
    });

  }, [getAllBlocks, handleUpdateBlock]);

  // Handle Enter key
  const handleEnter = useCallback(
    (blockId: string, contentFromEditor?: string) => {
      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      if (!block) return;

      // PRIORITY 1: Content passed directly from the editor component (Most reliable)
      let currentActualContent = contentFromEditor;

      const blockElement = blockRefsMap.current.get(blockId);
      const editor = blockElement?.querySelector<HTMLElement>('[contenteditable]');

      // PRIORITY 2: If not passed, try to read from DOM (Legacy/Fallback)
      if (currentActualContent === undefined) {
        currentActualContent = editor?.textContent || '';
      }

      // SAFETY CHECK: If content is empty but block state has content, 
      // check if we should fallback to state to prevent data loss.
      // We accept empty content ONLY if it was explicitly passed as empty string (user deleted text),
      // BUT if we are in a "lost text" race condition, relying on block.content is safer.
      if (currentActualContent === '' && block.content && block.content.length > 0) {
        // If contentFromEditor was explicitly "", user might have just deleted it?
        // But usually hitting Enter on empty content -> empty content.
        // Hitting Enter on "Hello" -> should be "Hello".
        // If we received "" but state is "Hello", it's suspicious.
        // Assume fallback.
        currentActualContent = block.content;
      }

      // Ensure string
      if (currentActualContent === undefined) currentActualContent = '';

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

      // CRITICAL FIX: Notion-like behavior for nested blocks
      // Default: new block is sibling (same parent as current block)
      let newParentId = block.parent_block_id;

      // SPECIAL CASE: Enter inside Expanded Toggle -> Create Child (not sibling)
      const isToggle = currentType === 'toggle';
      const isExpanded = isToggle && !(block.blockConfig?.collapsed || block.metadata?.collapsed);

      if (isExpanded) {
        // User pressed Enter inside an expanded toggle
        // New block should be CHILD of toggle
        newParentId = block.id;
        console.log('[Enter in Toggle]', {
          toggleId: block.id,
          creatingChildWith: { parent: newParentId },
          timestamp: new Date().toISOString(),
        });
      }
      // ELSE: Current block is NOT a toggle (or is collapsed)
      // Keep newParentId = block.parent_block_id
      // This creates a SIBLING at the same nesting level

      // 2. Enter on Empty Block inside Valid Parent -> Outdent (Break out)
      if (block.parent_block_id && currentActualContent === '' && currentType === 'text') {
        handleOutdent(blockId);
        return;
      }

      // Notion-like behavior for lists:
      // - If empty list item + Enter → create text block (break out of list)
      // - If list item with content + Enter → create new list item
      let newBlockType = 'text'; // Default to text

      if (currentType === 'numbered_list' || currentType === 'bulleted_list' || currentType === 'todo_list') {
        // Check if current list item has content
        if (currentActualContent && currentActualContent.trim().length > 0) {
          // Has content → create new list item of same type
          newBlockType = currentType;
        } else {
          // Empty list item → break out to text
          newBlockType = 'text';
        }
      }

      console.log('[DEBUG handleEnter] Creating new block with type:', newBlockType, 'parent:', newParentId);

      // CRITICAL FIX: Use queueMicrotask to ensure handleUpdateBlock state update completes
      // before creating new block. This prevents content loss during the transition.
      queueMicrotask(() => {
        createNewBlock(blockId, newBlockType, '', newParentId);
      });
    },
    [createNewBlock, getAllBlocks, handleUpdateBlock, handleOutdent] // Phải có handleUpdateBlock ở đây
  );

  // Handle Backspace on empty block
  const handleBackspace = useCallback(
    (blockId: string) => {
      console.log('[handleBackspace] Called with blockId:', blockId);

      const allBlocks = getAllBlocks();
      const block = allBlocks.find((b) => b.id === blockId);
      if (!block) {
        console.log('[handleBackspace] Block not found, returning');
        return;
      }

      const blockIndex = allBlocks.findIndex((b) => b.id === blockId);
      const prevBlock = blockIndex > 0 ? allBlocks[blockIndex - 1] : null;

      console.log('[handleBackspace] Block info:', {
        blockId,
        blockIndex,
        totalBlocks: allBlocks.length,
        hasPrevBlock: !!prevBlock,
        prevBlockId: prevBlock?.id,
        isFirstBlock: blockIndex === 0,
        hasParent: !!block.parent_block_id
      });

      // FIXED: Prevent deleting the first and only block on the page
      // Instead, convert it to an empty text block (Notion-like behavior)
      // Only top-level blocks count (blocks without parent)
      const topLevelBlocks = allBlocks.filter(b => !b.parent_block_id);
      console.log('[handleBackspace] Top-level blocks count:', topLevelBlocks.length);

      if (topLevelBlocks.length === 1 && !prevBlock && !block.parent_block_id) {
        // This is the only top-level block - don't delete, just clear it
        console.log('[handleBackspace] Preventing deletion of the only block on the page');

        // Convert to empty text block instead of deleting
        if (block.id.startsWith('temp-')) {
          // Update temp block
          setPendingCreates((prev) => {
            const next = new Map(prev);
            next.set(blockId, {
              ...block,
              type: 'text',
              type_id: 'text',
              content: '',
              blockConfig: {},
            });
            return next;
          });
        } else {
          // Update real block
          setBlocks((prevBlocks) =>
            prevBlocks.map((b) =>
              b.id === blockId
                ? { ...b, type: 'text', type_id: 'text', content: '', blockConfig: {} }
                : b
            )
          );

          // Enqueue update
          if (syncQueueRef.current) {
            syncQueueRef.current.enqueue({
              id: `update-${blockId}-${Date.now()}`,
              type: 'update',
              blockId: blockId,
              data: {
                type: 'text',
                content: '',
                blockConfig: {},
              },
              timestamp: Date.now(),
              priority: 'normal',
            });
          }
        }
        return; // Don't delete, just cleared
      }

      console.log('[handleBackspace] Proceeding with deletion');


      if (block.id.startsWith('temp-')) {
        console.log('[handleBackspace] Deleting temp block');
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
        console.log('[handleBackspace] Deleting real block');
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

      // Focus on previous block OR next block (if first block is deleted)
      console.log('[handleBackspace] Focusing logic - hasPrevBlock:', !!prevBlock, 'blockIndex:', blockIndex, 'totalBlocks:', allBlocks.length);

      if (prevBlock) {
        console.log('[handleBackspace] Focusing on previous block:', prevBlock.id);
        // Standard case: focus on previous block
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
            console.log('[handleBackspace] Focused on previous block editor');
          } else {
            console.log('[handleBackspace] Previous block element not found');
          }
        });
      } else if (blockIndex < allBlocks.length - 1) {
        // FIXED: When deleting first block, focus on next block instead
        const nextBlock = allBlocks[blockIndex + 1];
        console.log('[handleBackspace] Focusing on next block:', nextBlock?.id);
        if (nextBlock) {
          setFocusedBlockId(nextBlock.id);
          setCursorPositions((prev) => {
            const next = new Map(prev);
            next.set(nextBlock.id, 0); // Focus at start of next block
            return next;
          });

          queueMicrotask(() => {
            const blockElement = blockRefsMap.current.get(nextBlock.id);
            if (blockElement) {
              const editor = blockElement.querySelector<HTMLElement>('[contenteditable]');
              editor?.focus();
              console.log('[handleBackspace] Focused on next block editor');
            } else {
              console.log('[handleBackspace] Next block element not found');
            }
          });
        }
      } else {
        console.log('[handleBackspace] No block to focus on - last block deleted?');
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

      const blockTypeConfig = CONTENT_BLOCK_TYPES.find((t) => t.name === newType);
      if (!blockTypeConfig) {
        console.warn('Unknown block type:', newType);
        return;
      }

      let backendType = newType;
      let updatedBlockConfig = block.blockConfig || {};

      if (['kanban', 'habit'].includes(newType)) {
        backendType = 'framework_container';
        updatedBlockConfig = {
          ...blockTypeConfig.defaultMetadata,
          framework_type: newType,
        };
        console.log('[handleTypeChange] Converting to framework_container:', {
          requestedType: newType,
          backendType,
          blockConfig: updatedBlockConfig
        });
      }

      if (block.id.startsWith('temp-')) {
        setPendingCreates((prev) => {
          const next = new Map(prev);
          const updated = {
            ...block,
            type: backendType,
            type_id: backendType,
            blockConfig: updatedBlockConfig,
            metadata: updatedBlockConfig
          };
          next.set(blockId, updated);
          return next;
        });
      } else {
        const existingBlock = blocks.find((b) => b.id === blockId);
        if (!existingBlock) {
          console.warn('Block not found in server blocks:', blockId);
          return;
        }

        setBlocks((prevBlocks) =>
          prevBlocks.map((b) =>
            b.id === blockId ? {
              ...b,
              type: backendType,
              type_id: backendType,
              blockConfig: updatedBlockConfig,
              metadata: updatedBlockConfig
            } : b
          )
        );

        if (syncQueueRef.current) {
          syncQueueRef.current.enqueue({
            id: `update-type-${blockId}-${Date.now()}`,
            type: 'update',
            blockId: blockId,
            data: {
              type: backendType,
              blockConfig: updatedBlockConfig,
            },
            timestamp: Date.now(),
          });
        }
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

      // Security: If we don't know what is being dragged (internal state is null), ignore it
      // This prevents external files/text from being processed as blocks
      if (!draggedBlockId) {
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

      {/* Render only top-level blocks (blocks without parent_block_id) 
          Child blocks will be recursively rendered by BlockWithChildren
      */}
      {allBlocks
        .filter(block => !block.parent_block_id) // Only top-level blocks
        .map((block) => (
          <BlockWithChildren
            key={block.id}
            block={block}
            allBlocks={allBlocks}
            focusedBlockId={focusedBlockId}
            cursorPositions={cursorPositions}
            draggedBlockId={draggedBlockId}
            dragOverBlockId={dragOverBlockId}
            blockRefsMap={blockRefsMap}
            onContentChange={handleContentChange}
            onEnter={handleEnter}
            onBackspace={handleBackspace}
            onMerge={handleMerge}
            onTypeChange={handleTypeChange}
            onIndent={handleIndent}
            onOutdent={handleOutdent}
            onFocus={setFocusedBlockId}
            onArrowUp={handleArrowUp}
            onArrowDown={handleArrowDown}
            onUpdateBlock={handleUpdateBlock}
            onCreateBlock={createNewBlock as any}
            onDragStart={handleDragStart}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            calculateListNumber={calculateListNumber}
            indentLevel={0}
          />
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

