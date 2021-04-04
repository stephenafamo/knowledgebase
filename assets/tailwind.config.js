const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
  theme: {
    extend: {
      fontFamily: {
        heading: [
          ...defaultTheme.fontFamily.serif,
        ],
        body: [
          ...defaultTheme.fontFamily.sans,
        ],
      }
    }
  },
  purge: [
    '../templates/*.html',
  ],
  variants: {}
}
