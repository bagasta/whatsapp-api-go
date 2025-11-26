const template = `
<div class="ui stackable grid">
  <div class="sixteen wide column">
    <div class="ui two stackable cards">
      <div class="card">
        <div class="content">
          <div class="header">
            Default Webhook (Fallback)
            <div class="sub header" style="margin-top:4px;">Dipakai semua session baru jika tidak ada override</div>
          </div>
        </div>
        <div class="content">
          <div class="ui form">
            <div class="field required">
              <label>Webhook URL</label>
              <input type="url" v-model="defaultUrl" placeholder="https://your-n8n.example.com/webhook/xyz">
            </div>
            <div class="field">
              <label>Secret (opsional, HMAC header)</label>
              <input type="text" v-model="defaultSecret" placeholder="Shared secret untuk signature">
            </div>
            <div class="ui grid">
              <div class="eight wide column">
                <button class="ui primary button" type="button" @click="saveDefault" :class="{loading:loadingDefault}">
                  <i class="save icon"></i>
                  Simpan Default
                </button>
                <button class="ui button" type="button" @click="loadDefault" :disabled="loadingDefault">
                  Refresh
                </button>
              </div>
              <div class="eight wide column right aligned">
                <div class="meta" v-if="defaultUpdatedAt">
                  Terakhir disimpan: [[ formatDate(defaultUpdatedAt) ]]
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="content">
          <div class="header">
            Buat Session & Scan QR
            <div class="sub header" style="margin-top:4px;">Daftarkan agent baru, lalu scan QR di sini</div>
          </div>
        </div>
        <div class="content">
          <div class="ui form">
            <div class="two fields">
              <div class="field required">
                <label>User ID</label>
                <input type="text" v-model="form.userId" placeholder="USER_ID">
              </div>
              <div class="field required">
                <label>Agent ID</label>
                <input type="text" v-model="form.agentId" placeholder="AGENT_ID unik">
              </div>
            </div>
            <div class="field required">
              <label>Agent Name</label>
              <input type="text" v-model="form.agentName" placeholder="Contoh: Support Bot">
            </div>
            <div class="two fields">
              <div class="field">
                <label>API Key (opsional)</label>
                <input type="text" v-model="form.apikey" placeholder="Jika kosong akan pakai api_keys aktif">
              </div>
              <div class="field">
                <label>Endpoint AI (opsional)</label>
                <input type="url" v-model="form.endpointUrlRun" placeholder="Override AI endpoint">
              </div>
            </div>
            <div class="ui grid">
              <div class="eight wide column">
                <button class="ui primary button" type="button" @click="createSession" :class="{loading:creating}">
                  <i class="qrcode icon"></i>
                  Buat Session & Ambil QR
                </button>
              </div>
              <div class="eight wide column right aligned">
                <div class="meta" v-if="qrUpdatedAt">
                  QR diperbarui: [[ formatDate(qrUpdatedAt) ]]
                </div>
              </div>
            </div>
          </div>
          <div class="ui divider"></div>
          <div class="qr-wrapper" v-if="qrBase64">
            <img :src="'data:image/png;base64,' + qrBase64" alt="QR Code" class="ui bordered rounded image medium centered">
            <div class="ui info message" style="margin-top:12px;">Scan QR ini dengan WhatsApp untuk login.</div>
          </div>
          <div class="ui message" v-else>
            Belum ada QR. Isi data dan klik "Buat Session & Ambil QR".
          </div>
        </div>
      </div>
    </div>
  </div>

  <div class="sixteen wide column">
    <div class="ui fluid card">
      <div class="content">
        <div class="header">Cara pakai dengan n8n</div>
      </div>
      <div class="content">
        <div class="ui ordered list">
          <div class="item">Set <b>Default Webhook</b> — semua session (termasuk yang baru) otomatis kirim event ke URL ini.</div>
          <div class="item">Buat session baru di kartu "Buat Session & Scan QR", lalu scan QR untuk login.</div>
          <div class="item">Secret akan digunakan untuk header <code>X-Go-Wa-Signature</code> (HMAC-SHA256).</div>
          <div class="item">Klik <b>Simpan</b>, lalu uji dengan mengirim pesan ke nomor bot.</div>
        </div>
        <div class="ui divider"></div>
        <div class="ui small message">
          <div class="header">Format event</div>
          <p>Body JSON berisi detail event WhatsApp (message, receipt, delete, group info). Signature dihitung dari body menggunakan secret default atau override per session.</p>
        </div>
      </div>
    </div>
  </div>
</div>
`;

export default {
  name: 'AdminWebhook',
  template,
  delimiters: ['[[', ']]'],
  data() {
    return {
      defaultUrl: '',
      defaultSecret: '',
      defaultUpdatedAt: '',
      loadingDefault: false,
      form: {
        userId: '',
        agentId: '',
        agentName: '',
        apikey: '',
        endpointUrlRun: '',
      },
      creating: false,
      qrBase64: '',
      qrUpdatedAt: '',
      qrTimer: null,
    };
  },
  mounted() {
    this.loadDefault();
  },
  beforeUnmount() {
    this.stopQRRefresh();
  },
  methods: {
    async loadDefault() {
      try {
        this.loadingDefault = true;
        const { data } = await window.http.get('admin/webhook-config');
        this.defaultUrl = data?.url || '';
        this.defaultSecret = data?.secret || '';
        this.defaultUpdatedAt = data?.updated_at || '';
      } catch (err) {
        const msg = err?.response?.data?.error?.message || err?.message || 'Gagal memuat default webhook';
        window.showErrorInfo(msg);
      } finally {
        this.loadingDefault = false;
      }
    },
    async saveDefault() {
      if (!this.defaultUrl) {
        window.showErrorInfo('Webhook URL wajib diisi');
        return;
      }
      try {
        this.loadingDefault = true;
        const { data } = await window.http.post('admin/webhook-config', {
          url: this.defaultUrl,
          secret: this.defaultSecret,
        });
        this.defaultUpdatedAt = data?.updated_at || '';
        window.showSuccessInfo('Default webhook tersimpan');
      } catch (err) {
        const msg = err?.response?.data?.error?.message || err?.message || 'Gagal menyimpan default webhook';
        window.showErrorInfo(msg);
      } finally {
        this.loadingDefault = false;
      }
    },
    async createSession() {
      const { userId, agentId, agentName, apikey, endpointUrlRun } = this.form;
      if (!userId || !agentId || !agentName) {
        window.showErrorInfo('User ID, Agent ID, dan Agent Name wajib diisi');
        return;
      }
      this.stopQRRefresh();
      try {
        this.creating = true;
        const { data } = await window.http.post('sessions', {
          userId,
          agentId,
          agentName,
          apikey,
          endpointUrlRun,
        });
        if (data?.qr?.base64) {
          this.qrBase64 = data.qr.base64;
          this.qrUpdatedAt = new Date().toISOString();
        } else {
          await this.fetchQR(agentId);
        }
        this.startQRRefresh(agentId);
        window.showSuccessInfo('Session dibuat. Scan QR untuk login.');
      } catch (err) {
        const msg = err?.response?.data?.error?.message || err?.message || 'Gagal membuat session';
        window.showErrorInfo(msg);
      } finally {
        this.creating = false;
      }
    },
    async fetchQR(agentId) {
      if (!agentId) return;
      try {
        const { data } = await window.http.post(`sessions/${agentId}/qr`);
        this.qrBase64 = data?.qr?.base64 || '';
        this.qrUpdatedAt = data?.qrUpdatedAt || '';
      } catch (err) {
        const msg = err?.response?.data?.error?.message || err?.message || 'QR belum tersedia';
        window.showErrorInfo(msg);
        this.stopQRRefresh();
      }
    },
    startQRRefresh(agentId) {
      this.stopQRRefresh();
      this.qrTimer = setInterval(() => {
        this.fetchQR(agentId);
      }, 20000); // WA QR biasanya kadaluarsa ~20s
    },
    stopQRRefresh() {
      if (this.qrTimer) {
        clearInterval(this.qrTimer);
        this.qrTimer = null;
      }
    },
    formatDate(val) {
      if (!val) return '';
      return moment(val).format('YYYY-MM-DD HH:mm:ss');
    },
  },
};
