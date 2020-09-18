const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
    plugins: [
        require('@tailwindcss/ui')({
            layout: 'sidebar',
        }),
    ],
	theme: {
		extend: {
			screens: {
				'xxl': '1440px',
				'xxxl': '1920px',
			},
			padding: {
				'80': '20rem',
				'96': '24rem',
				'112': '28rem',
				'128': '32rem',
			},
			spacing: {
				'1/5': '20%',
				'1/4': '25%',
				'1/3': '33%',
				'1/2': '50%',
				'full': '100%',
			},
			colors: {
			},
			minWidth: {
				'1/10': '10%',
				'1/6': '17%',
				'1/5': '20%',
				'1/4': '25%',
				'1/3': '33%',
				'1/2': '50%',
				'2/3': '66%',
				'3/4': '75%',
				'4': '1rem',
				'6': '1.5rem',
				'8': '2rem',
			},
			maxWidth: {
				'0': '0',
				'1/10': '10%',
				'1/6': '17%',
				'1/5': '20%',
				'1/4': '25%',
				'1/3': '33%',
				'1/2': '50%',
				'2/3': '66%',
				'3/4': '75%',
				'16': '4rem',
				'24': '6rem',
				'32': '8rem',
			},
			minHeight: {
				'1/10': '10vh',
				'1/6': '17vh',
				'1/5': '20vh',
				'1/4': '25vh',
				'1/3': '33vh',
				'1/2': '50vh',
				'2/3': '66vh',
				'3/4': '75vh',
			},
			maxHeight: {
				'0': '0',
				'1/20': '5vh',
				'1/10': '10vh',
				'1/6': '17vh',
				'1/5': '20vh',
				'1/4': '25vh',
				'1/3': '33vh',
				'1/2': '50vh',
				'2/3': '66vh',
				'3/4': '75vh',
			},
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
