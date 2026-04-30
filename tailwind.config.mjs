/** @type {import('tailwindcss').Config} */
export default {
    content: ['./front-astro/**/*.{astro,html,js,jsx,md,mdx,sss,ts,tsx,vue}'],
    theme: {
        extend: {
            colors: {
                terra: {
                    100: '#f9eed9',
                    200: '#f1d9ae',
                    300: '#e6be7c',
                    400: '#d99e4a',
                    500: '#c8842a',
                    600: '#a8681f',
                    700: '#87501b',
                    800: '#6e3f1c',
                    900: '#5b3419',
                },
                verde: {
                    50: '#f2f7ef',
                    100: '#e1eeda',
                    200: '#c4ddb6',
                    300: '#9dc587',
                    400: '#77ab5c',
                    500: '#578f3e',
                    600: '#43712f',
                    700: '#355927',
                    800: '#2c4821',
                    900: '#263c1d',
                },
                tierra: {
                    50: '#faf7f2',
                    100: '#f2ebe0',
                    200: '#e5d4bf',
                    300: '#d4b894',
                    400: '#bf9668',
                    500: '#b07d4e',
                    600: '#966543',
                    700: '#7c5039',
                    800: '#664232',
                    900: '#55362a',
                },
            },
            fontFamily: {
                sans: ['Inter', 'system-ui', 'sans-serif'],
                display: ['Playfair Display', 'Georgia', 'serif'],
            },
        },
    },
    plugins: [],
};