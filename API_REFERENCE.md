# WhatsApp API (Go) - API Reference

This document provides a comprehensive reference for the WhatsApp API endpoints.

## Base URL
By default, the API is served at:
`http://localhost:8080` (or the port specified in `.env`)

## Authentication
If `APP_BASIC_AUTH` is set in `.env` (e.g., `user1:pass1`), you must provide the credentials using Basic Auth header.

## Response Format
All successful responses follow this standard JSON structure:
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Operation successful",
  "results": { ... } // or [...] or null
}
```

---

## 📱 App Management

### Login (QR Code)
Initiate a login session and get a QR code.
- **URL:** `/app/login`
- **Method:** `GET`
- **Response:**
  ```json
  {
    "qr_link": "http://localhost:8080/statics/qr.png",
    "qr_duration": 60
  }
  ```
**cURL:**
```bash
curl -u user1:pass1 http://localhost:8080/app/login
```

### Login with Code
Initiate login using a pairing code (for phone number login).
- **URL:** `/app/login-with-code`
- **Method:** `GET`
- **Query Params:**
  - `phone`: Phone number (e.g., `628123456789`)
- **Response:**
  ```json
  {
    "pair_code": "ABC-123"
  }
  ```
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/app/login-with-code?phone=628123456789"
```

### Logout
Logout the current session.
- **URL:** `/app/logout`
- **Method:** `GET`
**cURL:**
```bash
curl -u user1:pass1 http://localhost:8080/app/logout
```

### Reconnect
Force a reconnection to WhatsApp servers.
- **URL:** `/app/reconnect`
- **Method:** `GET`
**cURL:**
```bash
curl -u user1:pass1 http://localhost:8080/app/reconnect
```

### Fetch Devices
Get list of connected devices.
- **URL:** `/app/devices`
- **Method:** `GET`
**cURL:**
```bash
curl -u user1:pass1 http://localhost:8080/app/devices
```

### Connection Status
Check the current connection status.
- **URL:** `/app/status`
- **Method:** `GET`
- **Response:**
  ```json
  {
    "is_connected": true,
    "is_logged_in": true,
    "device_id": "..."
  }
  ```
**cURL:**
```bash
curl -u user1:pass1 http://localhost:8080/app/status
```

---

## 💬 Chat Management

### List Chats
Get a list of recent chats.
- **URL:** `/chats`
- **Method:** `GET`
- **Query Params:**
  - `limit`: (Optional) Number of chats (default: 25)
  - `offset`: (Optional) Pagination offset (default: 0)
  - `search`: (Optional) Search term
  - `has_media`: (Optional) Filter chats with media (true/false)
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/chats?limit=10"
```

### Get Chat Messages
Fetch messages from a specific chat.
- **URL:** `/chat/:chat_jid/messages`
- **Method:** `GET`
- **URL Params:**
  - `chat_jid`: JID of the chat (e.g., `628123456789@s.whatsapp.net`)
- **Query Params:**
  - `limit`: (Optional) Default 50
  - `offset`: (Optional) Default 0
  - `media_only`: (Optional) true/false
  - `search`: (Optional) Search within chat
  - `start_time`: (Optional) Filter by start time
  - `end_time`: (Optional) Filter by end time
  - `is_from_me`: (Optional) Filter by sender (true/false)
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/chat/628123456789@s.whatsapp.net/messages?limit=20"
```

### Pin Chat
Pin or unpin a chat.
- **URL:** `/chat/:chat_jid/pin`
- **Method:** `POST`
- **URL Params:**
  - `chat_jid`: JID of the chat
- **Body:**
  ```json
  {
    "pinned": true
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/chat/628123456789@s.whatsapp.net/pin \
  -H "Content-Type: application/json" \
  -d '{"pinned": true}'
```

---

## 📤 Sending Messages

### Send Text
- **URL:** `/send/message`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789",
    "message": "Hello World",
    "reply_message_id": "optional-message-id-to-reply-to"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/message \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "message": "Hello World"}'
```

### Send Image
- **URL:** `/send/image`
- **Method:** `POST`
- **Content-Type:** `multipart/form-data`
- **Form Fields:**
  - `image`: (File) The image file
  - `phone`: Recipient phone number
  - `caption`: (Optional) Image caption
  - `view_once`: (Optional) true/false
  - `compress`: (Optional) true/false
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/image \
  -F "image=@/path/to/image.jpg" \
  -F "phone=628123456789" \
  -F "caption=Check this out"
```

### Send File
- **URL:** `/send/file`
- **Method:** `POST`
- **Content-Type:** `multipart/form-data`
- **Form Fields:**
  - `file`: (File) The document file
  - `phone`: Recipient phone number
  - `caption`: (Optional) File caption
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/file \
  -F "file=@/path/to/document.pdf" \
  -F "phone=628123456789"
```

### Send Video
- **URL:** `/send/video`
- **Method:** `POST`
- **Content-Type:** `multipart/form-data`
- **Form Fields:**
  - `video`: (File) The video file
  - `phone`: Recipient phone number
  - `caption`: (Optional) Video caption
  - `view_once`: (Optional) true/false
  - `compress`: (Optional) true/false
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/video \
  -F "video=@/path/to/video.mp4" \
  -F "phone=628123456789"
```

### Send Sticker
- **URL:** `/send/sticker`
- **Method:** `POST`
- **Content-Type:** `multipart/form-data`
- **Form Fields:**
  - `sticker`: (File) The image/sticker file
  - `phone`: Recipient phone number
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/sticker \
  -F "sticker=@/path/to/sticker.webp" \
  -F "phone=628123456789"
```

### Send Contact
- **URL:** `/send/contact`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789",
    "contact_name": "John Doe",
    "contact_phone": "628987654321"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/contact \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "contact_name": "John Doe", "contact_phone": "628987654321"}'
```

### Send Link
- **URL:** `/send/link`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789",
    "link": "https://google.com",
    "caption": "Check this out"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/link \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "link": "https://google.com", "caption": "Search here"}'
```

### Send Location
- **URL:** `/send/location`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789",
    "latitude": "-6.200000",
    "longitude": "106.816666",
    "name": "Jakarta",
    "address": "Jakarta, Indonesia"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/location \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "latitude": "-6.2", "longitude": "106.8", "name": "Jakarta"}'
```

### Send Audio
- **URL:** `/send/audio`
- **Method:** `POST`
- **Content-Type:** `multipart/form-data`
- **Form Fields:**
  - `audio`: (File) The audio file
  - `phone`: Recipient phone number
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/audio \
  -F "audio=@/path/to/audio.mp3" \
  -F "phone=628123456789"
```

### Send Poll
- **URL:** `/send/poll`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789",
    "question": "Favorite color?",
    "options": ["Red", "Blue", "Green"],
    "max_vote_count": 1
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/poll \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "question": "Favorite color?", "options": ["Red", "Blue"], "max_answer": 1}'
```

### Send Presence
Set your presence status (e.g., "Online").
- **URL:** `/send/presence`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789",
    "presence": "available" // or "unavailable"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/presence \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "presence": "available"}'
```

### Send Chat Presence
Set chat-specific presence (e.g., "Typing...").
- **URL:** `/send/chat-presence`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789",
    "presence": "composing" // "composing", "paused", "recording"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/send/chat-presence \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "presence": "composing"}'
```

---

## 👤 User & Account

### User Info
Get information about a WhatsApp user.
- **URL:** `/user/info`
- **Method:** `GET`
- **Query Params:**
  - `phone`: Phone number
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/user/info?phone=628123456789"
```

### User Avatar
Get a user's profile picture.
- **URL:** `/user/avatar`
- **Method:** `GET`
- **Query Params:**
  - `phone`: Phone number
  - `is_community`: (Optional) true/false
  - `is_preview`: (Optional) true/false
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/user/avatar?phone=628123456789"
```

### Change Avatar
Change your own profile picture.
- **URL:** `/user/avatar`
- **Method:** `POST`
- **Content-Type:** `multipart/form-data`
- **Form Fields:**
  - `avatar`: (File) Image file
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/user/avatar \
  -F "avatar=@/path/to/profile.jpg"
```

### Change Push Name
Change your display name.
- **URL:** `/user/pushname`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "name": "New Name"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/user/pushname \
  -H "Content-Type: application/json" \
  -d '{"name": "My New Name"}'
```

### Check User
Check if a number is registered on WhatsApp.
- **URL:** `/user/check`
- **Method:** `GET`
- **Query Params:**
  - `phone`: Phone number
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/user/check?phone=628123456789"
```

### Business Profile
Get business profile information.
- **URL:** `/user/business-profile`
- **Method:** `GET`
- **Query Params:**
  - `phone`: Phone number
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/user/business-profile?phone=628123456789"
```

### My Data
- **Privacy Settings:** `GET /user/my/privacy`
- **My Groups:** `GET /user/my/groups`
- **My Newsletters:** `GET /user/my/newsletters`
- **My Contacts:** `GET /user/my/contacts`
**cURL:**
```bash
curl -u user1:pass1 http://localhost:8080/user/my/contacts
```

---

## 📩 Message Actions

### React to Message
- **URL:** `/message/:message_id/reaction`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789", // Chat JID where message is
    "emoji": "👍"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/message/MSGID123/reaction \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "emoji": "👍"}'
```

### Revoke (Delete for Everyone)
- **URL:** `/message/:message_id/revoke`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/message/MSGID123/revoke \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789"}'
```

### Delete (Delete for Me)
- **URL:** `/message/:message_id/delete`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/message/MSGID123/delete \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789"}'
```

### Update (Edit) Message
- **URL:** `/message/:message_id/update`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789",
    "message": "New text content"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/message/MSGID123/update \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789", "message": "Edited text"}'
```

### Mark as Read
- **URL:** `/message/:message_id/read`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "phone": "628123456789"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/message/MSGID123/read \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789"}'
```

### Star/Unstar Message
- **Star:** `POST /message/:message_id/star`
- **Unstar:** `POST /message/:message_id/unstar`
- **Body:**
  ```json
  {
    "phone": "628123456789"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/message/MSGID123/star \
  -H "Content-Type: application/json" \
  -d '{"phone": "628123456789"}'
```

### Download Media
- **URL:** `/message/:message_id/download`
- **Method:** `GET`
- **Query Params:**
  - `phone`: Chat JID
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/message/MSGID123/download?phone=628123456789"
```

---

## 👥 Group Management

### Create Group
- **URL:** `/group`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "title": "My New Group",
    "participants": ["628123456789", "628987654321"]
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/group \
  -H "Content-Type: application/json" \
  -d '{"title": "My Group", "participants": ["628123456789"]}'
```

### Join Group via Link
- **URL:** `/group/join-with-link`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "link": "https://chat.whatsapp.com/..."
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/group/join-with-link \
  -H "Content-Type: application/json" \
  -d '{"link": "https://chat.whatsapp.com/INVITELINK"}'
```

### Group Info
- **URL:** `/group/info`
- **Method:** `GET`
- **Query Params:**
  - `group_id`: Group JID (e.g., `123456789@g.us`)
**cURL:**
```bash
curl -u user1:pass1 "http://localhost:8080/group/info?group_id=123456789@g.us"
```

### Leave Group
- **URL:** `/group/leave`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "group_id": "123456789@g.us"
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/group/leave \
  -H "Content-Type: application/json" \
  -d '{"group_id": "123456789@g.us"}'
```

### Manage Participants
- **List:** `GET /group/participants?group_id=...`
- **Export CSV:** `GET /group/participants/export?group_id=...`
- **Add:** `POST /group/participants`
- **Remove:** `POST /group/participants/remove`
- **Promote:** `POST /group/participants/promote`
- **Demote:** `POST /group/participants/demote`
- **Body for POST actions:**
  ```json
  {
    "group_id": "123456789@g.us",
    "participants": ["628123456789"]
  }
  ```
**cURL (Add Participant):**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/group/participants \
  -H "Content-Type: application/json" \
  -d '{"group_id": "123456789@g.us", "participants": ["628123456789"]}'
```

### Group Settings
- **Set Photo:** `POST /group/photo` (Form-data: `group_id`, `photo`)
- **Set Name:** `POST /group/name` (Body: `group_id`, `name`)
- **Set Locked (Admins only):** `POST /group/locked` (Body: `group_id`, `locked`=true/false)
- **Set Announce (Only admins send):** `POST /group/announce` (Body: `group_id`, `announce`=true/false)
- **Set Topic:** `POST /group/topic` (Body: `group_id`, `topic`)
- **Get Invite Link:** `GET /group/invite-link?group_id=...`
**cURL (Set Name):**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/group/name \
  -H "Content-Type: application/json" \
  -d '{"group_id": "123456789@g.us", "name": "New Group Name"}'
```

---

## 📰 Newsletter

### Unfollow Newsletter
- **URL:** `/newsletter/unfollow`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "newsletter_id": "..."
  }
  ```
**cURL:**
```bash
curl -X POST -u user1:pass1 http://localhost:8080/newsletter/unfollow \
  -H "Content-Type: application/json" \
  -d '{"newsletter_id": "123456@newsletter"}'
```
