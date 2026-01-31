'use client';

import React from 'react';
import { Block } from '@/services/blockApi';
import { frameworkRegistry, FrameworkContainerProps } from '@/shared/frameworks/registry';
import { AlertCircle } from 'lucide-react';

interface FrameworkBlockProps extends FrameworkContainerProps {
    block: Block;
    children: Block[];
}

export function FrameworkBlock({
    block,
    children,
    onUpdate,
    onChildCreate,
    onChildUpdate,
    onChildDelete,
}: FrameworkBlockProps) {
    const frameworkType = (block.metadata?.framework_type || block.blockConfig?.framework_type) as string | undefined;

    if (!frameworkType) {
        return (
            <div className="border border-yellow-300 bg-yellow-50 rounded-lg p-4 my-2">
                <div className="flex items-center gap-2 text-yellow-800">
                    <AlertCircle className="h-5 w-5" />
                    <p className="text-sm font-medium">
                        Framework block is missing framework_type in metadata/blockConfig
                    </p>
                </div>
                <pre className="mt-2 text-xs text-yellow-700 overflow-auto">
                    metadata: {JSON.stringify(block.metadata, null, 2)}
                    {'\n'}blockConfig: {JSON.stringify(block.blockConfig, null, 2)}
                </pre>
            </div>
        );
    }

    const frameworkConfig = frameworkRegistry.get(frameworkType);

    if (!frameworkConfig) {
        return (
            <div className="border border-red-300 bg-red-50 rounded-lg p-4 my-2">
                <div className="flex items-center gap-2 text-red-800">
                    <AlertCircle className="h-5 w-5" />
                    <p className="text-sm font-medium">
                        Unknown framework type: <code className="font-mono">{frameworkType}</code>
                    </p>
                </div>
                <p className="mt-2 text-xs text-red-600">
                    Available frameworks: {frameworkRegistry.list().map(fw => fw.type).join(', ') || 'None registered'}
                </p>
            </div>
        );
    }

    const FrameworkComponent = frameworkConfig.containerComponent;

    return (
        <div
            className="framework-block-wrapper my-3"
            data-framework-type={frameworkType}
            data-block-id={block.id}
        >
            <FrameworkComponent
                block={block}
                children={children}
                onUpdate={onUpdate}
                onChildCreate={onChildCreate}
                onChildUpdate={onChildUpdate}
                onChildDelete={onChildDelete}
            />
        </div>
    );
}

export default FrameworkBlock;
