'use client';

import React, { useState, useRef, useEffect } from 'react';
import { FrameworkContainerProps } from '@/shared/frameworks/registry';
import {
    Plus,
    Calendar as CalendarIcon,
    List as ListIcon,
    TrendingUp,
    ChevronLeft,
    ChevronRight,
    MoreHorizontal,
    Trash2,
    Palette,
    Check
} from 'lucide-react';
import dayjs from 'dayjs';
import { Block } from '@/services/blockApi';

// Initial color palette with complete style definitions for Tailwind JIT safety
const HABIT_THEMES: Record<string, {
    base: string;
    text: string;
    ring: string;
    border: string;
    shadow: string;
    lightBorder: string;
    lightText: string;
}> = {
    blue: {
        base: 'bg-blue-500', text: 'text-blue-600', ring: 'ring-blue-500',
        border: 'border-blue-500', shadow: 'shadow-blue-200',
        lightBorder: 'hover:border-blue-200', lightText: 'hover:text-blue-400'
    },
    green: {
        base: 'bg-green-500', text: 'text-green-600', ring: 'ring-green-500',
        border: 'border-green-500', shadow: 'shadow-green-200',
        lightBorder: 'hover:border-green-200', lightText: 'hover:text-green-400'
    },
    red: {
        base: 'bg-red-500', text: 'text-red-600', ring: 'ring-red-500',
        border: 'border-red-500', shadow: 'shadow-red-200',
        lightBorder: 'hover:border-red-200', lightText: 'hover:text-red-400'
    },
    purple: {
        base: 'bg-purple-500', text: 'text-purple-600', ring: 'ring-purple-500',
        border: 'border-purple-500', shadow: 'shadow-purple-200',
        lightBorder: 'hover:border-purple-200', lightText: 'hover:text-purple-400'
    },
    orange: {
        base: 'bg-orange-500', text: 'text-orange-600', ring: 'ring-orange-500',
        border: 'border-orange-500', shadow: 'shadow-orange-200',
        lightBorder: 'hover:border-orange-200', lightText: 'hover:text-orange-400'
    },
    pink: {
        base: 'bg-pink-500', text: 'text-pink-600', ring: 'ring-pink-500',
        border: 'border-pink-500', shadow: 'shadow-pink-200',
        lightBorder: 'hover:border-pink-200', lightText: 'hover:text-pink-400'
    },
};

const COLOR_KEYS = Object.keys(HABIT_THEMES);

interface HabitRowProps {
    habit: Block;
    onUpdate: (block: Block) => void;
    onDelete: (blockId: string) => void;
    toggleHabit: (habit: Block, date: string) => void;
}

function HabitRow({ habit, onUpdate, onDelete, toggleHabit }: HabitRowProps) {
    const [isEditing, setIsEditing] = useState(false);
    const [showMenu, setShowMenu] = useState(false);
    const [tempContent, setTempContent] = useState(habit.content || '');
    const menuRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    const colorName = habit.metadata?.color || 'blue';
    const theme = HABIT_THEMES[colorName] || HABIT_THEMES['blue'];

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

    useEffect(() => {
        if (isEditing && inputRef.current) {
            inputRef.current.focus();
        }
    }, [isEditing]);

    const handleSave = () => {
        if (tempContent.trim() !== habit.content) {
            onUpdate({ ...habit, content: tempContent });
        }
        setIsEditing(false);
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
            handleSave();
        } else if (e.key === 'Escape') {
            setTempContent(habit.content || '');
            setIsEditing(false);
        }
    };

    const setColor = (color: string) => {
        onUpdate({
            ...habit,
            metadata: { ...habit.metadata, color },
            blockConfig: { ...(habit.blockConfig || {}), color }
        });
        setShowMenu(false);
    };

    // Calculate streak (simplified)
    const history = habit.metadata?.history || {};
    let streak = 0;
    const today = dayjs();
    for (let i = 0; i < 365; i++) {
        const date = today.subtract(i, 'day').format('YYYY-MM-DD');
        if (history[date]) {
            streak++;
        } else if (i > 0) {
            const yesterday = today.subtract(1, 'day').format('YYYY-MM-DD');
            if (i === 0 && !history[date]) continue;
            break;
        }
    }

    return (
        <div className="group">
            <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3 flex-1">
                    {/* Color Dot / Menu Trigger */}
                    <div className="relative" ref={menuRef}>
                        <button
                            onClick={(e) => { e.stopPropagation(); setShowMenu(!showMenu); }}
                            className={`w-4 h-4 rounded-full ${theme.base} transition-transform hover:scale-110 flex items-center justify-center`}
                        >
                            <div className="w-1.5 h-1.5 bg-white rounded-full opacity-50" />
                        </button>

                        {showMenu && (
                            <div className="absolute left-0 top-full mt-2 w-48 bg-white rounded-xl shadow-xl border border-gray-100 z-50 animate-in fade-in zoom-in-95 duration-100 p-2">
                                <div className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2 px-1">Habit Color</div>
                                <div className="grid grid-cols-6 gap-2 mb-3">
                                    {COLOR_KEYS.map(key => (
                                        <button
                                            key={key}
                                            onClick={() => setColor(key)}
                                            className={`w-6 h-6 rounded-full ${HABIT_THEMES[key].base} hover:scale-110 transition-transform flex items-center justify-center ring-2 ring-transparent ${colorName === key ? 'ring-gray-300' : ''}`}
                                            title={key}
                                        >
                                            {colorName === key && <Check className="w-3 h-3 text-white" />}
                                        </button>
                                    ))}
                                </div>
                                <div className="h-px bg-gray-100 my-1" />
                                <button
                                    onClick={() => onDelete(habit.id)}
                                    className="w-full flex items-center gap-2 px-2 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                                >
                                    <Trash2 className="h-4 w-4" />
                                    Delete Habit
                                </button>
                            </div>
                        )}
                    </div>

                    {/* Editable Name */}
                    <div className="flex-1">
                        {isEditing ? (
                            <input
                                ref={inputRef}
                                value={tempContent}
                                onChange={(e) => setTempContent(e.target.value)}
                                onBlur={handleSave}
                                onKeyDown={handleKeyDown}
                                className="w-full font-bold text-gray-700 bg-transparent border-b-2 border-primary/20 focus:border-primary outline-none px-1 py-0.5"
                                placeholder="Habit name..."
                                onClick={(e) => e.stopPropagation()}
                            />
                        ) : (
                            <h3
                                onClick={() => { setIsEditing(true); setTempContent(habit.content || ''); }}
                                className="font-bold text-gray-700 cursor-text hover:text-primary transition-colors truncate"
                            >
                                {habit.content || 'Untitled Habit'}
                            </h3>
                        )}
                    </div>
                </div>

                <div className="flex items-center gap-4 text-xs font-medium text-gray-400">
                    <div className="flex items-center gap-1 bg-gray-50 px-2 py-1 rounded-md">
                        <TrendingUp className={`h-3 w-3 ${streak > 0 ? 'text-green-500' : 'text-gray-400'}`} />
                        <span className={streak > 0 ? 'text-green-600' : ''}>{streak} Day Streak</span>
                    </div>
                </div>
            </div>

            {/* Weekly progress dots */}
            <div className="flex gap-2">
                {Array.from({ length: 14 }, (_, i) => {
                    const date = dayjs().subtract(13 - i, 'day').format('YYYY-MM-DD');
                    const isCompleted = habit.metadata?.history?.[date];
                    const isToday = dayjs().format('YYYY-MM-DD') === date;

                    return (
                        <button
                            key={date}
                            onClick={() => toggleHabit(habit, date)}
                            className={`flex-1 h-10 rounded-lg flex flex-col items-center justify-center transition-all duration-200 border ${isCompleted
                                ? `${theme.base} ${theme.border} text-white shadow-md ${theme.shadow} scale-105`
                                : `bg-white border-gray-100 text-gray-300 ${theme.lightBorder} ${theme.lightText}`
                                } ${isToday ? 'ring-2 ring-primary/20 ring-offset-1' : ''}`}
                            title={dayjs(date).format('dddd, MMM D')}
                        >
                            <span className="text-[10px] font-bold">{dayjs(date).format('D')}</span>
                            <span className="text-[8px] uppercase">{dayjs(date).format('dd')}</span>
                        </button>
                    );
                })}
            </div>
        </div>
    );
}

export function HabitTracker({
    block,
    children,
    onChildCreate,
    onChildUpdate,
    onChildDelete,
    onUpdate
}: FrameworkContainerProps) {
    const [view, setView] = useState<'list' | 'calendar'>(block.metadata?.view_mode || 'list');
    const [currentMonth, setCurrentMonth] = useState(dayjs());

    // Title Editing State
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

    const handleSetView = (newView: 'list' | 'calendar') => {
        setView(newView);
        onUpdate({
            ...block,
            metadata: { ...block.metadata, view_mode: newView },
            blockConfig: { ...(block.blockConfig || {}), view_mode: newView }
        });
    };

    const daysInMonth = currentMonth.daysInMonth();
    const firstDayOfMonth = currentMonth.startOf('month').day();

    // Create array for calendar days
    const calendarDays = Array.from({ length: daysInMonth }, (_, i) => i + 1);
    const blanks = Array.from({ length: firstDayOfMonth === 0 ? 6 : firstDayOfMonth - 1 }, (_, i) => i);

    const toggleHabit = (habitBlock: Block, day: string) => {
        const history = habitBlock.metadata?.history || {};
        const newState = !history[day];

        onChildUpdate({
            ...habitBlock,
            metadata: {
                ...habitBlock.metadata,
                history: {
                    ...history,
                    [day]: newState
                }
            },
            blockConfig: {
                ...(habitBlock.blockConfig || {}),
                history: {
                    ...history,
                    [day]: newState
                }
            }
        });
    };

    return (
        <div className="flex flex-col w-full bg-white rounded-2xl border border-gray-100 shadow-sm animate-in fade-in duration-500">
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 bg-gray-50/50 border-b border-gray-100 rounded-t-2xl">
                <div className="flex items-center gap-4">
                    {isEditingTitle ? (
                        <input
                            ref={titleInputRef}
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            onBlur={handleTitleSave}
                            onKeyDown={handleTitleKeyDown}
                            className="text-lg font-bold text-gray-800 bg-transparent outline-none border-b-2 border-primary/20 focus:border-primary px-1 min-w-[200px]"
                            placeholder="Habit Tracker"
                        />
                    ) : (
                        <h2
                            onClick={() => { setIsEditingTitle(true); setTitle(block.content || ''); }}
                            className="text-lg font-bold text-gray-800 cursor-text hover:text-gray-600 transition-colors px-1 border-b-2 border-transparent hover:border-gray-200"
                        >
                            {block.content || 'Habit Tracker'}
                        </h2>
                    )}

                    <div className="flex bg-gray-200/50 p-1 rounded-lg">
                        <button
                            onClick={() => handleSetView('list')}
                            className={`p-1.5 rounded-md transition-all ${view === 'list' ? 'bg-white shadow-sm text-primary' : 'text-gray-500 hover:text-gray-700'}`}
                            title="List View"
                        >
                            <ListIcon className="h-4 w-4" />
                        </button>
                        <button
                            onClick={() => handleSetView('calendar')}
                            className={`p-1.5 rounded-md transition-all ${view === 'calendar' ? 'bg-white shadow-sm text-primary' : 'text-gray-500 hover:text-gray-700'}`}
                            title="Calendar View"
                        >
                            <CalendarIcon className="h-4 w-4" />
                        </button>
                    </div>
                </div>

                <div className="flex items-center gap-3">
                    {view === 'calendar' && (
                        <div className="flex items-center gap-2">
                            <button
                                onClick={() => setCurrentMonth(dayjs())}
                                className="text-xs font-semibold text-primary hover:text-primary/80 hover:bg-primary/5 px-2 py-1 rounded-md transition-colors"
                            >
                                Today
                            </button>
                            <div className="flex items-center bg-white border border-gray-200 rounded-lg px-2 py-1">
                                <button onClick={() => setCurrentMonth(prev => prev.subtract(1, 'month'))} className="p-1 hover:bg-gray-50 rounded text-gray-500">
                                    <ChevronLeft className="h-4 w-4" />
                                </button>
                                <span className="text-xs font-bold text-gray-700 px-3 min-w-[100px] text-center">
                                    {currentMonth.format('MMMM YYYY')}
                                </span>
                                <button onClick={() => setCurrentMonth(prev => prev.add(1, 'month'))} className="p-1 hover:bg-gray-50 rounded text-gray-500">
                                    <ChevronRight className="h-4 w-4" />
                                </button>
                            </div>
                        </div>
                    )}

                    <button
                        onClick={() => onChildCreate({
                            type: 'habit_item',
                            content: 'New Habit',
                            metadata: { history: {}, color: 'blue' }
                        })}
                        className="flex items-center gap-2 px-3 py-1.5 bg-primary text-white text-xs font-bold rounded-lg hover:bg-primary/90 transition-all shadow-lg shadow-primary/20 active:scale-95"
                    >
                        <Plus className="h-4 w-4" />
                        Add Habit
                    </button>
                </div>
            </div>

            {/* Content */}
            <div className="p-6">
                {view === 'list' ? (
                    <div className="space-y-8">
                        {children.map(habit => (
                            <HabitRow
                                key={habit.id}
                                habit={habit}
                                onUpdate={onChildUpdate}
                                onDelete={onChildDelete}
                                toggleHabit={toggleHabit}
                            />
                        ))}

                        {children.length === 0 && (
                            <div className="flex flex-col items-center justify-center py-16 bg-gray-50/50 rounded-2xl border-2 border-dashed border-gray-100 group cursor-pointer"
                                onClick={() => onChildCreate({
                                    type: 'habit_item',
                                    content: 'New Habit',
                                    metadata: { history: {}, color: 'blue' }
                                })}>
                                <div className="p-4 bg-white rounded-full shadow-sm mb-3 group-hover:scale-110 transition-transform">
                                    <Plus className="h-6 w-6 text-primary" />
                                </div>
                                <h3 className="text-sm font-bold text-gray-700">No habits yet</h3>
                                <p className="text-xs text-gray-400 mt-1">Click to start tracking your first habit</p>
                            </div>
                        )}
                    </div>
                ) : (
                    <div className="grid grid-cols-7 gap-3">
                        {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map(day => (
                            <div key={day} className="text-center text-[10px] font-bold text-gray-400 uppercase tracking-widest py-2">
                                {day}
                            </div>
                        ))}
                        {blanks.map(i => <div key={`blank-${i}`} />)}
                        {calendarDays.map(day => {
                            const fullDate = currentMonth.date(day).format('YYYY-MM-DD');
                            const completions = children.filter(h => h.metadata?.history?.[fullDate]);
                            const isToday = dayjs().format('YYYY-MM-DD') === fullDate;

                            return (
                                <div
                                    key={day}
                                    className={`aspect-square rounded-2xl border p-2 flex flex-col items-center justify-between transition-all hover:shadow-md ${isToday
                                        ? 'bg-primary/5 border-primary/20'
                                        : 'bg-white border-gray-100 hover:border-gray-200'
                                        }`}
                                >
                                    <span className={`text-xs font-bold ${isToday ? 'text-primary' : 'text-gray-400'}`}>
                                        {day}
                                    </span>
                                    <div className="flex flex-wrap gap-1.5 justify-center content-end w-full">
                                        {completions.map(h => (
                                            <div
                                                key={h.id}
                                                className={`w-2 h-2 rounded-full cursor-help transition-transform hover:scale-150 bg-${h.metadata?.color || 'blue'}-500 shadow-sm`}
                                                title={h.content}
                                            />
                                        ))}
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
        </div>
    );
}
export default HabitTracker;
