# Integrasi WhatsApp API dengan n8n

Dokumen ini menjelaskan cara menghubungkan WhatsApp API (Go) dengan n8n untuk otomatisasi dua arah: mengirim pesan dari n8n dan menerima pesan masuk di n8n.

## 1. Mengirim Pesan WhatsApp dari n8n (Outbound)

Anda dapat menggunakan node **HTTP Request** di n8n untuk memanggil API ini.

### Konfigurasi Node HTTP Request
- **Method:** `POST`
- **URL:** `http://<ALAMAT-API-ANDA>:8080/send/message`
  - *Jika n8n dan API ada di mesin yang sama (Docker), gunakan `http://host.docker.internal:8080/send/message` atau nama service docker compose.*
  - *Jika n8n di cloud dan API di lokal, gunakan Tunnel (ngrok/cloudflared).*
- **Authentication:** `Basic Auth`
  - **Username:** `user1` (sesuai `.env`)
  - **Password:** `pass1` (sesuai `.env`)
- **Send Body:** `JSON`
- **Body Parameters:**
  ```json
  {
    "phone": "628123456789",
    "message": "Halo, ini pesan otomatis dari n8n! 🚀"
  }
  ```

### Contoh Penggunaan Lain
Anda bisa mengganti URL endpoint untuk mengirim media, lokasi, dll. Lihat `API_REFERENCE.md` untuk daftar lengkap endpoint.

---

## 2. Menerima Pesan Masuk di n8n (Inbound / Webhook)

Agar n8n dapat merespons pesan masuk (misal: Chatbot), Anda perlu mendaftarkan URL Webhook n8n ke dalam konfigurasi API ini.

### Langkah 1: Buat Webhook di n8n
1. Buka n8n, buat Workflow baru.
2. Tambahkan node **Webhook**.
3. Set **HTTP Method** ke `POST`.
4. Set **Path** (misal: `wa-incoming`).
5. Salin **Production URL** (atau Test URL untuk percobaan).
   - Contoh: `https://n8n.domain-anda.com/webhook/wa-incoming`

### Langkah 2: Konfigurasi API Go
Edit file `src/.env` dan tambahkan URL webhook n8n tersebut.

```dotenv
# src/.env

# Masukkan URL n8n di sini
WHATSAPP_WEBHOOK=https://n8n.domain-anda.com/webhook/wa-incoming

# Secret key untuk memverifikasi bahwa data benar-benar dari API ini (Opsional tapi disarankan)
WHATSAPP_WEBHOOK_SECRET=kunci-rahasia-anda
```

### Langkah 3: Restart API
Setelah mengubah `.env`, restart aplikasi Go.

```bash
go run . rest --port 8080 --debug true --basic-auth user1:pass1
```

### Langkah 4: Tes Data Masuk
Kirim pesan WhatsApp ke nomor yang terhubung. n8n akan menerima JSON dengan struktur umum seperti ini:

```json
{
    "id": "3EB0...",
    "from": "628123456789@s.whatsapp.net",
    "push_name": "Bagas",
    "message": {
        "conversation": "Halo bot"
    },
    "timestamp": 1678888888
}
```

### Keamanan Webhook (Opsional)
API ini mengirimkan header `X-Webhook-Secret` yang berisi nilai dari `WHATSAPP_WEBHOOK_SECRET`.
Di n8n, Anda bisa menambahkan node **If** untuk mengecek apakah header tersebut cocok, guna mencegah request palsu dari pihak lain.

---

## 3. Contoh Skenario Otomatisasi

### Auto-Reply Sederhana
1. **Webhook (n8n):** Menerima pesan masuk.
2. **If (n8n):** Cek apakah pesan berisi kata "halo".
3. **HTTP Request (n8n):** Panggil API `/send/message` untuk membalas "Halo juga! Ada yang bisa dibantu?".

### Kirim Notifikasi Order
1. **Webhook (WooCommerce/Shopify):** Ada order baru.
2. **HTTP Request (n8n):** Panggil API `/send/message` untuk kirim WA ke pembeli: "Terima kasih, pesanan #123 sedang diproses."

### Chatbot AI
1. **Webhook (n8n):** Terima pesan.
2. **OpenAI / LangChain (n8n):** Proses pesan user.
3. **HTTP Request (n8n):** Kirim balasan hasil AI ke user via WA.
