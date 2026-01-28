/**
 * BlockSyncQueue - Manages batched block operations for optimal performance
 * 
 * This queue batches create, update, and delete operations to reduce API calls
 * and improve UX during rapid editing (similar to Notion/Obsidian).
 */

export type BlockOperationType = 'create' | 'update' | 'delete';

export interface QueuedOperation {
  id: string; // Unique operation ID (can be temp ID for creates)
  type: BlockOperationType;
  blockId: string; // Real block ID (or temp ID for creates)
  data?: {
    pageId?: string;
    type?: string;
    content?: string;
    position?: number;
    blockConfig?: Record<string, any>;
  };
  timestamp: number;
  retryCount?: number;
  lastRetryAt?: number; // Timestamp of last retry attempt
  priority?: 'low' | 'normal' | 'high' | 'critical'; // Operation priority
}

export interface BatchSyncRequest {
  creates?: Array<{
    pageId: string;
    type: string;
    content?: string;
    position: number;
    blockConfig?: Record<string, any>;
    tempId: string; // Frontend temp ID to map back
  }>;
  updates?: Array<{
    id: string;
    content?: string;
    type?: string;
    position?: number;
    blockConfig?: Record<string, any>;
  }>;
  deletes?: string[]; // Block IDs to delete
}

export interface BatchSyncResponse {
  creates: Array<{
    tempId: string;
    block: {
      id: string;
      pageId: string;
      type: string;
      content: string;
      position: number;
      blockConfig?: Record<string, any>;
      created_at: string;
      updated_at: string;
    };
  }>;
  updates: Array<{
    id: string;
    block: {
      id: string;
      pageId: string;
      type: string;
      content: string;
      position: number;
      blockConfig?: Record<string, any>;
      updated_at: string;
    };
  }>;
  deletes: string[]; // Successfully deleted block IDs
  errors?: Array<{
    operationId: string;
    type: BlockOperationType;
    error: string;
  }>;
}

type SyncCallback = (response: BatchSyncResponse) => void;
type ErrorCallback = (error: Error) => void;

export class BlockSyncQueue {
  private queue: Map<string, QueuedOperation> = new Map();
  private syncTimer: ReturnType<typeof setTimeout> | null = null;
  private isSyncing: boolean = false;
  private syncInterval: number;
  private maxBatchSize: number;
  private onSyncSuccess?: SyncCallback;
  private onSyncError?: ErrorCallback;
  private maxRetries: number;
  private baseRetryDelay: number; // Base delay for exponential backoff (ms)
  private maxRetryDelay: number; // Maximum retry delay (ms)

  constructor(options: {
    syncInterval?: number; // ms between syncs (default: 2000)
    maxBatchSize?: number; // max operations per batch (default: 50)
    maxRetries?: number; // max retry attempts (default: 5)
    baseRetryDelay?: number; // base delay for exponential backoff (default: 1000)
    maxRetryDelay?: number; // max retry delay (default: 30000)
    onSyncSuccess?: SyncCallback;
    onSyncError?: ErrorCallback;
  } = {}) {
    this.syncInterval = options.syncInterval || 2000;
    this.maxBatchSize = options.maxBatchSize || 50;
    this.maxRetries = options.maxRetries || 5;
    this.baseRetryDelay = options.baseRetryDelay || 1000;
    this.maxRetryDelay = options.maxRetryDelay || 30000;
    this.onSyncSuccess = options.onSyncSuccess;
    this.onSyncError = options.onSyncError;
  }

  /**
   * Add or update an operation in the queue
   */
  enqueue(operation: QueuedOperation): void {
    // For updates, replace existing operation with same blockId
    if (operation.type === 'update' || operation.type === 'delete') {
      // Remove any pending operations for this block
      for (const [key, op] of this.queue.entries()) {
        if (op.blockId === operation.blockId) {
          // If we're deleting, remove all operations for this block
          if (operation.type === 'delete') {
            this.queue.delete(key);
          } else if (op.type === 'create') {
            // Can't update a block that's being created - keep create, update data
            const updatedOp: QueuedOperation = {
              ...op,
              data: { ...op.data, ...operation.data },
            };
            this.queue.set(key, updatedOp);
            return;
          }
        }
      }
    }

    // Add new operation
    this.queue.set(operation.id, {
      ...operation,
      timestamp: Date.now(),
    });

    // Schedule sync if not already scheduled
    this.scheduleSync();
  }

  /**
   * Remove an operation from the queue (e.g., when manually synced)
   */
  dequeue(operationId: string): void {
    this.queue.delete(operationId);
  }

  /**
   * Get all queued operations
   */
  getQueue(): QueuedOperation[] {
    return Array.from(this.queue.values());
  }

  /**
   * Get queue size
   */
  getSize(): number {
    return this.queue.size;
  }

  /**
   * Check if queue is empty
   */
  isEmpty(): boolean {
    return this.queue.size === 0;
  }

  /**
   * Clear all operations from queue
   */
  clear(): void {
    this.queue.clear();
    if (this.syncTimer) {
      clearTimeout(this.syncTimer);
      this.syncTimer = null;
    }
  }

  /**
   * Schedule next sync
   */
  private scheduleSync(): void {
    if (this.syncTimer || this.isSyncing) {
      return;
    }

    this.syncTimer = setTimeout(() => {
      this.syncTimer = null;
      this.sync();
    }, this.syncInterval);
  }

  /**
   * Force immediate sync (for critical operations)
   */
  async forceSync(): Promise<BatchSyncResponse | null> {
    if (this.isSyncing || this.queue.size === 0) {
      return null;
    }

    if (this.syncTimer) {
      clearTimeout(this.syncTimer);
      this.syncTimer = null;
    }

    return this.sync();
  }

  /**
   * Sync operations to server
   */
  private async sync(): Promise<BatchSyncResponse | null> {
    if (this.isSyncing || this.queue.size === 0) {
      return null;
    }

    this.isSyncing = true;

    try {
      // Get operations to sync (limit by maxBatchSize)
      // Sort by priority (critical > high > normal > low) then by timestamp
      const now = Date.now();
      const operations = Array.from(this.queue.values())
        .filter((op) => {
          // Only include operations that are ready (timestamp <= now)
          return op.timestamp <= now;
        })
        .sort((a, b) => {
          // Priority order: critical > high > normal > low
          const priorityOrder: Record<string, number> = {
            critical: 0,
            high: 1,
            normal: 2,
            low: 3,
          };
          const aPriority = priorityOrder[a.priority || 'normal'];
          const bPriority = priorityOrder[b.priority || 'normal'];
          
          if (aPriority !== bPriority) {
            return aPriority - bPriority;
          }
          // If same priority, sort by timestamp
          return a.timestamp - b.timestamp;
        })
        .slice(0, this.maxBatchSize);

      if (operations.length === 0) {
        this.isSyncing = false;
        return null;
      }

      // Build batch request
      const batchRequest: BatchSyncRequest = {
        creates: [],
        updates: [],
        deletes: [],
      };

      const operationIds: string[] = [];

      for (const op of operations) {
        operationIds.push(op.id);

        switch (op.type) {
          case 'create':
            if (op.data?.pageId && op.data?.type) {
              batchRequest.creates!.push({
                pageId: op.data.pageId,
                type: op.data.type,
                content: op.data.content,
                position: op.data.position ?? 0,
                blockConfig: op.data.blockConfig,
                tempId: op.blockId, // Use blockId as tempId
              });
            }
            break;

          case 'update':
            batchRequest.updates!.push({
              id: op.blockId,
              content: op.data?.content,
              type: op.data?.type,
              position: op.data?.position,
              blockConfig: op.data?.blockConfig,
            });
            break;

          case 'delete':
            batchRequest.deletes!.push(op.blockId);
            break;
        }
      }

      // Only sync if there are operations
      if (
        (batchRequest.creates?.length ?? 0) === 0 &&
        (batchRequest.updates?.length ?? 0) === 0 &&
        (batchRequest.deletes?.length ?? 0) === 0
      ) {
        this.isSyncing = false;
        return null;
      }

      // Call batch sync API (will be implemented in blockApi.ts)
      const response = await this.syncBatch(batchRequest);

      // Remove successfully synced operations
      const successfulOps = new Set<string>();
      
      // Track successful creates
      response.creates.forEach((create) => {
        // Find operation by tempId
        for (const [key, op] of this.queue.entries()) {
          if (op.type === 'create' && op.blockId === create.tempId) {
            successfulOps.add(key);
            break;
          }
        }
      });

      // Track successful updates
      response.updates.forEach((update) => {
        for (const [key, op] of this.queue.entries()) {
          if (op.type === 'update' && op.blockId === update.id) {
            successfulOps.add(key);
            break;
          }
        }
      });

      // Track successful deletes
      response.deletes.forEach((blockId) => {
        for (const [key, op] of this.queue.entries()) {
          if (op.type === 'delete' && op.blockId === blockId) {
            successfulOps.add(key);
            break;
          }
        }
      });

      // Remove successful operations
      successfulOps.forEach((id) => this.queue.delete(id));

      // Handle errors - retry with exponential backoff
      if (response.errors && response.errors.length > 0) {
        for (const error of response.errors) {
          const op = this.queue.get(error.operationId);
          if (op) {
            const retryCount = (op.retryCount || 0) + 1;
            if (retryCount < this.maxRetries) {
              // Calculate exponential backoff delay
              const delay = Math.min(
                this.baseRetryDelay * Math.pow(2, retryCount - 1),
                this.maxRetryDelay
              );
              
              // Schedule retry with exponential backoff
              this.queue.set(error.operationId, {
                ...op,
                retryCount,
                lastRetryAt: Date.now(),
                timestamp: Date.now() + delay, // Schedule retry in the future
              });
            } else {
              // Remove after max retries
              this.queue.delete(error.operationId);
              console.error('Operation failed after max retries:', error);
            }
          }
        }
      }

      // Call success callback
      if (this.onSyncSuccess) {
        this.onSyncSuccess(response);
      }

      // Schedule next sync if queue is not empty
      if (this.queue.size > 0) {
        this.scheduleSync();
      }

      this.isSyncing = false;
      return response;
    } catch (error) {
      this.isSyncing = false;
      const err = error instanceof Error ? error : new Error('Unknown sync error');
      
      if (this.onSyncError) {
        this.onSyncError(err);
      } else {
        console.error('Block sync failed:', err);
      }

      // Retry failed operations with exponential backoff
      const now = Date.now();
      for (const op of Array.from(this.queue.values())) {
        const retryCount = (op.retryCount || 0) + 1;
        if (retryCount < this.maxRetries) {
          // Calculate exponential backoff delay
          const delay = Math.min(
            this.baseRetryDelay * Math.pow(2, retryCount - 1),
            this.maxRetryDelay
          );
          
          this.queue.set(op.id, {
            ...op,
            retryCount,
            lastRetryAt: now,
            timestamp: now + delay, // Schedule retry in the future
          });
        } else {
          // Remove after max retries
          this.queue.delete(op.id);
        }
      }

      // Schedule retry (will check timestamps)
      this.scheduleSync();
      return null;
    }
  }

  /**
   * Call batch sync API
   */
  private async syncBatch(request: BatchSyncRequest): Promise<BatchSyncResponse> {
    // Import blockApi dynamically to avoid circular dependencies
    const { blockApi } = await import('./blockApi');
    return blockApi.batchSync(request);
  }

  /**
   * Set sync callback
   */
  setSyncCallbacks(callbacks: {
    onSyncSuccess?: SyncCallback;
    onSyncError?: ErrorCallback;
  }): void {
    if (callbacks.onSyncSuccess) {
      this.onSyncSuccess = callbacks.onSyncSuccess;
    }
    if (callbacks.onSyncError) {
      this.onSyncError = callbacks.onSyncError;
    }
  }

  /**
   * Destroy queue and cleanup
   */
  destroy(): void {
    this.clear();
    this.onSyncSuccess = undefined;
    this.onSyncError = undefined;
  }
}

