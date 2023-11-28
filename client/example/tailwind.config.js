/** @type {import('tailwindcss').Config} */
export default {
  content: ['./public/**/*.{html,js}'],
  theme: {
    extend: {
      fontFamily: {
        primary: ['Montserrat']
      },
      colors: {
        gray: '#E5E5E5',
        yellow: '#FCA311',
        blue: '#14213D'
      }
    }
  },
  plugins: []
}
