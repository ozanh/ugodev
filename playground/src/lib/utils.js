export function debounce(func, wait) {
  let timeout

  return function executedFunc(...args) {
    const later = () => {
      timeout = null
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}
