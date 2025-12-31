// env.js - Inyecta variables de entorno en window

(function() {
    try {
        console.log('env.js: iniciando');
        
        const isDev = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
        let backendUrl = '';
        
        if (isDev) {
            backendUrl = 'http://localhost:8080/api/v1';
        } else if (window.location.hostname.indexOf('qa-ecos') !== -1) {
            backendUrl = 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
        } else if (window.location.hostname.indexOf('ecos-ventas-pizzas') !== -1) {
            backendUrl = 'https://pizzas-ecos-backend-prod-872448320700.us-central1.run.app/api/v1';
        } else {
            backendUrl = 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
        }
        
        window.BACKEND_URL = backendUrl;
        console.log('env.js: BACKEND_URL = ' + backendUrl);
        
    } catch (e) {
        console.error('env.js error:', e);
    }
})();
