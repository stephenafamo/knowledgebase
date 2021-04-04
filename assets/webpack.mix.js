let mix = require('laravel-mix');
require('laravel-mix-clean');

mix.setPublicPath('build')
.webpackConfig({
    watchOptions: {
        ignored: /node_modules/
    }
})
.js('src/js/app.js', 'build/js')
.postCss('src/css/app.css', 'build/css', [
    require('postcss-import'),
    require('@tailwindcss/jit'),
    require('autoprefixer'),
])
.clean();
