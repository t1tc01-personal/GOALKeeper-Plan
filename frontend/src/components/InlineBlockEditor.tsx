'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { type Block } from '@/services/blockApi';
import { SlashCommandMenu } from './SlashCommandMenu';
import { type BlockTypeConfig } from '@/shared/types/blocks';
import { 
  Type, Heading1, Heading2, Heading3, List, ListOrdered, 
  CheckSquare, ChevronRight, Info 
} from 'lucide-react';

interface InlineBlockEditorProps {
  readonly block: Block;
  readonly isFocused: boolean;
  readonly onContentChange: (blockId: string, content: string) => void;
  readonly onEnter: (blockId: string) => void;
  readonly onBackspace: (blockId: string) => void;
  readonly onMerge?: (blockId: string) => void;
  readonly onTypeChange?: (blockId: string, newType: string) => void;
  readonly onDelete?: (blockId: string) => void;
  readonly onFocus: () => void;
  readonly onBlur?: () => void;
  readonly onArrowUp?: () => void;
  readonly onArrowDown?: () => void;
  readonly autoFocus?: boolean;
  readonly restoreCursorPosition?: number | null;
  readonly onUpdateBlock?: (blockId: string, data: { content?: string; type?: string; blockConfig?: Record<string, any> }) => void;
}

// Debounce helper with cancel support
function useDebounce<T extends (...args: any[]) => void>(
  callback: T,
  delay: number
): [T, () => void] {
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const callbackRef = useRef(callback);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  const debounced = useCallback(
    ((...args: Parameters<T>) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      timeoutRef.current = setTimeout(() => {
        callbackRef.current(...args);
      }, delay);
    }) as T,
    [delay]
  );

  const cancel = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  }, []);

  return [debounced as T, cancel];
}

export function InlineBlockEditor({
  block,
  isFocused,
  onContentChange,
  onEnter,
  onBackspace,
  onMerge,
  onTypeChange,
  onDelete,
  onFocus,
  onBlur,
  onArrowUp,
  onArrowDown,
  autoFocus = false,
  restoreCursorPosition = null,
  onUpdateBlock,
}: InlineBlockEditorProps) {
  const [content, setContent] = useState(block.content || '');
  
  // Sync content when block updates (from server or merge)
  useEffect(() => {
    if (!isFocused || block.content !== content) {
      setContent(block.content || '');
    }
  }, [block.content, block.id]);
  const [showSlashMenu, setShowSlashMenu] = useState(false);
  const [slashMenuPosition, setSlashMenuPosition] = useState<{ top: number; left: number } | undefined>();
  const [slashQuery, setSlashQuery] = useState('');
  const editorRef = useRef<HTMLDivElement>(null);
  const isTempBlock = block.id.startsWith('temp-');

  // Sync content when block updates (from server)
  useEffect(() => {
    if (!isFocused && !isTempBlock) {
      setContent(block.content || '');
    }
  }, [block.content, isFocused, isTempBlock]);

  // Debounced save function - direct API update via onUpdateBlock
  const [debouncedSave, cancelDebouncedSave] = useDebounce(
    (blockId: string, content: string, blockType: string) => {
    if (isTempBlock || blockId.startsWith('temp-')) {
      // Temp blocks will be saved when they get real IDs
      return;
    }

    if (!onUpdateBlock) {
      // Fallback: if no queue callback provided, skip (should not happen)
      console.warn('onUpdateBlock not provided, skipping save');
      return;
    }

    // Queue update to batch sync
    const currentMetadata = block.metadata || block.blockConfig || {};
    const updateData: { content?: string; type?: string; blockConfig?: Record<string, any> } = {
      content: content || undefined,
    };

    // Handle metadata for specific block types
    if (blockType === 'toggle' || blockType === 'callout') {
      updateData.blockConfig = currentMetadata;
    }

    onUpdateBlock(blockId, updateData);
  },
  1500); // 1.5s debounce

  // Cleanup debounce on unmount or when block is deleted
  useEffect(() => {
    return () => {
      cancelDebouncedSave();
    };
  }, [cancelDebouncedSave]);

  // Handle content change
  const handleContentChange = (newContent: string) => {
    setContent(newContent);
    onContentChange(block.id, newContent);
    
    // Trigger debounced save (only for non-temp blocks)
    // Check again in case block was deleted
    if (!block.id.startsWith('temp-') && block.id) {
      const blockType = block.type_id || block.type || 'text';
      debouncedSave(block.id, newContent, blockType);
    }
  };

  // Handle key events
  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    // Ctrl+Enter or Cmd+Enter: Insert newline in same block (markdown style)
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      // Allow default behavior (insert newline/line break)
      // The contentEditable will handle this naturally
      return;
    }

    // Enter: Always create new block (Notion-style)
    // This is the default behavior - Enter = new block
    if (e.key === 'Enter' && !e.shiftKey && !e.ctrlKey && !e.metaKey) {
      e.preventDefault();
      // Cursor position will be saved by PageEditor before creating new block
      onEnter(block.id);
      return;
    }

    // Shift+Enter: Insert line break in same block (Notion-style)
    // Allow default behavior - contentEditable will insert <br> or newline
    if (e.key === 'Enter' && e.shiftKey) {
      // Don't prevent default - let browser handle line break
      return;
    }

    // Backspace: Handle block deletion and merging
    if (e.key === 'Backspace') {
      const selection = globalThis.getSelection();
      if (selection && selection.rangeCount > 0) {
        const range = selection.getRangeAt(0);
        const editor = editorRef.current;
        if (editor) {
          // Check if cursor is at the start
          const isAtStart = range.startOffset === 0 && 
            (range.startContainer === editor || 
             (range.startContainer.nodeType === Node.TEXT_NODE && 
              range.startContainer.parentElement === editor && 
              range.startOffset === 0));
          
          if (isAtStart) {
            if (content === '' && !isTempBlock) {
              // Empty block: Delete block
              e.preventDefault();
              cancelDebouncedSave();
              onBackspace(block.id);
              return;
            } else if (content.length > 0 && onMerge) {
              // Block with content: Merge with previous block (Notion-style)
              e.preventDefault();
              cancelDebouncedSave();
              onMerge(block.id);
              return;
            }
          }
        }
      }
    }

    // Escape: Close slash menu
    if (e.key === 'Escape' && showSlashMenu) {
      e.preventDefault();
      setShowSlashMenu(false);
      setSlashQuery('');
      // Remove the "/" from content
      const newContent = content.slice(0, -1);
      handleContentChange(newContent);
      return;
    }

    // Arrow Up: Move to previous block
    if (e.key === 'ArrowUp') {
      const selection = globalThis.getSelection();
      if (selection && selection.rangeCount > 0) {
        const range = selection.getRangeAt(0);
        const editor = editorRef.current;
        if (editor) {
          // Check if cursor is at the start of content
          const isAtStart = range.startOffset === 0 && 
            (range.startContainer === editor || 
             (range.startContainer.nodeType === Node.TEXT_NODE && 
              range.startContainer.parentElement === editor && 
              range.startOffset === 0));
          
          if (isAtStart && onArrowUp) {
            e.preventDefault();
            onArrowUp();
            return;
          }
        }
      }
    }

    // Arrow Down: Move to next block
    if (e.key === 'ArrowDown') {
      const selection = globalThis.getSelection();
      if (selection && selection.rangeCount > 0) {
        const range = selection.getRangeAt(0);
        const editor = editorRef.current;
        if (editor) {
          // Check if cursor is at the end of content
          const textLength = editor.textContent?.length || 0;
          const isAtEnd = range.endOffset >= textLength;
          
          if (isAtEnd && onArrowDown) {
            e.preventDefault();
            onArrowDown();
            return;
          }
        }
      }
    }
  };

  // Handle slash menu visibility & query update
  useEffect(() => {
    const hasSlash = content.includes('/');

    // Open menu when user has just typed "/" (content ends with "/")
    if (!showSlashMenu && hasSlash && content.endsWith('/')) {
      const rect = editorRef.current?.getBoundingClientRect();
      if (rect) {
        setSlashMenuPosition({
          top: rect.bottom + 4,
          left: rect.left,
        });
        setShowSlashMenu(true);
      }
    }

    if (showSlashMenu && hasSlash) {
      const lastSlashIndex = content.lastIndexOf('/');
      const query = content.slice(lastSlashIndex + 1);
      setSlashQuery(query);
    } else if (!hasSlash) {
      setShowSlashMenu(false);
      setSlashQuery('');
    }
  }, [content, showSlashMenu]);

  // Handle slash menu selection
  const handleSlashMenuSelect = (blockType: BlockTypeConfig) => {
    // Remove "/" and query from content
    const newContent = content.replace(/\/.*$/, '').trim();
    handleContentChange(newContent);
    
    // Change block type if different
    const currentType = block.type_id || block.type || 'text';
    if (blockType.name !== currentType && onTypeChange) {
      onTypeChange(block.id, blockType.name);
    }
    
    setShowSlashMenu(false);
    setSlashQuery('');
  };

  // Render block based on type
  const renderBlock = () => {
    const blockType = block.type_id || block.type || 'text';
    // Remove border and focus ring - clean markdown-style appearance
    const baseClass =
      'w-full min-h-[1.5rem] px-2 py-1 outline-none border-0 focus:outline-none focus-visible:outline-none focus:ring-0 focus:ring-offset-0 focus:border-transparent';
    
    switch (blockType) {
      case 'heading_1':
        return (
          <div
            ref={editorRef}
            contentEditable
            suppressContentEditableWarning
            role="textbox"
            tabIndex={0}
            aria-label="Heading 1"
            onInput={(e) => {
              const newContent = e.currentTarget.textContent || '';
              handleContentChange(newContent);
            }}
            onKeyDown={handleKeyDown}
            onFocus={onFocus}
            onBlur={onBlur}
            className={`${baseClass} text-3xl font-bold`}
            data-placeholder="Heading 1"
            style={{
              minHeight: '2.5rem',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              border: 'none',
              outline: 'none',
              boxShadow: 'none',
            }}
          />
        );

      case 'heading_2':
        return (
          <div
            ref={editorRef}
            contentEditable
            suppressContentEditableWarning
            role="textbox"
            tabIndex={0}
            aria-label="Heading 2"
            onInput={(e) => {
              const newContent = e.currentTarget.textContent || '';
              handleContentChange(newContent);
            }}
            onKeyDown={handleKeyDown}
            onFocus={onFocus}
            onBlur={onBlur}
            className={`${baseClass} text-2xl font-bold`}
            data-placeholder="Heading 2"
            style={{
              minHeight: '2rem',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              border: 'none',
              outline: 'none',
              boxShadow: 'none',
            }}
          />
        );

      case 'heading_3':
        return (
          <div
            ref={editorRef}
            contentEditable
            suppressContentEditableWarning
            role="textbox"
            tabIndex={0}
            aria-label="Heading 3"
            onInput={(e) => {
              const newContent = e.currentTarget.textContent || '';
              handleContentChange(newContent);
            }}
            onKeyDown={handleKeyDown}
            onFocus={onFocus}
            onBlur={onBlur}
            className={`${baseClass} text-xl font-bold`}
            data-placeholder="Heading 3"
            style={{
              minHeight: '1.75rem',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              border: 'none',
              outline: 'none',
              boxShadow: 'none',
            }}
          />
        );

      case 'text':
      default:
        return (
          <div
            ref={editorRef}
            contentEditable
            suppressContentEditableWarning
            role="textbox"
            tabIndex={0}
            aria-label="Text block"
            onInput={(e) => {
              const newContent = e.currentTarget.textContent || '';
              handleContentChange(newContent);
            }}
            onKeyDown={handleKeyDown}
            onFocus={onFocus}
            onBlur={onBlur}
            className={baseClass}
            data-placeholder="Type '/' for commands"
            style={{
              minHeight: '1.5rem',
              whiteSpace: 'pre-wrap', // Allow line breaks
              wordBreak: 'break-word',
              border: 'none',
              outline: 'none',
              boxShadow: 'none',
            }}
          />
        );
    }
  };

  // Auto-focus when isFocused changes
  useEffect(() => {
    if (isFocused && editorRef.current) {
      // Use requestAnimationFrame to ensure DOM is ready
      requestAnimationFrame(() => {
        const editor = editorRef.current;
        if (!editor) return;
        
        editor.focus();
        
        // Restore cursor position if provided
        if (restoreCursorPosition !== null && restoreCursorPosition >= 0) {
          const range = document.createRange();
          const selection = globalThis.getSelection();
          
          // Try to set cursor at the specified position
          try {
            // Find the text node and set cursor position
            const textContent = editor.textContent || '';
            const targetPos = Math.min(restoreCursorPosition, textContent.length);
            
            // Walk through nodes to find the right position
            let currentPos = 0;
            const walker = document.createTreeWalker(
              editor,
              NodeFilter.SHOW_TEXT,
              null
            );
            
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
            } else {
              // Fallback: set at end
              const lastChild = editor.lastChild;
              if (lastChild && lastChild.nodeType === Node.TEXT_NODE) {
                const lastNode = lastChild as Text;
                range.setStart(lastNode, lastNode.length);
                range.setEnd(lastNode, lastNode.length);
              } else {
                range.selectNodeContents(editor);
                range.collapse(false);
              }
            }
            
            selection?.removeAllRanges();
            selection?.addRange(range);
          } catch (err) {
            // If restoration fails, set cursor at start
            try {
              const fallbackRange = document.createRange();
              const fallbackSelection = globalThis.getSelection();
              const firstChild = editor.firstChild;
              if (firstChild && firstChild.nodeType === Node.TEXT_NODE) {
                const firstNode = firstChild as Text;
                fallbackRange.setStart(firstNode, 0);
                fallbackRange.setEnd(firstNode, 0);
              } else {
                fallbackRange.setStart(editor, 0);
                fallbackRange.setEnd(editor, 0);
              }
              fallbackSelection?.removeAllRanges();
              fallbackSelection?.addRange(fallbackRange);
            } catch (error_) {
              // Ignore errors in cursor restoration
              console.debug('Cursor restoration failed:', error_);
            }
          }
        }
        
        if (autoFocus && (!restoreCursorPosition || restoreCursorPosition === null)) {
          // Move cursor to start if autoFocus and no restore position
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
      });
    }
  }, [isFocused, autoFocus, restoreCursorPosition]);

  // Set initial content and sync when block/content changes
  useEffect(() => {
    if (editorRef.current) {
      const currentText = editorRef.current.textContent || '';
      if (currentText !== content) {
        editorRef.current.textContent = content;
      }
    }
  }, [content]);

  return (
    <div className="relative group">
      <div className="flex items-start gap-2">
        {/* Block type icon (hidden by default, show on hover) */}
        <div className="opacity-0 group-hover:opacity-100 transition-opacity pt-1">
          {getBlockIcon(block.type_id || block.type || 'text')}
        </div>
        
        {/* Editor */}
        <div className="flex-1 relative" style={{ border: 'none', outline: 'none' }}>
          {renderBlock()}
          
          {/* Placeholder */}
          {!content && (
            <div
              className="absolute left-2 top-1 text-gray-400 pointer-events-none"
              style={{ fontSize: 'inherit' }}
            >
              {getPlaceholder(block.type_id || block.type || 'text')}
            </div>
          )}
        </div>
      </div>

      {/* Slash command menu */}
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
  );
}

function getBlockIcon(type: string) {
  const iconClass = 'h-4 w-4 text-gray-400';
  switch (type) {
    case 'heading_1':
      return <Heading1 className={iconClass} />;
    case 'heading_2':
      return <Heading2 className={iconClass} />;
    case 'heading_3':
      return <Heading3 className={iconClass} />;
    case 'bulleted_list':
      return <List className={iconClass} />;
    case 'numbered_list':
      return <ListOrdered className={iconClass} />;
    case 'todo_list':
      return <CheckSquare className={iconClass} />;
    case 'toggle':
      return <ChevronRight className={iconClass} />;
    case 'callout':
      return <Info className={iconClass} />;
    default:
      return <Type className={iconClass} />;
  }
}

function getPlaceholder(type: string): string {
  switch (type) {
    case 'heading_1':
      return 'Heading 1';
    case 'heading_2':
      return 'Heading 2';
    case 'heading_3':
      return 'Heading 3';
    case 'text':
    default:
      return "Type '/' for commands";
  }
}

