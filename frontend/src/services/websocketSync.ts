/**
 * WebSocketSync - Real-time synchronization via WebSocket
 * 
 * Handles bidirectional real-time updates for collaborative editing
 */

import type { Block } from './blockApi';
import type { BatchSyncResponse } from './blockSyncQueue';

export interface WebSocketMessage {
  type: 'block_created' | 'block_updated' | 'block_deleted' | 'batch_sync' | 'ping' | 'pong' | 'error';
  pageId?: string;
  blockId?: string;
  block?: Block;
  batchResponse?: BatchSyncResponse;
  timestamp: number;
  userId?: string;
}

export type WebSocketEventHandler = (message: WebSocketMessage) => void;

export class WebSocketSync {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 10;
  private reconnectDelay: number = 1000;
  private maxReconnectDelay: number = 30000;
  private pingInterval: ReturnType<typeof setInterval> | null = null;
  private eventHandlers: Map<string, Set<WebSocketEventHandler>> = new Map();
  private isConnected: boolean = false;
  private pageId: string | null = null;

  constructor(wsUrl: string) {
    this.url = wsUrl;
  }

  /**
   * Connect to WebSocket server
   */
  async connect(pageId: string, token?: string): Promise<void> {
    this.pageId = pageId;
    
    // Build WebSocket URL with auth
    const url = new URL(this.url);
    url.searchParams.set('pageId', pageId);
    if (token) {
      url.searchParams.set('token', token);
    }

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(url.toString());

        this.ws.onopen = () => {
          this.isConnected = true;
          this.reconnectAttempts = 0;
          console.log('WebSocket connected');
          this.startPing();
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };

        this.ws.onclose = () => {
          this.isConnected = false;
          this.stopPing();
          console.log('WebSocket disconnected');
          
          // Attempt to reconnect
          if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    this.stopPing();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
  }

  /**
   * Send message to server
   */
  send(message: Omit<WebSocketMessage, 'timestamp'>): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not connected, message not sent');
      return;
    }

    const fullMessage: WebSocketMessage = {
      ...message,
      timestamp: Date.now(),
    };

    this.ws.send(JSON.stringify(fullMessage));
  }

  /**
   * Subscribe to message events
   */
  on(eventType: string, handler: WebSocketEventHandler): () => void {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, new Set());
    }
    this.eventHandlers.get(eventType)!.add(handler);

    // Return unsubscribe function
    return () => {
      const handlers = this.eventHandlers.get(eventType);
      if (handlers) {
        handlers.delete(handler);
      }
    };
  }

  /**
   * Handle incoming messages
   */
  private handleMessage(message: WebSocketMessage): void {
    // Handle ping/pong
    if (message.type === 'ping') {
      this.send({ type: 'pong' });
      return;
    }

    // Notify event handlers
    const handlers = this.eventHandlers.get(message.type);
    if (handlers) {
      handlers.forEach((handler) => handler(message));
    }

    // Also notify 'all' handlers
    const allHandlers = this.eventHandlers.get('all');
    if (allHandlers) {
      allHandlers.forEach((handler) => handler(message));
    }
  }

  /**
   * Start ping interval to keep connection alive
   */
  private startPing(): void {
    this.pingInterval = setInterval(() => {
      if (this.isConnected && this.ws?.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping' });
      }
    }, 30000); // Ping every 30 seconds
  }

  /**
   * Stop ping interval
   */
  private stopPing(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    this.reconnectAttempts++;
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      this.maxReconnectDelay
    );

    setTimeout(() => {
      if (this.pageId) {
        this.connect(this.pageId).catch((error) => {
          console.error('Reconnection failed:', error);
        });
      }
    }, delay);
  }

  /**
   * Check if WebSocket is connected
   */
  get connected(): boolean {
    return this.isConnected && this.ws?.readyState === WebSocket.OPEN;
  }
}

// Helper to get WebSocket URL from API base URL
export function getWebSocketUrl(apiBaseUrl: string): string {
  // Convert http:// to ws:// and https:// to wss://
  const wsUrl = apiBaseUrl
    .replace(/^http:/, 'ws:')
    .replace(/^https:/, 'wss:');
  
  // Remove /api/v1 if present and add /ws
  return wsUrl.replace(/\/api\/v1$/, '') + '/ws';
}

