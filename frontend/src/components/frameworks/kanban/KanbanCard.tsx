'use client';

import React, { useState, useEffect, useRef } from 'react';
import { FrameworkItemProps } from '@/shared/frameworks/registry';
import { MoreHorizontal, GripVertical, Trash2, Clock } from 'lucide-react';

export function KanbanCard({
    block,
    onUpdate,
    onDelete
}: FrameworkItemProps) {
    const [isEditing, setIsEditing] = useState(false);
    const [showMenu, setShowMenu] = useState(false);
    const [tempContent, setTempContent] = useState(block.content || '');
    const menuRef = useRef<HTMLDivElement>(null);
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    const priority = block.metadata?.priority || 'medium';

    const priorityColors = {
        low: 'bg-blue-100 text-blue-700 border-blue-200',
        medium: 'bg-yellow-100 text-yellow-700 border-yellow-200',
        high: 'bg-red-100 text-red-700 border-red-200'
    } as Record<string, string>;

    const priorityConfig = {
        low: { label: 'Low', color: 'bg-blue-500' },
        medium: { label: 'Medium', color: 'bg-yellow-500' },
        high: { label: 'High', color: 'bg-red-500' }
    };

    // Close menu when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setShowMenu(false);
            }
        };

        if (showMenu) {
            document.addEventListener('mousedown', handleClickOutside);
        }
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [showMenu]);

    // Auto-resize textarea and focus
    useEffect(() => {
        if (isEditing && textareaRef.current) {
            textareaRef.current.focus();
            textareaRef.current.style.height = 'auto';
            textareaRef.current.style.height = textareaRef.current.scrollHeight + 'px';
        }
    }, [isEditing]);

    const handleDragStart = (e: React.DragEvent) => {
        if (isEditing) {
            e.preventDefault();
            return;
        }
        e.dataTransfer.setData('application/json', JSON.stringify({
            type: 'task',
            id: block.id,
            fromColumn: block.metadata?.column_key || block.metadata?.status
        }));
        e.dataTransfer.effectAllowed = 'move';

        const target = e.currentTarget as HTMLElement;
        target.style.opacity = '0.4';
        target.style.transform = 'scale(0.95)';
    };

    const handleDragEnd = (e: React.DragEvent) => {
        const target = e.currentTarget as HTMLElement;
        target.style.opacity = '1';
        target.style.transform = 'scale(1)';
    };

    const handleSaveContent = () => {
        if (tempContent.trim() !== block.content) {
            onUpdate({ ...block, content: tempContent });
        }
        setIsEditing(false);
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSaveContent();
        } else if (e.key === 'Escape') {
            setTempContent(block.content || '');
            setIsEditing(false);
        }
    };

    const setPriority = (newPriority: string) => {
        onUpdate({
            ...block,
            metadata: { ...block.metadata, priority: newPriority },
            blockConfig: { ...(block.blockConfig || {}), priority: newPriority }
        });
        setShowMenu(false);
    };

    return (
        <div
            draggable={!isEditing}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            className={`group bg-white p-3 rounded-xl border shadow-sm hover:shadow-md transition-all duration-200 relative ${isEditing ? 'ring-2 ring-primary/20 border-primary' : 'border-gray-200 cursor-grab active:cursor-grabbing'
                }`}
        >
            <div className="flex justify-between items-start mb-2">
                {/* Priority Badge */}
                <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full border ${priorityColors[priority] || priorityColors.medium}`}>
                    {priority}
                </span>

                {/* Menu Trigger */}
                <div className="relative" ref={menuRef}>
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            setShowMenu(!showMenu);
                        }}
                        className={`text-gray-400 hover:text-gray-600 p-1 rounded-md hover:bg-gray-100 transition-all ${showMenu ? 'opacity-100 bg-gray-100' : 'opacity-0 group-hover:opacity-100'}`}
                    >
                        <MoreHorizontal className="h-4 w-4" />
                    </button>

                    {/* Context Menu */}
                    {showMenu && (
                        <div className="absolute right-0 top-full mt-1 w-48 bg-white rounded-xl shadow-xl border border-gray-100 z-50 overflow-hidden animate-in fade-in zoom-in-95 duration-100 origin-top-right">
                            <div className="p-1">
                                <div className="px-2 py-1.5 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                                    Priority
                                </div>
                                <div className="grid grid-cols-1 gap-0.5 mb-1">
                                    {(Object.keys(priorityConfig) as Array<keyof typeof priorityConfig>).map((p) => (
                                        <button
                                            key={p}
                                            onClick={() => setPriority(p)}
                                            className={`flex items-center gap-2 px-2 py-1.5 text-sm rounded-lg hover:bg-gray-50 transition-colors ${priority === p ? 'bg-primary/5 text-primary font-medium' : 'text-gray-700'}`}
                                        >
                                            <div className={`w-2 h-2 rounded-full ${priorityConfig[p].color}`} />
                                            {priorityConfig[p].label}
                                            {priority === p && <span className="ml-auto text-xs">✓</span>}
                                        </button>
                                    ))}
                                </div>
                                <div className="h-px bg-gray-100 my-1" />
                                <button
                                    onClick={() => onDelete(block.id)}
                                    className="w-full flex items-center gap-2 px-2 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                                >
                                    <Trash2 className="h-4 w-4" />
                                    Delete Task
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {/* Content Area */}
            <div className="min-h-[20px] mb-2">
                {isEditing ? (
                    <textarea
                        ref={textareaRef}
                        value={tempContent}
                        onChange={(e) => {
                            setTempContent(e.target.value);
                            e.target.style.height = 'auto';
                            e.target.style.height = e.target.scrollHeight + 'px';
                        }}
                        onBlur={handleSaveContent}
                        onKeyDown={handleKeyDown}
                        className="w-full text-sm font-medium text-gray-800 bg-transparent resize-none outline-none placeholder:text-gray-3000 leading-relaxed"
                        placeholder="Task description..."
                        rows={1}
                        onClick={(e) => e.stopPropagation()}
                    />
                ) : (
                    <p
                        onClick={() => {
                            setTempContent(block.content || '');
                            setIsEditing(true);
                        }}
                        className="text-sm font-medium text-gray-800 leading-relaxed cursor-text hover:text-primary transition-colors empty:text-gray-400 empty:italic"
                    >
                        {block.content || 'Untitled Task'}
                    </p>
                )}
            </div>

            {/* Footer */}
            <div className="flex items-center justify-between text-xs text-gray-400">
                <div className="flex items-center gap-2">
                    {block.metadata?.due_date && (
                        <div className="flex items-center gap-1 text-orange-500 bg-orange-50 px-1.5 py-0.5 rounded">
                            <Clock className="h-3 w-3" />
                            <span>{new Date(block.metadata.due_date).toLocaleDateString()}</span>
                        </div>
                    )}
                </div>
                {!isEditing && (
                    <GripVertical className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity text-gray-300" />
                )}
            </div>

            {/* Hover border effect */}
            <div className={`absolute inset-0 border-2 border-primary/20 rounded-xl pointer-events-none transition-opacity ${isEditing ? 'opacity-0' : 'opacity-0 group-hover:opacity-100'}`} />
        </div>
    );
}
