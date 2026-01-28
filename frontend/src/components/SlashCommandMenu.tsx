'use client';

import React, { useEffect, useRef } from 'react';
import { Command } from 'cmdk';
import { CONTENT_BLOCK_TYPES, type BlockTypeConfig } from '@/shared/types/blocks';

interface SlashCommandMenuProps {
  open: boolean;
  onClose: () => void;
  onSelect: (blockType: BlockTypeConfig) => void;
  query?: string;
  position?: { top: number; left: number };
}

export function SlashCommandMenu({
  open,
  onClose,
  onSelect,
  query = '',
  position,
}: SlashCommandMenuProps) {
  const [search, setSearch] = React.useState(query);
  const commandRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setSearch(query);
  }, [query]);

  useEffect(() => {
    if (open && commandRef.current) {
      // Focus the command input when menu opens
      const input = commandRef.current.querySelector('input');
      input?.focus();
    }
  }, [open]);

  // Filter block types based on search
  const filteredTypes = React.useMemo(() => {
    if (!search) return CONTENT_BLOCK_TYPES;
    const lowerSearch = search.toLowerCase();
    return CONTENT_BLOCK_TYPES.filter(
      type =>
        type.name.toLowerCase().includes(lowerSearch) ||
        type.displayName.toLowerCase().includes(lowerSearch) ||
        type.description.toLowerCase().includes(lowerSearch)
    );
  }, [search]);

  const handleSelect = (blockType: BlockTypeConfig) => {
    onSelect(blockType);
    setSearch('');
    onClose();
  };

  useEffect(() => {
    if (open) {
      const handleClickOutside = (e: MouseEvent) => {
        if (commandRef.current && !commandRef.current.contains(e.target as Node)) {
          onClose();
        }
      };
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [open, onClose]);

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
      className="fixed z-50"
      style={menuStyle}
      onClick={e => e.stopPropagation()}
    >
      <Command
        ref={commandRef}
        className="w-64 rounded-lg border border-gray-200 bg-white shadow-lg"
        shouldFilter={false}
      >
        <Command.Input
          value={search}
          onValueChange={setSearch}
          placeholder="Search blocks..."
          className="w-full border-0 px-3 py-2 text-sm outline-none rounded-t-lg"
          autoFocus
        />
        <Command.List className="max-h-64 overflow-y-auto p-1">
          {filteredTypes.length === 0 ? (
            <Command.Empty className="px-3 py-2 text-sm text-gray-500">
              No blocks found
            </Command.Empty>
          ) : (
            filteredTypes.map(blockType => {
              const Icon = blockType.icon;
              return (
                <Command.Item
                  key={blockType.name}
                  onSelect={() => handleSelect(blockType)}
                  className="flex items-center gap-2 rounded px-3 py-2 text-sm hover:bg-gray-100 cursor-pointer aria-selected:bg-gray-100 data-[selected]:bg-gray-100"
                >
                  <Icon className="h-4 w-4 text-gray-500" />
                  <div className="flex-1">
                    <div className="font-medium">{blockType.displayName}</div>
                    <div className="text-xs text-gray-500">
                      {blockType.description}
                    </div>
                  </div>
                </Command.Item>
              );
            })
          )}
        </Command.List>
      </Command>
    </div>
  );
}

