/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
	theme: {
		extend: {
      fontFamily: {
        sans: ['Outfit', 'sans-serif'],
        serif: ['"Playfair Display"', 'serif'],
      },
      colors: {
        'kafapet-green': {
          500: '#10B981', // Neon Emerald
          700: '#047857', 
          900: '#064E3B', 
        },
        'luxury-gold': {
          300: '#FDE047',
          400: '#FACC15', // Neon Yellow/Gold
          500: '#EAB308',
        },
        'dark': {
          900: '#030712', // Black base
          800: '#111827', // Very dark gray
          700: '#1F2937', 
        }
      },
      backgroundImage: {
        'neon-gradient': 'linear-gradient(135deg, #10B981 0%, #3B82F6 100%)',
        'gold-glow': 'linear-gradient(135deg, #FACC15 0%, #EAB308 100%)',
        'glass-dark': 'linear-gradient(135deg, rgba(17, 24, 39, 0.7), rgba(17, 24, 39, 0.4))',
      },
      animation: {
        'blob': 'blob 7s infinite',
        'glow': 'glow 2s ease-in-out infinite alternate',
      },
      keyframes: {
        blob: {
          '0%': { transform: 'translate(0px, 0px) scale(1)' },
          '33%': { transform: 'translate(30px, -50px) scale(1.1)' },
          '66%': { transform: 'translate(-20px, 20px) scale(0.9)' },
          '100%': { transform: 'translate(0px, 0px) scale(1)' },
        },
        glow: {
          '0%': { boxShadow: '0 0 10px rgba(16, 185, 129, 0.2)' },
          '100%': { boxShadow: '0 0 20px rgba(16, 185, 129, 0.6), 0 0 40px rgba(16, 185, 129, 0.2)' }
        }
      }
    },
	},
	plugins: [],
}
