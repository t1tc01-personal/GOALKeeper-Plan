'use client';

import React, { useState } from 'react';
import { Block } from '@/services/blockApi';
import { KanbanCard } from './KanbanCard';
import { Plus, MoreVertical } from 'lucide-react';

interface KanbanColumnProps {
    column: {
        key: string;
        label: string;
        color: string;
    };
    tasks: Block[];
    frameworkContainer: Block;
    onTaskCreate: (columnKey: string) => void;
    onTaskUpdate: (task: Block) => void;
    onTaskDelete: (taskId: string) => void;
    onTaskMove: (taskId: string, targetColumnKey: string) => void;
}

export function KanbanColumn({
    column,
    tasks,
    frameworkContainer,
    onTaskCreate,
    onTaskUpdate,
    onTaskDelete,
    onTaskMove
}: KanbanColumnProps) {
    const [isOver, setIsOver] = useState(false);

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        // Security check: only allow if it contains our type
        if (!e.dataTransfer.types.includes('application/json')) {
            e.dataTransfer.dropEffect = 'none';
            return;
        }
        e.dataTransfer.dropEffect = 'move';
        setIsOver(true);
    };

    const handleDragLeave = () => {
        setIsOver(false);
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setIsOver(false);

        // Security: Only accept JSON data
        if (!e.dataTransfer.types.includes('application/json')) {
            return;
        }

        try {
            const data = JSON.parse(e.dataTransfer.getData('application/json'));
            if (data.type === 'task' && data.fromColumn !== column.key) {
                onTaskMove(data.id, column.key);
            }
        } catch (err) {
            console.error('Failed to parse drop data', err);
        }
    };

    const columnColorMap = {
        gray: 'bg-gray-100',
        blue: 'bg-blue-50',
        green: 'bg-green-50',
        yellow: 'bg-yellow-50',
        purple: 'bg-purple-50',
        red: 'bg-red-50'
    } as Record<string, string>;

    const headerColorMap = {
        gray: 'bg-gray-400',
        blue: 'bg-blue-500',
        green: 'bg-green-500',
        yellow: 'bg-yellow-500',
        purple: 'bg-purple-500',
        red: 'bg-red-500'
    } as Record<string, string>;

    return (
        <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            className={`flex-shrink-0 w-80 flex flex-col h-full rounded-2xl transition-colors duration-200 ${isOver ? 'bg-primary/5 ring-2 ring-primary/20 ring-inset' : 'bg-transparent'
                }`}
        >
            {/* Column Header */}
            <div className="flex items-center justify-between px-3 py-4">
                <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${headerColorMap[column.color] || headerColorMap.gray}`} />
                    <h3 className="font-bold text-gray-700 text-sm">{column.label}</h3>
                    <span className="text-[10px] font-bold text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">
                        {tasks.length}
                    </span>
                </div>
                <button className="text-gray-400 hover:text-gray-600 p-1 rounded-lg hover:bg-gray-100 transition-colors">
                    <MoreVertical className="h-4 w-4" />
                </button>
            </div>

            {/* Task List */}
            <div className={`flex-1 overflow-y-auto p-2 space-y-3 min-h-[150px] transition-all duration-200 ${isOver ? 'bg-primary/5' : ''}`}>
                {tasks.map(task => (
                    <KanbanCard
                        key={task.id}
                        block={task}
                        frameworkContainer={frameworkContainer}
                        onUpdate={onTaskUpdate}
                        onDelete={onTaskDelete}
                    />
                ))}

                {/* Drop target placeholder */}
                {isOver && tasks.length === 0 && (
                    <div className="border-2 border-dashed border-primary/20 rounded-xl h-24 flex items-center justify-center text-primary/40 text-xs font-medium">
                        Drop here
                    </div>
                )}
            </div>

            {/* Add Task Button */}
            <div className="p-2">
                <button
                    onClick={() => onTaskCreate(column.key)}
                    className="w-full flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-xl transition-all duration-200 group"
                >
                    <div className="bg-gray-100 group-hover:bg-white p-1 rounded-md transition-colors shadow-sm ring-1 ring-gray-200">
                        <Plus className="h-3 w-3" />
                    </div>
                    New task
                </button>
            </div>
        </div>
    );
}
