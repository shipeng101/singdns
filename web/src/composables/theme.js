import { ref } from 'vue'
import { useDark, useToggle } from '@vueuse/core'

export function useTheme() {
  const isDark = useDark({
    selector: 'html',
    attribute: 'class',
    valueDark: 'dark',
    valueLight: 'light',
  })
  const toggleTheme = useToggle(isDark)

  return {
    isDark,
    toggleTheme
  }
} 