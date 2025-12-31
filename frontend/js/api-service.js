/**
 * API Service - Capa de comunicación con backend
 * Centraliza todos los endpoints de API v1
 */

class APIService {
    constructor(baseURL) {
        this._baseURL = baseURL;
        this.token = this.getStoredToken();
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
        // Verificar si window.BACKEND_URL fue establecida por env.js
        if (window.BACKEND_URL) {
            console.log('📡 Usando BACKEND_URL:', window.BACKEND_URL);
            return window.BACKEND_URL;
        }
        
        const isDev = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
        const fallback = isDev ? 'http://localhost:8080/api/v1' : 'https://pizzas-ecos-backend-qa-872448320700.us-central1.run.app/api/v1';
        console.log('📡 Fallback URL:', fallback);
        return fallback;
    }

    /**
     * Obtiene token JWT del sessionStorage (más seguro que localStorage)
     */
    getStoredToken() {
        return sessionStorage.getItem('authToken');
    }

    /**
     * Guarda token JWT en sessionStorage
     * Nota: sessionStorage se limpia automáticamente al cerrar la pestaña
     */
    setToken(token) {
        this.token = token;
        if (token) {
            sessionStorage.setItem('authToken', token);
        } else {
            sessionStorage.removeItem('authToken');
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
     * Wrapper para fetch con manejo de errores
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
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
    async obtenerVentas() {
        return this.request('/estadisticas');
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
    async obtenerEstadisticas() {
        return this.request('/estadisticas-sheet');
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
