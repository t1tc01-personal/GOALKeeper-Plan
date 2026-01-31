import React from 'react';
import { Block } from '@/services/blockApi';

export interface FrameworkContainerProps {
    block: Block;
    children: Block[];
    onUpdate: (block: Block) => void;
    onChildCreate: (child: Partial<Block>) => Promise<Block>;
    onChildUpdate: (child: Block) => void;
    onChildDelete: (childId: string) => void;
}

export interface FrameworkItemProps {
    block: Block;
    frameworkContainer: Block;
    onUpdate: (block: Block) => void;
    onDelete: (blockId: string) => void;
}

export interface FrameworkConfig {
    type: string;
    name: string;
    displayName: string;
    description: string;
    icon: string;
    category: 'framework';
    containerComponent: React.ComponentType<FrameworkContainerProps>;
    itemComponent?: React.ComponentType<FrameworkItemProps>;
    defaultMetadata: Record<string, any>;
    supportsChildren: boolean;
}

class FrameworkRegistry {
    private frameworks = new Map<string, FrameworkConfig>();

    register(config: FrameworkConfig): void {
        if (this.frameworks.has(config.type)) {
            console.warn(`Framework ${config.type} is already registered. Overwriting.`);
        }
        this.frameworks.set(config.type, config);
    }

    get(type: string): FrameworkConfig | undefined {
        return this.frameworks.get(type);
    }

    has(type: string): boolean {
        return this.frameworks.has(type);
    }

    list(): FrameworkConfig[] {
        return Array.from(this.frameworks.values());
    }

    listByCategory(category: string): FrameworkConfig[] {
        return this.list().filter(fw => fw.category === category);
    }

    unregister(type: string): boolean {
        return this.frameworks.delete(type);
    }
}

export const frameworkRegistry = new FrameworkRegistry();

export default frameworkRegistry;
