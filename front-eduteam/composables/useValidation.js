export const validateRUT = (rut) => {
  if (!rut || typeof rut !== 'string') return false
  const parts = rut.split('-')
  if (parts.length !== 2) return false
  const [body, dv] = parts
  if (dv.length !== 1) return false
  const dvChar = dv.toLowerCase()
  if (dvChar !== 'k' && (dvChar < '0' || dvChar > '9')) return false
  const bodyParts = body.split('.')
  if (bodyParts.length !== 3) return false
  if (bodyParts[0].length < 1 || bodyParts[0].length > 2) return false
  if (bodyParts[1].length !== 3 || bodyParts[2].length !== 3) return false
  return bodyParts.every(p => /^\d+$/.test(p))
}

export const isValidEmail = (email) => {
  if (!email || typeof email !== 'string') return false
  const atIndex = email.indexOf('@')
  if (atIndex <= 0) return false
  const local = email.slice(0, atIndex)
  const domain = email.slice(atIndex + 1)
  if (!local || !domain) return false
  if (local.includes(' ') || domain.includes(' ')) return false
  const dotIndex = domain.lastIndexOf('.')
  if (dotIndex <= 0 || dotIndex === domain.length - 1) return false
  return true
}

export const addDotsToRUT = (rutBody) => {
  let result = ''
  const digits = rutBody
  for (let i = 0; i < digits.length; i++) {
    result += digits[i]
    const remaining = digits.length - i - 1
    if (remaining > 0 && remaining % 3 === 0) {
      result += '.'
    }
  }
  return result
}

export const formatRUT = (rut) => {
  if (!rut || typeof rut !== 'string') return rut
  let cleanRUT = ''
  for (const ch of rut) {
    if ((ch >= '0' && ch <= '9') || ch === 'k' || ch === 'K') {
      cleanRUT += ch
    }
  }

  if (cleanRUT.length > 1) {
    const rutBody = cleanRUT.slice(0, -1)
    const rutDV = cleanRUT.slice(-1)
    return addDotsToRUT(rutBody) + '-' + rutDV
  }

  return rut
}

export default {
  validateRUT,
  isValidEmail,
  addDotsToRUT,
  formatRUT
}
