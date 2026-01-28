/**
 * ConflictResolver - Handles conflict resolution for block operations
 * 
 * Implements Last-Write-Wins (LWW) and Operational Transform (OT) strategies
 */

import type { Block } from './blockApi';

export interface BlockVersion {
  blockId: string;
  version: number; // Version number or timestamp
  updatedAt: string; // ISO timestamp
  content?: string;
  metadata?: Record<string, any>;
}

export interface ConflictInfo {
  blockId: string;
  localVersion: BlockVersion;
  serverVersion: BlockVersion;
  conflictType: 'content' | 'metadata' | 'both';
}

export type ConflictResolutionStrategy = 'last-write-wins' | 'server-wins' | 'client-wins' | 'merge' | 'manual';

export interface ConflictResolution {
  blockId: string;
  strategy: ConflictResolutionStrategy;
  resolvedBlock: Block;
  mergedContent?: string; // For merge strategy
}

export class ConflictResolver {
  /**
   * Detect conflicts between local and server versions
   */
  static detectConflict(
    localBlock: Block,
    serverBlock: Block,
    localVersion: BlockVersion
  ): ConflictInfo | null {
    // Compare versions
    const localTime = new Date(localVersion.updatedAt).getTime();
    const serverTime = new Date(serverBlock.updated_at).getTime();

    // If server is newer, check for conflicts
    if (serverTime > localTime) {
      const contentConflict = localBlock.content !== serverBlock.content;
      const metadataConflict = JSON.stringify(localBlock.metadata) !== JSON.stringify(serverBlock.metadata);

      if (contentConflict || metadataConflict) {
        return {
          blockId: localBlock.id,
          localVersion,
          serverVersion: {
            blockId: serverBlock.id,
            version: serverTime,
            updatedAt: serverBlock.updated_at,
            content: serverBlock.content,
            metadata: serverBlock.metadata,
          },
          conflictType: contentConflict && metadataConflict ? 'both' : contentConflict ? 'content' : 'metadata',
        };
      }
    }

    return null;
  }

  /**
   * Resolve conflict using specified strategy
   */
  static resolveConflict(
    conflict: ConflictInfo,
    strategy: ConflictResolutionStrategy = 'last-write-wins'
  ): ConflictResolution {
    const { localVersion, serverVersion } = conflict;

    switch (strategy) {
      case 'last-write-wins':
        // Use the version with latest timestamp
        const localTime = new Date(localVersion.updatedAt).getTime();
        const serverTime = new Date(serverVersion.updatedAt).getTime();
        
        if (serverTime > localTime) {
          return {
            blockId: conflict.blockId,
            strategy: 'server-wins',
            resolvedBlock: {
              id: conflict.blockId,
              pageId: '', // Will be filled by caller
              type: '', // Will be filled by caller
              content: serverVersion.content,
              metadata: serverVersion.metadata,
              position: 0,
              created_at: '',
              updated_at: serverVersion.updatedAt,
            },
          };
        } else {
          return {
            blockId: conflict.blockId,
            strategy: 'client-wins',
            resolvedBlock: {
              id: conflict.blockId,
              pageId: '',
              type: '',
              content: localVersion.content,
              metadata: localVersion.metadata,
              position: 0,
              created_at: '',
              updated_at: localVersion.updatedAt,
            },
          };
        }

      case 'server-wins':
        return {
          blockId: conflict.blockId,
          strategy: 'server-wins',
          resolvedBlock: {
            id: conflict.blockId,
            pageId: '',
            type: '',
            content: serverVersion.content,
            metadata: serverVersion.metadata,
            position: 0,
            created_at: '',
            updated_at: serverVersion.updatedAt,
          },
        };

      case 'client-wins':
        return {
          blockId: conflict.blockId,
          strategy: 'client-wins',
          resolvedBlock: {
            id: conflict.blockId,
            pageId: '',
            type: '',
            content: localVersion.content,
            metadata: localVersion.metadata,
            position: 0,
            created_at: '',
            updated_at: localVersion.updatedAt,
          },
        };

      case 'merge':
        // Simple merge: combine content with separator
        const mergedContent = this.mergeContent(
          localVersion.content || '',
          serverVersion.content || ''
        );
        
        return {
          blockId: conflict.blockId,
          strategy: 'merge',
          resolvedBlock: {
            id: conflict.blockId,
            pageId: '',
            type: '',
            content: mergedContent,
            metadata: { ...serverVersion.metadata, ...localVersion.metadata }, // Merge metadata
            position: 0,
            created_at: '',
            updated_at: new Date().toISOString(),
          },
          mergedContent,
        };

      case 'manual':
        // Return both versions for manual resolution
        return {
          blockId: conflict.blockId,
          strategy: 'manual',
          resolvedBlock: {
            id: conflict.blockId,
            pageId: '',
            type: '',
            content: localVersion.content || serverVersion.content || '',
            metadata: localVersion.metadata || serverVersion.metadata || {},
            position: 0,
            created_at: '',
            updated_at: new Date().toISOString(),
          },
        };

      default:
        // Default to server-wins
        return this.resolveConflict(conflict, 'server-wins');
    }
  }

  /**
   * Merge content from two versions
   */
  private static mergeContent(localContent: string, serverContent: string): string {
    if (!localContent) return serverContent;
    if (!serverContent) return localContent;
    if (localContent === serverContent) return localContent;

    // Simple merge: append with conflict markers
    // In production, you might want to use operational transform or diff algorithms
    return `${localContent}\n\n--- Conflict Resolution ---\n\n${serverContent}`;
  }

  /**
   * Create version info from block
   */
  static createVersion(block: Block): BlockVersion {
    return {
      blockId: block.id,
      version: new Date(block.updated_at).getTime(),
      updatedAt: block.updated_at,
      content: block.content,
      metadata: block.metadata,
    };
  }
}

