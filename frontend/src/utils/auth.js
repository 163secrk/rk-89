const TOKEN_KEY = 'meal_token'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  return localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken() {
  return localStorage.removeItem(TOKEN_KEY)
}

export function getUser() {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
}

export function setUser(user) {
  return localStorage.setItem('user', JSON.stringify(user))
}

export function removeUser() {
  return localStorage.removeItem('user')
}

export function isAdmin() {
  const user = getUser()
  return user && user.role === 'admin'
}
