'use client';

import React, { useState, useRef, useCallback, useEffect } from 'react';
import { blockApi, type Block, type UpdateBlockRequest } from '@/services/blockApi';

interface BlockEditorProps {
  block: Block;
  onUpdate: (block: Block) => void;
  onDelete: (blockId: string) => void;
  onSaveError: (error: string) => void;
}

export function BlockEditor({ block, onUpdate, onDelete, onSaveError }: BlockEditorProps) {
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
  }, [block.content]);

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
      announceToScreenReader(`Saving ${block.type_id}...`);

      const updateReq: UpdateBlockRequest = {
        content: content || undefined,
      };

      const updatedBlock = await blockApi.updateBlock(block.id, updateReq);
      onUpdate(updatedBlock);
      setIsEditing(false);
      announceToScreenReader(`${block.type_id} saved successfully.`);
      contentDisplayRef.current?.focus();
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to save block';
      setError(errorMsg);
      onSaveError(errorMsg);
      announceToScreenReader(`Error saving ${block.type_id}: ${errorMsg}`);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm(`Are you sure you want to delete this ${block.type_id}? This action cannot be undone.`)) {
      announceToScreenReader('Delete cancelled.');
      return;
    }

    try {
      setIsSaving(true);
      announceToScreenReader(`Deleting ${block.type_id}...`);
      await blockApi.deleteBlock(block.id);
      onDelete(block.id);
      announceToScreenReader(`${block.type_id} deleted successfully.`);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete block';
      setError(errorMsg);
      onSaveError(errorMsg);
      announceToScreenReader(`Error deleting ${block.type_id}: ${errorMsg}`);
    } finally {
      setIsSaving(false);
    }
  };

  const renderBlockContent = () => {
    const baseClass = 'w-full min-h-[2rem] px-3 py-2 rounded border border-gray-200 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 transition-colors';

    const blockTypeLabel = {
      heading: 'Block heading',
      paragraph: 'Block paragraph',
      checklist: 'Block checklist item',
    }[block.type_id] || 'Block content';

    const blockTypeInstructions = {
      heading: 'Edit heading. Press Ctrl+Enter to save, Escape to cancel.',
      paragraph: 'Edit paragraph. Press Ctrl+Enter to save, Escape to cancel.',
      checklist: 'Edit checklist item. Press Ctrl+Enter to save, Escape to cancel.',
    }[block.type_id] || 'Edit block content. Press Ctrl+Enter to save, Escape to cancel.';

    switch (block.type_id) {
      case 'heading':
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
                className={`${baseClass} text-xl font-bold`}
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
              className="text-xl font-bold cursor-pointer hover:bg-gray-100 px-3 py-2 rounded transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0"
              aria-label={`Heading: ${content || '(empty)'}`}
              aria-describedby={`heading-edit-${block.id}`}
            >
              {content || 'Untitled heading'}
            </h3>
          )
        );

      case 'paragraph':
      default:
        return (
          isEditing ? (
            <div>
              <label htmlFor={`paragraph-${block.id}`} className="sr-only">
                {blockTypeLabel}
              </label>
              <textarea
                id={`paragraph-${block.id}`}
                ref={contentRef as React.Ref<HTMLTextAreaElement>}
                value={content}
                onChange={e => setContent(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Type something..."
                className={`${baseClass} resize-none`}
                rows={3}
                aria-label={blockTypeLabel}
                aria-describedby={`paragraph-instructions-${block.id}`}
                disabled={isSaving}
              />
              <p id={`paragraph-instructions-${block.id}`} className="sr-only">
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
              aria-label={`Paragraph: ${content || '(empty)'}`}
              aria-describedby={`paragraph-edit-${block.id}`}
            >
              {content || 'Click to add text...'}
            </p>
          )
        );

      case 'checklist':
        return (
          isEditing ? (
            <div>
              <label htmlFor={`checklist-${block.id}`} className="sr-only">
                {blockTypeLabel}
              </label>
              <input
                id={`checklist-${block.id}`}
                ref={contentRef as React.Ref<HTMLInputElement>}
                type="text"
                value={content}
                onChange={e => setContent(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Checklist item..."
                className={baseClass}
                aria-label={blockTypeLabel}
                aria-describedby={`checklist-instructions-${block.id}`}
                disabled={isSaving}
              />
              <p id={`checklist-instructions-${block.id}`} className="sr-only">
                {blockTypeInstructions}
              </p>
            </div>
          ) : (
            <label
              ref={contentDisplayRef as React.Ref<HTMLLabelElement>}
              className="cursor-pointer hover:bg-gray-100 px-3 py-2 rounded transition-colors flex items-center gap-2 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0"
              aria-label={`Checklist item: ${content || '(empty)'}`}
            >
              <input
                type="checkbox"
                onClick={() => setIsEditing(true)}
                className="w-4 h-4 focus:ring-2 focus:ring-primary"
                aria-label={`Select checklist item: ${content || '(empty)'}`}
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
                {content || 'Click to add checklist item...'}
              </span>
            </label>
          )
        );
    }
  };

  return (
    <div
      ref={blockContainerRef}
      className="mb-2 p-2 border border-gray-100 rounded bg-white focus-within:ring-2 focus-within:ring-primary focus-within:ring-offset-0 transition-all"
      role="region"
      aria-label={`${block.type_id} block, created ${new Date(block.created_at).toLocaleDateString()}`}
      aria-describedby={`block-help-${block.id}`}
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
        Block of type {block.type_id}. Double-click or press Enter to edit. Tab to navigate to buttons.
      </div>

      {renderBlockContent()}

      {isEditing && (
        <div className="mt-2 flex gap-2 flex-wrap" role="toolbar" aria-label="Block editing actions">
          <button
            ref={saveButtonRef}
            onClick={handleSave}
            className="px-3 py-1 text-sm bg-primary text-white rounded hover:bg-primary/90 transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-0 disabled:opacity-60 disabled:cursor-not-allowed min-h-[44px] min-w-[44px]"
            disabled={isSaving}
            aria-label={`Save ${block.type_id}`}
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
            aria-label={`Cancel editing ${block.type_id}`}
            title="Cancel (Escape)"
          >
            Cancel
          </button>
          <button
            onClick={handleDelete}
            className="px-3 py-1 text-sm text-red-600 hover:bg-red-50 rounded transition-colors focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-0 disabled:opacity-60 disabled:cursor-not-allowed min-h-[44px] min-w-[44px]"
            disabled={isSaving}
            aria-label={`Delete ${block.type_id}`}
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
          {block.type_id} • Created {new Date(block.created_at).toLocaleDateString()}
        </div>
      )}
    </div>
  );
}
