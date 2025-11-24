# WhatsApp API Go – Full Setup Guide

## Overview
This repository provides a **WhatsApp Web Multi‑Device API** written in Go. It includes a fully‑featured **dashboard** built with vanilla HTML, CSS, and Vue 3 (via CDN). The dashboard lets you:
- Connect/disconnect devices
- Send messages, media, contacts, locations, etc.
- Manage groups, newsletters, and account settings
- View real‑time device status

The UI already uses a modern, premium design with glass‑morphism, smooth animations, and a dark‑blue color palette.

---

## Prerequisites
- **Linux** (Ubuntu/Debian based) – you are on this already.
- **Go 1.22+** (installed via `sudo apt install golang-go`).
- **FFmpeg** – required for media handling (`sudo apt install ffmpeg`).
- **Docker & Docker‑Compose** (optional, for containerised deployment).

---

## 1️⃣ Clone the repository (already done)
```bash
git clone https://github.com/aldinokemal/go-whatsapp-web-multidevice.git .
```

---

## 2️⃣ Install Go dependencies
```bash
cd /home/bagas/Whatsapp-api-go/src
go mod tidy   # downloads all required modules
```
> The command may take a minute as it pulls many dependencies.

---

## 3️⃣ Configure environment variables
A template file exists at `src/.env.example`. Copy it to `.env` and edit the values to suit your setup:
```bash
cp src/.env.example src/.env
```
Edit `src/.env` with your favourite editor (e.g., `nano src/.env`). Important fields:
- `APP_PORT` – port the dashboard will listen on (default **3000**).
- `APP_BASIC_AUTH` – `user:pass` for HTTP basic auth (optional).
- `WHATSAPP_WEBHOOK` – URL(s) to receive webhook events (leave empty if not needed).
- `WHATSAPP_AUTO_REPLY` – default auto‑reply text.

---

## 4️⃣ Run the server (development mode)
```bash
# From the repository root
go run ./src/main.go
```
The server starts on `http://localhost:3000`. Open that URL in a browser – you should see the splash screen, then the dashboard once a device connects.

---

## 5️⃣ Docker‑Compose (quick start)
If you prefer containerised execution, a `docker-compose.yml` is already provided.
```bash
# Build and start the container
docker compose up --build -d
```
The container will automatically copy the `.env` file, install Go modules, and launch the app. Access the dashboard at `http://localhost:3000`.

---

## 6️⃣ Using the Dashboard
1. **Connect a device** – click **Login** → a QR code appears in WhatsApp on your phone. Scan it.
2. **Send messages** – use the *Send* cards to dispatch text, images, files, etc.
3. **Manage groups** – create, join, add participants, change photo/name, etc.
4. **Account settings** – change avatar, push name, business profile, privacy, etc.
5. **Webhooks** – if you configured `WHATSAPP_WEBHOOK`, events will be POSTed there.

All UI components are defined in `src/views/components/*.js`. Feel free to customise them – the CSS lives in `src/views/assets/app.css` and already follows a premium design system (gradient background, glass‑morphism, micro‑animations).

---

## 7️⃣ Production build (optional)
For a production‑ready binary:
```bash
cd /home/bagas/Whatsapp-api-go/src
go build -o whatsapp-api
./whatsapp-api
```
You can then run it behind a reverse proxy (NGINX, Caddy, etc.) and enable TLS.

---

## 8️⃣ Troubleshooting
- **Port already in use** – change `APP_PORT` in `.env`.
- **FFmpeg missing** – install with `sudo apt install ffmpeg`.
- **WebSocket connection fails** – ensure the port is open and you are accessing the correct host (`http://localhost:3000`).
- **Docker build fails** – make sure Docker is running and you have enough memory (the Go build can be memory‑heavy).

---

## 9️⃣ Next steps / Customisation
- Replace the logo (`src/views/assets/gowa.svg`) with your own branding.
- Extend the UI by adding new Vue components in `src/views/components/`.
- Integrate a database (SQLite is default) for persisting messages.
- Deploy to a cloud VM or Kubernetes – the Docker image can be pushed to any registry.

---

**Enjoy building powerful WhatsApp integrations with Go!**
