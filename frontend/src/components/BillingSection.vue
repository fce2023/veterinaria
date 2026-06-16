<template>
  <div class="billing-section animate-in">

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
        <p class="section-sub">Configura la integración con FacturaAPI, define tus series y correlativos, y administra todos tus comprobantes electrónicos ante SUNAT.</p>
      </div>
    </div>

    <!-- ===== TABS NAVIGATION ===== -->
    <div class="tabs-nav">
      <button 
        type="button"
        @click="activeTab = 'config'" 
        :class="['tab-nav-btn', activeTab === 'config' ? 'tab-nav-btn--active' : '']"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="tab-icon">
          <circle cx="12" cy="12" r="3"/>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
        </svg>
        Configuración de API
      </button>
      <button 
        type="button"
        @click="activeTab = 'documents'" 
        :class="['tab-nav-btn', activeTab === 'documents' ? 'tab-nav-btn--active' : '']"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="tab-icon">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
          <line x1="16" y1="13" x2="8" y2="13"/>
          <line x1="16" y1="17" x2="8" y2="17"/>
          <polyline points="10 9 9 9 8 9"/>
        </svg>
        Comprobantes Emitidos
      </button>
      <button 
        type="button"
        @click="activeTab = 'series'" 
        :class="['tab-nav-btn', activeTab === 'series' ? 'tab-nav-btn--active' : '']"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="tab-icon">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
          <line x1="9" y1="3" x2="9" y2="21"/>
          <line x1="15" y1="3" x2="15" y2="21"/>
          <line x1="3" y1="9" x2="21" y2="9"/>
          <line x1="3" y1="15" x2="21" y2="15"/>
        </svg>
        Series y Correlativos
      </button>
    </div>

    <!-- ===== LOADING STATE ===== -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Cargando datos del módulo...</p>
    </div>

    <div v-else class="content-wrapper">
      
      <!-- ================= TAB 1: CONFIG ================= -->
      <div v-if="activeTab === 'config'" class="tab-panel animate-in">
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
          <div class="form-grid form-grid--2" style="margin-top:20px;">
            <div class="field">
              <label class="field-label">Entorno de Emisión (SUNAT)</label>
              <select v-model="config.modo" class="field-input field-input--select">
                <option value="dev">Entorno de Pruebas (BETA - Sin Validez)</option>
                <option value="prod">Entorno de Producción (Legal - Oficial)</option>
              </select>
              <p class="field-hint" v-if="config.modo === 'prod'" style="color: #ef4444; font-weight: bold;">¡Atención! Documentos emitidos aquí tienen validez tributaria.</p>
              <p class="field-hint" v-else>Para realizar pruebas de homologación y desarrollo.</p>
            </div>
            <div class="field">
              <label class="field-label">Modo de Envío</label>
              <select v-model="config.emision_diferida" class="field-input field-input--select">
                <option :value="false">Instantáneo (Al cerrar la venta)</option>
                <option :value="true">Diferido (Al final de la jornada)</option>
              </select>
              <p class="field-hint" v-if="config.emision_diferida">Las ventas se guardarán como borradores. Deberás emitirlas manualmente al cerrar el día.</p>
              <p class="field-hint" v-else>Los comprobantes se envían a SUNAT inmediatamente tras cada venta.</p>
            </div>
          </div>

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
                required
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
                  required
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

          <!-- Estado del modulo -->
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

      <!-- ================= TAB 2: DOCUMENTS ================= -->
      <div v-if="activeTab === 'documents'" class="tab-panel animate-in">
        
        <!-- Stats Cards -->
        <div class="doc-stats-grid">
          <div class="doc-stat-card doc-stat-card--total">
            <div class="doc-stat-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:18px;height:18px"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
            </div>
            <div>
              <p class="doc-stat-label">Total Comprobantes</p>
              <p class="doc-stat-value">{{ docStats.grand_total || 0 }}</p>
            </div>
          </div>
          <div class="doc-stat-card doc-stat-card--factura">
            <div class="doc-stat-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:18px;height:18px"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg>
            </div>
            <div>
              <p class="doc-stat-label">Facturas</p>
              <p class="doc-stat-value">{{ getStatTotal('01') }}</p>
            </div>
          </div>
          <div class="doc-stat-card doc-stat-card--boleta">
            <div class="doc-stat-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:18px;height:18px"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
            </div>
            <div>
              <p class="doc-stat-label">Boletas</p>
              <p class="doc-stat-value">{{ getStatTotal('03') }}</p>
            </div>
          </div>
          <div class="doc-stat-card doc-stat-card--nc">
            <div class="doc-stat-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:18px;height:18px"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>
            </div>
            <div>
              <p class="doc-stat-label">Notas de Crédito</p>
              <p class="doc-stat-value">{{ getStatTotal('07') }}</p>
            </div>
          </div>
        </div>

        <!-- Filters Controls -->
        <div class="filter-controls-wrap">
          <div class="search-box">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="search-icon">
              <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
            </svg>
            <input 
              v-model="searchQuery" 
              type="text" 
              placeholder="Buscar por serie o número..." 
              @input="loadDocuments" 
              class="search-input"
            />
          </div>

          <button 
            v-if="documents.some(d => d.estado === 'draft')" 
            @click="batchEmitDrafts" 
            class="btn-primary" 
            style="background: #2563eb; padding: 0 20px; font-weight: 700; display: flex; align-items: center; gap: 8px;"
            :disabled="batchEmitting"
          >
            <i v-if="batchEmitting" class="ti ti-loader animate-spin"></i>
            <i v-else class="ti ti-send"></i>
            Emitir Pendientes (Cierre)
          </button>

          <div class="select-filter-group">
            <select v-model="filterType" @change="loadDocuments" class="filter-select">
              <option value="">Todos los Tipos</option>
              <option value="01">Facturas</option>
              <option value="03">Boletas</option>
              <option value="07">Notas de Crédito</option>
            </select>

            <select v-model="filterEstado" @change="loadDocuments" class="filter-select">
              <option value="">Todos los Estados</option>
              <option value="accepted">Aceptados (SUNAT)</option>
              <option value="pending">Pendientes</option>
              <option value="rejected">Rechazados</option>
              <option value="voided">Anulados</option>
            </select>
          </div>

          <div class="date-filter-group">
            <div class="date-field">
              <label class="date-label">Desde</label>
              <input v-model="filterFechaDesde" type="date" class="filter-select" @change="loadDocuments" />
            </div>
            <div class="date-field">
              <label class="date-label">Hasta</label>
              <input v-model="filterFechaHasta" type="date" class="filter-select" @change="loadDocuments" />
            </div>
            <button v-if="filterFechaDesde || filterFechaHasta" type="button" @click="clearDateFilter" class="clear-date-btn">✕ Limpiar fechas</button>
          </div>
        </div>

        <!-- Results Count -->
        <div v-if="!loadingDocs && documents.length > 0" class="results-count">
          <span>Mostrando <strong>{{ documents.length }}</strong> comprobante(s)</span>
        </div>

        <!-- Documents Table -->
        <div v-if="loadingDocs" class="table-loading">
          <div class="spinner"></div>
          <p>Cargando comprobantes...</p>
        </div>

        <div v-else-if="documents.length === 0" class="empty-table-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="empty-icon">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
          </svg>
          <p class="empty-title">No se encontraron comprobantes</p>
          <p class="empty-sub">Aún no se han emitido comprobantes electrónicos que coincidan con los filtros seleccionados.</p>
        </div>

        <div v-else class="table-responsive">
          <table class="billing-table">
            <thead>
              <tr>
                <th>Fecha Emisión</th>
                <th>Comprobante</th>
                <th>Cliente</th>
                <th>Monto Total</th>
                <th>Estado SUNAT</th>
                <th class="text-right">Descargas (FacturaAPI)</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="doc in documents" :key="doc.id">
                <td class="font-mono text-xs">{{ formatDate(doc.created_at) }}</td>
                <td>
                  <div class="doc-ident-wrap">
                    <span class="doc-type-pill">{{ formatDocType(doc.tipo_documento) }}</span>
                    <strong class="font-mono text-slate-800">{{ doc.serie }}-{{ doc.numero }}</strong>
                  </div>
                </td>
                <td>
                  <div v-if="doc.sale && doc.sale.customer" class="cust-info">
                    <p class="cust-name">{{ doc.sale.customer.nombre }}</p>
                    <p class="cust-doc text-slate-400 font-mono text-[10px]">{{ doc.sale.customer.tipo_documento }}: {{ doc.sale.customer.numero_documento }}</p>
                  </div>
                  <span v-else class="text-slate-400">-</span>
                </td>
                <td>
                  <strong v-if="doc.sale" class="text-slate-800 font-bold">S/. {{ doc.sale.total.toFixed(2) }}</strong>
                  <span v-else class="text-slate-400">-</span>
                </td>
                <td>
                  <span 
                    :class="['badge', 'badge--' + (doc.observaciones && doc.estado === 'accepted' ? 'warning' : doc.estado)]"
                    @click="showSunatMessage(doc)"
                    style="cursor: pointer;"
                    title="Click para ver detalle de SUNAT"
                  >
                    <span class="badge-dot"></span>
                    {{ (doc.observaciones && doc.estado === 'accepted' ? 'OBSERVADO' : doc.estado).toUpperCase() }}
                  </span>
                  <button
                    v-if="doc.estado === 'pending' || doc.estado === 'error' || doc.estado === 'rejected'"
                    type="button"
                    @click="syncDocumentStatus(doc.id)"
                    class="sync-status-inline-btn"
                    title="Actualizar estado desde SUNAT"
                    :disabled="syncingId === doc.id"
                  >
                    <i class="ti ti-refresh" :class="{ 'animate-spin': syncingId === doc.id }"></i>
                  </button>
                </td>

                <td class="text-right">
                  <div class="action-btn-group">
                    <button 
                      v-if="doc.estado !== 'accepted' && doc.estado !== 'voided'"
                      type="button"
                      @click="resendDocument(doc.id)"
                      class="action-btn action-btn--resend"
                      title="Intentar enviar nuevamente"
                      :disabled="resendingId === doc.id"
                    >
                      <i v-if="resendingId === doc.id" class="ti ti-loader animate-spin"></i>
                      <span v-else>Reenviar</span>
                    </button>
                    <button 
                      type="button"
                      @click="showSunatMessage(doc)"
                      class="action-btn action-btn--msg"
                      title="Ver mensajes de SUNAT/API"
                    >
                      Mensajes
                    </button>
                    <button 
                      v-if="doc.estado === 'draft'"
                      type="button"
                      @click="openEditDraftModal(doc)"
                      class="action-btn action-btn--edit"
                      title="Editar datos antes de emitir"
                    >
                      Editar
                    </button>
                    <button 
                      v-if="doc.estado === 'accepted' && doc.tipo_documento !== '03' && doc.tipo_documento !== 'NV'"
                      type="button"
                      @click="openVoidModal(doc)"
                      class="action-btn action-btn--void"
                      title="Solicitar Comunicación de Baja (Anulación Legal)"
                    >
                      Anular
                    </button>
                    <button 
                      v-if="(config.modo === 'dev' || config.modo === 'beta') && doc.estado !== 'accepted' && doc.estado !== 'voided'"
                      type="button"
                      @click="deleteDocumentTest(doc.id)"
                      class="action-btn action-btn--void"
                      title="Eliminar registro (Solo Modo Prueba)"
                    >
                      Eliminar
                    </button>
                    <button 
                      type="button"
                      @click="downloadFile(doc.document_uuid, 'pdf')" 
                      class="action-btn action-btn--pdf" 
                      title="Descargar PDF Representación Impresa"
                    >
                      PDF
                    </button>
                    <button 
                      type="button"
                      @click="downloadFile(doc.document_uuid, 'xml')" 
                      class="action-btn action-btn--xml" 
                      title="Descargar XML firmado por SUNAT"
                    >
                      XML
                    </button>
                    <button 
                      type="button"
                      @click="downloadFile(doc.document_uuid, 'cdr')" 
                      class="action-btn action-btn--cdr" 
                      title="Descargar CDR (Constancia de Recepción)"
                      :disabled="doc.estado !== 'accepted'"
                    >
                      CDR
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- SUNAT Message Modal -->
        <div v-if="showMsgModal" class="modal-overlay" @click="showMsgModal = false">
          <div class="modal-mini" @click.stop>
            <div class="modal-mini-header">
              <h3 class="modal-mini-title">Mensaje de SUNAT</h3>
              <button @click="showMsgModal = false" class="modal-close-btn">✕</button>
            </div>
            <div class="modal-mini-body">
              <div v-if="selectedDoc" class="msg-box" :class="'msg-box--' + (selectedDoc.observaciones && selectedDoc.estado === 'accepted' ? 'warning' : selectedDoc.estado)">
                <div class="msg-box-header">
                  <span class="badge" :class="'badge--' + (selectedDoc.observaciones && selectedDoc.estado === 'accepted' ? 'warning' : selectedDoc.estado)">
                    {{ (selectedDoc.observaciones && selectedDoc.estado === 'accepted' ? 'OBSERVADO' : selectedDoc.estado).toUpperCase() }}
                  </span>
                  <span class="font-mono text-[11px]">{{ selectedDoc.serie }}-{{ selectedDoc.numero }}</span>
                </div>
                
                <!-- Main Message Content -->
                <div v-if="selectedDoc.sunat_response || selectedDoc.facturacion_error" class="msg-content">
                  
                  <!-- Technical API Error -->
                  <div v-if="selectedDoc.facturacion_error" class="technical-error">
                    <p class="font-bold text-red-600 mb-1">Error de Comunicación/Validación:</p>
                    <pre class="msg-text-raw">{{ cleanErrorMessage(selectedDoc.facturacion_error) }}</pre>
                  </div>

                  <!-- SUNAT Notes/Observations -->
                  <div v-if="selectedDoc.sunat_response" class="sunat-notes">
                    <p v-if="selectedDoc.facturacion_error" class="font-bold text-slate-700 mt-2 mb-1">Respuesta de SUNAT:</p>
                    <div v-if="selectedDoc.sunat_response.includes('\n')" class="msg-list">
                      <div v-for="(line, idx) in selectedDoc.sunat_response.split('\n')" :key="idx" class="msg-item">
                        <span class="msg-item-bullet"></span>
                        <p class="msg-text">{{ line }}</p>
                      </div>
                    </div>
                    <p v-else class="msg-text">{{ selectedDoc.sunat_response }}</p>
                  </div>

                </div>
                <p v-else class="msg-text text-slate-400 italic">No hay mensajes detallados para este comprobante.</p>
              </div>
            </div>
            <div class="modal-mini-footer">
              <button 
                v-if="config.modo === 'dev' || config.modo === 'beta'"
                @click="deleteDocumentTest(selectedDoc.id)"
                class="btn-danger-outline btn-primary--sm mr-auto"
              >
                Eliminar (Prueba)
              </button>
              <button 
                v-if="selectedDoc.estado === 'accepted' && selectedDoc.tipo_documento !== '03' && selectedDoc.tipo_documento !== 'NV'"
                @click="openVoidModal(selectedDoc)"
                class="btn-danger btn-primary--sm"
                style="margin-right: 8px;"
              >
                Anular Comprobante
              </button>
              <button @click="showMsgModal = false" class="btn-primary btn-primary--sm">Cerrar</button>
            </div>
          </div>
          </div>
          </div>

    <!-- === MODAL: Editar Borrador (Antes de emitir) === -->
    <div v-if="isEditDraftModalOpen" class="modal-mini-overlay" @click.self="isEditDraftModalOpen = false">
      <div class="modal-mini animate-in" style="max-width: 550px;">
        <div class="modal-mini-header">
          <h3>Editar Borrador de Comprobante</h3>
          <button @click="isEditDraftModalOpen = false" class="modal-close-btn">×</button>
        </div>
        <div class="modal-mini-body">
          <p style="font-size: 13px; color: #64748b; margin-bottom: 16px;">
            Revisa y corrige los datos antes de la emisión final a SUNAT.
          </p>
          
          <div class="form-section-label">Identificación del Cliente</div>
          <div class="form-grid form-grid--2">
            <div class="field">
              <label class="field-label">Tipo Documento</label>
              <select v-model="draftForm.tipo_documento_identidad" class="field-input field-input--select">
                <option value="DNI">DNI</option>
                <option value="RUC">RUC</option>
                <option value="CE">Carné de Extranjería</option>
                <option value="PASAPORTE">Pasaporte</option>
                <option value="SIN_DOCUMENTO">Sin Documento</option>
              </select>
            </div>
            <div class="field">
              <label class="field-label">Nro. Documento</label>
              <input v-model="draftForm.numero_documento" type="text" class="field-input" placeholder="Ej. 12345678" />
            </div>
          </div>

          <div class="field" style="margin-top: 14px;">
            <label class="field-label">Nombre / Razón Social <span class="required">*</span></label>
            <input 
              v-model="draftForm.razon_social" 
              type="text" 
              class="field-input" 
              placeholder="Nombre legal del cliente..."
            />
          </div>

          <div class="field" style="margin-top: 14px;">
            <label class="field-label">Dirección Fiscal</label>
            <input 
              v-model="draftForm.direccion" 
              type="text" 
              class="field-input" 
              placeholder="Opcional..."
            />
          </div>

          <div class="form-section-label" style="margin-top: 20px;">Datos del Comprobante</div>
          <div class="form-grid form-grid--3">
            <div class="field">
              <label class="field-label">Serie</label>
              <input v-model="draftForm.serie" type="text" class="field-input" maxlength="4" style="text-transform: uppercase;" />
            </div>
            <div class="field">
              <label class="field-label">Número</label>
              <input v-model="draftForm.numero" type="text" class="field-input" />
            </div>
            <div class="field">
              <label class="field-label">Fecha Emisión</label>
              <input v-model="draftForm.fecha_emision" type="date" class="field-input" />
            </div>
          </div>
          <div class="form-section-label" style="margin-top: 20px;">Detalle de Productos & Servicios</div>
          <div class="draft-items-table-wrap">
            <table class="draft-items-table">
              <thead>
                <tr>
                  <th>Descripción</th>
                  <th width="80">Cant.</th>
                  <th width="100">Precio (C/IGV)</th>
                  <th width="100" class="text-right">Total</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in draftForm.items" :key="item.id">
                  <td class="item-name">{{ item.nombre }}</td>
                  <td>
                    <input v-model.number="item.cantidad" type="number" class="item-input" step="0.01" />
                  </td>
                  <td>
                    <input v-model.number="item.precio_unitario" type="number" class="item-input" step="0.01" />
                  </td>
                  <td class="text-right item-total">
                    S/ {{ (item.cantidad * item.precio_unitario).toFixed(2) }}
                  </td>
                </tr>
              </tbody>
              <tfoot>
                <tr>
                  <td colspan="3" class="text-right" style="font-weight: 700; padding-top: 12px;">TOTAL COMPROBANTE:</td>
                  <td class="text-right" style="font-weight: 800; color: #2563eb; font-size: 15px; padding-top: 12px;">
                    S/ {{ computedDraftTotal.toFixed(2) }}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
        <div class="modal-mini-footer">
          <button @click="isEditDraftModalOpen = false" class="btn-secondary-mini" style="margin-right: 8px;">Cancelar</button>
          <button 
            @click="submitDraftUpdate" 
            :disabled="editingDraft || !draftForm.razon_social" 
            class="btn-primary-mini"
          >
            <i v-if="editingDraft" class="ti ti-loader animate-spin"></i>
            Actualizar Borrador
          </button>
        </div>
      </div>
    </div>

          <!-- === MODAL: Anulación Legal (Baja) === -->
      <div v-if="isVoidModalOpen" class="modal-mini-overlay" @click.self="isVoidModalOpen = false">
        <div class="modal-mini animate-in" style="max-width: 450px;">
          <div class="modal-mini-header" style="background: #fef2f2; border-bottom-color: #fee2e2;">
            <h3 style="color: #b91c1c;">Solicitar Anulación (Baja)</h3>
            <button @click="isVoidModalOpen = false" class="modal-close-btn">×</button>
          </div>
          <div class="modal-mini-body">
            <p style="font-size: 13.5px; color: #475569; margin-bottom: 20px; line-height: 1.5;">
              Estás por anular legalmente la <strong>{{ docToVoid?.serie }}-{{ docToVoid?.numero }}</strong>. 
              Esta acción enviará una <strong>Comunicación de Baja</strong> a SUNAT.
            </p>
            
            <div class="field">
              <label class="field-label" style="font-weight: 700; color: #1e293b;">Motivo de Anulación <span class="required">*</span></label>
              <textarea 
                v-model="voidReason" 
                class="field-input" 
                rows="3" 
                placeholder="Ej: Error en datos del cliente, Anulación de la operación..."
                style="resize: none; border-color: #cbd5e1;"
              ></textarea>
              <p class="field-hint" style="margin-top: 6px;">Mínimo 10 caracteres.</p>
            </div>

            <div class="alert-banner alert-banner--warning" style="margin-top:20px; padding: 12px; border-radius: 10px;">
              <div style="display: flex; gap: 10px;">
                <i class="ti ti-info-circle" style="font-size: 18px;"></i>
                <p style="font-size: 12px; line-height: 1.4; margin: 0;">
                  Esta operación es irreversible. Una vez aceptada la baja por SUNAT, el comprobante quedará sin valor legal.
                </p>
              </div>
            </div>
          </div>
          <div class="modal-mini-footer" style="padding: 16px 24px 24px;">
            <button @click="isVoidModalOpen = false" class="btn-secondary" style="margin-right: 12px; font-weight: 600;">Cancelar</button>
            <button 
              @click="submitVoidRequest" 
              :disabled="voidingDoc || voidReason.length < 10" 
              class="btn-primary"
              style="background: #ef4444; box-shadow: 0 4px 12px rgba(239, 68, 68, 0.25);"
            >
              <i v-if="voidingDoc" class="ti ti-loader animate-spin" style="margin-right: 6px;"></i>
              Confirmar Anulación
            </button>
          </div>
        </div>
      </div>

      <!-- ================= TAB 3: SERIES & CORRELATIVOS ================= -->
      <div v-if="activeTab === 'series'" class="tab-panel animate-in">
        
        <div v-if="loadingSeries" class="loading-state">
          <div class="spinner"></div>
          <p>Cargando configuración de sucursales...</p>
        </div>

        <div v-else>
          <div class="info-box info-box--neutral">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:16px;height:16px;flex-shrink:0;"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="8"/><line x1="12" y1="12" x2="12" y2="16"/></svg>
            Las series y correlativos se configuran por sucursal. Configura las series oficiales autorizadas por SUNAT. El número correlativo inicial se usará como punto de partida si no existen comprobantes emitidos previamente con esa serie.
          </div>

          <transition name="fade-slide">
            <div v-if="branchSuccessMsg" class="success-banner">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="width:16px;height:16px;"><polyline points="20 6 9 17 4 12"/></svg>
              {{ branchSuccessMsg }}
            </div>
            <div v-else-if="branchErrorMsg" class="error-banner">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="width:16px;height:16px;"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
              {{ branchErrorMsg }}
            </div>
          </transition>

          <div v-for="branch in allBranches" :key="branch.id" class="branch-series-block">
            <div class="branch-series-header">
              <div class="branch-series-title">
                <i class="ti ti-building-store"></i>
                <span>{{ branch.nombre }}</span>
                <span v-if="branch.is_main" class="main-branch-badge">Principal</span>
              </div>
              <button 
                type="button"
                @click="saveSingleBranch(branch)" 
                :disabled="savingBranchId === branch.id"
                class="btn-primary btn-primary--sm"
              >
                <div v-if="savingBranchId === branch.id" class="btn-spinner"></div>
                <i v-else class="ti ti-device-floppy"></i>
                {{ savingBranchId === branch.id ? 'Guardando...' : 'Guardar' }}
              </button>
            </div>

            <div class="series-block-grid">
              <!-- FACTURAS -->
              <div class="series-card">
                <div class="series-card-header">
                  <span class="series-badge">FACTURAS</span>
                </div>
                <div class="series-card-body">
                  <div class="field">
                    <label class="field-label">Serie de Factura</label>
                    <input 
                      v-model="branch.serie_factura" 
                      type="text" 
                      placeholder="Ej. F001" 
                      maxlength="4" 
                      class="field-input field-input--mono" 
                    />
                    <p class="field-hint">4 chars — F001 (prod) / FL01 (beta)</p>
                  </div>
                  <div class="field" style="margin-top:12px;">
                    <label class="field-label">Siguiente Correlativo</label>
                    <input 
                      v-model.number="branch.correlativo_factura" 
                      type="number" 
                      placeholder="Ej. 1" 
                      min="1" 
                      class="field-input field-input--mono" 
                    />
                    <p class="field-hint">Autoincrementa con cada venta</p>
                  </div>
                </div>
              </div>

              <!-- BOLETAS -->
              <div class="series-card">
                <div class="series-card-header">
                  <span class="series-badge series-badge--boleta">BOLETAS</span>
                </div>
                <div class="series-card-body">
                  <div class="field">
                    <label class="field-label">Serie de Boleta</label>
                    <input 
                      v-model="branch.serie_boleta" 
                      type="text" 
                      placeholder="Ej. B001" 
                      maxlength="4" 
                      class="field-input field-input--mono" 
                    />
                    <p class="field-hint">4 chars — B001 (prod) / BL01 (beta)</p>
                  </div>
                  <div class="field" style="margin-top:12px;">
                    <label class="field-label">Siguiente Correlativo</label>
                    <input 
                      v-model.number="branch.correlativo_boleta" 
                      type="number" 
                      placeholder="Ej. 1" 
                      min="1" 
                      class="field-input field-input--mono" 
                    />
                    <p class="field-hint">Autoincrementa con cada venta</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="allBranches.length === 0" class="empty-table-state">
            <p class="empty-title">No hay sucursales configuradas</p>
          </div>
        </div>

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
import { ref, onMounted, onUnmounted, reactive, watch, computed } from 'vue'
import { useAuthStore } from '../store/auth'
import axios from 'axios'

const authStore = useAuthStore()

const loading = ref(true)
const saving = ref(false)
const showSolPass = ref(false)
const showClientSecret = ref(false)
const showCertPass = ref(false)
const certInput = ref<HTMLInputElement | null>(null)

const props = defineProps({
  initialTab: {
    type: String,
    default: 'documents'
  }
})

const activeTab = ref(props.initialTab) // config, documents, series

// Void modal state
const isVoidModalOpen = ref(false)
const voidReason = ref('')
const docToVoid = ref<any>(null)
const voidingDoc = ref(false)
const batchEmitting = ref(false)

// Edit draft state
const isEditDraftModalOpen = ref(false)
const editingDraft = ref(false)
const draftForm = reactive({
  id: '',
  tipo_documento_identidad: 'DNI',
  numero_documento: '',
  razon_social: '',
  direccion: '',
  serie: '',
  numero: '',
  fecha_emision: '',
  items: [] as any[]
})

const computedDraftTotal = computed(() => {
  return draftForm.items.reduce((acc, item) => acc + (item.cantidad * item.precio_unitario), 0)
})

const emit = defineEmits(['tabChange'])

watch(activeTab, (newTab) => {
  emit('tabChange', newTab)
})

// API Config state
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
  logo_base64: '',
  emision_diferida: false
})

const saveLogs = ref<string[]>([])
const apiError = ref<string>('')

async function batchEmitDrafts() {
  if (!confirm('¿Estás seguro de emitir todos los comprobantes pendientes? Esta acción procesará todos los borradores de la jornada.')) return
  
  batchEmitting.value = true
  try {
    const res = await axios.post('/billing/documents/batch-emit')
    if (res.data.success) {
      alert(res.data.message)
      await loadDocuments()
      await loadDocumentStats()
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error al procesar emisión por lotes')
  } finally {
    batchEmitting.value = false
  }
}

// Documents state
const documents = ref<any[]>([])
const loadingDocs = ref(false)
const searchQuery = ref('')
const filterType = ref('')
const filterEstado = ref('')
const filterFechaDesde = ref('')
const filterFechaHasta = ref('')

// Document stats
const docStats = ref<any>({ grand_total: 0, totals: [], data: [] })

// Resend and Message logic
const resendingId = ref<string | null>(null)
const syncingId = ref<string | null>(null)
const showMsgModal = ref(false)
const selectedDoc = ref<any>(null)

function cleanErrorMessage(error: string) {
  if (!error) return ""
  try {
    const decoded = JSON.parse(error)
    return decoded.message || decoded.sunat_description || error
  } catch (e) {
    return error
  }
}

function showSunatMessage(doc: any) {
  selectedDoc.value = doc
  showMsgModal.value = true
}

async function syncDocumentStatus(id: string) {
  syncingId.value = id
  try {
    const res = await axios.get(`/billing/documents/${id}/sync`)
    if (res.data.success) {
      // Find and update the document in the list
      const idx = documents.value.findIndex(d => d.id === id)
      if (idx !== -1) documents.value[idx] = res.data.data
    }
  } catch (err: any) {
    console.error('Error syncing status', err)
  } finally {
    syncingId.value = null
  }
}

// Polling for pending docs
let pollingInterval: any = null
const startPolling = () => {
  if (pollingInterval) return
  pollingInterval = setInterval(() => {
    const pendingDocs = documents.value.filter(d => d.estado === 'pending' || d.estado === 'error')
    if (pendingDocs.length > 0) {
      pendingDocs.forEach(d => syncDocumentStatus(d.id))
    } else {
      stopPolling()
    }
  }, 5000)
}

const stopPolling = () => {
  if (pollingInterval) {
    clearInterval(pollingInterval)
    pollingInterval = null
  }
}

async function resendDocument(id: string) {
  resendingId.value = id
  try {
    const res = await axios.post(`/billing/documents/${id}/resend`)
    if (res.data.success) {
      alert('¡Comprobante re-enviado con éxito!')
      await loadDocuments()
      await loadDocumentStats()
    }
  } catch (err: any) {
    alert('Error al re-enviar: ' + (err.response?.data?.error || err.message))
  } finally {
    resendingId.value = null
  }
}

async function deleteDocumentTest(id: string) {
  if (!confirm('¿Estás seguro de eliminar este registro? Esta acción solo está permitida en MODO PRUEBA para corregir errores de envío.')) return
  try {
    const res = await axios.delete(`/billing/documents/${id}`)
    if (res.data.success) {
      showMsgModal.value = false
      await loadDocuments()
      await loadDocumentStats()
    }
  } catch (err: any) {
    alert('Error: ' + (err.response?.data?.error || err.message))
  }
}

function openVoidModal(doc: any) {
  docToVoid.value = doc
  voidReason.value = ''
  isVoidModalOpen.value = true
}

async function submitVoidRequest() {
  if (!docToVoid.value) return
  if (voidReason.value.length < 10) {
    alert('El motivo debe tener al menos 10 caracteres.')
    return
  }

  voidingDoc.value = true
  try {
    const res = await axios.post(`/billing/documents/${docToVoid.value.id}/void`, {
      motivo: voidReason.value
    })
    if (res.data.success) {
      isVoidModalOpen.value = false
      alert('Solicitud de anulación enviada correctamente. SUNAT procesará la baja en unos momentos.')
      await loadDocuments()
      await loadDocumentStats()
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error al enviar solicitud de anulación')
  } finally {
    voidingDoc.value = false
  }
}

function openEditDraftModal(doc: any) {
  if (!doc) return
  draftForm.id = doc.id
  draftForm.tipo_documento_identidad = doc.sale?.customer?.tipo_documento || 'DNI'
  draftForm.numero_documento = doc.sale?.customer?.numero_documento || ''
  draftForm.razon_social = doc.sale?.customer?.nombre || 'Público General'
  draftForm.direccion = doc.sale?.customer?.direccion || ''
  draftForm.serie = doc.serie
  draftForm.numero = doc.numero
  draftForm.fecha_emision = doc.created_at ? new Date(doc.created_at).toISOString().split('T')[0] : new Date().toISOString().split('T')[0]
  
  // Load items from sale
  draftForm.items = doc.sale?.items?.map((item: any) => ({
    id: item.id,
    nombre: item.product?.nombre || 'Producto',
    cantidad: item.cantidad,
    precio_unitario: item.precio_unitario
  })) || []

  isEditDraftModalOpen.value = true
}

async function submitDraftUpdate() {
  editingDraft.value = true
  try {
    const res = await axios.patch(`/billing/documents/${draftForm.id}/draft`, {
      tipo_documento_identidad: draftForm.tipo_documento_identidad,
      numero_documento: draftForm.numero_documento,
      razon_social: draftForm.razon_social,
      direccion: draftForm.direccion,
      serie: draftForm.serie,
      numero: draftForm.numero,
      fecha_emision: draftForm.fecha_emision,
      items: draftForm.items.map(i => ({
        id: i.id,
        cantidad: i.cantidad,
        precio_unitario: i.precio_unitario
      }))
    })
    if (res.data.success) {
      isEditDraftModalOpen.value = false
      alert('Comprobante borrador actualizado correctamente.')
      await loadDocuments()
    }
  } catch (err: any) {
    alert(err.response?.data?.error || 'Error al actualizar borrador')
  } finally {
    editingDraft.value = false
  }
}

// Series/Branch state
const activeBranch = ref<any>(null)
const allBranches = ref<any[]>([])
const loadingSeries = ref(false)
const savingBranch = ref(false)
const savingBranchId = ref<string | null>(null)
const branchSuccessMsg = ref('')
const branchErrorMsg = ref('')

onMounted(async () => {
  await Promise.all([
    loadConfig(),
    loadDocuments(),
    loadDocumentStats(),
    loadAllBranches()
  ])
})

onUnmounted(() => {
  stopPolling()
})

watch(documents, (newDocs) => {
  const hasPending = newDocs.some(d => d.estado === "pending" || d.estado === "error")
  if (hasPending) {
    startPolling()
  } else {
    stopPolling()
  }
}, { deep: true })

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
  saveLogs.value = []
  apiError.value = ''
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
    const data = err.response?.data
    apiError.value = data?.error || data?.api_error || 'Error de red o de servidor al intentar guardar.'
    if (data?.logs && data.logs.length > 0) {
      saveLogs.value = data.logs
    } else {
      saveLogs.value = ['Error al conectar con el servidor.']
    }
  } finally {
    saving.value = false
  }
}

// Fetch Electronic Documents
async function loadDocuments() {
  loadingDocs.value = true
  try {
    const params: any = {}
    if (filterType.value) params.tipo_documento = filterType.value
    if (filterEstado.value) params.estado = filterEstado.value
    if (searchQuery.value) params.search = searchQuery.value
    if (filterFechaDesde.value) params.fecha_desde = filterFechaDesde.value
    if (filterFechaHasta.value) params.fecha_hasta = filterFechaHasta.value

    const res = await axios.get('/billing/documents', { params })
    if (res.data.success) {
      documents.value = res.data.data
    }
  } catch (err) {
    console.error('Error loading electronic documents', err)
  } finally {
    loadingDocs.value = false
  }
}

// Fetch Document Stats
async function loadDocumentStats() {
  try {
    const res = await axios.get('/billing/documents/stats')
    if (res.data.success) {
      docStats.value = res.data
    }
  } catch (err) {
    console.error('Error loading document stats', err)
  }
}

function getStatTotal(tipo: string): number {
  if (!docStats.value.totals) return 0
  const found = docStats.value.totals.find((t: any) => t.tipo_documento === tipo)
  return found ? found.total : 0
}

// Fetch All Branches with series config
async function loadAllBranches() {
  loadingSeries.value = true
  try {
    const res = await axios.get('/billing/series')
    if (res.data.success && res.data.data) {
      allBranches.value = res.data.data
      // Also set activeBranch for legacy compat
      const activeBranchId = authStore.user?.branch_id
      activeBranch.value = res.data.data.find((b: any) => b.id === activeBranchId) || res.data.data[0]
    }
  } catch (err) {
    console.error('Error loading branch series', err)
  } finally {
    loadingSeries.value = false
  }
}

// Save single branch series via dedicated endpoint
async function saveSingleBranch(branch: any) {
  savingBranchId.value = branch.id
  branchSuccessMsg.value = ''
  branchErrorMsg.value = ''
  try {
    const res = await axios.patch(`/billing/series/${branch.id}`, {
      serie_factura: branch.serie_factura,
      serie_boleta: branch.serie_boleta,
      correlativo_factura: branch.correlativo_factura,
      correlativo_boleta: branch.correlativo_boleta,
    })
    if (res.data.success) {
      branchSuccessMsg.value = `Series de "${branch.nombre}" guardadas correctamente.`
      // Update local data
      const idx = allBranches.value.findIndex((b: any) => b.id === branch.id)
      if (idx !== -1) allBranches.value[idx] = res.data.data
      setTimeout(() => { branchSuccessMsg.value = '' }, 4000)
    }
  } catch (err: any) {
    branchErrorMsg.value = err.response?.data?.error || 'Error al guardar series.'
  } finally {
    savingBranchId.value = null
  }
}

function clearDateFilter() {
  filterFechaDesde.value = ''
  filterFechaHasta.value = ''
  loadDocuments()
}

// Helper: Download files from FacturaAPI
async function downloadFile(uuid: string, type: 'pdf' | 'xml' | 'cdr') {
  try {
    const res = await axios.get(`/billing/files/${uuid}`)
    if (res.data.success && res.data.data && res.data.data.files && res.data.data.files[type]) {
      window.open(res.data.data.files[type], '_blank')
    } else if (res.data.files && res.data.files[type]) {
      window.open(res.data.files[type], '_blank')
    } else {
      alert('El comprobante no tiene este archivo disponible en el microservicio.')
    }
  } catch (err: any) {
    if (err.response?.status === 404 || (err.response?.data?.error && err.response.data.error.includes('404'))) {
      alert('Error 404: El archivo no existe en el servidor de SUNAT/API.\nSi esto es un registro huérfano de pruebas, puedes usar el botón "Eliminar" para limpiar tu sistema local.')
    } else {
      alert('No se pudo recuperar el archivo: ' + (err.response?.data?.error || 'el comprobante está en modo pruebas y no tiene archivos guardados en el servidor.'))
    }
  }
}

function formatDocType(type: string): string {
  switch (type) {
    case '01': return 'Factura'
    case '03': return 'Boleta'
    case '07': return 'Nota Crédito'
    case '08': return 'Nota Débito'
    default: return 'Doc'
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('es-PE', { timeZone: 'America/Lima' })
}

const clearLogs = () => {
  saveLogs.value = []
  apiError.value = ''
}
</script>

<style scoped>
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

/* Tabs Navigation */
.tabs-nav {
  display: flex;
  gap: 8px;
  background: #f1f5f9;
  padding: 6px;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
}
.tab-nav-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 16px;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.25s ease;
}
.tab-nav-btn:hover {
  color: #1e293b;
  background: rgba(255, 255, 255, 0.4);
}
.tab-nav-btn--active {
  background: white;
  color: #6366f1;
  box-shadow: 0 4px 12px rgba(0,0,0,0.04);
}
.tab-icon {
  width: 15px;
  height: 15px;
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
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
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

/* Filters for Vouchers List */
.filter-controls-wrap {
  display: flex;
  gap: 14px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.search-box {
  flex: 1;
  min-width: 260px;
  position: relative;
}
.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: #94a3b8;
}
.search-input {
  width: 100%;
  padding: 9px 12px 9px 38px;
  border: 1.5px solid #e2e8f0;
  border-radius: 9px;
  font-size: 13px;
  outline: none;
  background: #f8fafc;
  box-sizing: border-box;
  transition: all 0.2s;
}
.search-input:focus {
  border-color: #6366f1;
  background: white;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.12);
}
.select-filter-group {
  display: flex;
  gap: 10px;
}
.filter-select {
  padding: 9px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 9px;
  font-size: 13px;
  color: #475569;
  background: #f8fafc;
  outline: none;
  cursor: pointer;
  transition: all 0.2s;
}
.filter-select:focus {
  border-color: #6366f1;
  background: white;
}

/* Date filter group */
.date-filter-group {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.date-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.date-label {
  font-size: 10px;
  font-weight: 700;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.clear-date-btn {
  padding: 7px 12px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
  background: white;
  cursor: pointer;
  transition: all 0.15s;
  margin-top: 13px;
}
.clear-date-btn:hover {
  background: #fef2f2;
  border-color: #fca5a5;
  color: #dc2626;
}

/* Results count */
.results-count {
  font-size: 12px;
  color: #64748b;
  margin-bottom: 10px;
}
.results-count strong {
  color: #6366f1;
}

/* Document Stats Grid */
.doc-stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}
@media (max-width: 768px) {
  .doc-stats-grid { grid-template-columns: repeat(2, 1fr); }
}
.doc-stat-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 12px;
  border: 1.5px solid #e2e8f0;
  background: white;
  transition: box-shadow 0.2s;
}
.doc-stat-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.06); }
.doc-stat-card--total { border-color: #c7d2fe; background: #f0f4ff; }
.doc-stat-card--factura { border-color: #a7f3d0; background: #f0fdf4; }
.doc-stat-card--boleta { border-color: #bae6fd; background: #f0f9ff; }
.doc-stat-card--nc { border-color: #fde68a; background: #fffbeb; }
.doc-stat-icon {
  width: 36px;
  height: 36px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.doc-stat-card--total .doc-stat-icon { background: #818cf8; color: white; }
.doc-stat-card--factura .doc-stat-icon { background: #34d399; color: white; }
.doc-stat-card--boleta .doc-stat-icon { background: #38bdf8; color: white; }
.doc-stat-card--nc .doc-stat-icon { background: #f59e0b; color: white; }
.doc-stat-label {
  font-size: 10px;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.doc-stat-value {
  font-size: 22px;
  font-weight: 900;
  color: #1e293b;
  line-height: 1.2;
}

/* Branch series block */
.branch-series-block {
  border: 1.5px solid #e2e8f0;
  border-radius: 14px;
  overflow: hidden;
  margin-bottom: 16px;
}
.branch-series-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  background: #f8fafc;
  border-bottom: 1.5px solid #e2e8f0;
}
.branch-series-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 700;
  color: #1e293b;
}
.main-branch-badge {
  font-size: 9px;
  font-weight: 800;
  padding: 2px 8px;
  border-radius: 99px;
  background: #ede9fe;
  color: #7c3aed;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}
.btn-primary--sm {
  padding: 7px 14px !important;
  font-size: 11px !important;
  gap: 5px !important;
}
.branch-series-block .series-block-grid {
  padding: 16px;
  background: white;
}


.table-responsive {
  width: 100%;
  overflow-x: auto;
  border: 1.5px solid #f1f5f9;
  border-radius: 12px;
}
.billing-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 13px;
}
.billing-table th {
  background: #f8fafc;
  padding: 12px 16px;
  font-weight: 700;
  color: #475569;
  border-bottom: 1.5px solid #f1f5f9;
}
.billing-table td {
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
  color: #1e293b;
  vertical-align: middle;
}
.billing-table tr:last-child td {
  border-bottom: none;
}
.billing-table tr:hover td {
  background: #fafafa;
}
.table-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
  color: #94a3b8;
  gap: 10px;
  font-size: 12.5px;
}
.doc-ident-wrap {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.doc-type-pill {
  font-size: 9px;
  font-weight: 800;
  text-transform: uppercase;
  color: #6366f1;
  background: #f5f3ff;
  padding: 2px 6px;
  border-radius: 4px;
  align-self: flex-start;
  letter-spacing: 0.2px;
}
.cust-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.cust-name {
  font-weight: 600;
  color: #334155;
  margin: 0;
}
.cust-doc {
  margin: 0;
}

/* Badge states */
.badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 99px;
  font-size: 10.5px;
  font-weight: 800;
}
.badge-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: currentColor;
}
.badge--accepted { background: #f0fdf4; color: #16a34a; }
.badge--rejected { background: #fef2f2; color: #ef4444; }
.badge--pending { background: #fffbeb; color: #d97706; }
.badge--voided { background: #f8fafc; color: #64748b; }
.badge--error { background: #fff1f2; color: #e11d48; }

.sync-status-inline-btn {
  background: none;
  border: none;
  color: #6366f1;
  cursor: pointer;
  padding: 4px;
  margin-left: 4px;
  border-radius: 4px;
  transition: background 0.2s;
  display: inline-flex;
  align-items: center;
  vertical-align: middle;
}
.sync-status-inline-btn:hover:not(:disabled) {
  background: #f5f3ff;
}
.sync-status-inline-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Download Actions Group */
.action-btn-group {
  display: inline-flex;
  gap: 4px;
}
.draft-items-table-wrap {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 12px;
  margin-top: 10px;
}
.draft-items-table {
  width: 100%;
  border-collapse: collapse;
}
.draft-items-table th {
  text-align: left;
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e2e8f0;
}
.draft-items-table td { padding: 8px 0; border-bottom: 1px dashed #f1f5f9; }
.item-name { font-size: 13px; font-weight: 600; color: #334155; }
.item-input {
  width: 90%;
  padding: 6px 8px;
  border: 1.5px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 700;
  color: #0f172a;
}
.item-total { font-size: 13px; font-weight: 700; color: #475569; }

.action-btn {
  padding: 5px 10px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
  border: 1px solid;
  transition: all 0.2s;
  background: white;
}
.action-btn--pdf {
  border-color: #fca5a5;
  color: #dc2626;
}
.action-btn--pdf:hover {
  background: #fef2f2;
}
.action-btn--xml {
  border-color: #93c5fd;
  color: #2563eb;
}
.action-btn--xml:hover {
  background: #eff6ff;
}
.action-btn--resend {
  border-color: #6366f1;
  color: #6366f1;
}
.action-btn--resend:hover {
  background: #f5f3ff;
}
.action-btn--msg {
  border-color: #64748b;
  color: #64748b;
}
.action-btn--msg:hover {
  background: #f8fafc;
}
.action-btn--cdr {
  border-color: #cbd5e1;
  color: #475569;
}
.action-btn--cdr:hover:not(:disabled) {
  background: #f1f5f9;
}
.action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  background: #f8fafc;
}

/* Series Cards configuration */
.series-block-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-top: 18px;
}
@media (max-width: 768px) {
  .series-block-grid {
    grid-template-columns: 1fr;
  }
}
.series-card {
  border: 1.5px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  background: #f8fafc;
  transition: border-color 0.25s;
}
.series-card:hover {
  border-color: #cbd5e1;
}
.series-card-header {
  padding: 12px 16px;
  border-bottom: 1.5px solid #e2e8f0;
  background: white;
  display: flex;
}
.series-badge {
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.5px;
  background: #eef2ff;
  color: #4f46e5;
  padding: 3px 8px;
  border-radius: 5px;
}
.series-badge--boleta {
  background: #f0fdf4;
  color: #16a34a;
}
.series-card-body {
  padding: 18px 16px;
}

/* Banners success & error */
.success-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  font-size: 12.5px;
  color: #15803d;
  margin-bottom: 16px;
}
.error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: #fef2f2;
  border: 1px solid #fca5a5;
  border-radius: 8px;
  font-size: 12.5px;
  color: #b91c1c;
  margin-bottom: 16px;
}
.info-box--neutral {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #475569;
  margin-bottom: 16px;
}

/* Empty State */
.empty-table-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 60px 20px;
  color: #94a3b8;
}
.empty-icon {
  width: 48px;
  height: 48px;
  stroke: #cbd5e1;
  margin-bottom: 12px;
}
.empty-title {
  font-size: 14px;
  font-weight: 700;
  color: #475569;
  margin: 0 0 4px;
}
.empty-sub {
  font-size: 12px;
  max-width: 380px;
  margin: 0;
  line-height: 1.5;
}

/* Config forms styling */
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
  border-radius: 99px;
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
.hidden-file { display: none; }

/* SUNAT Message Modal */
.modal-overlay { 
  position: fixed; inset: 0; background: rgba(15, 23, 42, 0.6); 
  backdrop-filter: blur(4px); z-index: 300; display: flex; 
  align-items: center; justify-content: center; padding: 20px; 
}
.modal-mini { 
  background: white; padding: 24px; border-radius: 24px; 
  width: 100%; max-width: 420px; display: flex; 
  flex-direction: column; gap: 16px; box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}
.modal-mini-header { display: flex; align-items: center; justify-content: space-between; }
.modal-mini-title { font-size: 16px; font-weight: 800; color: #1e293b; margin: 0; }
.modal-close-btn { background: none; border: none; font-size: 18px; color: #94a3b8; cursor: pointer; }
.modal-mini-body { padding: 4px 0; }
.msg-box { padding: 18px; border-radius: 16px; border: 1px solid; display: flex; flex-direction: column; gap: 12px; }
.msg-box--accepted { border-color: #bbf7d0; background: #f0fdf4; }
.msg-box--pending { border-color: #fde68a; background: #fffbeb; }
.msg-box--rejected { border-color: #fca5a5; background: #fef2f2; }
.msg-box--warning { border-color: #fde68a; background: #fffbeb; }
.msg-box-header { display: flex; align-items: center; justify-content: space-between; }
.msg-text { font-size: 13.5px; color: #334155; line-height: 1.6; margin: 0; white-space: pre-wrap; }
.msg-text-raw { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; font-size: 12px; color: #ef4444; background: #fee2e2; padding: 12px; border-radius: 8px; overflow-x: auto; white-space: pre-wrap; border: 1px solid #fecaca; }
.msg-content { display: flex; flex-direction: column; gap: 14px; }
.msg-list { display: flex; flex-direction: column; gap: 8px; }
.msg-item { display: flex; gap: 10px; align-items: flex-start; }
.msg-item-bullet { width: 6px; height: 6px; border-radius: 50%; background: #6366f1; margin-top: 8px; flex-shrink: 0; }
.msg-box--rejected .msg-item-bullet { background: #ef4444; }
.msg-box--warning .msg-item-bullet { background: #f59e0b; }
.badge--warning { background: #fef3c7; color: #92400e; }
.modal-mini-footer { display: flex; justify-content: flex-end; padding-top: 8px; }

</style>
