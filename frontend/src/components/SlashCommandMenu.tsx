'use client';

import React, { useEffect, useRef, useState } from 'react';
import { CONTENT_BLOCK_TYPES, type BlockTypeConfig } from '@/shared/types/blocks';

interface SlashCommandMenuProps {
  open: boolean;
  onClose: () => void;
  onSelect: (blockType: BlockTypeConfig) => void;
  onInsertSlash?: () => void;
  query?: string;
  position?: { top: number; left: number };
}

export function SlashCommandMenu({
  open,
  onClose,
  onSelect,
  onInsertSlash,
  query = '',
  position,
}: SlashCommandMenuProps) {
  const commandRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const itemsRef = useRef<Map<number, HTMLButtonElement>>(new Map());
  const [selectedIndex, setSelectedIndex] = useState(0);

  const handleSelect = (blockType: BlockTypeConfig) => {
    onSelect(blockType);
    onClose();
  };

  // Scroll selected item into view
  useEffect(() => {
    const selectedElement = itemsRef.current.get(selectedIndex);
    if (selectedElement && listRef.current) {
      selectedElement.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex]);

  useEffect(() => {
    if (!open) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        e.stopPropagation();
        // Call onInsertSlash to keep the "/" in the text
        onInsertSlash?.();
        onClose();
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        e.stopPropagation();
        setSelectedIndex(prev => (prev - 1 + CONTENT_BLOCK_TYPES.length) % CONTENT_BLOCK_TYPES.length);
      } else if (e.key === 'ArrowDown') {
        e.preventDefault();
        e.stopPropagation();
        setSelectedIndex(prev => (prev + 1) % CONTENT_BLOCK_TYPES.length);
      } else if (e.key === 'Enter') {
        e.preventDefault();
        e.stopPropagation();
        handleSelect(CONTENT_BLOCK_TYPES[selectedIndex]);
      }
    };

    const handleClickOutside = (e: MouseEvent) => {
      if (commandRef.current && !commandRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('mousedown', handleClickOutside);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [open, onClose, selectedIndex]);

  // Reset selection when menu opens
  useEffect(() => {
    if (open) {
      setSelectedIndex(0);
    }
  }, [open]);

  if (!open) return null;

  const menuStyle = position
    ? {
        position: 'fixed' as const,
        top: `${position.top}px`,
        left: `${position.left}px`,
      }
    : {};

  return (
    <div
      ref={commandRef}
      className="fixed z-50 w-96"
      style={menuStyle}
      onClick={e => e.stopPropagation()}
    >
      <div className="rounded-lg border border-gray-200 bg-white shadow-lg overflow-hidden">
        <div className="border-b border-gray-200 px-4 py-3">
          <div className="text-xs font-semibold text-gray-600 uppercase tracking-wide">Basic blocks</div>
        </div>
        <div ref={listRef} className="max-h-96 overflow-y-auto p-2">
          {CONTENT_BLOCK_TYPES.map((blockType, index) => {
            const Icon = blockType.icon;
            const isSelected = index === selectedIndex;
            return (
              <button
                key={blockType.name}
                ref={(el) => {
                  if (el) {
                    itemsRef.current.set(index, el);
                  }
                }}
                onClick={() => handleSelect(blockType)}
                onMouseEnter={() => setSelectedIndex(index)}
                className={`w-full flex items-center gap-3 rounded-md px-3 py-2.5 text-sm cursor-pointer transition-colors mb-1 ${
                  isSelected
                    ? 'bg-blue-100'
                    : 'hover:bg-gray-100'
                }`}
              >
                <Icon className="h-5 w-5 text-gray-600" />
                <div className="flex-1 text-left">
                  <div className="font-medium text-gray-900">{blockType.displayName}</div>
                  <div className="text-xs text-gray-500">
                    {blockType.description}
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}

