import { ref } from 'vue'

// Thème d'interface, choisi dans la config (ui.theme) : 'default' (Gatus) ou 'utd' (système de design UTD,
// du gouvernement du Québec). Aucun commutateur de thème dans l'interface : la config décide.
// Avec le thème UTD, ui.theme-mode choisit 'system' (défaut, suit l'appareil), 'light' ou 'dark' ;
// le bouton clair/foncé de la page peut ensuite le changer pour ce visiteur (mémorisé dans le navigateur).
const UTD_VERSION = '5.0.0'
const UTD_BASE = `https://cdn.jsdelivr.net/gh/MTESSDev/utd-webcomponents@${UTD_VERSION}/dist`
const MODE_KEY = 'gatus:utd-mode'

const cfg = () => window.config || {}

export const uiTheme = ref(cfg().uiTheme === 'utd' ? 'utd' : 'default')

const configuredMode = () => {
  const mode = cfg().uiThemeMode
  return mode === 'light' || mode === 'dark' ? mode : 'system'
}

const visitorMode = () => {
  try {
    const saved = localStorage.getItem(MODE_KEY)
    return saved === 'light' || saved === 'dark' ? saved : null
  } catch (e) {
    return null
  }
}

const themeMode = () => visitorMode() || configuredMode()

export const utdDark = ref(false)

// Charge le CSS puis le JS d'UTD, une seule fois. Le CSS d'UTD est placé avant le CSS du site (règle UTD).
const loadUtd = () => {
  if (!document.getElementById('utd-css')) {
    const link = document.createElement('link')
    link.id = 'utd-css'
    link.rel = 'stylesheet'
    link.href = `${UTD_BASE}/css/utd-webcomponents.min.css`
    document.head.insertBefore(link, document.head.firstChild)
  }
  if (!document.getElementById('utd-js')) {
    const script = document.createElement('script')
    script.id = 'utd-js'
    script.src = `${UTD_BASE}/js/utd-webcomponents.min.js`
    document.body.appendChild(script)
  }
}

const applyUtdMode = () => {
  const mode = themeMode()
  const dark = mode === 'dark' || (mode === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  utdDark.value = dark
  const root = document.documentElement
  root.classList.toggle('utd-dark', dark)
  // La classe 'dark' de Gatus suit, pour que les variantes dark: de Tailwind restent cohérentes
  root.classList.toggle('dark', dark)
}

export const applyUiTheme = () => {
  const isUtd = uiTheme.value === 'utd'
  document.documentElement.classList.toggle('utd-theme', isUtd)
  if (!isUtd) return
  loadUtd()
  applyUtdMode()
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyUtdMode)
}

// Bouton clair/foncé de la page, avec le thème UTD
export const toggleUtdMode = () => {
  try {
    localStorage.setItem(MODE_KEY, utdDark.value ? 'light' : 'dark')
  } catch (e) { /* stockage indisponible : le choix vaut jusqu'au rechargement */ }
  const root = document.documentElement
  utdDark.value = !utdDark.value
  root.classList.toggle('utd-dark', utdDark.value)
  root.classList.toggle('dark', utdDark.value)
}
