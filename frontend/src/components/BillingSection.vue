<template>
  <div class="billing-section">

    <!-- ===== HEADER ===== -->
    <div class="section-header">
      <div class="header-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="header-svg">
          <rect x="2" y="7" width="20" height="14" rx="2"/>
          <path d="M16 7V5a2 2 0 00-2-2h-4a2 2 0 00-2 2v2"/>
          <line x1="12" y1="12" x2="12" y2="16"/>
          <line x1="10" y1="14" x2="14" y2="14"/>
        </svg>
      </div>
      <div>
        <h2 class="section-title">Facturación Electrónica</h2>
        <p class="section-sub">Configura la integración con FacturaAPI para emitir comprobantes electrónicos válidos ante SUNAT.</p>
      </div>
    </div>

    <!-- ===== LOADING STATE ===== -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Cargando configuración...</p>
    </div>

    <div v-else class="content-wrapper">
      <div class="tab-panel animate-in">
        <!-- Sync Logs Console -->
        <transition name="fade-slide">
          <div v-if="saveLogs.length > 0 || apiError" class="sync-console" :class="apiError ? 'sync-console--error' : 'sync-console--info'">
            <div class="sync-console-header">
              <div class="sync-console-title">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:15px;height:15px;">
                  <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
                </svg>
                Consola de Sincronización con FacturaAPI
              </div>
              <button type="button" @click="clearLogs" class="sync-clear">Limpiar</button>
            </div>
            <ul class="sync-log-list">
              <li v-for="(log, idx) in saveLogs" :key="idx" class="sync-log-item">
                <span class="sync-log-bullet">›</span> {{ log }}
              </li>
            </ul>
            <div v-if="apiError" class="sync-error-detail">
              <strong>⚠ Error:</strong> {{ apiError }}
            </div>
          </div>
        </transition>

        <!-- Status Bar -->
        <div class="status-bar">
          <div :class="['status-indicator', config.estado === 'active' ? 'status-indicator--on' : 'status-indicator--off']">
            <span class="status-dot"></span>
            {{ config.estado === 'active' ? 'Servicio Activo' : 'Servicio Desactivado' }}
          </div>
          <div v-if="config.tenant_uuid" class="tenant-uuid-chip">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:11px;height:11px;"><circle cx="12" cy="12" r="3"/><path d="M19.07 4.93a10 10 0 010 14.14M4.93 4.93a10 10 0 000 14.14"/></svg>
            UUID: {{ config.tenant_uuid.substring(0, 16) }}…
          </div>
        </div>

        <form @submit.prevent="saveConfig" class="biz-form">
          <!-- SOL Credentials -->
          <div class="form-section-label">Credenciales SOL (SUNAT)</div>
          <div class="form-grid form-grid--2">
            <div class="field">
              <label class="field-label">Usuario SOL <span class="required">*</span></label>
              <input
                v-model="config.sol_user"
                type="text"
                placeholder="Ej. MODDATOS"
                class="field-input field-input--mono"
              />
              <p class="field-hint">Usuario secundario SOL con permisos de emisión</p>
            </div>
            <div class="field">
              <label class="field-label">Clave SOL <span class="required">*</span></label>
              <div class="input-password-wrap">
                <input
                  v-model="config.sol_pass"
                  :type="showSolPass ? 'text' : 'password'"
                  placeholder="Clave SOL..."
                  class="field-input field-input--mono"
                />
                <button type="button" @click="showSolPass = !showSolPass" class="input-eye-btn">
                  <svg v-if="!showSolPass" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
                </button>
              </div>
            </div>
          </div>

          <!-- Certificate -->
          <div class="form-section-label" style="margin-top:20px;">Certificado Digital</div>
          <div class="field">
            <label class="field-label">Certificado .p12 / .pfx</label>
            <div class="file-upload-zone" @click="triggerCertInput" :class="config.certificado_base64 ? 'file-upload-zone--loaded' : ''">
              <input
                ref="certInput"
                type="file"
                accept=".p12,.pfx"
                @change="handleCertificateUpload"
                class="hidden-file"
              />
              <div v-if="!config.certificado_base64">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="file-upload-icon">
                  <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
                </svg>
                <p class="file-upload-text">Haz clic o arrastra tu certificado digital aquí</p>
                <p class="file-upload-hint">Formatos soportados: .p12, .pfx</p>
              </div>
              <div v-else class="file-loaded-state">
                <div class="file-loaded-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="width:20px;height:20px;"><polyline points="20 6 9 17 4 12"/></svg>
                </div>
                <div>
                  <p class="file-loaded-title">Certificado digital cargado</p>
                  <p class="file-loaded-sub">Haz clic para reemplazarlo</p>
                </div>
                <button type="button" @click.stop="config.certificado_base64 = ''" class="file-remove-btn">Remover</button>
              </div>
            </div>
            <p class="field-hint">Tu certificado digital emitido por SUNAT o un proveedor autorizado.</p>
          </div>

          <!-- Certificate Password -->
          <div class="field" style="margin-top:16px;">
            <label class="field-label">Contraseña del Certificado</label>
            <div class="input-password-wrap">
              <input
                v-model="config.certificado_password"
                :type="showCertPass ? 'text' : 'password'"
                placeholder="Contraseña del certificado..."
                class="field-input field-input--mono"
              />
              <button type="button" @click="showCertPass = !showCertPass" class="input-eye-btn">
                <svg v-if="!showCertPass" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
              </button>
            </div>
            <p class="field-hint">La clave de importación necesaria para desencriptar el certificado digital (.p12 / .pfx).</p>
          </div>

          <!-- GRE Credentials -->
          <div class="form-section-label" style="margin-top:20px;">Credenciales Guía de Remisión (API GRE)</div>
          <div class="form-grid form-grid--2">
            <div class="field">
              <label class="field-label">Client ID</label>
              <input
                v-model="config.client_id"
                type="text"
                placeholder="Client ID..."
                class="field-input field-input--mono"
              />
            </div>
            <div class="field">
              <label class="field-label">Client Secret</label>
              <div class="input-password-wrap">
                <input
                  v-model="config.client_secret"
                  :type="showClientSecret ? 'text' : 'password'"
                  placeholder="Client Secret..."
                  class="field-input field-input--mono"
                />
                <button type="button" @click="showClientSecret = !showClientSecret" class="input-eye-btn">
                  <svg v-if="!showClientSecret" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;"><path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
                </button>
              </div>
            </div>
          </div>
          <p class="field-hint" style="margin-top:4px;">Credenciales del portal SOL de SUNAT para emitir Guías de Remisión Electrónicas.</p>

          <!-- API Settings -->
          <div class="form-section-label" style="margin-top:20px;">Configuración API & Entorno</div>
          <div class="form-grid form-grid--2">
            <div class="field">
              <label class="field-label">Tenant UUID (FacturaAPI)</label>
              <input
                v-model="config.tenant_uuid"
                type="text"
                placeholder="Se genera automáticamente al guardar..."
                class="field-input field-input--mono"
              />
              <p class="field-hint">Identificador único en FacturaAPI. Se autogenera si está vacío y existe una API Key global.</p>
            </div>
            <div class="field">
              <label class="field-label">Modo de Operación</label>
              <div class="mode-toggle">
                <button
                  type="button"
                  @click="config.modo = 'dev'"
                  :class="['mode-btn', config.modo === 'dev' ? 'mode-btn--dev' : '']"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px;"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                  Desarrollo / Pruebas
                </button>
                <button
                  type="button"
                  @click="config.modo = 'prod'"
                  :class="['mode-btn', config.modo === 'prod' ? 'mode-btn--prod' : '']"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px;"><polyline points="20 6 9 17 4 12"/></svg>
                  Producción
                </button>
              </div>
            </div>
          </div>

          <!-- Estado -->
          <div class="field" style="margin-top:16px;">
            <label class="field-label">Estado del Servicio de Facturación</label>
            <div class="toggle-row">
              <button
                type="button"
                @click="config.estado = 'active'"
                :class="['estado-btn', config.estado === 'active' ? 'estado-btn--active' : '']"
              >Activo</button>
              <button
                type="button"
                @click="config.estado = 'inactive'"
                :class="['estado-btn', config.estado === 'inactive' ? 'estado-btn--inactive' : '']"
              >Inactivo (No emitir)</button>
            </div>
          </div>

          <!-- Overrides Avanzados -->
          <details class="advanced-details">
            <summary class="advanced-summary">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px;"><circle cx="12" cy="12" r="3"/><path d="M19.07 4.93A10 10 0 1 0 4.93 19.07M12 2v2M12 20v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M2 12h2M20 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
              Personalizar API (Overrides opcionales)
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="chevron-icon"><polyline points="6 9 12 15 18 9"/></svg>
            </summary>
            <div class="advanced-body">
              <div class="info-box">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;flex-shrink:0;"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="8"/><line x1="12" y1="12" x2="12" y2="16"/></svg>
                Por defecto, el sistema usa la URL y API Key configuradas globalmente por el administrador del SaaS. Usa los campos de abajo solo si necesitas sobrescribirlos para esta empresa.
              </div>
              <div class="form-grid form-grid--2" style="margin-top:12px;">
                <div class="field">
                  <label class="field-label">URL API Personalizada</label>
                  <input v-model="config.api_url" type="url" placeholder="Usar URL global por defecto..." class="field-input field-input--mono" />
                  <p class="field-hint">Ejemplo: https://apifacturacion.sehuacho.com</p>
                </div>
                <div class="field">
                  <label class="field-label">API Key Personalizada</label>
                  <input v-model="config.api_key" type="password" placeholder="Usar API Key global por defecto..." class="field-input field-input--mono" />
                </div>
              </div>
            </div>
          </details>

          <!-- Form Actions -->
          <div class="form-actions" style="margin-top:24px;">
            <div class="action-info">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;flex-shrink:0;"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
              Al guardar, los datos se sincronizarán automáticamente con FacturaAPI (registro, certificado, credenciales).
            </div>
            <button type="submit" :disabled="saving" class="btn-primary">
              <div v-if="saving" class="btn-spinner"></div>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="width:15px;height:15px;"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
              {{ saving ? 'Sincronizando...' : 'Guardar & Sincronizar con FacturaAPI' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- ===== HELP CARD ===== -->
    <div class="help-card">
      <div class="help-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" style="width:22px;height:22px;"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 015.83 1c0 2-3 3-3 3"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
      </div>
      <div class="help-body">
        <h4 class="help-title">¿Necesitas ayuda con la integración?</h4>
        <p class="help-text">Consulta la documentación de FacturaAPI para obtener tus credenciales SOL, configurar el certificado digital y registrar tu empresa.</p>
      </div>
      <div class="help-links">
        <a href="https://apifacturacion.sehuacho.com" target="_blank" class="help-link">
          Ver API <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:11px;height:11px;"><path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
        </a>
        <a href="https://ww1.sunat.gob.pe" target="_blank" class="help-link help-link--secondary">
          Portal SOL <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:11px;height:11px;"><path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
        </a>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import axios from 'axios'

const loading = ref(true)
const saving = ref(false)
const showSolPass = ref(false)
const showClientSecret = ref(false)
const showCertPass = ref(false)
const certInput = ref<HTMLInputElement | null>(null)

const config = reactive({
  api_url: '',
  api_key: '',
  tenant_uuid: '',
  modo: 'dev',
  estado: 'active',
  sol_user: '',
  sol_pass: '',
  certificado_base64: '',
  certificado_password: '',
  client_id: '',
  client_secret: '',
  logo_base64: ''
})

const saveLogs = ref<string[]>([])
const apiError = ref<string>('')

const clearLogs = () => {
  saveLogs.value = []
  apiError.value = ''
}

onMounted(async () => {
  await loadConfig()
})

const triggerCertInput = () => certInput.value?.click()

const handleCertificateUpload = (event: any) => {
  const file = event.target.files[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e: any) => {
    const base64Data = e.target.result.split(',')[1]
    config.certificado_base64 = base64Data
  }
  reader.readAsDataURL(file)
}

async function loadConfig() {
  loading.value = true
  try {
    const res = await axios.get('/billing/config')
    if (res.data.success && res.data.data) {
      Object.assign(config, res.data.data)
    }
  } catch (err) {
    console.error('Error loading billing config', err)
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  clearLogs()
  try {
    const res = await axios.post('/billing/config', config)
    if (res.data.success) {
      if (res.data.logs) saveLogs.value = res.data.logs
      if (res.data.api_error) {
        apiError.value = res.data.api_error
      }
      await loadConfig()
    }
  } catch (err: any) {
    apiError.value = err.response?.data?.error || 'Error de red o de servidor al intentar guardar.'
    saveLogs.value = ['Error al conectar con el servidor.']
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
/* Copied exact scoped styles from original BillingSection.vue for design consistency */
.billing-section {
  max-width: 960px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.section-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 24px;
  background: linear-gradient(135deg, #f8faff 0%, #eef2ff 100%);
  border: 1px solid #e0e7ff;
  border-radius: 16px;
}
.header-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(99,102,241,0.25);
}
.header-svg {
  width: 24px;
  height: 24px;
  stroke: white;
}
.section-title {
  font-size: 17px;
  font-weight: 800;
  color: #1e1b4b;
  margin: 0 0 4px;
  letter-spacing: -0.3px;
}
.section-sub {
  font-size: 12.5px;
  color: #6366f1;
  margin: 0;
  line-height: 1.5;
}
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 60px;
  color: #94a3b8;
  font-size: 13px;
}
.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e0e7ff;
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.content-wrapper {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  overflow: hidden;
}
.tab-panel {
  padding: 28px 28px 32px;
}
.animate-in {
  animation: fadeSlideIn 0.25s ease-out;
}
@keyframes fadeSlideIn {
  from { opacity: 0; transform: translateY(6px); }
  to   { opacity: 1; transform: translateY(0); }
}
.biz-form { display: flex; flex-direction: column; }
.form-section-label {
  font-size: 10.5px;
  font-weight: 800;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.form-section-label::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #f1f5f9;
}
.form-grid { display: grid; gap: 14px; }
.form-grid--2 { grid-template-columns: repeat(2, 1fr); }
@media (max-width: 680px) {
  .form-grid--2 { grid-template-columns: 1fr; }
}
.field { display: flex; flex-direction: column; gap: 5px; }
.field-label { font-size: 11.5px; font-weight: 700; color: #374151; }
.required { color: #ef4444; }
.field-input {
  width: 100%;
  padding: 9px 12px;
  border: 1.5px solid #e2e8f0;
  border-radius: 9px;
  font-size: 13px;
  color: #1e293b;
  background: #f8fafc;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s, background 0.2s;
  box-sizing: border-box;
}
.field-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.12);
  background: white;
}
.field-input::placeholder { color: #94a3b8; }
.field-input--mono { font-family: 'JetBrains Mono', 'Fira Mono', monospace; font-size: 12.5px; }
.field-hint { font-size: 10.5px; color: #94a3b8; line-height: 1.4; }
.input-password-wrap { position: relative; }
.input-password-wrap .field-input { padding-right: 36px; }
.input-eye-btn {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  color: #94a3b8;
  padding: 2px;
  display: flex;
  align-items: center;
}
.input-eye-btn:hover { color: #6366f1; }
.file-upload-zone {
  border: 2px dashed #cbd5e1;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  text-align: center;
  background: #f8fafc;
  transition: all 0.2s;
}
.file-upload-zone:hover, .file-upload-zone--loaded {
  border-color: #6366f1;
  background: #f5f3ff;
}
.file-upload-icon { width: 32px; height: 32px; stroke: #94a3b8; margin: 0 auto 8px; display: block; }
.file-upload-text { font-size: 13px; font-weight: 600; color: #475569; margin: 0 0 4px; }
.file-upload-hint { font-size: 11px; color: #94a3b8; margin: 0; }
.file-loaded-state {
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: space-between;
}
.file-loaded-icon {
  width: 36px;
  height: 36px;
  background: #dcfce7;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #16a34a;
  flex-shrink: 0;
}
.file-loaded-title { font-size: 13px; font-weight: 700; color: #16a34a; margin: 0 0 2px; }
.file-loaded-sub { font-size: 11px; color: #94a3b8; margin: 0; }
.file-remove-btn {
  font-size: 11px;
  font-weight: 700;
  color: #ef4444;
  background: #fee2e2;
  border: none;
  border-radius: 6px;
  padding: 4px 10px;
  cursor: pointer;
  flex-shrink: 0;
}
.file-remove-btn:hover { background: #fecaca; }
.mode-toggle { display: flex; gap: 6px; }
.mode-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 9px 12px;
  border: 1.5px solid #e2e8f0;
  border-radius: 99px;
  font-size: 11.5px;
  font-weight: 700;
  background: white;
  cursor: pointer;
  color: #64748b;
  transition: all 0.2s;
}
.mode-btn:hover { border-color: #94a3b8; }
.mode-btn--dev { background: #fffbeb; border-color: #fbbf24; color: #92400e; }
.mode-btn--prod { background: #f0fdf4; border-color: #22c55e; color: #15803d; }
.toggle-row { display: flex; gap: 6px; }
.estado-btn {
  flex: 1;
  padding: 9px;
  border: 1.5px solid #e2e8f0;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  background: white;
  color: #64748b;
  transition: all 0.2s;
}
.estado-btn--active { background: #f0fdf4; border-color: #22c55e; color: #15803d; }
.estado-btn--inactive { background: #fef2f2; border-color: #fca5a5; color: #b91c1c; }
.advanced-details {
  margin-top: 20px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  overflow: hidden;
}
.advanced-summary {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 12px 14px;
  font-size: 11.5px;
  font-weight: 700;
  color: #475569;
  cursor: pointer;
  list-style: none;
  background: #f8fafc;
  user-select: none;
}
.advanced-summary::-webkit-details-marker { display: none; }
.chevron-icon {
  width: 13px;
  height: 13px;
  margin-left: auto;
  transition: transform 0.2s;
}
details[open] .chevron-icon { transform: rotate(180deg); }
.advanced-body {
  padding: 16px 14px;
  border-top: 1px solid #e2e8f0;
}
.info-box {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  font-size: 11.5px;
  color: #1d4ed8;
  line-height: 1.5;
}
.sync-console {
  border-radius: 10px;
  margin-bottom: 20px;
  overflow: hidden;
  border: 1px solid;
}
.sync-console--info { border-color: #bfdbfe; background: #eff6ff; }
.sync-console--error { border-color: #fecaca; background: #fef2f2; }
.sync-console-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid rgba(0,0,0,0.06);
}
.sync-console-title {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 11.5px;
  font-weight: 800;
  color: #1d4ed8;
}
.sync-console--error .sync-console-title { color: #b91c1c; }
.sync-clear {
  font-size: 10.5px;
  font-weight: 700;
  color: #94a3b8;
  background: none;
  border: none;
  cursor: pointer;
  text-decoration: underline;
}
.sync-clear:hover { color: #475569; }
.sync-log-list { margin: 0; padding: 10px 14px; list-style: none; display: flex; flex-direction: column; gap: 5px; }
.sync-log-item { display: flex; align-items: flex-start; gap: 6px; font-size: 11.5px; color: #1e293b; font-family: monospace; line-height: 1.5; }
.sync-log-bullet { color: #6366f1; font-weight: 900; flex-shrink: 0; }
.sync-error-detail {
  padding: 10px 14px;
  border-top: 1px solid #fecaca;
  background: #fef2f2;
  font-size: 11px;
  font-family: monospace;
  color: #b91c1c;
  white-space: pre-wrap;
  word-break: break-all;
}
.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.status-indicator {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
  font-weight: 700;
}
.status-indicator--on { color: #15803d; }
.status-indicator--off { color: #94a3b8; }
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
}
.status-indicator--on .status-dot { animation: pulse 1.5s ease-in-out infinite; }
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
.tenant-uuid-chip {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 10.5px;
  font-weight: 600;
  color: #64748b;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 99px;
  padding: 3px 10px;
  font-family: monospace;
}
.form-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 20px;
  margin-top: 8px;
  border-top: 1px solid #f1f5f9;
  gap: 12px;
}
.action-info {
  display: flex;
  align-items: flex-start;
  gap: 7px;
  font-size: 11px;
  color: #94a3b8;
  line-height: 1.5;
  max-width: 340px;
}
.btn-primary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 22px;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(99,102,241,0.3);
  white-space: nowrap;
  flex-shrink: 0;
}
.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #4f46e5, #4338ca);
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(99,102,241,0.4);
}
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }
.btn-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}
.help-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 22px;
  background: linear-gradient(135deg, #1e1b4b, #312e81);
  border-radius: 14px;
  flex-wrap: wrap;
}
.help-icon {
  width: 40px;
  height: 40px;
  background: rgba(255,255,255,0.1);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #c7d2fe;
  flex-shrink: 0;
}
.help-body { flex: 1; min-width: 180px; }
.help-title { font-size: 14px; font-weight: 700; color: white; margin: 0 0 4px; }
.help-text { font-size: 11.5px; color: #a5b4fc; margin: 0; line-height: 1.5; }
.help-links { display: flex; gap: 8px; flex-shrink: 0; }
.help-link {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 8px 16px;
  background: rgba(99,102,241,0.5);
  border: 1px solid rgba(165,180,252,0.3);
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;
  color: white;
  text-decoration: none;
  transition: all 0.2s;
}
.help-link:hover { background: rgba(99,102,241,0.8); }
.help-link--secondary {
  background: rgba(255,255,255,0.08);
  color: #c7d2fe;
}
.help-link--secondary:hover { background: rgba(255,255,255,0.15); }
.fade-slide-enter-active, .fade-slide-leave-active { transition: all 0.3s ease; }
.fade-slide-enter-from { opacity: 0; transform: translateY(-8px); }
.fade-slide-leave-to   { opacity: 0; transform: translateY(-4px); }
</style>
