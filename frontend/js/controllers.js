/**
 * Venta Controller - Lógica de negocio para ventas
 */

class VentaController {
    constructor() {
        this.api = api;
        this.state = appState;
    }

    /**
     * Carga datos iniciales (vendedores, clientes, productos)
     */
    async cargarDatos() {
        try {
            UIUtils.showSpinner(true);
            const data = await this.api.getData();
            this.state.cargarDatos(data);
            return data;
        } catch (error) {
            UIUtils.showMessage(`Error cargando datos: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Crea una nueva venta
     */
    async crearVenta(ventaData) {
        try {
            UIUtils.showSpinner(true);

            // Validar datos de venta
            Validators.validateVenta(ventaData);

            const response = await this.api.crearVenta(ventaData);
            UIUtils.showMessage('Venta guardada exitosamente', 'success');
            return response;
        } catch (error) {
            UIUtils.showMessage(`Error: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Obtiene todas las ventas
     */
    async obtenerVentas() {
        try {
            UIUtils.showSpinner(true);
            const ventas = await this.api.obtenerVentas();
            this.state.setVentas(ventas);
            return ventas;
        } catch (error) {
            UIUtils.showMessage(`Error obteniendo ventas: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Obtiene estadísticas
     */
    async obtenerEstadisticas() {
        try {
            UIUtils.showSpinner(true);
            return await this.api.obtenerEstadisticas();
        } catch (error) {
            UIUtils.showMessage(`Error obteniendo estadísticas: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Actualiza una venta
     */
    async actualizarVenta(id, ventaData) {
        try {
            UIUtils.showSpinner(true);
            const response = await this.api.actualizarVenta(id, ventaData);
            UIUtils.showMessage('Venta actualizada exitosamente', 'success');
            return response;
        } catch (error) {
            UIUtils.showMessage(`Error: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Obtiene clientes para un vendedor específico
     */
    getClientesPorVendedor(vendedorNombre) {
        return this.state.clientesPorVendedor[vendedorNombre] || [];
    }

    /**
     * Obtiene un producto por ID
     */
    getProducto(id) {
        return this.state.productos.find(p => p.id === id);
    }

    /**
     * Obtiene todos los productos
     */
    getProductos() {
        return this.state.productos;
    }

    /**
     * Obtiene todos los vendedores
     */
    getVendedores() {
        return this.state.vendedores;
    }
}

/**
 * Producto Controller - Lógica para gestión de productos
 */
class ProductoController {
    constructor() {
        this.api = api;
        this.state = appState;
    }

    /**
     * Crea un nuevo producto
     */
    async crearProducto(productoData) {
        try {
            // Validar datos del producto
            Validators.validateProducto(productoData);

            UIUtils.showSpinner(true);
            const response = await this.api.crearProducto(productoData);
            UIUtils.showMessage('Producto creado exitosamente', 'success');
            
            // Recargar lista
            await this.obtenerProductos();
            return response;
        } catch (error) {
            UIUtils.showMessage(`Error: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Obtiene todos los productos
     */
    async obtenerProductos() {
        try {
            UIUtils.showSpinner(true);
            const productos = await this.api.obtenerProductos();
            this.state.productos = productos;
            return productos;
        } catch (error) {
            UIUtils.showMessage(`Error obteniendo productos: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Actualiza un producto
     */
    async actualizarProducto(id, productoData) {
        try {
            // Validar datos del producto
            Validators.validateProducto(productoData);

            UIUtils.showSpinner(true);
            const response = await this.api.actualizarProducto(id, productoData);
            UIUtils.showMessage('Producto actualizado exitosamente', 'success');
            
            // Recargar lista
            await this.obtenerProductos();
            return response;
        } catch (error) {
            UIUtils.showMessage(`Error: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Elimina un producto
     */
    async eliminarProducto(id) {
        try {
            const confirmado = await UIUtils.confirmAction('¿Estás seguro de que deseas eliminar este producto?');
            if (!confirmado) return;

            UIUtils.showSpinner(true);
            const response = await this.api.eliminarProducto(id);
            UIUtils.showMessage('Producto eliminado exitosamente', 'success');
            
            // Recargar lista
            await this.obtenerProductos();
            return response;
        } catch (error) {
            UIUtils.showMessage(`Error: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }
}

/**
 * Vendedor Controller - Lógica para gestión de vendedores
 */
class VendedorController {
    constructor() {
        this.api = api;
        this.state = appState;
    }

    /**
     * Crea un nuevo vendedor
     */
    async crearVendedor(vendedorData) {
        try {
            // Validar datos del vendedor
            Validators.validateVendedor(vendedorData);

            UIUtils.showSpinner(true);
            const response = await this.api.crearVendedor(vendedorData);
            UIUtils.showMessage('Vendedor creado exitosamente', 'success');
            
            // Recargar lista
            await this.obtenerVendedores();
            return response;
        } catch (error) {
            UIUtils.showMessage(`Error: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Obtiene todos los vendedores
     */
    async obtenerVendedores() {
        try {
            UIUtils.showSpinner(true);
            const vendedores = await this.api.obtenerVendedores();
            this.state.vendedores = vendedores;
            return vendedores;
        } catch (error) {
            UIUtils.showMessage(`Error obteniendo vendedores: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Actualiza un vendedor
     */
    async actualizarVendedor(id, vendedorData) {
        try {
            // Validar datos del vendedor
            Validators.validateVendedor(vendedorData);

            UIUtils.showSpinner(true);
            const response = await this.api.actualizarVendedor(id, vendedorData);
            UIUtils.showMessage('Vendedor actualizado exitosamente', 'success');
            
            // Recargar lista
            await this.obtenerVendedores();
            return response;
        } catch (error) {
            UIUtils.showMessage(`Error: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Elimina un vendedor
     */
    async eliminarVendedor(id) {
        try {
            const confirmado = await UIUtils.confirmAction('¿Estás seguro de que deseas eliminar este vendedor?');
            if (!confirmado) return;

            UIUtils.showSpinner(true);
            const response = await this.api.eliminarVendedor(id);
            UIUtils.showMessage('Vendedor eliminado exitosamente', 'success');
            
            // Recargar lista
            await this.obtenerVendedores();
            return response;
        } catch (error) {
            UIUtils.showMessage(`Error: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }
}

/**
 * Auth Controller - Lógica de autenticación
 */
class AuthController {
    constructor() {
        this.api = api;
    }

    /**
     * Realiza login
     */
    async login(username, password) {
        try {
            // Validar credenciales de login
            Validators.validateLogin({ username, password });

            UIUtils.showSpinner(true);
            const response = await this.api.login(username, password);
            
            if (response && response.token) {
                UIUtils.showMessage('Login exitoso', 'success');
                return response;
            } else {
                throw new Error('Respuesta inválida del servidor');
            }
        } catch (error) {
            UIUtils.showMessage(`Error de login: ${error.message}`, 'error');
            throw error;
        } finally {
            UIUtils.showSpinner(false);
        }
    }

    /**
     * Realiza logout
     */
    logout() {
        this.api.logout();
        UIUtils.showMessage('Sesión cerrada', 'info');
        window.location.href = '/login.html';
    }

    /**
     * Verifica si hay token válido
     */
    isAuthenticated() {
        return !!this.api.getStoredToken();
    }
}

// Instancias globales de los controladores
const ventaController = new VentaController();
const productoController = new ProductoController();
const vendedorController = new VendedorController();
const authController = new AuthController();

// Exports para testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        VentaController,
        ProductoController,
        VendedorController,
        AuthController,
        ventaController,
        productoController,
        vendedorController,
        authController
    };
}
