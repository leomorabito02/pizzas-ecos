/**
 * Models - DTOs y estructuras de datos
 * Define la forma de los datos de la aplicación
 */

class ProductoItem {
    constructor(productoId, tipo, cantidad, precio) {
        this.product_id = productoId;
        this.tipo = tipo;
        this.cantidad = cantidad;
        this.precio = precio;
        this.total = cantidad * precio;
    }
}

class Producto {
    constructor(id, tipo_pizza, descripcion, precio, activo = true) {
        this.id = id;
        this.tipo_pizza = tipo_pizza;
        this.descripcion = descripcion;
        this.precio = precio;
        this.activo = activo;
    }
}

class Vendedor {
    constructor(id, nombre) {
        this.id = id;
        this.nombre = nombre;
    }
}

class Cliente {
    constructor(id, nombre) {
        this.id = id;
        this.nombre = nombre;
    }
}

class Venta {
    constructor(vendedor, cliente, items = [], paymentMethod = 'efectivo', tipoEntrega = 'retiro') {
        this.vendedor = vendedor;
        this.cliente = cliente;
        this.items = items;
        this.payment_method = paymentMethod;
        this.tipo_entrega = tipoEntrega;
        this.estado = 'pendiente';
        this.total = this.calcularTotal();
    }

    agregarItem(item) {
        this.items.push(item);
        this.total = this.calcularTotal();
    }

    eliminarItem(index) {
        this.items.splice(index, 1);
        this.total = this.calcularTotal();
    }

    calcularTotal() {
        return this.items.reduce((sum, item) => sum + item.total, 0);
    }

    toJSON() {
        return {
            vendedor: this.vendedor,
            cliente: this.cliente,
            items: this.items,
            payment_method: this.payment_method,
            tipo_entrega: this.tipo_entrega,
            estado: this.estado
        };
    }
}

class AppState {
    constructor() {
        this.productos = [];
        this.vendedores = [];
        this.clientesPorVendedor = {};
        this.ventaActual = null;
        this.ventasListado = [];
    }

    cargarDatos(data) {
        this.productos = (data.productos || []).map(p => new Producto(p.id, p.tipo_pizza, p.descripcion, p.precio, p.activo));
        this.vendedores = (data.vendedores || []).map(v => new Vendedor(v.id, v.nombre));
        this.clientesPorVendedor = data.clientesPorVendedor || {};
    }

    crearVenta() {
        this.ventaActual = new Venta('', '', []);
        return this.ventaActual;
    }

    limpiarVenta() {
        this.ventaActual = null;
    }

    setVentas(ventas) {
        this.ventasListado = ventas || [];
    }
}

// Instancia global del estado
const appState = new AppState();

// Exports para testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        ProductoItem,
        Producto,
        Vendedor,
        Cliente,
        Venta,
        AppState,
        appState
    };
}
