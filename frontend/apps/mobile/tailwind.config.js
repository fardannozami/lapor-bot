const { colors, typography, shadows } = require('@lapor-bot/design-system/src/tokens');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
    "../../packages/ui/src/**/*.{js,jsx,ts,tsx}",
  ],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors,
      fontFamily: typography.fontFamily,
      boxShadow: shadows,
    },
  },
  plugins: [],
};
