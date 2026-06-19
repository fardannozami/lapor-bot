import { colors, typography, shadows } from '@lapor-bot/design-system';

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors,
      fontFamily: typography.fontFamily,
      boxShadow: shadows,
    },
  },
  plugins: [],
}
