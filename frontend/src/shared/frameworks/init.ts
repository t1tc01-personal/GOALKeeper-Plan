import { frameworkRegistry } from './registry';
import { KanbanBoard } from '@/components/frameworks/kanban/KanbanBoard';
import { HabitTracker } from '@/components/frameworks/habit/HabitTracker';

export function initializeFrameworks() {
    // Register Kanban
    frameworkRegistry.register({
        type: 'kanban',
        name: 'kanban',
        displayName: 'Kanban Board',
        description: 'Visual task management board with columns',
        icon: 'layout',
        category: 'framework',
        containerComponent: KanbanBoard,
        supportsChildren: true,
        defaultMetadata: {
            framework_type: 'kanban',
            columns: [
                { key: 'todo', label: 'To Do', color: 'gray', position: 0 },
                { key: 'doing', label: 'In Progress', color: 'blue', position: 1 },
                { key: 'done', label: 'Done', color: 'green', position: 2 }
            ],
            settings: {
                allow_drag_drop: true,
                show_card_count: true
            }
        }
    });

    // Register Habit Tracker
    frameworkRegistry.register({
        type: 'habit',
        name: 'habit',
        displayName: 'Habit Tracker',
        description: 'Track your daily progress and streaks',
        icon: 'calendar',
        category: 'framework',
        containerComponent: HabitTracker,
        supportsChildren: true,
        defaultMetadata: {
            framework_type: 'habit',
            display_mode: 'list',
            settings: {
                show_stats: true,
                allow_custom_habits: true
            }
        }
    });
}
