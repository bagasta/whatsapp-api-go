# Deploy with PM2

Panduan singkat menjalankan WhatsApp API dengan PM2 di VPS (Ubuntu/Debian serupa).

## Prasyarat
- OS: Ubuntu/Debian atau serupa.
- Paket dasar: `curl`, `git`, `build-essential`, `pkg-config`.
- Go 1.24 (`go version`).
- Node.js + npm (hanya untuk PM2), lalu `npm install -g pm2`.
- FFmpeg (wajib untuk kirim media), SQLite sudah disertakan di repo (default storage).
- Port aplikasi terbuka (default `3000`) atau diproksikan lewat Nginx.

Contoh instalasi cepat (Ubuntu):
```bash
sudo apt update
sudo apt install -y curl git build-essential pkg-config ffmpeg
# Go 1.24 (ubah versi jika perlu)
curl -L https://go.dev/dl/go1.24.0.linux-amd64.tar.gz -o /tmp/go.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
# PM2
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
sudo npm install -g pm2
```

## Siapkan kode dan binary
```bash
git clone https://github.com/aldinokemal/go-whatsapp-web-multidevice.git
cd go-whatsapp-web-multidevice/src
go build -o ../bin/whatsapp .
```

## Konfigurasi lingkungan
1. Salin contoh env dan sesuaikan:
   ```bash
   cp src/.env.example .env
   ```
   Edit `.env` (contoh variabel penting):
   - `APP_PORT=3000`
   - `APP_BASE_PATH=` (opsional prefix, mis. `/wa`)
   - `APP_BASIC_AUTH=user:password` (opsional, pisahkan dengan koma untuk multi user)
   - `AI_BACKEND_URL=https://...` (jika pakai AI)
   - `FFMPEG_PATH=` (isi jika binary FFmpeg tidak di PATH)

2. Pastikan direktori `storages/` dapat ditulis (default SQLite disimpan di situ).

## Buat konfigurasi PM2
Buat `ecosystem.config.js` di root repo:
```js
module.exports = {
  apps: [{
    name: "whatsapp-api",
    script: "./bin/whatsapp",
    args: "rest",
    cwd: "/home/<user>/go-whatsapp-web-multidevice/src",
    env: {
      APP_PORT: "3000",
      APP_BASE_PATH: "",
      // Atau gunakan PM2 --env-file untuk memakai .env
    },
    out_file: "/var/log/whatsapp-api/out.log",
    error_file: "/var/log/whatsapp-api/err.log",
    max_restarts: 10,
    restart_delay: 2000
  }]
};
```

Alternatif: gunakan file env langsung dengan PM2:
```bash
pm2 start ecosystem.config.js --env-file .env
```

## Jalankan dengan PM2
```bash
pm2 start ecosystem.config.js --env-file .env
pm2 logs whatsapp-api     # lihat log
pm2 save                  # simpan proses untuk reboot
pm2 startup               # generate perintah auto start (ikuti instruksi)
```

## Reverse proxy (opsional)
Jika memakai Nginx:
```nginx
location / {
    proxy_pass http://127.0.0.1:3000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## Periksa kesehatan
- Health check: `curl http://localhost:3000/health`
- Metrics: `curl http://localhost:3000/metrics`
- Admin dashboard: `http://<host>:3000/admin` (perlu basic auth jika diaktifkan).

## Update versi
```bash
git pull
cd src
go build -o ../bin/whatsapp .
pm2 restart whatsapp-api
```
