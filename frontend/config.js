/**
 * Configuración global de la aplicación
 * Detecta automáticamente dónde debe conectarse el backend según el ambiente
 * 
 * En Netlify: Usar variable REACT_APP_API_URL
 * En localhost: Auto-detecta localhost:8080
 * En otro servidor: Auto-detecta mismo dominio:8080
 */

const CONFIG = {
    // Detectar URL del API según el ambiente
    getAPIUrl: function() {
        // 1. Si hay variable de entorno (Netlify, Vercel, etc)
        if (typeof window !== 'undefined') {
            // Netlify inyecta como window.REACT_APP_API_URL
            if (window.REACT_APP_API_URL) {
                console.log('✅ API URL from environment variable:', window.REACT_APP_API_URL);
                return window.REACT_APP_API_URL;
            }
            
            // Vercel/otros pueden usar window.ENV
            if (window.ENV?.API_URL) {
                console.log('✅ API URL from window.ENV:', window.ENV.API_URL);
                return window.ENV.API_URL;
            }
        }
        
        // 2. Si está en localhost, usar localhost:8080
        if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
            const url = 'http://localhost:8080/api';
            console.log('ℹ️  Using localhost API:', url);
            return url;
        }
        
        // 3. En producción, asumir backend en puerto 8080 del mismo dominio
        const protocol = window.location.protocol; // http: o https:
        const host = window.location.hostname;
        const port = 8080;
        const url = `${protocol}//${host}:${port}/api`;
        console.log('ℹ️  Using same-server API:', url);
        return url;
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
