import { ref } from 'vue'

const isDark = ref(false)

const apply = (dark) => {
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
}

export const initTheme = () => {
  apply(document.documentElement.classList.contains('dark'))
}

export const useTheme = () => ({
  isDark,
  toggle: () => {
    apply(!isDark.value)
    try {
      localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
    } catch (e) {
      // the choice just won't persist
    }
  }
})
