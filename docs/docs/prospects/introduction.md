# Introduction

## Overview
Prospects in Frontier stores and manage subscription preferences for various activities like newsletters, blog updates, and marketing communications. Unlike user preferences which are tied to registered accounts (registered account means record is present in users table), Prospects is designed to handle subscriptions for non-registered users.

A key feature of Prospects is its ability to manage subscriptions without requiring user registration. This makes it ideal for:
- Newsletter subscriptions from website visitors
- Blog update notifications for readers
- Marketing communication preferences
- Event notifications etc

Each email address can be subscribed to multiple activities simultaneously, allowing granular control over different types of communications. The system maintains subscription status and related details independently of Frontier's user authentication system.

### Key Fields
- **Email**: Subscriber's email address
- **Activity**: Type of subscription (newsletters, marketing emails, blog articles)
- **Status**: Current subscription state (`subscribed` or `unsubscribed`)
- **ChangedAt**: Timestamp of last status change (useful for metrics like monthly subscriber growth)
- **Source**: Origin of subscription (`website`, `email`, `social-media`)
- **Verified**: Email verification status (boolean)

## Access Control
Admin users have full CRUD access.

Public access limited to prospect creation via the unauthenticated endpoint.

## API Reference

### Public Endpoints

#### Create a prospect (Unauthenticated)
Endpoint: POST `/v1beta1/prospects`

RPC: `CreateProspectPublic`

Request:

```json
{
  "email": "test@example.com",
  "activity": "newsletter",
  "source": "website"
}
```

Response:
```json
{
  "prospect": {
    "id": "id-1",
    "name": "",
    "email": "test@example.com",
    "phone": "",
    "activity": "newsletter",
    "status": "STATUS_SUBSCRIBED",
    "changed_at": "2025-03-04T07:23:52.611644Z",
    "source": "website",
    "verified": true,
    "created_at": "2025-03-04T07:23:52.611644Z",
    "updated_at": "2025-03-04T07:23:52.611644Z",
    "metadata": {}
  }
}
```
### Admin endpoints

#### Create a prospect 
Endpoint: POST `/v1beta1/admin/prospects`

RPC: `CreateProspect`

Request:

```json5
{
  "email": "test@example.com",
  "activity": "blog",
  "status": 1, // subscribed
  "source": "admin",
  "metadata": {
    "medium": "test"
  }
}
```

Response:
```json
{
  "prospect": {
    "id": "id-2",
    "name": "",
    "email": "test@example.com",
    "phone": "",
    "activity": "blog",
    "status": "STATUS_SUBSCRIBED",
    "changed_at": "2025-03-04T07:23:52.611644Z",
    "source": "admin",
    "verified": true,
    "created_at": "2025-03-04T07:23:52.611644Z",
    "updated_at": "2025-03-04T07:23:52.611644Z",
    "metadata": {"medium": "test"}
  }
}
```

#### Get Prospect
Endpoint: GET `v1beta1/admin/prospects/{id}`

RPC: `GetProspect`

Request: `v1beta1/admin/prospects/id-1`

Response:
```json
{
  "prospect": {
    "id": "id-1",
    "name": "",
    "email": "test@example.com",
    "phone": "",
    "activity": "newsletter",
    "status": "STATUS_SUBSCRIBED",
    "changed_at": "2025-03-04T07:23:52.611644Z",
    "source": "website",
    "verified": true,
    "created_at": "2025-03-04T07:23:52.611644Z",
    "updated_at": "2025-03-04T07:23:52.611644Z",
    "metadata": {}
  }
}
```

#### List Prospects

Endpoint: POST `/v1beta1/admin/prospects/list`

RPC: `ListProspects`

Request:
```json5
{
  // filters, sorting, pagination and groups supported
}
```

Response:

```json
{
  "prospects": [
    {
        "id": "id-1",
        "name": "",
        "email": "test@example.com",
        "phone": "",
        "activity": "newsletter",
        "status": "STATUS_SUBSCRIBED",
        "changed_at": "2025-03-04T07:23:52.611644Z",
        "source": "website",
        "verified": true,
        "created_at": "2025-03-04T07:23:52.611644Z",
        "updated_at": "2025-03-04T07:23:52.611644Z",
        "metadata": {}
    },
    {
        "id": "id-2",
        "name": "",
        "email": "test@example.com",
        "phone": "",
        "activity": "blog",
        "status": "STATUS_SUBSCRIBED",
        "changed_at": "2025-03-04T07:23:52.611644Z",
        "source": "admin",
        "verified": true,
        "created_at": "2025-03-04T07:23:52.611644Z",
        "updated_at": "2025-03-04T07:23:52.611644Z",
        "metadata": {"medium": "test"}
    }
  ],
  "pagination": {
    "offset": 0,
    "limit": 2,
    "total_count": 5
  },
  "group": null
}
```

#### Update Prospect
Endpoint: PUT `/v1beta1/admin/prospects/{id}`

RPC: `UpdateProspect`

Request:
```json
{
    "id": "id-1",
    "name": "updated-name",
    "email": "test@example.com",
    "phone": "",
    "activity": "newsletter",
    "status": 1,
    "source": "website",
    "verified": true,
    "metadata": {}
}
```

Response:
```json
{
  "prospect": {
    "id": "id-1",
    "name": "updated-name",
    "email": "test@example.com",
    "phone": "",
    "activity": "newsletter",
    "status": "STATUS_SUBSCRIBED",
    "changed_at": "2025-03-04T07:23:52.611644Z",
    "source": "website",
    "verified": true,
    "created_at": "2025-03-04T07:23:52.611644Z",
    "updated_at": "2025-03-05T07:23:52.611644Z",
    "metadata": {}
  }
}
```

#### Delete Prospect
Endpoint: DELETE `/v1beta1/admin/prospects/{id}`

RPC: `DeleteProspect`

Request: `v1beta1/admin/prospects/id-1`

Response: `{}`






