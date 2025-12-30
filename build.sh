#!/bin/bash
# Script que Netlify ejecuta durante el build para inyectar variables de entorno

# Crear un archivo JavaScript que inyecte las variables
cat > frontend/env-inject.js << 'EOF'
// Variables de entorno inyectadas durante el build de Netlify
window.__ENV = {
  REACT_APP_API_URL: "$REACT_APP_API_URL"
};
EOF

echo "✅ Variables de entorno inyectadas en env-inject.js"
