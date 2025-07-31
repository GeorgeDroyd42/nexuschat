# Nexus

**Nexus** is a privacy-focused, real-time chat platform built for performance and scalability.

Currently in development, Nexus includes:

- Guild and channel creation
- Real-time messaging (WebSocket-based)
- Custom profile pictures and presence indicators
- Admin dashboard for managing bans, webhooks, users, and logs
- Settings pages for guilds, channels, and users
- User bios
- Modular frontend (vanilla JS + Tailwind CSS)
- Backend in Go using Echo, with Redis and PostgreSQL
- Fully responsive design for desktop and mobile
- Tested up to 50,000 concurrent WebSocket connections

---

## Screenshots

### Desktop

<img src="images/desktopguild.png" alt="Desktop UI" width="600"/>

### Mobile

<img src="images/mobileguild.png" alt="Mobile UI" width="300"/>

---

## Stack

- **Backend**: Go (Echo)
- **Frontend**: Vanilla JavaScript + Tailwind CSS
- **Database**: PostgreSQL
- **Cache / Presence**: Redis
- **Real-time**: Native WebSocket support

---

## Performance

Nexus is optimized for high-throughput, low-latency messaging at scale:

- Load tested with 50,000 concurrent WebSocket connections
- Sub-20ms message delivery latency under moderate CPU load
- Redis-backed presence and pub/sub for real-time coordination
- Horizontal scalability with stateless workers and shared cache

---

## Status

This project is under active development.  
No license is applied yet.  
Contributions and self-hosting instructions will be available at a later stage.
