#!/bin/bash
# Script que Netlify ejecuta durante el build para inyectar variables de entorno

# Crear un archivo JavaScript que inyecte la URL del backend
cat > frontend/backend-config.js << 'EOF'
(function() {
    var hostname = window.location.hostname;
    var backendUrl;
    
    if (hostname.indexOf('localhost') !== -1 || hostname.indexOf('127.0.0.1') !== -1) {
        backendUrl = 'http://localhost:8080/api/v1';
    } else if (hostname.indexOf('qa-ecos') !== -1) {
        backendUrl = 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
    } else if (hostname.indexOf('ecos-ventas-pizzas') !== -1) {
        backendUrl = 'https://pizzas-ecos-backend-prod-872448320700.us-central1.run.app/api/v1';
    } else {
        backendUrl = 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
    }
    
    window.BACKEND_URL = backendUrl;
    console.log('Backend URL configured: ' + backendUrl);
})();
EOF

echo "Backend configuration injected in backend-config.js"
