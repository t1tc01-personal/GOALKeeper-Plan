import { blockApi, type BatchSyncRequest, type BatchSyncResponse } from './blockApi';

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
    parent_block_id?: string | null;
    blockConfig?: Record<string, any>;
    tempId?: string;
  };
  timestamp: number;
  retryCount?: number;
  lastRetryAt?: number; // Timestamp of last retry attempt
  priority?: 'low' | 'normal' | 'high' | 'critical'; // Operation priority
  status?: 'pending' | 'syncing'; // Sync status
}

export type { BatchSyncRequest, BatchSyncResponse };

type SyncCallback = (response: BatchSyncResponse) => void;
type ErrorCallback = (error: Error) => void;

export class BlockSyncManager {
  private queue: Map<string, QueuedOperation> = new Map();
  private syncInterval: number;
  private maxBatchSize: number;
  private maxRetries: number;
  private baseRetryDelay: number;
  private maxRetryDelay: number;
  private onSyncSuccess?: SyncCallback;
  private onSyncError?: ErrorCallback;

  private syncTimer: ReturnType<typeof setTimeout> | null = null;
  private isSyncing: boolean = false;

  constructor(options: {
    syncInterval?: number;
    maxBatchSize?: number;
    maxRetries?: number;
    baseRetryDelay?: number;
    maxRetryDelay?: number;
    onSyncSuccess?: SyncCallback;
    onSyncError?: ErrorCallback;
  } = {}) {
    console.log('[BlockSyncManager] v2.0 INSTANCE CREATED (Fresh Code)');
    // ...
    this.syncInterval = options.syncInterval || 2000;
    this.maxBatchSize = options.maxBatchSize || 50;
    this.maxRetries = options.maxRetries || 5;
    this.baseRetryDelay = options.baseRetryDelay || 1000;
    this.maxRetryDelay = options.maxRetryDelay || 30000;
    this.onSyncSuccess = options.onSyncSuccess;
    this.onSyncError = options.onSyncError;
    console.log('[BlockSyncQueue] New Instance Created');
  }

  /**
   * Add or update an operation in the queue
   */
  enqueue(operation: QueuedOperation): void {
    try {
      console.log('[BlockSyncManager] Enqueue called for:', operation.type, operation.blockId);

      let shouldEnqueue = true;

      // Check against existing operations
      for (const [key, op] of this.queue.entries()) {
        if (op.blockId === operation.blockId) {

          if (operation.type === 'delete') {
            // CASE: We are deleting a block that has pending operations

            if (op.status === 'syncing') {
              // If operating is syncing, we cannot remove it.
              // We MUST enqueue the delete execution to happen after sync finishes.
              // Note: If op is 'create' and syncing, our delete op (temp-id) will be skipped by syncBatch,
              // but remapBlockId will update it to (real-id) later, which will then sync. Correct.
              console.log('[BlockSyncQueue] Deleting syncing block - enqueuing delete for later:', key);
              continue; // Keep scanning (rare, but maybe duplicates)
            }

            // If operation is pending, we can remove it.
            if (op.type === 'create') {
              // CASE: Deleting a pending CREATE.
              // Cancellation: We don't need to create, and we don't need to delete.
              // Just remove the create op and Do Not enqueue the delete op.
              console.log('[BlockSyncQueue] Cancelling pending create due to delete:', key);
              this.queue.delete(key);
              shouldEnqueue = false; // Cancel the delete op too
              break; // Optimization: one create per block usually
            } else {
              // CASE: Deleting a pending UPDATE.
              // Remove the update, but we STILL need to enqueue the DELETE.
              console.log('[BlockSyncQueue] Replacing pending update with delete:', key);
              this.queue.delete(key);
            }

          } else if (op.type === 'create') {
            // CASE: Creating a block that already has ops? (Shouldn't happen for same ID)
            if (op.status === 'syncing') {
              console.log('[BlockSyncQueue] Skipping merge into syncing create:', key);
              continue;
            }
            // Can't update a block that's being created - keep create, update data
            // (Handling logical merge of data)
            const updatedOp: QueuedOperation = {
              ...op,
              data: { ...op.data, ...operation.data },
              timestamp: Date.now(), // Reset debounce
            };
            this.queue.set(key, updatedOp);
            return; // Done

          } else if (op.type === 'update' && operation.type === 'update') {
            // CASE: coalescing updates
            if (op.status === 'syncing') {
              continue;
            }
            // Coalesce updates
            const updatedOp: QueuedOperation = {
              ...op,
              data: { ...op.data, ...operation.data },
              timestamp: Date.now(), // Reset debounce
            };
            this.queue.set(key, updatedOp);
            return; // Done
          }
        }
      }

      if (shouldEnqueue) {
        // Add new operation
        this.queue.set(operation.id, {
          ...operation,
          timestamp: Date.now(),
        });
        // Schedule sync if not already scheduled
        this.scheduleSync();
      }

    } catch (error) {
      console.error('[BlockSyncManager] Enqueue failed:', error);
    }
  }

  /**
   * Remove an operation from the queue (e.g., when manually synced)
   */
  dequeue(operationId: string): void {
    this.queue.delete(operationId);
  }

  /**
   * Remap block ID for pending operations (e.g. when temp ID becomes real ID)
   */
  remapBlockId(oldId: string, newId: string): void {
    console.log('[BlockSyncQueue] Remapping ID:', oldId, '->', newId);
    for (const [opId, op] of this.queue.entries()) {
      if (op.blockId === oldId) {
        // Create updated operation with new block ID
        const updatedOp: QueuedOperation = {
          ...op,
          blockId: newId,
        };

        // Update data payload if it contains the ID
        if (updatedOp.data) {
          if (updatedOp.type === 'create' && updatedOp.data.tempId === oldId) {
            updatedOp.data.tempId = newId;
          } else if (updatedOp.type === 'update' && updatedOp.id === oldId) {
            // This shouldn't happen for update ops (id is usually "update-..."), 
            // but just in case the op ID matches
          }
          // Note: for updates, the ID in the payload is usually implicit or separate
        }

        this.queue.set(opId, updatedOp);
      }
    }
    // Schedule sync to process remapped operations
    this.scheduleSync();
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
    if (this.syncTimer) {
      return;
    }

    console.log('[BlockSyncQueue] Scheduling sync in', this.syncInterval, 'ms');
    this.syncTimer = setTimeout(() => {
      this.syncTimer = null;
      this.sync();
    }, this.syncInterval);
  }

  /**
   * Force immediate sync (for critical operations)
   */
  async forceSync(): Promise<BatchSyncResponse | null> {
    if (this.queue.size === 0) {
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
    // Note: We allow parallel syncs now (status tracking handles concurrency)
    // But we still block if queue is empty to avoid useless calls
    if (this.queue.size === 0) {
      return null;
    }

    this.isSyncing = true;
    console.log('[BlockSyncQueue] Sync started. Queue size:', this.queue.size);

    // Capture operations for this sync execution scope
    let operations: QueuedOperation[] = [];

    try {
      // Get operations to sync (limit by maxBatchSize)
      // Filter for PENDING operations only
      const now = Date.now();
      operations = Array.from(this.queue.values())
        .filter((op) => {
          // Include operations that are ready (timestamp <= now) AND pending
          const shouldSync = op.timestamp <= now && (!op.status || op.status === 'pending');
          if (!shouldSync) return false;

          if ((op.type === 'update' || op.type === 'delete') && op.blockId.startsWith('temp-')) {
            console.log('[BlockSyncQueue] Skipping temp-ID op:', op.id);
            return false;
          }
          return true;
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
        // Check if there are ANY syncing ops globally to update isSyncing flag correctly
        const hasSyncingOps = Array.from(this.queue.values()).some(op => op.status === 'syncing');
        this.isSyncing = hasSyncingOps;
        console.log('[BlockSyncQueue] No operations to sync. isSyncing:', this.isSyncing);

        // Schedule next sync if pending operations exist
        const pendingOps = Array.from(this.queue.values()).filter(op => !op.status || op.status === 'pending');
        if (pendingOps.length > 0) {
          console.log('[BlockSyncQueue] Pending ops exist (retry/wait). Scheduling next sync.');
          this.scheduleSync();
        }

        return null;
      }

      // Mark operations as syncing to prevent other syncs from picking them up
      for (const op of operations) {
        op.status = 'syncing';
        this.queue.set(op.id, op);
      }

      // Build batch request
      const batchRequest: BatchSyncRequest = {
        creates: [],
        updates: [],
        deletes: [],
      };

      for (const op of operations) {
        switch (op.type) {
          case 'create':
            if (op.data?.pageId && op.data?.type) {
              batchRequest.creates!.push({
                pageId: op.data.pageId,
                type: op.data.type,
                content: op.data.content,
                position: op.data.position ?? 0,
                parent_block_id: op.data.parent_block_id,
                blockConfig: op.data.blockConfig,
                tempId: op.blockId, // Use blockId as tempId
              });
            } else {
              console.error('[BlockSyncQueue] Dropping INVALID create op (missing pageId/type):', op);
              this.queue.delete(op.id); // Remove invalid op to prevent retry loop
            }
            break;

          case 'update':
            // Validation for update?
            batchRequest.updates!.push({
              id: op.blockId,
              content: op.data?.content,
              type: op.data?.type,
              position: op.data?.position,
              parent_block_id: op.data?.parent_block_id,
              blockConfig: op.data?.blockConfig,
            });
            break;

          case 'delete':
            batchRequest.deletes!.push(op.blockId);
            break;
        }
      }

      // Only sync if there are operations (double check though filtering implies there are)
      if (
        (batchRequest.creates?.length ?? 0) === 0 &&
        (batchRequest.updates?.length ?? 0) === 0 &&
        (batchRequest.deletes?.length ?? 0) === 0
      ) {
        // Reset status if no valid ops found
        for (const op of operations) {
          op.status = 'pending';
          this.queue.set(op.id, op);
        }

        const hasSyncingOps = Array.from(this.queue.values()).some(op => op.status === 'syncing');
        this.isSyncing = hasSyncingOps;
        return null;
      }

      // Call batch sync API
      const response = await this.syncBatch(batchRequest);

      // Remove successfully synced operations
      const successfulOps = new Set<string>();

      // Track successful creates
      response.creates.forEach((create) => {
        // Find operation by tempId (which is blockId)
        for (const op of operations) {
          if (op.type === 'create' && op.blockId === create.tempId) {
            successfulOps.add(op.id);
            break;
          }
        }
      });

      // Track successful updates
      response.updates.forEach((update) => {
        for (const op of operations) {
          if (op.type === 'update' && op.blockId === update.id) {
            successfulOps.add(op.id);
            break;
          }
        }
      });

      // Track successful deletes
      response.deletes.forEach((blockId) => {
        for (const op of operations) {
          if (op.type === 'delete' && op.blockId === blockId) {
            successfulOps.add(op.id);
            break;
          }
        }
      });

      // Remove successful operations
      successfulOps.forEach((id) => this.queue.delete(id));

      // Handle errors - retry with exponential backoff
      if (response.errors && response.errors.length > 0) {
        for (const error of response.errors) {
          let op = this.queue.get(error.operationId);
          // If not found by ID directly, search in current batch operations
          if (!op) {
            op = operations.find(o => o.blockId === error.operationId || o.id === error.operationId);
          }

          if (op) {
            const retryCount = (op.retryCount || 0) + 1;
            if (retryCount < this.maxRetries) {
              // Calculate exponential backoff delay
              const delay = Math.min(
                this.baseRetryDelay * Math.pow(2, retryCount - 1),
                this.maxRetryDelay
              );

              // Schedule retry with exponential backoff AND RESET STATUS
              this.queue.set(op.id, {
                ...op,
                retryCount,
                lastRetryAt: Date.now(),
                timestamp: Date.now() + delay,
                status: 'pending', // Reset to pending
              });
            } else {
              // Remove after max retries
              this.queue.delete(op.id);
              console.error('Operation failed after max retries:', error);
            }
          }
        }
      }

      // Call success callback
      if (this.onSyncSuccess) {
        this.onSyncSuccess(response);
      }

      // Update sync state before scheduling next
      const hasSyncingOps = Array.from(this.queue.values()).some(op => op.status === 'syncing');
      this.isSyncing = hasSyncingOps;

      // Schedule next sync if pending operations exist
      const pendingOps = Array.from(this.queue.values()).filter(op => !op.status || op.status === 'pending');
      if (pendingOps.length > 0) {
        this.scheduleSync();
      }

      return response;
    } catch (error) {
      // Handle generic transport errors (e.g. network offline)
      const err = error instanceof Error ? error : new Error('Unknown sync error');

      if (this.onSyncError) {
        this.onSyncError(err);
      } else {
        console.error('Block sync failed:', err);
      }

      // Retry ALL operations in this batch with exponential backoff
      const now = Date.now();
      for (const op of operations) {
        if (this.queue.has(op.id)) {
          const retryCount = (op.retryCount || 0) + 1;
          if (retryCount < this.maxRetries) {
            const delay = Math.min(
              this.baseRetryDelay * Math.pow(2, retryCount - 1),
              this.maxRetryDelay
            );

            this.queue.set(op.id, {
              ...op,
              retryCount,
              lastRetryAt: now,
              timestamp: now + delay,
              status: 'pending', // IMPORTANT: Reset to pending
            });
          } else {
            this.queue.delete(op.id);
          }
        }
      }

      // Update sync state before scheduling retry
      const hasSyncingOps = Array.from(this.queue.values()).some(op => op.status === 'syncing');
      this.isSyncing = hasSyncingOps;

      // Schedule retry
      this.scheduleSync();
      return null;
    }
  }

  /**
   * Call batch sync API
   */
  private async syncBatch(request: BatchSyncRequest): Promise<BatchSyncResponse> {
    console.log('[BlockSyncQueue] Calling API batchSync with:', request);
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
