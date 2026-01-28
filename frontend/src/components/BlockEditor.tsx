'use client';

import React, { useState, useRef, useCallback, useEffect } from 'react';
import { blockApi, type Block, type UpdateBlockRequest } from '@/services/blockApi';

interface BlockEditorProps {
  block: Block;
  onUpdate: (block: Block) => void;
  onDelete: (blockId: string) => void;
  onSaveError: (error: string) => void;
  /** Optional: Block index for keyboard navigation between blocks */
  blockIndex?: number;
  /** Optional: Total number of blocks for navigation announcements */
  totalBlocks?: number;
  /** Optional: Callback when block requests focus to previous block */
  onFocusPrevious?: () => void;
  /** Optional: Callback when block requests focus to next block */
  onFocusNext?: () => void;
}

export function BlockEditor({ 
  block, 
  onUpdate, 
  onDelete, 
  onSaveError,
  blockIndex,
  totalBlocks,
  onFocusPrevious,
  onFocusNext,
}: BlockEditorProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [content, setContent] = useState(block.content || '');
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Refs for focus management
  const contentRef = useRef<HTMLTextAreaElement | HTMLInputElement | null>(null);
  const saveButtonRef = useRef<HTMLButtonElement>(null);
  const blockContainerRef = useRef<HTMLDivElement>(null);
  const contentDisplayRef = useRef<HTMLParagraphElement | HTMLHeadingElement | HTMLLabelElement>(null);

  // Auto-focus edit input when entering edit mode
  useEffect(() => {
    if (isEditing && contentRef.current) {
      setTimeout(() => contentRef.current?.focus(), 0);
    }
  }, [isEditing]);

  // Keyboard shortcuts for common actions
  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLTextAreaElement | HTMLInputElement>) => {
    // Ctrl/Cmd + Enter: Save
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault();
      handleSave();
      return;
    }

    // Escape: Cancel editing
    if (e.key === 'Escape') {
      e.preventDefault();
      setIsEditing(false);
      setContent(block.content || '');
      setError(null);
      contentDisplayRef.current?.focus();
      return;
    }

    // Ctrl/Cmd + Shift + D: Delete (with confirmation)
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'd') {
      e.preventDefault();
      handleDelete();
      return;
    }

    // Arrow Up: Navigate to previous block (when not editing or at start of content)
    if (e.key === 'ArrowUp' && !isEditing && onFocusPrevious) {
      const input = e.currentTarget;
      if (input.selectionStart === 0 && input.selectionEnd === 0) {
        e.preventDefault();
        onFocusPrevious();
        announceToScreenReader(`Navigated to previous block. Block ${blockIndex} of ${totalBlocks || '?'}.`);
        return;
      }
    }

    // Arrow Down: Navigate to next block (when not editing or at end of content)
    if (e.key === 'ArrowDown' && !isEditing && onFocusNext) {
      const input = e.currentTarget;
      const isAtEnd = input.selectionStart === input.value.length && input.selectionEnd === input.value.length;
      if (isAtEnd) {
        e.preventDefault();
        onFocusNext();
        const nextIndex = blockIndex !== undefined ? blockIndex + 2 : undefined;
        announceToScreenReader(`Navigated to next block. Block ${nextIndex} of ${totalBlocks || '?'}.`);
        return;
      }
    }

    // Ctrl/Cmd + B: Bold (announcement for screen readers)
    if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
      e.preventDefault();
      announceToScreenReader('Bold formatting applied. Note: This editor stores plain text.');
      return;
    }

    // Ctrl/Cmd + I: Italic (announcement for screen readers)
    if ((e.ctrlKey || e.metaKey) && e.key === 'i') {
      e.preventDefault();
      announceToScreenReader('Italic formatting applied. Note: This editor stores plain text.');
      return;
    }
  }, [block.content, isEditing, blockIndex, totalBlocks, onFocusPrevious, onFocusNext]);

  // Helper to announce messages to screen readers
  const announceToScreenReader = (message: string) => {
    const announcement = document.createElement('div');
    announcement.setAttribute('role', 'status');
    announcement.setAttribute('aria-live', 'polite');
    announcement.setAttribute('aria-atomic', 'true');
    announcement.className = 'sr-only';
    announcement.textContent = message;
    document.body.appendChild(announcement);
    setTimeout(() => announcement.remove(), 1000);
  };

  const handleSave = async () => {
    if (!content.trim() && block.content !== '') {
      // Allow empty content for some block types
      setIsEditing(false);
      announceToScreenReader('Block edit cancelled.');
      return;
    }

    try {
      setIsSaving(true);
      setError(null);
      announceToScreenReader(`Saving ${blockType}...`);

      const updateReq: UpdateBlockRequest = {
        content: content || undefined,
      };

      const updatedBlock = await blockApi.updateBlock(block.id, updateReq);
      onUpdate(updatedBlock);
      setIsEditing(false);
      announceToScreenReader(`${blockType} saved successfully.`);
      contentDisplayRef.current?.focus();
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to save block';
      setError(errorMsg);
      onSaveError(errorMsg);
      announceToScreenReader(`Error saving ${blockType}: ${errorMsg}`);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm(`Are you sure you want to delete this ${blockType}? This action cannot be undone.`)) {
      announceToScreenReader('Delete cancelled.');
      return;
    }

    try {
      setIsSaving(true);
      announceToScreenReader(`Deleting ${blockType}...`);
      await blockApi.deleteBlock(block.id);
      onDelete(block.id);
      announceToScreenReader(`${blockType} deleted successfully.`);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete block';
      setError(errorMsg);
      onSaveError(errorMsg);
      announceToScreenReader(`Error deleting ${blockType}: ${errorMsg}`);
    } finally {
      setIsSaving(false);
    }
  };

  const renderBlockContent = () => {
    const baseClass = 'w-full min-h-[2rem] px-3 py-2 rounded border border-gray-200 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 transition-colors';

    // Get block type (normalize from backend response)
    const currentBlockType = block.type_id || block.type || 'text';

    const blockTypeLabel = {
      text: 'Text block',
      heading_1: 'Heading 1 block',
      heading_2: 'Heading 2 block',
      heading_3: 'Heading 3 block',
      todo_list: 'Todo list block',
      bulleted_list: 'Bulleted list block',
      numbered_list: 'Numbered list block',
    }[currentBlockType] || 'Content block';

    const blockTypeInstructions = {
      text: 'Edit text. Press Ctrl+Enter to save, Escape to cancel.',
      heading_1: 'Edit heading 1. Press Ctrl+Enter to save, Escape to cancel.',
      heading_2: 'Edit heading 2. Press Ctrl+Enter to save, Escape to cancel.',
      heading_3: 'Edit heading 3. Press Ctrl+Enter to save, Escape to cancel.',
      todo_list: 'Edit todo item. Press Ctrl+Enter to save, Escape to cancel.',
      bulleted_list: 'Edit bulleted list item. Press Ctrl+Enter to save, Escape to cancel.',
      numbered_list: 'Edit numbered list item. Press Ctrl+Enter to save, Escape to cancel.',
    }[currentBlockType] || 'Edit block content. Press Ctrl+Enter to save, Escape to cancel.';

    switch (currentBlockType) {
      case 'heading_1':
      case 'heading_2':
      case 'heading_3':
        const headingLevel = currentBlockType === 'heading_1' ? 1 : currentBlockType === 'heading_2' ? 2 : 3;
        const headingSize = headingLevel === 1 ? 'text-3xl' : headingLevel === 2 ? 'text-2xl' : 'text-xl';
        return (
          isEditing ? (
            <div>
              <label htmlFor={`heading-${block.id}`} className="sr-only">
                {blockTypeLabel}
              </label>
              <input
                id={`heading-${block.id}`}
                ref={contentRef as React.Ref<HTMLInputElement>}
                type="text"
                value={content}
                onChange={e => setContent(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Heading..."
                className={`${baseClass} ${headingSize} font-bold`}
                aria-label={blockTypeLabel}
                aria-describedby={`heading-instructions-${block.id}`}
                disabled={isSaving}
              />
              <p id={`heading-instructions-${block.id}`} className="sr-only">
                {blockTypeInstructions}
              </p>
            </div>
          ) : (
            <h3
              ref={contentDisplayRef as React.Ref<HTMLHeadingElement>}
              onClick={() => setIsEditing(true)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  setIsEditing(true);
                }
              }}
              role="button"
              tabIndex={0}
              className={`${headingSize} font-bold cursor-pointer hover:bg-gray-100 px-3 py-2 rounded transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0`}
              aria-label={`Heading ${headingLevel}: ${content || '(empty)'}`}
              aria-describedby={`heading-edit-${block.id}`}
            >
              {content || 'Untitled heading'}
            </h3>
          )
        );

      case 'todo_list':
        return (
          isEditing ? (
            <div>
              <label htmlFor={`todo-${block.id}`} className="sr-only">
                {blockTypeLabel}
              </label>
              <input
                id={`todo-${block.id}`}
                ref={contentRef as React.Ref<HTMLInputElement>}
                type="text"
                value={content}
                onChange={e => setContent(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Todo item..."
                className={baseClass}
                aria-label={blockTypeLabel}
                aria-describedby={`todo-instructions-${block.id}`}
                disabled={isSaving}
              />
              <p id={`todo-instructions-${block.id}`} className="sr-only">
                {blockTypeInstructions}
              </p>
            </div>
          ) : (
            <label
              ref={contentDisplayRef as React.Ref<HTMLLabelElement>}
              className="cursor-pointer hover:bg-gray-100 px-3 py-2 rounded transition-colors flex items-center gap-2 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0"
              aria-label={`Todo item: ${content || '(empty)'}`}
            >
              <input
                type="checkbox"
                onClick={() => setIsEditing(true)}
                className="w-4 h-4 focus:ring-2 focus:ring-primary"
                aria-label={`Select todo item: ${content || '(empty)'}`}
                disabled={isSaving}
              />
              <span
                onClick={() => setIsEditing(true)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    setIsEditing(true);
                  }
                }}
                role="button"
                tabIndex={0}
              >
                {content || 'Click to add todo item...'}
              </span>
            </label>
          )
        );

      case 'text':
      default:
        return (
          isEditing ? (
            <div>
              <label htmlFor={`text-${block.id}`} className="sr-only">
                {blockTypeLabel}
              </label>
              <textarea
                id={`text-${block.id}`}
                ref={contentRef as React.Ref<HTMLTextAreaElement>}
                value={content}
                onChange={e => setContent(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Type something..."
                className={`${baseClass} resize-none`}
                rows={3}
                aria-label={blockTypeLabel}
                aria-describedby={`text-instructions-${block.id}`}
                disabled={isSaving}
              />
              <p id={`text-instructions-${block.id}`} className="sr-only">
                {blockTypeInstructions}
              </p>
            </div>
          ) : (
            <p
              ref={contentDisplayRef as React.Ref<HTMLParagraphElement>}
              onClick={() => setIsEditing(true)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  setIsEditing(true);
                }
              }}
              role="button"
              tabIndex={0}
              className="cursor-pointer hover:bg-gray-100 px-3 py-2 rounded transition-colors min-h-[2rem] flex items-center focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0"
              aria-label={`Text: ${content || '(empty)'}`}
              aria-describedby={`text-edit-${block.id}`}
            >
              {content || 'Click to add text...'}
            </p>
          )
        );
    }
  };

  // Build accessible label with position information
  const blockPositionLabel = blockIndex !== undefined && totalBlocks !== undefined
    ? `Block ${blockIndex + 1} of ${totalBlocks}`
    : 'Block';
  
  const blockType = block.type_id || block.type || 'text';
  const blockTypeLabel = {
    text: 'Text block',
    heading_1: 'Heading 1 block',
    heading_2: 'Heading 2 block',
    heading_3: 'Heading 3 block',
    todo_list: 'Todo list block',
    bulleted_list: 'Bulleted list block',
    numbered_list: 'Numbered list block',
  }[blockType] || 'Content block';

  return (
    <div
      ref={blockContainerRef}
      className="mb-2 p-2 border border-gray-100 rounded bg-white focus-within:ring-2 focus-within:ring-primary focus-within:ring-offset-0 transition-all"
      role="article"
      aria-label={`${blockTypeLabel}, ${blockPositionLabel}`}
      aria-describedby={`block-help-${block.id}`}
      data-block-id={block.id}
      data-block-type={blockType}
      data-block-index={blockIndex}
    >
      {error && (
        <div
          role="alert"
          aria-live="assertive"
          aria-atomic="true"
          className="mb-2 p-2 bg-red-50 text-red-700 text-sm rounded border border-red-200"
        >
          <span className="font-semibold">Error:</span> {error}
        </div>
      )}

      <div id={`block-help-${block.id}`} className="sr-only">
        {blockTypeLabel}. {blockPositionLabel}. 
        {blockIndex !== undefined && totalBlocks !== undefined && (
          <>
            {blockIndex > 0 && 'Press Arrow Up to navigate to previous block. '}
            {blockIndex < totalBlocks - 1 && 'Press Arrow Down to navigate to next block. '}
          </>
        )}
        Double-click or press Enter or Space to edit. 
        Press Ctrl+Enter to save, Escape to cancel. 
        Press Ctrl+Shift+D to delete. 
        Tab to navigate to action buttons.
      </div>

      {renderBlockContent()}

      {isEditing && (
        <div className="mt-2 flex gap-2 flex-wrap" role="toolbar" aria-label="Block editing actions">
          <button
            ref={saveButtonRef}
            onClick={handleSave}
            className="px-3 py-1 text-sm bg-primary text-white rounded hover:bg-primary/90 transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 disabled:opacity-60 disabled:cursor-not-allowed min-h-[44px] min-w-[44px]"
            disabled={isSaving}
            aria-label={`Save ${blockType}`}
            title="Save block (Ctrl+Enter)"
          >
            {isSaving ? 'Saving...' : 'Save'}
          </button>
          <button
            onClick={() => {
              setIsEditing(false);
              setContent(block.content || '');
              setError(null);
              contentDisplayRef.current?.focus();
            }}
            className="px-3 py-1 text-sm border rounded hover:bg-gray-100 transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 disabled:opacity-60 disabled:cursor-not-allowed min-h-[44px] min-w-[44px]"
            disabled={isSaving}
            aria-label={`Cancel editing ${blockType}`}
            title="Cancel (Escape)"
          >
            Cancel
          </button>
          <button
            onClick={handleDelete}
            className="px-3 py-1 text-sm text-red-600 hover:bg-red-50 rounded transition-colors focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-0 disabled:opacity-60 disabled:cursor-not-allowed min-h-[44px] min-w-[44px]"
            disabled={isSaving}
            aria-label={`Delete ${blockType}`}
            title="Delete block (Ctrl+Shift+D)"
          >
            Delete
          </button>
          <div className="sr-only" role="status" aria-live="polite" aria-atomic="true">
            Keyboard shortcuts: Ctrl+Enter to save, Escape to cancel, Ctrl+Shift+D to delete
          </div>
        </div>
      )}

      {!isEditing && (
        <div className="mt-1 text-xs text-gray-500" aria-label="Block metadata">
          {blockType} • Created {new Date(block.created_at).toLocaleDateString()}
        </div>
      )}
    </div>
  );
}
