// env.js - Inyecta variables de entorno en window
// Netlify inyecta las variables en window.__ENV, así que las copiamos a window

const IsDev = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

if (window.__ENV) {
    Object.keys(window.__ENV).forEach(key => {
        window[key] = window.__ENV[key];
    });
    if (IsDev) {
        console.log('✅ Variables de entorno cargadas desde Netlify');
    }
} else if (IsDev) {
    // En desarrollo local, la variable no existe
    console.log('ℹ️  Sin variables de entorno de Netlify (desarrollo local)');
}
