<template>
  <div class="login-shell">
    <!-- Topographic background -->
    <svg class="topo-bg" viewBox="0 0 1440 900" preserveAspectRatio="none">
      <defs>
        <pattern id="topo-login" x="0" y="0" width="200" height="200" patternUnits="userSpaceOnUse">
          <path d="M0 80 Q50 60 100 80 T200 80"    fill="none" stroke="rgba(232,168,56,0.05)" stroke-width="1"/>
          <path d="M0 120 Q50 100 100 120 T200 120" fill="none" stroke="rgba(232,168,56,0.03)" stroke-width="1"/>
          <path d="M0 160 Q50 140 100 160 T200 160" fill="none" stroke="rgba(45,212,191,0.03)"  stroke-width="1"/>
          <circle cx="150" cy="50"  r="1"   fill="rgba(232,168,56,0.07)"/>
          <circle cx="30"  cy="140" r="0.8" fill="rgba(45,212,191,0.06)"/>
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#topo-login)"/>
    </svg>

    <div class="login-card">
      <!-- Brand -->
      <div class="brand">
        <div class="brand-icon">
          <svg viewBox="0 0 40 40" fill="none">
            <rect x="3"  y="18" width="7" height="19" rx="1.5" fill="#e8a838"/>
            <rect x="13" y="10" width="7" height="27" rx="1.5" fill="#2dd4bf"/>
            <rect x="23" y="5"  width="7" height="32" rx="1.5" fill="#e8a838" opacity="0.75"/>
            <rect x="33" y="13" width="5" height="24" rx="1.5" fill="#2dd4bf" opacity="0.65"/>
            <line x1="0" y1="37" x2="40" y2="37" stroke="rgba(255,255,255,0.2)" stroke-width="1"/>
          </svg>
        </div>
        <div class="brand-text">
          <span class="brand-name">{{ t('brand.platform') }}</span>
          <span class="brand-sub">{{ t('brand.ivdp') }}</span>
        </div>
      </div>

      <h2 class="form-title">{{ t('login.title') }}</h2>

      <form @submit.prevent="handleLogin" novalidate>

        <!-- Global error banner (API / network errors) -->
        <div v-if="globalError" class="error-banner" role="alert">
          <svg viewBox="0 0 20 20" fill="currentColor" class="error-icon">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
          </svg>
          {{ globalError }}
        </div>

        <!-- ── Username ── -->
        <div class="field" :class="{ 'has-error': fieldErrors.username }">
          <label for="username">{{ t('login.username') }}</label>
          <div class="input-wrap">
            <svg class="input-icon" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"/>
            </svg>
            <input
              id="username"
              v-model="username"
              type="text"
              :placeholder="t('login.usernamePlaceholder')"
              autocomplete="username"
              autofocus
              spellcheck="false"
              autocorrect="off"
              autocapitalize="none"
              @keydown.space.prevent
              @input="onUsernameInput"
              @paste.prevent="onUsernamePaste"
              @dragover.prevent
              @drop.prevent
            />
          </div>
          <span v-if="fieldErrors.username" class="field-error">{{ fieldErrors.username }}</span>
        </div>

        <!-- ── Password ── -->
        <div class="field" :class="{ 'has-error': fieldErrors.password }">
          <label for="password">{{ t('login.password') }}</label>
          <div class="input-wrap">
            <svg class="input-icon" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
            </svg>
            <input
              id="password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              :placeholder="t('login.passwordPlaceholder')"
              autocomplete="current-password"
              @input="onPasswordInput"
              @dragover.prevent
              @drop.prevent
            />
            <button
              type="button"
              class="toggle-pw"
              @click="showPassword = !showPassword"
              tabindex="-1"
              :title="showPassword ? t('login.hidePassword') : t('login.showPassword')"
            >
              <!-- eye-open -->
              <svg v-if="showPassword" viewBox="0 0 20 20" fill="currentColor">
                <path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
                <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd"/>
              </svg>
              <!-- eye-closed -->
              <svg v-else viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M3.707 2.293a1 1 0 00-1.414 1.414l14 14a1 1 0 001.414-1.414l-1.473-1.473A10.014 10.014 0 0019.542 10C18.268 5.943 14.478 3 10 3a9.958 9.958 0 00-4.512 1.074l-1.78-1.781zm4.261 4.26l1.514 1.515a2.003 2.003 0 012.45 2.45l1.514 1.514a4 4 0 00-5.478-5.478z" clip-rule="evenodd"/>
                <path d="M12.454 16.697L9.75 13.992a4 4 0 01-3.742-3.741L2.335 6.578A9.98 9.98 0 00.458 10c1.274 4.057 5.064 7 9.542 7 .847 0 1.669-.105 2.454-.303z"/>
              </svg>
            </button>
          </div>
          <span v-if="fieldErrors.password" class="field-error">{{ fieldErrors.password }}</span>
        </div>

        <!-- ── Captcha row ── -->
        <div class="captcha-row">
          <!-- Generated captcha (read-only display) -->
          <div class="field captcha-display-field" :class="{ 'has-error': fieldErrors.captcha }">
            <label>{{ t('login.captcha') }}</label>
            <div class="input-wrap">
              <input
                type="text"
                :value="generatedCaptcha"
                readonly
                tabindex="-1"
                class="captcha-input"
                aria-label="Generated captcha code"
              />
              <button
                type="button"
                class="captcha-refresh"
                @click="refreshCaptcha"
                :title="t('login.refreshCaptcha')"
                tabindex="-1"
              >
                <!-- refresh icon -->
                <svg viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v4a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- User-entered captcha -->
          <div class="field captcha-entry-field" :class="{ 'has-error': fieldErrors.captcha }">
            <label for="captcha-entry">{{ t('login.captchaEntry') }}</label>
            <div class="input-wrap">
              <input
                id="captcha-entry"
                v-model="enteredCaptcha"
                type="text"
                :placeholder="t('login.captchaPlaceholder')"
                maxlength="6"
                autocomplete="off"
                spellcheck="false"
                @input="onCaptchaInput"
                @dragover.prevent
                @drop.prevent
                @paste.prevent="onCaptchaPaste"
              />
            </div>
            <span v-if="fieldErrors.captcha" class="field-error">{{ fieldErrors.captcha }}</span>
          </div>
        </div>

        <!-- Submit -->
        <button type="submit" class="submit-btn" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          <span>{{ loading ? t('login.signingIn') : t('login.signIn') }}</span>
        </button>
      </form>

      <p class="footer-note">{{ t('login.footerNote') }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { login } from '../api/index.js'

const { t } = useI18n()

// ── Captcha config (mirrors IVDP captcha.js exactly) ─────────────────────────
const CAPTCHA_CHARS  = '23456789QWERTYUPASDFGHJKZXCVBNM'
const CAPTCHA_LENGTH = 6

function generateCaptcha() {
  let code = ''
  for (let i = 0; i < CAPTCHA_LENGTH; i++) {
    code += CAPTCHA_CHARS[Math.floor(Math.random() * CAPTCHA_CHARS.length)]
  }
  return code
}

// ── State ─────────────────────────────────────────────────────────────────────
const router           = useRouter()
const route            = useRoute()
const username         = ref('')
const password         = ref('')
const showPassword     = ref(false)
const generatedCaptcha = ref('')
const enteredCaptcha   = ref('')
const loading          = ref(false)
const globalError      = ref('')

const fieldErrors = reactive({
  username: '',
  password: '',
  captcha:  '',
})

onMounted(() => {
  generatedCaptcha.value = generateCaptcha()

  // Show a clear message when the user arrives here after a failed SSO attempt.
  // The router guard appends ?error=sso_failed before redirecting to /login.
  if (route.query.error === 'sso_failed') {
    globalError.value = t('login.invalidSSO')
  }
})

// ── Input handlers ────────────────────────────────────────────────────────────

// Strip any spaces that somehow get into the username field
function onUsernameInput(e) {
  const clean = e.target.value.replace(/\s/g, '')
  if (clean !== e.target.value) {
    username.value = clean
  }
  if (fieldErrors.username) fieldErrors.username = ''
}

// Strip ONLY leading spaces from the password field as the user types.
// Spaces in the middle of a password are valid and must not be removed.
// Trailing spaces are silently trimmed on submit via password.value.trim().
function onPasswordInput() {
  const stripped = password.value.replace(/^\s+/, '')
  if (stripped !== password.value) {
    password.value = stripped
  }
  if (fieldErrors.password) fieldErrors.password = ''
}

// Block pasting text that contains spaces into username
function onUsernamePaste(e) {
  const text = (e.clipboardData || window.clipboardData).getData('text')
  const clean = text.replace(/\s/g, '')
  username.value = (username.value + clean).trim()
}

// Force captcha entry to uppercase letters/digits only
function onCaptchaInput(e) {
  const clean = e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, '')
  enteredCaptcha.value = clean
  if (fieldErrors.captcha) fieldErrors.captcha = ''
}

function onCaptchaPaste(e) {
  const text = (e.clipboardData || window.clipboardData).getData('text')
  const clean = text.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, CAPTCHA_LENGTH)
  enteredCaptcha.value = (enteredCaptcha.value + clean).slice(0, CAPTCHA_LENGTH)
}

function refreshCaptcha() {
  generatedCaptcha.value = generateCaptcha()
  enteredCaptcha.value   = ''
  fieldErrors.captcha    = ''
}

// ── Validation ────────────────────────────────────────────────────────────────
function validate() {
  let valid = true

  fieldErrors.username = ''
  fieldErrors.password = ''
  fieldErrors.captcha  = ''

  const trimmedUser = username.value.replace(/\s/g, '')

  if (!trimmedUser) {
    fieldErrors.username = t('login.usernameRequired')
    valid = false
  }

  if (!password.value.trim()) {
    fieldErrors.password = t('login.passwordRequired')
    valid = false
  }

  if (!enteredCaptcha.value) {
    fieldErrors.captcha = t('login.captchaRequired')
    valid = false
  } else if (enteredCaptcha.value.toUpperCase() !== generatedCaptcha.value.toUpperCase()) {
    fieldErrors.captcha = t('login.captchaIncorrect')
    valid = false
  }

  return valid
}

// ── Submit ────────────────────────────────────────────────────────────────────
async function handleLogin() {
  globalError.value = ''

  if (!validate()) {
    // Only refresh the captcha when the user typed something WRONG.
    // If the field is empty we just show the error — refreshCaptcha() would
    // clear fieldErrors.captcha immediately, making the message disappear.
    if (fieldErrors.captcha && enteredCaptcha.value.trim() !== '') {
      refreshCaptcha()
    }
    return
  }

  loading.value = true
  try {
    const data = await login(username.value.replace(/\s/g, ''), password.value.trim(), enteredCaptcha.value)
    localStorage.setItem('auth_token',    data.token)
    localStorage.setItem('auth_username', data.username)
    localStorage.setItem('auth_expires',  data.expires_at)
    router.push('/')
  } catch (e) {
    globalError.value = e.message || 'Login failed. Please check your credentials.'
    refreshCaptcha()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* ── Shell ──────────────────────────────────────────────────────────────── */
.login-shell {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0a1f18 0%, #0f2a22 50%, #0d2318 100%);
  position: relative;
  overflow: hidden;
  padding: 1.5rem;
}

.topo-bg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

/* ── Card ───────────────────────────────────────────────────────────────── */
.login-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 440px;
  background: rgba(15, 42, 34, 0.88);
  border: 1px solid rgba(255, 255, 255, 0.10);
  border-radius: 16px;
  padding: 2.4rem 2.2rem 2rem;
  backdrop-filter: blur(12px);
  box-shadow:
    0 4px 6px rgba(0,0,0,0.3),
    0 20px 60px rgba(0,0,0,0.4),
    0 0 0 1px rgba(232,168,56,0.06);
}

/* ── Brand ──────────────────────────────────────────────────────────────── */
.brand {
  display: flex;
  align-items: center;
  gap: 0.9rem;
  margin-bottom: 1.8rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}
.brand-icon svg { width: 40px; height: 40px; }
.brand-text { display: flex; flex-direction: column; }
.brand-name {
  font-size: 1.05rem;
  font-weight: 700;
  color: #ffffff;
  letter-spacing: -0.01em;
}
.brand-sub {
  font-size: 0.72rem;
  color: #8bb3a1;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  margin-top: 2px;
}

/* ── Form title ─────────────────────────────────────────────────────────── */
.form-title {
  font-size: 1.05rem;
  font-weight: 600;
  color: #d5e5dd;
  margin: 0 0 1.4rem;
}

/* ── Global error banner ────────────────────────────────────────────────── */
.error-banner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(239,68,68,0.12);
  border: 1px solid rgba(239,68,68,0.30);
  border-radius: 8px;
  padding: 0.65rem 0.85rem;
  color: #fca5a5;
  font-size: 0.85rem;
  margin-bottom: 1.2rem;
}
.error-icon { width: 16px; height: 16px; flex-shrink: 0; }

/* ── Fields ─────────────────────────────────────────────────────────────── */
.field { margin-bottom: 1rem; }

.field label {
  display: block;
  font-size: 0.78rem;
  font-weight: 500;
  color: #8bb3a1;
  margin-bottom: 0.4rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.input-wrap { position: relative; }

.input-icon {
  position: absolute;
  left: 0.7rem;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: #8bb3a1;
  pointer-events: none;
}

.input-wrap input {
  width: 100%;
  background: rgba(255,255,255,0.05);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 8px;
  padding: 0.62rem 2.4rem 0.62rem 2.4rem;
  font-size: 0.9rem;
  color: #e8f4ee;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
  font-family: inherit;
}

/* field without a left icon (captcha entry) */
.captcha-entry-field .input-wrap input {
  padding-left: 0.85rem;
}

.input-wrap input::placeholder { color: rgba(139,179,161,0.45); }
.input-wrap input:focus {
  border-color: #e8a838;
  box-shadow: 0 0 0 3px rgba(232,168,56,0.12);
}

/* Error state */
.has-error .input-wrap input {
  border-color: rgba(239,68,68,0.55);
}
.has-error .input-wrap input:focus {
  border-color: #ef4444;
  box-shadow: 0 0 0 3px rgba(239,68,68,0.12);
}

.field-error {
  display: block;
  font-size: 0.75rem;
  color: #fca5a5;
  margin-top: 0.3rem;
}

/* ── Show/hide password toggle ──────────────────────────────────────────── */
.toggle-pw {
  position: absolute;
  right: 0.7rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  color: #8bb3a1;
  display: flex;
  align-items: center;
}
.toggle-pw:hover { color: #e8a838; }
.toggle-pw svg { width: 16px; height: 16px; }

/* ── Captcha row ────────────────────────────────────────────────────────── */
.captcha-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  margin-bottom: 0;
}

/* Generated captcha display */
.captcha-input {
  font-family: 'Courier New', Courier, monospace !important;
  font-size: 1.05rem !important;
  font-weight: 700 !important;
  letter-spacing: 0.25em !important;
  color: #e8a838 !important;
  cursor: default !important;
  padding-right: 2.4rem !important;
  background: rgba(232,168,56,0.06) !important;
  border-color: rgba(232,168,56,0.20) !important;
  user-select: none;
}
.captcha-input:focus {
  border-color: rgba(232,168,56,0.20) !important;
  box-shadow: none !important;
}

/* Refresh button inside captcha display field */
.captcha-refresh {
  position: absolute;
  right: 0.6rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  color: #8bb3a1;
  display: flex;
  align-items: center;
}
.captcha-refresh:hover { color: #e8a838; }
.captcha-refresh svg { width: 15px; height: 15px; }

/* ── Submit button ──────────────────────────────────────────────────────── */
.submit-btn {
  width: 100%;
  margin-top: 1rem;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, #d09b2a 0%, #e8a838 100%);
  color: #1c1302;
  border: none;
  border-radius: 9px;
  font-size: 0.95rem;
  font-weight: 700;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.1s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.55rem;
  letter-spacing: 0.01em;
}
.submit-btn:hover:not(:disabled) { opacity: 0.92; transform: translateY(-1px); }
.submit-btn:active:not(:disabled) { transform: translateY(0); }
.submit-btn:disabled { opacity: 0.55; cursor: not-allowed; }

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(28,19,2,0.3);
  border-top-color: #1c1302;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  flex-shrink: 0;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Footer note ────────────────────────────────────────────────────────── */
.footer-note {
  margin-top: 1.5rem;
  text-align: center;
  font-size: 0.75rem;
  color: rgba(139,179,161,0.5);
  line-height: 1.5;
}

@media (max-width: 480px) {
  .login-card { padding: 1.8rem 1.4rem 1.6rem; }
  .captcha-row { grid-template-columns: 1fr; }
}
</style>
