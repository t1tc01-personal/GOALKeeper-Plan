'use client';

import React from 'react';
import { FrameworkContainerProps } from '@/shared/frameworks/registry';

export function DummyKanbanBoard({
    block,
    children,
    onUpdate,
    onChildCreate,
}: FrameworkContainerProps) {
    const columns = block.metadata?.columns || [];

    return (
        <div className="p-4 bg-gray-50 min-h-[300px]">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-bold">{block.content || 'Untitled Kanban'}</h2>
                <button
                    onClick={() => onUpdate({ ...block, content: 'Updated Kanban ' + Date.now() })}
                    className="text-xs bg-blue-500 text-white px-2 py-1 rounded"
                >
                    Rename
                </button>
            </div>

            <div className="flex gap-4 overflow-x-auto pb-4">
                {columns.map((column: any) => (
                    <div key={column.key} className="flex-shrink-0 w-72 bg-gray-200 rounded-lg p-3">
                        <div className="flex justify-between items-center mb-3">
                            <h3 className="font-semibold">{column.label}</h3>
                            <span className="text-xs bg-gray-300 px-1.5 py-0.5 rounded-full">
                                {children.filter(c => c.metadata?.column_key === column.key).length}
                            </span>
                        </div>

                        <div className="space-y-2 mb-3">
                            {children
                                .filter(c => c.metadata?.column_key === column.key)
                                .map(task => (
                                    <div key={task.id} className="bg-white p-2 rounded shadow-sm text-sm">
                                        {task.content || 'Empty task'}
                                    </div>
                                ))}
                        </div>

                        <button
                            onClick={() => onChildCreate({
                                type: 'task_item',
                                content: 'New Task',
                                metadata: { column_key: column.key, status: column.key }
                            })}
                            className="w-full py-1 text-xs text-gray-500 hover:bg-gray-300 rounded transition-colors text-left px-2"
                        >
                            + Add Task
                        </button>
                    </div>
                ))}

                <div className="flex-shrink-0 w-72 h-12 border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center text-gray-400 text-sm">
                    + Add Column
                </div>
            </div>
        </div>
    );
}
