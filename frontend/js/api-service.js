/**
 * API Service - Capa de comunicación con backend
 * Centraliza todos los endpoints de API v1
 */

/**
 * Helpers para el manejo de Cookies
 */
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

function setCookie(name, value, days = 7) {
    let expires = "";
    if (days) {
        const date = new Date();
        date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
        expires = `; expires=${date.toUTCString()}`;
    }
    const secure = window.location.protocol === 'https:' ? '; Secure' : '';
    document.cookie = `${name}=${value || ""}${expires}; path=/; SameSite=Strict${secure}`;
}

function eraseCookie(name) {
    document.cookie = `${name}=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT; SameSite=Strict`;
}

class APIService {
    constructor(baseURL) {
        this._baseURL = baseURL;
        this.token = this.getStoredToken();
        
        // Limpiar caché si la página fue recargada manualmente (F5)
        try {
            const navEntries = performance.getEntriesByType("navigation");
            if (navEntries.length > 0 && navEntries[0].type === "reload") {
                this.clearCache();
            }
        } catch (e) {
            console.warn("Performance API no soportada");
        }
    }

    /**
     * Limpia la caché almacenada en sessionStorage
     */
    clearCache() {
        console.log('🧹 Limpiando caché de la API...');
        const keysToRemove = [];
        for (let i = 0; i < sessionStorage.length; i++) {
            const key = sessionStorage.key(i);
            if (key && key.startsWith('api_cache_')) {
                keysToRemove.push(key);
            }
        }
        keysToRemove.forEach(key => sessionStorage.removeItem(key));
    }

    // Getter para baseURL que siempre usa la URL actualizada
    get baseURL() {
        if (this._baseURL) return this._baseURL;
        return this.getDefaultURL();
    }

    /**
     * Determina URL del backend según ambiente
     */
    getDefaultURL() {
        // Verificar si window.BACKEND_URL fue establecida por build.sh en Netlify
        if (window.BACKEND_URL) {
            console.log('📡 Usando BACKEND_URL:', window.BACKEND_URL);
            return window.BACKEND_URL;
        }
        
        const hostname = window.location.hostname;
        const isDev = hostname === 'localhost' || hostname === '127.0.0.1';
        
        // Detectar ambiente según el hostname del frontend
        let backendUrl;
        if (isDev) {
            backendUrl = 'http://localhost:8080/api/v1';
        } else if (hostname.includes('qa-ecos')) {
            backendUrl = 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
        } else if (hostname.includes('ecos-ventas-pizzas')) {
            backendUrl = 'https://pizzas-ecos-backend-prod-872448320700.us-central1.run.app/api/v1';
        } else {
            // Fallback a QA si no se reconoce el hostname
            backendUrl = 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
        }
        
        console.log('📡 Backend URL detectada:', backendUrl);
        return backendUrl;
    }

    /**
     * Obtiene token JWT de las cookies
     */
    getStoredToken() {
        return getCookie('authToken');
    }

    /**
     * Guarda token JWT en las cookies
     */
    setToken(token) {
        this.token = token;
        if (token) {
            setCookie('authToken', token, 7);
        } else {
            eraseCookie('authToken');
        }
    }

    /**
     * Headers por defecto (incluye token si existe)
     */
    getHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        return headers;
    }

    /**
     * Wrapper para fetch con manejo de errores y caché
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const method = (options.method || 'GET').toUpperCase();
        
        // Lógica de caché para peticiones GET
        const cacheKey = `api_cache_${endpoint}`;
        if (method === 'GET') {
            const cached = sessionStorage.getItem(cacheKey);
            if (cached) {
                try {
                    console.log(`🚀 [Cache Hit] ${endpoint}`);
                    return JSON.parse(cached);
                } catch (e) {
                    console.warn('Error procesando caché', e);
                }
            }
        } else if (['POST', 'PUT', 'DELETE'].includes(method)) {
            // Invalidar caché cuando se modifican datos
            this.clearCache();
        }

        const config = {
            ...options,
            headers: {
                ...this.getHeaders(),
                ...options.headers
            }
        };

        try {
            const response = await fetch(url, config);

            // Si recibimos 401, token expiró
            if (response.status === 401) {
                this.setToken(null);
                window.location.href = '/login.html';
                return null;
            }

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || `HTTP Error: ${response.status}`);
            }

            // Guardar en caché si fue exitoso
            if (method === 'GET') {
                sessionStorage.setItem(cacheKey, JSON.stringify(data));
            }

            return data;
        } catch (error) {
            console.error(`API Error [${endpoint}]:`, error);
            throw error;
        }
    }

    // ============= AUTENTICACIÓN =============

    /**
     * Login - Obtiene JWT token
     */
    async login(username, password) {
        const data = await this.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        });
        if (data && data.token) {
            this.setToken(data.token);
        }
        return data;
    }

    /**
     * Logout - Limpia token
     */
    logout() {
        this.setToken(null);
    }

    // ============= DATA ENDPOINTS =============

    /**
     * GET /data - Obtiene vendedores, clientes y productos
     */
    async getData() {
        return this.request('/data');
    }

    // ============= VENTAS =============

    /**
     * POST /ventas - Crear nueva venta
     */
    async crearVenta(ventaData) {
        return this.request('/ventas', {
            method: 'POST',
            body: JSON.stringify(ventaData)
        });
    }

    /**
     * GET /estadisticas - Obtener todas las ventas
     */
    async obtenerVentas(limit = 10, page = 1) {
        return this.request(`/estadisticas?limit=${limit}&page=${page}`);
    }

    /**
     * GET /ventas/:id - Obtener venta específica
     */
    async obtenerVenta(id) {
        return this.request(`/ventas/${id}`);
    }

    /**
     * PUT /ventas/:id - Actualizar venta
     */
    async actualizarVenta(id, ventaData) {
        return this.request(`/ventas/${id}`, {
            method: 'PUT',
            body: JSON.stringify(ventaData)
        });
    }

    /**
     * GET /estadisticas-sheet - Obtener estadísticas resumidas
     */
    async obtenerEstadisticas(limit = 10, page = 1) {
        return this.request(`/estadisticas-sheet?limit=${limit}&page=${page}`);
    }

    // ============= PRODUCTOS =============

    /**
     * GET /productos - Listar productos
     */
    async obtenerProductos() {
        return this.request('/productos');
    }

    /**
     * POST /productos - Crear producto
     */
    async crearProducto(productoData) {
        return this.request('/productos', {
            method: 'POST',
            body: JSON.stringify(productoData)
        });
    }

    /**
     * PUT /productos/:id - Actualizar producto
     */
    async actualizarProducto(id, productoData) {
        return this.request(`/productos/${id}`, {
            method: 'PUT',
            body: JSON.stringify(productoData)
        });
    }

    /**
     * DELETE /productos/:id - Eliminar producto
     */
    async eliminarProducto(id) {
        return this.request(`/productos/${id}`, {
            method: 'DELETE'
        });
    }

    // ============= VENDEDORES =============

    /**
     * GET /vendedores - Listar vendedores
     */
    async obtenerVendedores() {
        return this.request('/vendedores');
    }

    /**
     * POST /vendedores - Crear vendedor
     */
    async crearVendedor(vendedorData) {
        return this.request('/vendedores', {
            method: 'POST',
            body: JSON.stringify(vendedorData)
        });
    }

    /**
     * PUT /vendedores/:id - Actualizar vendedor
     */
    async actualizarVendedor(id, vendedorData) {
        return this.request(`/vendedores/${id}`, {
            method: 'PUT',
            body: JSON.stringify(vendedorData)
        });
    }

    /**
     * DELETE /vendedores/:id - Eliminar vendedor
     */
    async eliminarVendedor(id) {
        return this.request(`/vendedores/${id}`, {
            method: 'DELETE'
        });
    }
}

// Exportar instancia global
const api = new APIService();

// Exports para testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        APIService,
        api,
        getCookie,
        setCookie,
        eraseCookie
    };
}
