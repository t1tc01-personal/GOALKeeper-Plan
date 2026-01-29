'use client';

import React, { useState, useRef, useEffect } from 'react';
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
  readonly listNumber?: number; // For numbered_list: the item number (1, 2, 3, ...)
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
  listNumber = 1,
}: InlineBlockEditorProps) {
  const [content, setContent] = useState(block.content || '');
  const lastKnownContentRef = useRef<string>(block.content || ''); // Track last content from user typing
  const [showSlashMenu, setShowSlashMenu] = useState(false);
  const [slashMenuPosition, setSlashMenuPosition] = useState<{ top: number; left: number } | undefined>();
  const [slashQuery, setSlashQuery] = useState('');
  const [justClosedMenu, setJustClosedMenu] = useState(false);
  const editorRef = useRef<HTMLDivElement>(null);
  const isTempBlock = block.id.startsWith('temp-');

  // Ref để lưu nội dung cuối cùng được đồng bộ thành công
  const lastSyncedContentRef = useRef(block.content);

  // GIẢI PHÁP "MẠNH TAY": Quản lý nội dung contenteditable thông qua DOM trực tiếp
  // Cập nhật nội dung vào DOM khi props thay đổi
  useEffect(() => {
    if (editorRef.current) {
      // Lấy nội dung hiện tại từ DOM
      const currentDOMContent = editorRef.current.textContent || '';
      // Nội dung từ props
      const incomingContent = block.content || '';

      // Trường hợp 1: Block không focused -> Cập nhật thoải mái
      // Trường hợp 2: Block đang focused NHƯNG nội dung DOM khác với nội dung Props 
      // VÀ nội dung Props khác với nội dung ta vừa DB (để tránh feedback loop)
      if (currentDOMContent !== incomingContent) {
        if (!isFocused || incomingContent !== lastKnownContentRef.current) {
          editorRef.current.textContent = incomingContent;
          lastSyncedContentRef.current = incomingContent;
          setContent(incomingContent);
        }
      }
    }
  }, [block.content, isFocused]);

  // Cleanup on unmount (no longer using debounced save)

  // Handle content change - sử dụng uncontrolled component pattern
  const handleContentChange = (newContent: string) => {
    const currentType = block.type_id || block.type || 'text';

    // Auto-detect numbered list pattern: "1. " at start (only for text blocks)
    const numberedListPattern = /^1\.\s/;
    if (currentType === 'text' && numberedListPattern.exec(newContent)) {
      // Remove "1. " prefix
      const cleanContent = newContent.replace(/^1\.\s/, '');
      lastKnownContentRef.current = cleanContent;

      // Update DOM trực tiếp (không qua state)
      if (editorRef.current) {
        editorRef.current.textContent = cleanContent;
      }

      onContentChange(block.id, cleanContent);

      // Convert to numbered_list
      if (onTypeChange) {
        onTypeChange(block.id, 'numbered_list');
      }
      return;
    }

    // Auto-detect bulleted list pattern: "* " at start (only for text blocks)
    const bulletedListPattern = /^\*\s/;
    if (currentType === 'text' && bulletedListPattern.exec(newContent)) {
      // Remove "* " prefix
      const cleanContent = newContent.replace(/^\*\s/, '');
      lastKnownContentRef.current = cleanContent;

      // Update DOM trực tiếp (không qua state)
      if (editorRef.current) {
        editorRef.current.textContent = cleanContent;
      }

      onContentChange(block.id, cleanContent);

      // Convert to bulleted_list
      if (onTypeChange) {
        onTypeChange(block.id, 'bulleted_list');
      }
      return;
    }

    // Normal content change - Lưu vào state CHỈ để theo dõi (không phải để render)
    // Nội dung thực tế được quản lý bởi contenteditable DOM
    setContent(newContent);
    lastKnownContentRef.current = newContent;
    onContentChange(block.id, newContent);

    // Immediate sync - call onUpdateBlock directly on every keystroke
    // (only for non-temp blocks)
    if (!block.id.startsWith('temp-') && block.id && onUpdateBlock) {
      const blockType = block.type_id || block.type || 'text';
      const currentMetadata = block.metadata || block.blockConfig || {};
      const updateData: { content?: string; type?: string; blockConfig?: Record<string, any> } = {
        content: newContent || undefined,
      };

      // Handle metadata for specific block types
      if (blockType === 'toggle' || blockType === 'callout') {
        updateData.blockConfig = currentMetadata;
      }

      // Call immediately for instant batch sync
      onUpdateBlock(block.id, updateData);
    }
  };

  // Handle key events
  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    // When slash menu is open, let SlashCommandMenu handle navigation & selection
    if (showSlashMenu) {
      if (e.key === 'Enter' || e.key === 'ArrowUp' || e.key === 'ArrowDown') {
        // Do not trigger block-level enter / arrow navigation
        return;
      }

      if (e.key === 'Backspace') {
        // Avoid triggering block delete/merge while menu is open
        return;
      }
    }

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
            // QUAN TRỌNG: Kiểm tra DOM textContent thực tế, không phải state
            const actualDOMContent = editor.textContent || '';
            if (actualDOMContent === '') {
              // Empty block: Delete block (both temp and real blocks)
              e.preventDefault();
              onBackspace(block.id);
              return;
            } else if (actualDOMContent.length > 0 && onMerge) {
              // Block with content: Merge with previous block (Notion-style)
              e.preventDefault();
              onMerge(block.id);
              return;
            }
          }
        }
      }
    }

    // Escape: Close slash menu (handled by SlashCommandMenu component)
    // The SlashCommandMenu will prevent default and call onInsertSlash to keep the "/"
    if (e.key === 'Escape' && showSlashMenu) {
      // Let SlashCommandMenu handle the escape key
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
    // But NOT if menu was just closed via ESC (justClosedMenu = true)
    if (!showSlashMenu && hasSlash && content.endsWith('/') && !justClosedMenu) {
      const rect = editorRef.current?.getBoundingClientRect();
      if (rect) {
        setSlashMenuPosition({
          top: rect.bottom + 4,
          left: rect.left,
        });
        setShowSlashMenu(true);
      }
    }

    // Reset justClosedMenu flag when user continues typing after "/"
    if (justClosedMenu && !content.endsWith('/')) {
      setJustClosedMenu(false);
    }

    if (showSlashMenu && hasSlash) {
      const lastSlashIndex = content.lastIndexOf('/');
      const query = content.slice(lastSlashIndex + 1);
      setSlashQuery(query);
    } else if (!hasSlash) {
      setShowSlashMenu(false);
      setSlashQuery('');
    }
  }, [content, showSlashMenu, justClosedMenu]);

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

      case 'numbered_list':
        return (
          <div className="flex items-start gap-2 w-full">
            {/* Block type icon - chỉ hiện khi hover */}
            <div className="opacity-0 group-hover:opacity-100 transition-opacity pt-1">
              {getBlockIcon('numbered_list')}
            </div>

            {/* List prefix */}
            <span className="text-gray-500 select-none pt-1 min-w-[1.5rem] text-right text-sm">
              {listNumber}.
            </span>

            {/* Editor */}
            <div className="flex-1 relative">
              <div
                ref={editorRef}
                contentEditable
                suppressContentEditableWarning
                role="textbox"
                tabIndex={0}
                aria-label="Numbered list item"
                onInput={(e) => {
                  const newContent = e.currentTarget.textContent || '';
                  handleContentChange(newContent);
                }}
                onKeyDown={handleKeyDown}
                onFocus={onFocus}
                onBlur={onBlur}
                className={baseClass}
                data-placeholder="List item"
                style={{
                  minHeight: '1.5rem',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  border: 'none',
                  outline: 'none',
                  boxShadow: 'none',
                }}
              />

              {/* Placeholder - adjust position for list */}
              {!content && (
                <div
                  className="absolute left-2 top-1 text-gray-400 pointer-events-none"
                  style={{ fontSize: 'inherit' }}
                >
                  List item
                </div>
              )}
            </div>
          </div>
        );

      case 'bulleted_list':
        return (
          <div className="flex items-start gap-2 w-full">
            {/* Block type icon - chỉ hiện khi hover */}
            <div className="opacity-0 group-hover:opacity-100 transition-opacity pt-1">
              {getBlockIcon('bulleted_list')}
            </div>

            {/* List prefix */}
            <span className="text-gray-500 select-none pt-1 min-w-[1rem] text-center text-sm">
              •
            </span>

            {/* Editor */}
            <div className="flex-1 relative">
              <div
                ref={editorRef}
                contentEditable
                suppressContentEditableWarning
                role="textbox"
                tabIndex={0}
                aria-label="Bulleted list item"
                onInput={(e) => {
                  const newContent = e.currentTarget.textContent || '';
                  handleContentChange(newContent);
                }}
                onKeyDown={handleKeyDown}
                onFocus={onFocus}
                onBlur={onBlur}
                className={baseClass}
                data-placeholder="List item"
                style={{
                  minHeight: '1.5rem',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  border: 'none',
                  outline: 'none',
                  boxShadow: 'none',
                }}
              />

              {/* Placeholder - adjust position for list */}
              {!content && (
                <div
                  className="absolute left-2 top-1 text-gray-400 pointer-events-none"
                  style={{ fontSize: 'inherit' }}
                >
                  List item
                </div>
              )}
            </div>
          </div>
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

  // Restore focus after type change (especially for list conversion from "1. " or "* ")
  useEffect(() => {
    // Only restore if block is focused and editor exists
    if (isFocused && editorRef.current) {
      requestAnimationFrame(() => {
        const editor = editorRef.current;
        if (!editor) return;

        // Only restore if editor is not already focused
        if (document.activeElement !== editor) {
          editor.focus();

          // Set cursor at end of content (user was typing)
          const range = document.createRange();
          const selection = globalThis.getSelection();
          const textContent = editor.textContent || '';

          if (textContent.length > 0) {
            // Find last text node and set cursor at end
            const walker = document.createTreeWalker(
              editor,
              NodeFilter.SHOW_TEXT,
              null
            );

            let lastNode: Text | null = null;
            while (walker.nextNode()) {
              lastNode = walker.currentNode as Text;
            }

            if (lastNode) {
              range.setStart(lastNode, lastNode.length);
              range.setEnd(lastNode, lastNode.length);
            } else {
              range.selectNodeContents(editor);
              range.collapse(false);
            }
          } else if (editor.firstChild && editor.firstChild.nodeType === Node.TEXT_NODE) {
            // Empty content - set cursor at start of first text node
            range.setStart(editor.firstChild, 0);
            range.setEnd(editor.firstChild, 0);
          } else {
            // Empty content - set cursor at start of editor
            range.setStart(editor, 0);
            range.setEnd(editor, 0);
          }

          selection?.removeAllRanges();
          selection?.addRange(range);
        }
      });
    }
  }, [block.type_id, block.type, isFocused]); // Trigger on type change

  // Set initial content and sync when block/content changes
  useEffect(() => {
    if (editorRef.current) {
      const currentText = editorRef.current.textContent || '';
      // Only update if content state is different from DOM (avoids recursion)
      if (currentText !== content) {
        editorRef.current.textContent = content;
      }
    }
  }, [content]);

  // Khởi tạo nội dung DOM khi component mount hoặc block thay đổi
  useEffect(() => {
    // Luôn khởi tạo khi mount, bất kể focus (vì remount do key change có thể xảy ra)
    if (editorRef.current) {
      const initialContent = block.content || '';
      // Chỉ cập nhật nếu thực sự khác để tránh mất focus/cursor bất ngờ
      if (editorRef.current.textContent !== initialContent) {
        editorRef.current.textContent = initialContent;
        setContent(initialContent);
      }
    }
  }, [block.id]); // Chạy khi ID thay đổi (remount/transition)

  const blockType = block.type_id || block.type || 'text';
  const isListType = blockType === 'numbered_list' || blockType === 'bulleted_list';

  return (
    <div className="relative group">
      {/* For list types, renderBlock() already includes icon and structure */}
      {/* For other types, wrap with icon and placeholder */}
      {isListType ? (
        renderBlock()
      ) : (
        <div className="flex items-start gap-2">
          {/* Block type icon (hidden by default, show on hover) */}
          <div className="opacity-0 group-hover:opacity-100 transition-opacity pt-1">
            {getBlockIcon(blockType)}
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
                {getPlaceholder(blockType)}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Slash command menu */}
      {showSlashMenu && slashMenuPosition && (
        <SlashCommandMenu
          open={showSlashMenu}
          onClose={() => {
            setShowSlashMenu(false);
            setSlashQuery('');
          }}
          onSelect={handleSlashMenuSelect}
          onInsertSlash={() => {
            // Keep the "/" and close menu, set flag to prevent reopening
            setShowSlashMenu(false);
            setSlashQuery('');
            setJustClosedMenu(true);
          }}
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
    case 'numbered_list':
    case 'bulleted_list':
      return 'List item';
    case 'text':
    default:
      return "Type '/' for commands";
  }
}

