'use client';

import React, { useState, useRef, useEffect } from 'react';
import { FrameworkContainerProps } from '@/shared/frameworks/registry';
import { KanbanColumn } from './KanbanColumn';
import { MoreHorizontal, Plus, Settings2, Filter } from 'lucide-react';
import { Block } from '@/services/blockApi';

export function KanbanBoard({
    block,
    children,
    onUpdate,
    onChildCreate,
    onChildUpdate,
    onChildDelete,
}: FrameworkContainerProps) {
    const columns = block.metadata?.columns || [
        { key: 'todo', label: 'To Do', color: 'gray', position: 0 },
        { key: 'in_progress', label: 'In Progress', color: 'blue', position: 1 },
        { key: 'done', label: 'Done', color: 'green', position: 2 }
    ];

    const handleTaskCreate = async (columnKey: string) => {
        await onChildCreate({
            type: 'task_item',
            content: '', // Empty means start editing
            metadata: {
                column_key: columnKey,
                status: columnKey,
                priority: 'medium'
            }
        });
    };

    const handleTaskMove = (taskId: string, targetColumnKey: string) => {
        const task = children.find(c => c.id === taskId);
        if (task) {
            onChildUpdate({
                ...task,
                metadata: {
                    ...task.metadata,
                    column_key: targetColumnKey,
                    status: targetColumnKey
                },
                blockConfig: {
                    ...(task.blockConfig || {}),
                    column_key: targetColumnKey,
                    status: targetColumnKey
                }
            });
        }
    };

    const [isEditingTitle, setIsEditingTitle] = useState(false);
    const [title, setTitle] = useState(block.content || '');
    const titleInputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (isEditingTitle && titleInputRef.current) {
            titleInputRef.current.focus();
        }
    }, [isEditingTitle]);

    const handleTitleSave = () => {
        if (title.trim() !== block.content) {
            onUpdate({ ...block, content: title });
        }
        setIsEditingTitle(false);
    };

    const handleTitleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
            handleTitleSave();
        } else if (e.key === 'Escape') {
            setTitle(block.content || '');
            setIsEditingTitle(false);
        }
    };

    const handleAddColumn = () => {
        const label = window.prompt("Enter column name:", "New Column");
        if (!label) return;

        const key = label.toLowerCase().replace(/[^a-z0-9]/g, '_') + '_' + Date.now();
        const colors = ['gray', 'blue', 'green', 'yellow', 'purple', 'red'];
        const color = colors[Math.floor(Math.random() * colors.length)];

        const newColumn = {
            key,
            label,
            color,
            position: columns.length
        };

        const newColumns = [...columns, newColumn];

        onUpdate({
            ...block,
            metadata: { ...block.metadata, columns: newColumns },
            blockConfig: { ...(block.blockConfig || {}), columns: newColumns }
        });
    };

    return (
        <div className="flex flex-col w-full bg-white rounded-2xl border border-gray-200 shadow-sm">
            {/* Board Header/Toolbar */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100 bg-white/80 backdrop-blur-md sticky top-0 z-10">
                <div className="flex items-center gap-4">
                    {isEditingTitle ? (
                        <input
                            ref={titleInputRef}
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            onBlur={handleTitleSave}
                            onKeyDown={handleTitleKeyDown}
                            className="text-xl font-bold text-gray-900 leading-tight bg-transparent outline-none border-b-2 border-primary/20 focus:border-primary px-1 min-w-[200px]"
                            placeholder="Untitled Project"
                        />
                    ) : (
                        <h2
                            onClick={() => { setIsEditingTitle(true); setTitle(block.content || ''); }}
                            className="text-xl font-bold text-gray-900 leading-tight cursor-text hover:text-gray-600 transition-colors px-1 border-b-2 border-transparent hover:border-gray-200"
                        >
                            {block.content || 'Untitled Project'}
                        </h2>
                    )}
                    <div className="flex items-center gap-1 ml-2">
                        <button className="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-50 rounded-lg transition-all">
                            <Filter className="h-4 w-4" />
                        </button>
                        <button className="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-50 rounded-lg transition-all">
                            <Settings2 className="h-4 w-4" />
                        </button>
                    </div>
                </div>

                <div className="flex items-center gap-2">
                    <button className="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-50 rounded-lg transition-all ml-1">
                        <MoreHorizontal className="h-5 w-5" />
                    </button>
                </div>
            </div>

            {/* Columns Container */}
            <div className="flex-1 overflow-x-auto min-h-[600px] p-6">
                <div className="flex gap-8 h-full min-w-max">
                    {columns.sort((a: any, b: any) => a.position - b.position).map((column: any) => (
                        <KanbanColumn
                            key={column.key}
                            column={column}
                            tasks={children.filter(c =>
                                (c.metadata?.column_key === column.key) ||
                                (c.blockConfig?.column_key === column.key) ||
                                (c.metadata?.status === column.key)
                            )}
                            frameworkContainer={block}
                            onTaskCreate={handleTaskCreate}
                            onTaskUpdate={onChildUpdate}
                            onTaskDelete={onChildDelete}
                            onTaskMove={handleTaskMove}
                        />
                    ))}

                    {/* New Column Button */}
                    <button
                        onClick={handleAddColumn}
                        className="flex-shrink-0 w-80 h-[100px] border-2 border-dashed border-gray-100 rounded-2xl flex flex-col items-center justify-center text-gray-400 hover:text-gray-600 hover:border-gray-200 hover:bg-gray-50 transition-all group mt-12"
                    >
                        <Plus className="h-6 w-6 mb-2 group-hover:scale-110 transition-transform" />
                        <span className="text-sm font-medium">Add another column</span>
                    </button>
                </div>
            </div>

            {/* Board Footer / Status bar (optional) */}
            <div className="px-6 py-2 border-t border-gray-100 flex items-center justify-between text-[10px] text-gray-400 font-medium">
                <div className="flex items-center gap-4">
                    <span>{children.length} Tasks</span>
                    <span>Last sync: Just now</span>
                </div>
                <div className="flex items-center gap-1">
                    <div className="w-1.5 h-1.5 rounded-full bg-green-500" />
                    <span>Online</span>
                </div>
            </div>
        </div>
    );
}
