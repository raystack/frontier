---
title: Session Management
---

# Session Management

Sessions are created automatically when a user successfully authenticates with Frontier. They allow the system to remember that a user is authenticated without requiring them to authenticate again for every request. Sessions are stored as encrypted cookies in the user's browser and are managed entirely by the Frontier SDK.

:::note
Session management APIs are used internally by the SDK and are not exposed for direct client use. The SDK automatically handles session creation, validation, tracking, and revocation through its built-in components and dialogs.
:::

## What is a Session?

A session represents an authenticated user's active connection to Frontier. Each session contains:

- **Session ID**: Unique identifier for the session
- **User ID**: The authenticated user
- **Metadata**: Information about the device and location (browser, OS, IP address, location)
- **Timestamps**: Created at, updated at, authenticated at, and expiration time
- **Validity**: Whether the session is still active and not expired

Sessions are created automatically when a user successfully completes authentication (via social login, email OTP, etc.) through the SDK. Once created, the session is stored as an encrypted cookie that is sent with every subsequent request.

## Session Lifecycle

### 1. Creation
Sessions are created automatically after successful authentication. The session is stored as an encrypted cookie in the user's browser.

### 2. Validation
On each request, Frontier validates the session cookie to verify:
- The session exists and hasn't been deleted
- The session hasn't expired (based on the configured validity period)
- The session has a valid cookie associated with the user account.

### 3. Activity Tracking
The session's `updatedAt` timestamp is updated when:
- The user makes authenticated requests
- The frontend SDK's `useLastActiveTracker` hook pings the session (every 10 minutes)

:::note
Updating the `updatedAt` timestamp does **not** extend the session validity. Sessions expire based on the configured `validity` period regardless of activity.
:::

### 4. Expiration
Sessions expire when:
- The configured validity period has elapsed (e.g., 7 days from creation)
- The session is manually revoked by the user or admin

When a session expires, the system marks it with a `deleted_at` timestamp (which is `null` by default for active sessions). A cron job runs daily at midnight UTC to permanently delete sessions from the database that have been expired or soft-deleted for 24+ hours.

## Session Metadata

Each session stores metadata about the device and location, as well as tracking information:

### Metadata Fields
These are stored in the `metadata` object and track device/location information:

- **Browser**: The web browser being used (e.g., "Chrome", "Firefox", "Safari")
- **Operating System**: The OS of the device (e.g., "Windows", "macOS", "Linux")
- **IP Address**: The IP address from which the session was created
- **Location**: Geographic information including:
  - City
  - Country
  - Latitude
  - Longitude

### Session Tracking Fields
These are session-level fields for tracking activity:

- **Last Active**: Displays when the session was last used (derived from the `updatedAt` timestamp)
- **Is Current**: Indicates if this is the current browser session

The metadata is automatically extracted from HTTP headers when the session is created or updated. It's useful for:
- Security monitoring (detecting suspicious logins from new locations)
- User experience (showing users where they're logged in)
- Audit trails (tracking where actions were performed from)

### HTTP Headers Used

Frontier extracts session metadata from the following HTTP headers, all of which are configurable:

| Metadata Field | Header Name (Default) | Description |
|----------------|----------------------|-------------|
| IP Address | `x-forwarded-for` | Client's IP address. If the header contains multiple comma-separated values (common with proxies), the first value is used. |
| Country | `x-frontier-country` | Country code or name from the client's location |
| City | `x-frontier-city` | City name from the client's location |
| Latitude | `x-frontier-latitude` | Geographic latitude coordinate |
| Longitude | `x-frontier-longitude` | Geographic longitude coordinate |
| Browser & OS | `User-Agent` | Browser and operating system information. Parsed using the `uap-go` library to extract browser family and OS family. |

### Configuring Headers

All header names can be customized through the `authentication.session.headers` configuration section. This is useful when:
- Using a reverse proxy or CDN (like CloudFront) that sets different header names
- Integrating with infrastructure that already provides location headers with different names
- Customizing header names for security or compatibility reasons

Example configuration:

```yaml
authentication:
  session:
    headers:
      client_ip: "X-Forwarded-For"
      client_country: "X-Country"
      client_city: "X-City"
      client_latitude: "CloudFront-Viewer-Latitude"
      client_longitude: "CloudFront-Viewer-Longitude"
      client_user_agent: "User-Agent"
```

If a header is not present in the request, the corresponding metadata field will be empty. The system gracefully handles missing headers without errors.

## Viewing Active Sessions

The SDK provides built-in UI components that allow users to view all their active sessions. This functionality is integrated into the SDK's account management dialogs and displays:

- All active sessions for the authenticated user
- Each session's metadata (browser, OS, IP, location)
- The "Last Active" time (when the session was last used)
- Which session is the current one

:::info
The session listing is handled automatically by the SDK's UI components. You don't need to implement this functionality yourself - it's available out-of-the-box when you integrate the Frontier SDK.
:::

### Internal Implementation

For reference, the SDK internally uses the `useSessions` hook to fetch and display session data:

```typescript
import { useSessions } from '@raystack/frontier/react';

// This is used internally by SDK dialogs
function SessionsList() {
  const { sessions, isLoading, error } = useSessions();
  
  return (
    <div>
      {sessions.map(session => (
        <div key={session.id}>
          <p>Browser: {session.browser} on {session.operatingSystem}</p>
          <p>IP: {session.ipAddress}</p>
          <p>Location: {session.location}</p>
          <p>Last Active: {session.lastActive}</p>
          {session.isCurrent && <span>Current Session</span>}
        </div>
      ))}
    </div>
  );
}
```

## Revoking Sessions

The SDK's built-in session management UI allows users to revoke any of their active sessions. This immediately invalidates the session and logs out the user from that device. This is useful if:
- A user suspects their account has been compromised
- A user wants to log out from a specific device
- A user wants to clean up old sessions

:::warning
Revoking a session immediately logs out the user from that device. They will need to authenticate again to access Frontier from that device.
:::

:::info
Session revocation is handled automatically by the SDK's UI components. The revoke functionality is available in the session management dialog that comes with the SDK.
:::

### Internal Implementation

For reference, the SDK internally uses the `useSessions` hook to handle session revocation:

```typescript
import { useSessions } from '@raystack/frontier/react';

// This is used internally by SDK dialogs
function RevokeSessionButton({ sessionId }) {
  const { revokeSession, isRevokingSession } = useSessions();
  
  return (
    <button 
      onClick={() => revokeSession(sessionId)}
      disabled={isRevokingSession}
    >
      {isRevokingSession ? 'Revoking...' : 'Revoke Session'}
    </button>
  );
}
```

## Automatic Activity Tracking

The Frontier React SDK automatically tracks user activity to maintain accurate "Last Active" timestamps for each session. This is handled internally and requires no setup.

### How It Works

The SDK uses an internal `useLastActiveTracker` hook that:

- **Pings the session API every 10 minutes** to update the session's `updatedAt` timestamp
- **Updates session metadata** (browser, OS, IP, location) with current information
- **Continues tracking in background** even when the browser tab is not active
- **Is automatically enabled** when a user is logged in

Each ping:
1. Calls the internal session API endpoint
2. Updates the session's `updatedAt` timestamp
3. Updates session metadata with current device/location information
4. Returns the updated session metadata to the SDK
5. Stores the metadata in `FrontierContext` as `sessionMetadata` for application-wide access

### Purpose

The primary purpose of this tracking is to display accurate "Last Active" times in the session management UI. This helps users:
- See when they were last active on each device
- Identify which sessions are currently in use
- Detect suspicious activity (e.g., a session showing recent activity when they haven't been using that device)

:::important
**Important**: Activity tracking does **not** extend session validity. Sessions expire based on the configured validity period regardless of activity. The tracking only updates the `updatedAt` timestamp for display purposes.
:::

### Accessing Current Session Metadata

You can access the current session's metadata through the `useFrontier` hook:

```typescript
import { useFrontier } from '@raystack/frontier/react';

function MyComponent() {
  const { sessionMetadata } = useFrontier();
  
  // sessionMetadata contains current session information:
  // - browser: string
  // - operatingSystem: string
  // - ipAddress: string
  // - location: { city?, country?, latitude?, longitude? }
  
  return (
    <div>
      <p>Current Browser: {sessionMetadata?.browser}</p>
      <p>Current OS: {sessionMetadata?.operatingSystem}</p>
      <p>IP Address: {sessionMetadata?.ipAddress}</p>
    </div>
  );
}
```

## gRPC APIs

Frontier provides gRPC APIs for session management, split between user-facing and admin-only operations:

### User APIs

These APIs are available to authenticated users to manage their own sessions:

1. **`ListSessions`** - Returns a list of all active sessions for the current authenticated user
2. **`RevokeSession`** - Revokes a specific session for the current authenticated user
3. **`PingUserSession`** - Pings the user's current active session to update last active timestamp and metadata

:::info
These user APIs are used internally by the SDK and are not meant to be called directly by client applications. The SDK handles all session management through its built-in components.
:::

### Admin APIs

These APIs require admin privileges and allow administrators to manage sessions for any user:

1. **`ListUserSessions`** - Returns a list of all sessions for a specific user (Admin access required)
2. **`RevokeUserSession`** - Revokes a specific session for a specific user (admin only)

:::warning
Admin APIs are restricted to users with admin privileges and are typically used in administrative interfaces or internal tools for security and support purposes.
:::

## Related Documentation

- [User Authentication](./user.md) - Learn about authentication strategies
- [Configuration Reference](../reference/configurations.md) - Detailed configuration options
- [API Authentication](../reference/api-auth.md) - How to authenticate API requests

