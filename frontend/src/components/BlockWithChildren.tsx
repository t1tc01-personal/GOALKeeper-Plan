import { type Block } from '@/services/blockApi';
import { InlineBlockEditor } from './InlineBlockEditor';
import { FrameworkBlock } from './frameworks/FrameworkBlock';
import { isFrameworkContainerBlock } from '@/shared/frameworks';

interface BlockWithChildrenProps {
    block: Block;
    allBlocks: Block[];
    focusedBlockId: string | null;
    cursorPositions: Map<string, number>;
    draggedBlockId: string | null;
    dragOverBlockId: string | null;
    blockRefsMap: React.MutableRefObject<Map<string, HTMLDivElement>>;

    // Handlers
    onContentChange: (blockId: string, content: string) => void;
    onEnter: (blockId: string, content?: string) => void;
    onBackspace: (blockId: string) => void;
    onMerge: (blockId: string) => void;
    onTypeChange: (blockId: string, newType: string) => void;
    onIndent: (blockId: string) => void;
    onOutdent: (blockId: string) => void;
    onFocus: (blockId: string) => void;
    onArrowUp: (blockId: string) => void;
    onArrowDown: (blockId: string) => void;
    onUpdateBlock: (blockId: string, data: { content?: string; type?: string; blockConfig?: Record<string, any> }) => void;
    onCreateBlock: (afterBlockId: string, type?: string, content?: string, parentBlockId?: string | null) => Block;
    onDragStart: (blockId: string) => void;
    onDragOver: (e: React.DragEvent<HTMLDivElement>, blockId: string) => void;
    onDragLeave: (e: React.DragEvent<HTMLDivElement>) => void;
    onDrop: (e: React.DragEvent<HTMLDivElement>, blockId: string) => void;

    // For numbered list
    calculateListNumber: (blockId: string) => number;

    // Indentation level (for nested rendering)
    indentLevel?: number;
}

/**
 * BlockWithChildren - Recursively renders a block and its children
 * Supports Notion-style toggle blocks with collapsible child content
 */
export function BlockWithChildren({
    block,
    allBlocks,
    focusedBlockId,
    cursorPositions,
    draggedBlockId,
    dragOverBlockId,
    blockRefsMap,
    onContentChange,
    onEnter,
    onBackspace,
    onMerge,
    onTypeChange,
    onIndent,
    onOutdent,
    onFocus,
    onArrowUp,
    onArrowDown,
    onUpdateBlock,
    onCreateBlock,
    onDragStart,
    onDragOver,
    onDragLeave,
    onDrop,
    calculateListNumber,
    indentLevel = 0,
}: BlockWithChildrenProps) {
    // Find child blocks of this block
    const childBlocks = allBlocks.filter(b => b.parent_block_id === block.id);

    // Get block type and collapsed state
    const blockType = block.type_id || block.type || 'text';
    const isToggle = blockType === 'toggle';
    const isCollapsed = isToggle && (block.blockConfig?.collapsed ?? block.metadata?.collapsed ?? false);
    const isFramework = isFrameworkContainerBlock(block);

    // DEBUG: Log collapse state for toggles with children
    if (isToggle && childBlocks.length > 0) {
        console.log('[BlockWithChildren] Toggle rendering:', {
            blockId: block.id,
            isToggle,
            isCollapsed,
            blockConfig: block.blockConfig,
            metadata: block.metadata,
            childCount: childBlocks.length,
            willRenderChildren: !isCollapsed,
        });
    }

    return (
        <div className="block-with-children">
            {/* Render the block itself */}
            <div
                ref={(el) => {
                    if (el) blockRefsMap.current.set(block.id, el);
                }}
                draggable={true}
                onDragStart={() => onDragStart(block.id)}
                onDragOver={(e) => onDragOver(e, block.id)}
                onDragLeave={(e) => onDragLeave(e)}
                onDrop={(e) => onDrop(e, block.id)}
                className={`block-editor-item transition-all duration-150 ${draggedBlockId === block.id ? 'opacity-50 cursor-grabbing' : ''
                    } ${dragOverBlockId === block.id ? 'border-t-2 border-blue-400 bg-blue-50' : ''
                    }`}
                style={{
                    paddingLeft: indentLevel > 0 ? `${indentLevel * 24}px` : '0',
                    border: dragOverBlockId === block.id ? 'none' : 'none',
                    outline: 'none',
                    boxShadow: 'none',
                }}
            >
                {isFramework ? (
                    <FrameworkBlock
                        block={block}
                        children={childBlocks}
                        onUpdate={(updatedBlock) => onUpdateBlock(block.id, {
                            content: updatedBlock.content,
                            blockConfig: updatedBlock.blockConfig
                        })}
                        onChildCreate={async (childData) => {
                            console.log('[BlockWithChildren] Creating framework child:', childData);
                            // Find the last child in this parent to insert after
                            const lastChild = childBlocks.length > 0
                                ? childBlocks[childBlocks.length - 1]
                                : null;

                            const newBlock = onCreateBlock(
                                lastChild?.id || block.id,
                                childData.type,
                                childData.content,
                                block.id // Explicit parent
                            );

                            // Immediately update metadata if provided
                            if (childData.blockConfig || childData.metadata) {
                                onUpdateBlock(newBlock.id, {
                                    blockConfig: childData.blockConfig || childData.metadata
                                });
                            }

                            return newBlock;
                        }}
                        onChildUpdate={(childBlock) => onUpdateBlock(childBlock.id, {
                            content: childBlock.content,
                            blockConfig: childBlock.blockConfig
                        })}
                        onChildDelete={(childId) => onBackspace(childId)}
                    />
                ) : (
                    <InlineBlockEditor
                        block={block}
                        isFocused={focusedBlockId === block.id}
                        onContentChange={onContentChange}
                        onEnter={onEnter}
                        onBackspace={onBackspace}
                        onMerge={onMerge}
                        onTypeChange={onTypeChange}
                        onIndent={onIndent}
                        onOutdent={onOutdent}
                        onFocus={() => onFocus(block.id)}
                        onArrowUp={() => onArrowUp(block.id)}
                        onArrowDown={() => onArrowDown(block.id)}
                        autoFocus={false}
                        restoreCursorPosition={cursorPositions.get(block.id) ?? null}
                        onUpdateBlock={onUpdateBlock}
                        listNumber={calculateListNumber(block.id)}
                    />
                )}
            </div>

            {/* Render child blocks IF it's not a framework container (framework handles its own children)
                OR if it's a toggle that is not collapsed.
            */}
            {childBlocks.length > 0 && !isCollapsed && !isFramework && (
                <div className="block-children">
                    {childBlocks
                        .sort((a, b) => (a.position || 0) - (b.position || 0))
                        .map((childBlock) => (
                            <BlockWithChildren
                                key={childBlock.id}
                                block={childBlock}
                                allBlocks={allBlocks}
                                focusedBlockId={focusedBlockId}
                                cursorPositions={cursorPositions}
                                draggedBlockId={draggedBlockId}
                                dragOverBlockId={dragOverBlockId}
                                blockRefsMap={blockRefsMap}
                                onContentChange={onContentChange}
                                onEnter={onEnter}
                                onBackspace={onBackspace}
                                onMerge={onMerge}
                                onTypeChange={onTypeChange}
                                onIndent={onIndent}
                                onOutdent={onOutdent}
                                onFocus={onFocus}
                                onArrowUp={onArrowUp}
                                onArrowDown={onArrowDown}
                                onUpdateBlock={onUpdateBlock}
                                onCreateBlock={onCreateBlock}
                                onDragStart={onDragStart}
                                onDragOver={onDragOver}
                                onDragLeave={onDragLeave}
                                onDrop={onDrop}
                                calculateListNumber={calculateListNumber}
                                indentLevel={indentLevel + 1}
                            />
                        ))}
                </div>
            )}
        </div>
    );
}
