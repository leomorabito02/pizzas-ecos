// env.js - Inyecta variables de entorno en window
// Netlify inyecta las variables en window.__ENV, así que las copiamos a window

console.log('📌 env.js cargando...');

const __isDev = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Determinar URL del backend según el hostname
let BACKEND_URL;
if (__isDev) {
    // Desarrollo local
    BACKEND_URL = 'http://localhost:8080/api/v1';
    console.log('✅ env.js: Modo DEV');
} else if (window.location.hostname.includes('qa-ecos')) {
    // QA (detecta por el hostname de Netlify QA)
    BACKEND_URL = 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
    console.log('✅ env.js: Modo QA');
} else if (window.location.hostname.includes('ecos-ventas-pizzas')) {
    // Production (detecta por el hostname de Netlify Production)
    BACKEND_URL = 'https://pizzas-ecos-backend-prod-872448320700.us-central1.run.app/api/v1';
    console.log('✅ env.js: Modo PROD');
} else {
    // Fallback a QA si no reconoce el hostname
    BACKEND_URL = 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
    console.log('✅ env.js: Fallback a QA');
}

window.BACKEND_URL = BACKEND_URL;
console.log('✅ window.BACKEND_URL:', window.BACKEND_URL);

if (window.__ENV) {
    Object.keys(window.__ENV).forEach(key => {
        window[key] = window.__ENV[key];
    });
    if (__isDev) {
        console.log('✅ Variables de entorno cargadas desde Netlify');
        console.log('Backend URL:', BACKEND_URL);
    }
} else if (__isDev) {
    // En desarrollo local, la variable no existe
    console.log('ℹ️  Sin variables de entorno de Netlify (desarrollo local)');
}

console.log('✅ env.js completado');
