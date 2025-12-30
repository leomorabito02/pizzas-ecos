// env.js - Inyecta variables de entorno en window
// Netlify inyecta las variables en window.__ENV, así que las copiamos a window
if (window.__ENV) {
    Object.keys(window.__ENV).forEach(key => {
        window[key] = window.__ENV[key];
    });
    console.log('✅ Variables de entorno cargadas desde Netlify');
    console.log('REACT_APP_API_URL:', window.REACT_APP_API_URL);
} else if (typeof window !== 'undefined') {
    // En desarrollo local, la variable no existe
    console.log('ℹ️  Sin variables de entorno de Netlify (desarrollo local)');
}
