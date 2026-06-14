<template>
  <div class="billing-section">

    <!-- ===== HEADER ===== -->
    <div class="section-header">
      <div class="header-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="header-svg">
          <rect x="4" y="2" width="16" height="20" rx="2" ry="2"/>
          <line x1="9" y1="22" x2="9" y2="16"/>
          <line x1="15" y1="22" x2="15" y2="16"/>
          <line x1="9" y1="16" x2="15" y2="16"/>
          <path d="M9 6h6"/>
          <path d="M9 10h6"/>
        </svg>
      </div>
      <div>
        <h2 class="section-title">Configuración del Negocio</h2>
        <p class="section-sub">Gestiona la información comercial, logotipo y datos de contacto de tu empresa.</p>
      </div>
    </div>

    <!-- ===== LOADING STATE ===== -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Cargando configuración...</p>
    </div>

    <div v-else class="content-wrapper">
      <div class="tab-panel animate-in">
        <!-- === TOP: Logo + Nombre Comercial === -->
        <div class="business-hero">
          <!-- Logo Area -->
          <div class="logo-area">
            <div class="logo-frame" @click="triggerLogoInput">
              <img
                v-if="business.logo_base64"
                :src="'data:image/png;base64,' + business.logo_base64"
                class="logo-img"
                alt="Logo del negocio"
              />
              <div v-else class="logo-placeholder">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="logo-placeholder-icon">
                  <rect x="3" y="3" width="18" height="18" rx="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <polyline points="21 15 16 10 5 21"/>
                </svg>
                <span>Click para subir logo</span>
              </div>
              <div class="logo-overlay">
                <svg viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" class="logo-overlay-icon">
                  <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
                  <polyline points="17 8 12 3 7 8"/>
                  <line x1="12" y1="3" x2="12" y2="15"/>
                </svg>
                <span>Cambiar logo</span>
              </div>
            </div>
            <input
              ref="logoInput"
              type="file"
              accept="image/png,image/jpeg,image/webp,image/gif"
              @change="handleBusinessLogoUpload"
              class="hidden-file"
            />
            <div v-if="business.logo_base64" class="logo-actions">
              <button type="button" @click="removeLogo" class="btn-remove-logo">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:12px;height:12px;">
                  <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a1 1 0 011-1h4a1 1 0 011 1v2"/>
                </svg>
                Remover
              </button>
            </div>
            <div v-if="logoCompressing" class="logo-compressing">
              <svg class="spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px;"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
              Optimizando imagen...
            </div>
            <p class="logo-hint">
              PNG / JPG / WebP &bull; Máx. 800×800px<br/>
              <span v-if="logoFileSize" :class="logoFileSize > 200 ? 'hint-warn' : 'hint-ok'">
                {{ logoFileSize }}KB comprimido
              </span>
              <span v-else>Se optimiza automáticamente</span>
            </p>
          </div>

          <!-- Hero Info -->
          <div class="business-hero-info">
            <div class="ruc-badge">
              <span class="ruc-label">RUC</span>
              <span class="ruc-value">{{ business.ruc || '—' }}</span>
              <span class="ruc-lock-hint">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:11px;height:11px;vertical-align:middle;">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/>
                </svg>
                No editable
              </span>
            </div>
            <h3 class="hero-razon">{{ business.razon_social || 'Sin Razón Social' }}</h3>
            <p class="hero-comercial">{{ business.nombre_comercial || 'Sin nombre comercial' }}</p>
            <div class="hero-contact-row">
              <span v-if="business.telefono" class="hero-contact-chip">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:12px;height:12px;"><path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07A19.5 19.5 0 013.07 9.81 19.79 19.79 0 01.15 1.22 2 2 0 012.13 0h3a2 2 0 012 1.72c.127.96.361 1.903.7 2.81a2 2 0 01-.45 2.11L6.91 7.09a16 16 0 006 6l.56-.56a2 2 0 012.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0122 14.92v2z"/></svg>
                {{ business.telefono }}
              </span>
              <span v-if="business.email" class="hero-contact-chip">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:12px;height:12px;"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
                {{ business.email }}
              </span>
            </div>
          </div>
        </div>

        <!-- === FORM: Datos del negocio === -->
        <form @submit.prevent="saveBusinessProfile" class="biz-form">
          <div class="form-section-label">Información Legal</div>
          <div class="form-grid form-grid--3">
            <div class="field">
              <label class="field-label">Razón Social <span class="required">*</span></label>
              <input
                v-model="business.razon_social"
                type="text"
                placeholder="Nombre legal completo..."
                class="field-input"
                required
              />
              <p class="field-hint">Denominación registrada ante SUNAT</p>
            </div>
            <div class="field">
              <label class="field-label">Nombre Comercial</label>
              <input
                v-model="business.nombre_comercial"
                type="text"
                placeholder="Nombre de fantasía..."
                class="field-input"
              />
              <p class="field-hint">Aparece en documentos y recibos</p>
            </div>
            <div class="field">
              <label class="field-label">RUC (11 dígitos)</label>
              <input
                :value="business.ruc"
                type="text"
                class="field-input field-input--locked"
                disabled
              />
              <p class="field-hint">Fijado en la creación de la cuenta</p>
            </div>
          </div>

          <div class="form-section-label" style="margin-top:24px;">Contacto & Ubicación</div>
          <div class="form-grid form-grid--3">
            <div class="field">
              <label class="field-label">Dirección Fiscal</label>
              <input
                v-model="business.direccion"
                type="text"
                placeholder="Av. Ejemplo 123, Lima..."
                class="field-input"
              />
              <p class="field-hint">Dirección que aparece en los comprobantes</p>
            </div>
            <div class="field">
              <label class="field-label">Teléfono</label>
              <input
                v-model="business.telefono"
                type="text"
                placeholder="01-234-5678..."
                class="field-input"
              />
            </div>
            <div class="field">
              <label class="field-label">Email de Contacto</label>
              <input
                v-model="business.email"
                type="email"
                placeholder="contacto@empresa.com..."
                class="field-input"
              />
            </div>
          </div>

          <!-- ===== Alerta de éxito/error del negocio ===== -->
          <transition name="fade-slide">
            <div v-if="businessAlert.msg" :class="['alert-banner', `alert-banner--${businessAlert.type}`]">
              <svg v-if="businessAlert.type === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="alert-icon"><polyline points="20 6 9 17 4 12"/></svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="alert-icon"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
              <span>{{ businessAlert.msg }}</span>
              <button type="button" @click="businessAlert.msg = ''" class="alert-close">×</button>
            </div>
          </transition>

          <div class="form-actions">
            <div class="action-info">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:14px;height:14px;flex-shrink:0;">
                <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="8"/><line x1="12" y1="12" x2="12" y2="16"/>
              </svg>
              Los datos de razón social y dirección se usarán en los comprobantes electrónicos.
            </div>
            <button type="submit" :disabled="savingBusiness" class="btn-primary">
              <div v-if="savingBusiness" class="btn-spinner"></div>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="width:15px;height:15px;"><path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
              {{ savingBusiness ? 'Guardando...' : 'Guardar Datos del Negocio' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import axios from 'axios'

const loading = ref(true)
const savingBusiness = ref(false)
const logoInput = ref<HTMLInputElement | null>(null)
const logoCompressing = ref(false)
const logoFileSize = ref<number | null>(null)
const businessAlert = reactive({ msg: '', type: 'success' as 'success' | 'error' })

const business = reactive({
  ruc: '',
  razon_social: '',
  nombre_comercial: '',
  direccion: '',
  telefono: '',
  email: '',
  logo_base64: ''
})

onMounted(async () => {
  await loadBusinessProfile()
})

const triggerLogoInput = () => logoInput.value?.click()

const removeLogo = () => {
  business.logo_base64 = ''
  logoFileSize.value = null
}

/**
 * Compress image using an offscreen canvas before storing as base64.
 * - Resizes to max 800x800 preserving aspect ratio
 * - Tries JPEG @ 80% quality first; falls back to PNG if smaller
 * - Returns base64 string WITHOUT the data:...;base64, prefix
 */
async function compressImage(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    const objectUrl = URL.createObjectURL(file)
    img.onload = () => {
      URL.revokeObjectURL(objectUrl)
      const MAX_DIM = 400
      let { width, height } = img
      if (width > MAX_DIM || height > MAX_DIM) {
        if (width > height) {
          height = Math.round((height * MAX_DIM) / width)
          width = MAX_DIM
        } else {
          width = Math.round((width * MAX_DIM) / height)
          height = MAX_DIM
        }
      }
      const canvas = document.createElement('canvas')
      canvas.width = width
      canvas.height = height
      const ctx = canvas.getContext('2d')!
      // White background for JPEG (avoids black on transparency)
      ctx.fillStyle = '#FFFFFF'
      ctx.fillRect(0, 0, width, height)
      ctx.drawImage(img, 0, 0, width, height)

      // Try JPEG first
      const jpegData = canvas.toDataURL('image/jpeg', 0.75)
      // Try PNG (better for logos with few colors)
      const pngData = canvas.toDataURL('image/png')

      // Pick the smaller one
      const chosen = jpegData.length <= pngData.length ? jpegData : pngData
      // Strip the data URI prefix (e.g. "data:image/jpeg;base64,")
      resolve(chosen.split(',')[1])
    }
    img.onerror = () => {
      URL.revokeObjectURL(objectUrl)
      reject(new Error('No se pudo cargar la imagen'))
    }
    img.src = objectUrl
  })
}

const handleBusinessLogoUpload = async (event: any) => {
  const file: File = event.target.files[0]
  if (!file) return

  // Validate file type
  if (!file.type.startsWith('image/')) {
    businessAlert.msg = 'Solo se permiten archivos de imagen (PNG, JPG, WebP).'
    businessAlert.type = 'error'
    return
  }

  // Validate raw size (hard limit: 20MB to prevent browser freeze)
  if (file.size > 20 * 1024 * 1024) {
    businessAlert.msg = 'La imagen es demasiado grande (máx. 20MB). Por favor, usa una imagen más pequeña.'
    businessAlert.type = 'error'
    return
  }

  logoCompressing.value = true
  logoFileSize.value = null
  try {
    const compressed = await compressImage(file)
    business.logo_base64 = compressed
    // Estimate KB: base64 length × 0.75
    logoFileSize.value = Math.round((compressed.length * 0.75) / 1024)
  } catch (err) {
    businessAlert.msg = 'Error al procesar la imagen. Intenta con otro archivo.'
    businessAlert.type = 'error'
  } finally {
    logoCompressing.value = false
    // Reset input so same file can be re-selected
    if (logoInput.value) logoInput.value.value = ''
  }
}

async function loadBusinessProfile() {
  loading.value = true
  try {
    const res = await axios.get('/companies/me')
    if (res.data.success && res.data.data) {
      Object.assign(business, res.data.data)
    }
  } catch (err) {
    console.error('Error loading business profile', err)
  } finally {
    loading.value = false
  }
}

async function saveBusinessProfile() {
  savingBusiness.value = true
  businessAlert.msg = ''
  try {
    const res = await axios.put('/companies/me', {
      razon_social: business.razon_social,
      nombre_comercial: business.nombre_comercial,
      direccion: business.direccion,
      telefono: business.telefono,
      email: business.email,
      logo_base64: business.logo_base64
    })
    if (res.data.success) {
      businessAlert.msg = '✓ Datos del negocio guardados correctamente.'
      businessAlert.type = 'success'
      await loadBusinessProfile()
      setTimeout(() => { businessAlert.msg = '' }, 4000)
    }
  } catch (err: any) {
    businessAlert.msg = err.response?.data?.error || 'Error al guardar datos del negocio.'
    businessAlert.type = 'error'
  } finally {
    savingBusiness.value = false
  }
}
</script>

<style scoped>
/* Copied exact scoped styles from BillingSection.vue for design consistency */
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
.business-hero {
  display: flex;
  gap: 28px;
  align-items: flex-start;
  margin-bottom: 28px;
  padding-bottom: 28px;
  border-bottom: 1px solid #f1f5f9;
}
.logo-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.logo-frame {
  position: relative;
  width: 100px;
  height: 100px;
  border: 2px dashed #cbd5e1;
  border-radius: 14px;
  cursor: pointer;
  overflow: hidden;
  background: #f8fafc;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.logo-frame:hover { border-color: #6366f1; box-shadow: 0 0 0 4px rgba(99,102,241,0.08); }
.logo-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  padding: 8px;
}
.logo-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #94a3b8;
}
.logo-placeholder-icon { width: 28px; height: 28px; }
.logo-placeholder span { font-size: 9.5px; font-weight: 600; text-align: center; line-height: 1.3; }
.logo-overlay {
  position: absolute;
  inset: 0;
  background: rgba(99,102,241,0.88);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  color: white;
  font-size: 10px;
  font-weight: 700;
  opacity: 0;
  transition: opacity 0.2s;
}
.logo-frame:hover .logo-overlay { opacity: 1; }
.logo-overlay-icon { width: 20px; height: 20px; }
.hidden-file { display: none; }
.logo-actions { display: flex; gap: 8px; }
.btn-remove-logo {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  font-weight: 700;
  color: #ef4444;
  background: none;
  border: none;
  cursor: pointer;
  padding: 2px 4px;
}
.btn-remove-logo:hover { text-decoration: underline; }
.logo-hint { font-size: 9.5px; color: #94a3b8; text-align: center; max-width: 100px; line-height: 1.3; }
.business-hero-info { flex: 1; }
.ruc-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 4px 10px;
  margin-bottom: 10px;
}
.ruc-label { font-size: 10px; font-weight: 800; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.05em; }
.ruc-value { font-size: 13px; font-weight: 900; color: #1e293b; font-family: monospace; }
.ruc-lock-hint { font-size: 9.5px; color: #94a3b8; }
.hero-razon { font-size: 18px; font-weight: 800; color: #1e293b; margin: 0 0 4px; letter-spacing: -0.3px; }
.hero-comercial { font-size: 13px; color: #64748b; font-weight: 500; margin: 0 0 12px; }
.hero-contact-row { display: flex; flex-wrap: wrap; gap: 8px; }
.hero-contact-chip {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 11.5px;
  color: #475569;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 99px;
  padding: 3px 10px;
  font-weight: 500;
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
.form-grid--3 { grid-template-columns: repeat(3, 1fr); }
@media (max-width: 680px) {
  .form-grid--3 { grid-template-columns: 1fr; }
  .business-hero { flex-direction: column; align-items: center; }
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
.field-input--locked {
  background: #f1f5f9;
  color: #94a3b8;
  cursor: not-allowed;
  border-style: dashed;
}
.field-hint { font-size: 10.5px; color: #94a3b8; line-height: 1.4; }
.alert-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 10px;
  font-size: 12.5px;
  font-weight: 600;
  margin-top: 16px;
}
.alert-banner--success { background: #f0fdf4; border: 1px solid #bbf7d0; color: #15803d; }
.alert-banner--error   { background: #fef2f2; border: 1px solid #fecaca; color: #b91c1c; }
.alert-icon { width: 15px; height: 15px; flex-shrink: 0; }
.alert-close {
  margin-left: auto;
  font-size: 16px;
  font-weight: 700;
  background: none;
  border: none;
  cursor: pointer;
  color: inherit;
  opacity: 0.6;
  line-height: 1;
}
.alert-close:hover { opacity: 1; }
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
.fade-slide-enter-active, .fade-slide-leave-active { transition: all 0.3s ease; }
.fade-slide-enter-from { opacity: 0; transform: translateY(-8px); }
.fade-slide-leave-to   { opacity: 0; transform: translateY(-4px); }
</style>
