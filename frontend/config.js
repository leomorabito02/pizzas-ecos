/**
 * Configuración global de la aplicación
 * Define la URL del backend de forma centralizada
 */

// URL del backend - Define aquí una sola vez
const BACKEND_URL = 'https://pizzas-ecos.onrender.com/api';

const CONFIG = {
    // Obtener URL del API (centralizado en BACKEND_URL)
    getAPIUrl: function() {
        console.log('🔗 Backend URL:', BACKEND_URL);
        return BACKEND_URL;
    },
    
    API_BASE: null // Se inicializa al cargar
};

// Inicializar API_BASE
CONFIG.API_BASE = CONFIG.getAPIUrl();

console.log('🚀 API Base URL:', CONFIG.API_BASE);
console.log('🌍 Current environment:', {
    hostname: window.location.hostname,
    protocol: window.location.protocol,
    pathname: window.location.pathname
});

// Exportar para usar en otros archivos
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CONFIG;
}
