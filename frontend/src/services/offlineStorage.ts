/**
 * OfflineStorage - Manages local storage for offline support
 * 
 * Uses IndexedDB for persistent storage and localStorage for quick access
 */

import type { Block } from './blockApi';
import type { QueuedOperation } from './blockSyncQueue';

const DB_NAME = 'goalkeeper-plan';
const DB_VERSION = 1;
const STORE_BLOCKS = 'blocks';
const STORE_QUEUE = 'queue';
const STORE_VERSIONS = 'versions';

export interface StoredBlock extends Block {
  _localVersion?: number; // Local version timestamp
  _isDirty?: boolean; // Has unsaved changes
  _lastSynced?: string; // Last sync timestamp
}

export class OfflineStorage {
  private db: IDBDatabase | null = null;
  private initPromise: Promise<void> | null = null;

  constructor() {
    this.init();
  }

  /**
   * Initialize IndexedDB
   */
  private async init(): Promise<void> {
    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);

      request.onerror = () => {
        console.error('Failed to open IndexedDB');
        reject(new Error('Failed to open IndexedDB'));
      };

      request.onsuccess = () => {
        this.db = request.result;
        resolve();
      };

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;

        // Create blocks store
        if (!db.objectStoreNames.contains(STORE_BLOCKS)) {
          const blocksStore = db.createObjectStore(STORE_BLOCKS, { keyPath: 'id' });
          blocksStore.createIndex('pageId', 'pageId', { unique: false });
          blocksStore.createIndex('updated_at', 'updated_at', { unique: false });
        }

        // Create queue store
        if (!db.objectStoreNames.contains(STORE_QUEUE)) {
          const queueStore = db.createObjectStore(STORE_QUEUE, { keyPath: 'id' });
          queueStore.createIndex('timestamp', 'timestamp', { unique: false });
        }

        // Create versions store
        if (!db.objectStoreNames.contains(STORE_VERSIONS)) {
          const versionsStore = db.createObjectStore(STORE_VERSIONS, { keyPath: 'blockId' });
          versionsStore.createIndex('updatedAt', 'updatedAt', { unique: false });
        }
      };
    });

    return this.initPromise;
  }

  /**
   * Wait for DB to be ready
   */
  private async ensureDB(): Promise<IDBDatabase> {
    await this.init();
    if (!this.db) {
      throw new Error('Database not initialized');
    }
    return this.db;
  }

  /**
   * Save blocks to local storage
   */
  async saveBlocks(pageId: string, blocks: StoredBlock[]): Promise<void> {
    const db = await this.ensureDB();
    const transaction = db.transaction([STORE_BLOCKS], 'readwrite');
    const store = transaction.objectStore(STORE_BLOCKS);

    for (const block of blocks) {
      const storedBlock: StoredBlock = {
        ...block,
        _lastSynced: new Date().toISOString(),
        _isDirty: false,
      };
      await new Promise<void>((resolve, reject) => {
        const request = store.put(storedBlock);
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
      });
    }
  }

  /**
   * Load blocks from local storage
   */
  async loadBlocks(pageId: string): Promise<StoredBlock[]> {
    const db = await this.ensureDB();
    const transaction = db.transaction([STORE_BLOCKS], 'readonly');
    const store = transaction.objectStore(STORE_BLOCKS);
    const index = store.index('pageId');

    return new Promise((resolve, reject) => {
      const request = index.getAll(pageId);
      request.onsuccess = () => {
        resolve(request.result || []);
      };
      request.onerror = () => reject(request.error);
    });
  }

  /**
   * Save queued operations
   */
  async saveQueue(operations: QueuedOperation[]): Promise<void> {
    const db = await this.ensureDB();
    const transaction = db.transaction([STORE_QUEUE], 'readwrite');
    const store = transaction.objectStore(STORE_QUEUE);

    // Clear existing queue
    await new Promise<void>((resolve, reject) => {
      const clearRequest = store.clear();
      clearRequest.onsuccess = () => resolve();
      clearRequest.onerror = () => reject(clearRequest.error);
    });

    // Save new operations
    for (const op of operations) {
      await new Promise<void>((resolve, reject) => {
        const request = store.put(op);
        request.onsuccess = () => resolve();
        request.onerror = () => reject(request.error);
      });
    }
  }

  /**
   * Load queued operations
   */
  async loadQueue(): Promise<QueuedOperation[]> {
    const db = await this.ensureDB();
    const transaction = db.transaction([STORE_QUEUE], 'readonly');
    const store = transaction.objectStore(STORE_QUEUE);

    return new Promise((resolve, reject) => {
      const request = store.getAll();
      request.onsuccess = () => {
        resolve(request.result || []);
      };
      request.onerror = () => reject(request.error);
    });
  }

  /**
   * Mark block as dirty (has unsaved changes)
   */
  async markDirty(blockId: string): Promise<void> {
    const db = await this.ensureDB();
    const transaction = db.transaction([STORE_BLOCKS], 'readwrite');
    const store = transaction.objectStore(STORE_BLOCKS);

    return new Promise((resolve, reject) => {
      const getRequest = store.get(blockId);
      getRequest.onsuccess = () => {
        const block = getRequest.result;
        if (block) {
          block._isDirty = true;
          block._localVersion = Date.now();
          const putRequest = store.put(block);
          putRequest.onsuccess = () => resolve();
          putRequest.onerror = () => reject(putRequest.error);
        } else {
          resolve();
        }
      };
      getRequest.onerror = () => reject(getRequest.error);
    });
  }

  /**
   * Get dirty blocks (blocks with unsaved changes)
   */
  async getDirtyBlocks(pageId: string): Promise<StoredBlock[]> {
    const blocks = await this.loadBlocks(pageId);
    return blocks.filter((b) => b._isDirty);
  }

  /**
   * Clear all data for a page
   */
  async clearPage(pageId: string): Promise<void> {
    const db = await this.ensureDB();
    const transaction = db.transaction([STORE_BLOCKS], 'readwrite');
    const store = transaction.objectStore(STORE_BLOCKS);
    const index = store.index('pageId');

    return new Promise((resolve, reject) => {
      const request = index.openKeyCursor(pageId);
      request.onsuccess = () => {
        const cursor = request.result;
        if (cursor) {
          store.delete(cursor.primaryKey);
          cursor.continue();
        } else {
          resolve();
        }
      };
      request.onerror = () => reject(request.error);
    });
  }

  /**
   * Check if offline storage is available
   */
  static isAvailable(): boolean {
    return typeof indexedDB !== 'undefined';
  }
}

// Singleton instance
export const offlineStorage = new OfflineStorage();

