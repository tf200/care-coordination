# WebSocket Notification System - Frontend Integration Guide

## Overview

The notification system provides real-time push notifications via WebSocket and REST APIs for managing notifications. This document describes how to integrate the notification system in your frontend application.

---

## Table of Contents

1. [Authentication Flow](#authentication-flow)
2. [WebSocket Connection](#websocket-connection)
3. [Message Types](#message-types)
4. [REST API Endpoints](#rest-api-endpoints)
5. [TypeScript/JavaScript Implementation](#typescriptjavascript-implementation)
6. [React Hook Example](#react-hook-example)
7. [Error Handling](#error-handling)
8. [Reconnection Strategy](#reconnection-strategy)

---

## Authentication Flow

The WebSocket uses **ticket-based authentication** to avoid exposing JWTs in URLs.

```
┌─────────┐         ┌─────────┐         ┌─────────┐
│ Client  │         │   API   │         │  Redis  │
└────┬────┘         └────┬────┘         └────┬────┘
     │                   │                   │
     │ 1. POST /ws/auth  │                   │
     │ (JWT in header)   │                   │
     ├──────────────────>│                   │
     │                   │ 2. Store ticket   │
     │                   ├──────────────────>│
     │ 3. Return ticket  │                   │
     │<──────────────────┤                   │
     │                   │                   │
     │ 4. WS /ws/notifications?ticket=xxx    │
     ├──────────────────>│                   │
     │                   │ 5. Validate+Delete│
     │                   ├──────────────────>│
     │ 6. WebSocket Connected                │
     │<═══════════════════                   │
```

### Step 1: Get a Ticket

```http
POST /ws/auth
Authorization: Bearer <your-jwt-token>
Content-Type: application/json
```

**Response:**
```json
{
  "data": {
    "ticket": "Jqjj_GL6RY6DrJkFRjrj"
  },
  "message": "Ticket created"
}
```

> ⚠️ **Important:** Tickets expire in **30 seconds** and can only be used **once**.

### Step 2: Connect to WebSocket

```javascript
const ws = new WebSocket(`wss://api.example.com/ws/notifications?ticket=${ticket}`);
```

---

## WebSocket Connection

### URL Format
```
wss://<api-host>/ws/notifications?ticket=<ticket>
```

### Connection Lifecycle

| Event | Description |
|-------|-------------|
| `onopen` | Connection established |
| `onmessage` | Server sent a message |
| `onclose` | Connection closed |
| `onerror` | Connection error |

---

## Message Types

### Server → Client Messages

#### 1. `connected`
Sent immediately after successful connection.

```json
{
  "type": "connected"
}
```

#### 2. `notification`
A new notification was created.

```json
{
  "type": "notification",
  "payload": {
    "id": "abc123",
    "type": "incident_created",
    "priority": "high",
    "title": "New Incident Reported",
    "message": "A severe incident was reported at Location A",
    "resourceType": "incident",
    "resourceId": "inc_456",
    "isRead": false,
    "createdAt": "2026-01-12T12:00:00Z"
  }
}
```

#### 3. `ping`
Server heartbeat (every ~54 seconds). Respond with `pong`.

```json
{
  "type": "ping"
}
```

### Client → Server Messages

#### 1. `pong`
Response to server ping.

```json
{
  "type": "pong"
}
```

#### 2. `mark_read`
Mark a specific notification as read.

```json
{
  "type": "mark_read",
  "payload": "notification_id_here"
}
```

#### 3. `mark_all_read`
Mark all notifications as read.

```json
{
  "type": "mark_all_read"
}
```

---

## Notification Types

| Type | Description | Priority |
|------|-------------|----------|
| `incident_created` | New incident reported | Based on severity |
| `location_transfer_approved` | Transfer request approved | normal |
| `location_transfer_rejected` | Transfer request rejected | normal |
| `evaluation_due` | Client evaluation due soon | normal/high |
| `appointment_reminder` | Upcoming appointment | normal |
| `client_status_changed` | Client status changed | normal |
| `system_alert` | System-wide alert | varies |

### Priority Levels

- `low` - Informational
- `normal` - Standard notifications
- `high` - Important, needs attention
- `urgent` - Critical, immediate action required

---

## REST API Endpoints

### List Notifications

```http
GET /notifications?page=1&page_size=10&is_read=false
Authorization: Bearer <jwt>
```

**Response:**
```json
{
  "data": {
    "data": [
      {
        "id": "abc123",
        "type": "incident_created",
        "priority": "high",
        "title": "New Incident",
        "message": "Details here",
        "resourceType": "incident",
        "resourceId": "inc_456",
        "isRead": false,
        "createdAt": "2026-01-12T12:00:00Z"
      }
    ],
    "page": 1,
    "pageSize": 10,
    "totalCount": 42,
    "totalPages": 5
  },
  "message": "Notifications listed successfully"
}
```

### Get Unread Count

```http
GET /notifications/unread-count
Authorization: Bearer <jwt>
```

**Response:**
```json
{
  "data": {
    "count": 5
  },
  "message": "Unread count retrieved"
}
```

### Mark as Read

```http
PATCH /notifications/:id/read
Authorization: Bearer <jwt>
```

### Mark All as Read

```http
PATCH /notifications/read-all
Authorization: Bearer <jwt>
```

### Delete Notification

```http
DELETE /notifications/:id
Authorization: Bearer <jwt>
```

---

## TypeScript/JavaScript Implementation

### NotificationService Class

```typescript
interface Notification {
  id: string;
  type: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  title: string;
  message: string;
  resourceType?: string;
  resourceId?: string;
  isRead: boolean;
  createdAt: string;
}

interface WebSocketMessage {
  type: string;
  payload?: any;
}

class NotificationService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners: Map<string, Set<(data: any) => void>> = new Map();

  constructor(
    private apiBaseUrl: string,
    private getAccessToken: () => string
  ) {}

  // Connect to WebSocket
  async connect(): Promise<void> {
    try {
      // Step 1: Get ticket
      const ticket = await this.getTicket();
      
      // Step 2: Connect WebSocket
      const wsUrl = this.apiBaseUrl
        .replace('https://', 'wss://')
        .replace('http://', 'ws://');
      
      this.ws = new WebSocket(`${wsUrl}/ws/notifications?ticket=${ticket}`);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.emit('connected', null);
      };
      
      this.ws.onmessage = (event) => {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.handleMessage(message);
      };
      
      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.emit('disconnected', { code: event.code, reason: event.reason });
        this.attemptReconnect();
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.emit('error', error);
      };
    } catch (error) {
      console.error('Failed to connect:', error);
      this.attemptReconnect();
    }
  }

  // Get authentication ticket
  private async getTicket(): Promise<string> {
    const response = await fetch(`${this.apiBaseUrl}/ws/auth`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.getAccessToken()}`,
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      throw new Error('Failed to get WebSocket ticket');
    }
    
    const data = await response.json();
    return data.data.ticket;
  }

  // Handle incoming messages
  private handleMessage(message: WebSocketMessage): void {
    switch (message.type) {
      case 'connected':
        console.log('WebSocket authenticated');
        break;
        
      case 'notification':
        this.emit('notification', message.payload as Notification);
        break;
        
      case 'ping':
        this.send({ type: 'pong' });
        break;
    }
  }

  // Send message to server
  private send(message: WebSocketMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  // Mark notification as read via WebSocket
  markAsRead(notificationId: string): void {
    this.send({ type: 'mark_read', payload: notificationId });
  }

  // Mark all as read via WebSocket
  markAllAsRead(): void {
    this.send({ type: 'mark_all_read' });
  }

  // Reconnection logic
  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      this.emit('reconnect_failed', null);
      return;
    }
    
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectAttempts++;
    
    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
    
    setTimeout(() => {
      this.connect();
    }, delay);
  }

  // Event handling
  on(event: string, callback: (data: any) => void): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);
  }

  off(event: string, callback: (data: any) => void): void {
    this.listeners.get(event)?.delete(callback);
  }

  private emit(event: string, data: any): void {
    this.listeners.get(event)?.forEach(callback => callback(data));
  }

  // Disconnect
  disconnect(): void {
    this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection
    this.ws?.close();
    this.ws = null;
  }
}

// Usage
const notificationService = new NotificationService(
  'https://api.example.com',
  () => localStorage.getItem('accessToken') || ''
);

notificationService.on('notification', (notification: Notification) => {
  console.log('New notification:', notification);
  // Show toast, update badge, etc.
});

notificationService.connect();
```

---

## React Hook Example

```typescript
import { useEffect, useState, useCallback, useRef } from 'react';

interface Notification {
  id: string;
  type: string;
  priority: string;
  title: string;
  message: string;
  isRead: boolean;
  createdAt: string;
}

export function useNotifications(apiBaseUrl: string, accessToken: string) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

  // Fetch initial notifications
  const fetchNotifications = useCallback(async () => {
    try {
      const response = await fetch(`${apiBaseUrl}/notifications?page_size=20`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      const data = await response.json();
      setNotifications(data.data.data);
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
    }
  }, [apiBaseUrl, accessToken]);

  // Fetch unread count
  const fetchUnreadCount = useCallback(async () => {
    try {
      const response = await fetch(`${apiBaseUrl}/notifications/unread-count`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      const data = await response.json();
      setUnreadCount(data.data.count);
    } catch (error) {
      console.error('Failed to fetch unread count:', error);
    }
  }, [apiBaseUrl, accessToken]);

  // Connect to WebSocket
  const connect = useCallback(async () => {
    try {
      // Get ticket
      const ticketResponse = await fetch(`${apiBaseUrl}/ws/auth`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'application/json',
        },
      });
      const ticketData = await ticketResponse.json();
      const ticket = ticketData.data.ticket;

      // Connect WebSocket
      const wsUrl = apiBaseUrl.replace('https://', 'wss://').replace('http://', 'ws://');
      const ws = new WebSocket(`${wsUrl}/ws/notifications?ticket=${ticket}`);

      ws.onopen = () => {
        setIsConnected(true);
        console.log('Notifications WebSocket connected');
      };

      ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        
        if (message.type === 'notification') {
          // Prepend new notification
          setNotifications(prev => [message.payload, ...prev]);
          setUnreadCount(prev => prev + 1);
        } else if (message.type === 'ping') {
          ws.send(JSON.stringify({ type: 'pong' }));
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Reconnect after 3 seconds
        reconnectTimeoutRef.current = setTimeout(connect, 3000);
      };

      wsRef.current = ws;
    } catch (error) {
      console.error('WebSocket connection failed:', error);
      reconnectTimeoutRef.current = setTimeout(connect, 3000);
    }
  }, [apiBaseUrl, accessToken]);

  // Mark as read
  const markAsRead = useCallback(async (id: string) => {
    await fetch(`${apiBaseUrl}/notifications/${id}/read`, {
      method: 'PATCH',
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    setNotifications(prev =>
      prev.map(n => (n.id === id ? { ...n, isRead: true } : n))
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
  }, [apiBaseUrl, accessToken]);

  // Mark all as read
  const markAllAsRead = useCallback(async () => {
    await fetch(`${apiBaseUrl}/notifications/read-all`, {
      method: 'PATCH',
      headers: { Authorization: `Bearer ${accessToken}` },
    });
    setNotifications(prev => prev.map(n => ({ ...n, isRead: true })));
    setUnreadCount(0);
  }, [apiBaseUrl, accessToken]);

  // Initialize
  useEffect(() => {
    if (accessToken) {
      fetchNotifications();
      fetchUnreadCount();
      connect();
    }

    return () => {
      wsRef.current?.close();
      clearTimeout(reconnectTimeoutRef.current);
    };
  }, [accessToken, connect, fetchNotifications, fetchUnreadCount]);

  return {
    notifications,
    unreadCount,
    isConnected,
    markAsRead,
    markAllAsRead,
    refetch: fetchNotifications,
  };
}

// Usage in component
function NotificationBell() {
  const { notifications, unreadCount, isConnected, markAsRead, markAllAsRead } = 
    useNotifications('https://api.example.com', accessToken);

  return (
    <div>
      <span className="badge">{unreadCount}</span>
      <span className={isConnected ? 'status-online' : 'status-offline'} />
      
      {notifications.map(notification => (
        <div 
          key={notification.id}
          onClick={() => markAsRead(notification.id)}
          className={notification.isRead ? 'read' : 'unread'}
        >
          <strong>{notification.title}</strong>
          <p>{notification.message}</p>
        </div>
      ))}
      
      <button onClick={markAllAsRead}>Mark all as read</button>
    </div>
  );
}
```

---

## Error Handling

### Common Error Codes

| HTTP Code | WebSocket Code | Meaning |
|-----------|----------------|---------|
| 401 | - | JWT expired, re-authenticate |
| 401 (ticket) | - | Ticket expired or already used |
| - | 1000 | Normal closure |
| - | 1006 | Abnormal closure (network issue) |
| - | 1008 | Policy violation |

### Handling Token Refresh

```typescript
// When you refresh your access token, reconnect the WebSocket
async function onTokenRefresh(newAccessToken: string) {
  // Close existing connection
  notificationService.disconnect();
  
  // Update token getter
  // ... 
  
  // Reconnect with new token
  await notificationService.connect();
}
```

---

## Reconnection Strategy

The recommended reconnection strategy uses **exponential backoff**:

```
Attempt 1: Wait 1 second
Attempt 2: Wait 2 seconds
Attempt 3: Wait 4 seconds
Attempt 4: Wait 8 seconds
Attempt 5: Wait 16 seconds
Attempt 6+: Give up, show "reconnect" button
```

### Key Points

1. **Always get a fresh ticket** before reconnecting (tickets are one-time use)
2. **Reset attempts** on successful connection
3. **Show UI feedback** when disconnected
4. **Fallback to polling** if WebSocket consistently fails

---

## Testing

### Manual Testing Steps

1. Open browser DevTools → Network → WS tab
2. Call `POST /ws/auth` to get a ticket
3. Connect to `ws://localhost:8080/ws/notifications?ticket=<ticket>`
4. Verify you receive `{"type":"connected"}`
5. Create an incident → Verify notification arrives in WebSocket

### Test Invalid Ticket

```javascript
// Should return 401
const ws = new WebSocket('ws://localhost:8080/ws/notifications?ticket=invalid');
```

### Test Ticket Reuse

```javascript
// Get ticket, use it, try to use again → should fail
const ticket = await getTicket();
const ws1 = new WebSocket(`ws://...?ticket=${ticket}`); // ✓ Works
const ws2 = new WebSocket(`ws://...?ticket=${ticket}`); // ✗ Fails (401)
```

---

## Summary

| Action | Endpoint | Auth |
|--------|----------|------|
| Get ticket | `POST /ws/auth` | JWT |
| Connect WebSocket | `GET /ws/notifications?ticket=xxx` | Ticket |
| List notifications | `GET /notifications` | JWT |
| Get unread count | `GET /notifications/unread-count` | JWT |
| Mark as read | `PATCH /notifications/:id/read` | JWT |
| Mark all as read | `PATCH /notifications/read-all` | JWT |
| Delete | `DELETE /notifications/:id` | JWT |

For questions, contact the backend team.
