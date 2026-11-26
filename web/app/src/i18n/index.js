import fr from './fr'

// Langue choisie dans la config (ui.language) : 'fr' (défaut) ou 'en'. En anglais, les textes d'origine sont affichés tels quels.
const configured = () => (window.config && window.config.language === 'en' ? 'en' : 'fr')

export const language = configured()
export const locale = language === 'fr' ? 'fr-CA' : 'en-US'

export const t = (text, params = {}) => {
  let out = language === 'fr' && fr[text] ? fr[text] : text
  for (const key of Object.keys(params)) {
    out = out.split(`{${key}}`).join(params[key])
  }
  return out
}
