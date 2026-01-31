import { Block } from '@/services/blockApi';

export function isFrameworkContainerBlock(block: Block): boolean {
    const frameworkTypes = [
        'framework_container',
        'kanban',
        'gantt',
        'wbs',
        'action_plan',
        'habit_tracker',
        'learning_roadmap'
    ];

    const blockType = block.type_id || block.type || '';

    return frameworkTypes.includes(blockType);
}

export function getFrameworkType(block: Block): string | undefined {
    return block.metadata?.framework_type as string | undefined;
}

export function isFrameworkChildBlock(block: Block): boolean {
    const childTypes = [
        'task_item',
        'habit_item',
        'wbs_node',
        'learning_node',
        'milestone'
    ];

    const blockType = block.type_id || block.type || '';

    return childTypes.includes(blockType);
}

export function createFrameworkMetadata(frameworkType: string, customMetadata: Record<string, any> = {}): Record<string, any> {
    const baseMetadata = {
        framework_type: frameworkType,
        ...customMetadata
    };

    const defaults: Record<string, Record<string, any>> = {
        kanban: {
            columns: [
                { key: 'todo', label: 'To Do', color: 'gray', position: 0 },
                { key: 'in_progress', label: 'In Progress', color: 'blue', position: 1 },
                { key: 'done', label: 'Done', color: 'green', position: 2 }
            ],
            settings: {
                allow_drag_drop: true,
                show_card_count: true
            }
        },
        habit: {
            display_mode: 'calendar',
            settings: {
                start_of_week: 'monday',
                show_streak_badge: true
            }
        }
    };

    return {
        ...baseMetadata,
        ...(defaults[frameworkType] || {})
    };
}
