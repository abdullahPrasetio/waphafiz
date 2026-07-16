import type { Config } from 'tailwindcss'

export default {
  content: [],
  theme: {
    extend: {
      colors: {
        green: {
          50:  '#E1F5EE',
          100: '#C3EBD8',
          400: '#5DCAA5',
          500: '#1D9E75',
          600: '#0F6E56',
          700: '#085041',
        },
        amber: {
          50:  '#FAEEDA',
          900: '#633806',
        },
        purple: {
          50:  '#EEEDFE',
          800: '#3C3489',
        },
        blue: {
          50:  '#E6F1FB',
          700: '#185FA5',
        },
      },
      fontFamily: {
        arabic: ['Amiri', 'serif'],
      },
      borderRadius: {
        sm: '6px',
        DEFAULT: '10px',
        lg: '14px',
        xl: '20px',
        '2xl': '24px',
      },
    },
  },
  plugins: [],
} satisfies Config
