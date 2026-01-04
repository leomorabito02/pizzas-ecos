/**
 * Validadores - Funciones de validación de datos
 * Centraliza toda la lógica de validación del frontend
 */

class Validators {
    /**
     * Valida que un campo no esté vacío
     */
    static required(value, fieldName) {
        if (!value || value.toString().trim() === '') {
            throw new Error(`${fieldName} es requerido`);
        }
        return true;
    }

    /**
     * Valida que un número sea positivo
     */
    static positive(value, fieldName) {
        const num = parseFloat(value);
        if (isNaN(num) || num <= 0) {
            throw new Error(`${fieldName} debe ser mayor a 0`);
        }
        return true;
    }

    /**
     * Valida que un número esté dentro de un rango
     */
    static range(value, min, max, fieldName) {
        const num = parseFloat(value);
        if (isNaN(num)) {
            throw new Error(`${fieldName} debe ser un número válido`);
        }
        if (num < min || num > max) {
            throw new Error(`${fieldName} debe estar entre ${min} y ${max}`);
        }
        return true;
    }

    /**
     * Valida longitud mínima de string
     */
    static minLength(value, minLength, fieldName) {
        if (!value || value.toString().length < minLength) {
            throw new Error(`${fieldName} debe tener al menos ${minLength} caracteres`);
        }
        return true;
    }

    /**
     * Valida longitud máxima de string
     */
    static maxLength(value, maxLength, fieldName) {
        if (value && value.toString().length > maxLength) {
            throw new Error(`${fieldName} no puede tener más de ${maxLength} caracteres`);
        }
        return true;
    }

    /**
     * Valida formato de email
     */
    static email(value, fieldName) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!value || !emailRegex.test(value)) {
            throw new Error(`${fieldName} debe ser un email válido`);
        }
        return true;
    }

    /**
     * Valida formato de teléfono argentino
     */
    static phone(value, fieldName) {
        // Acepta formatos: +54912345678, 01112345678, 12345678
        const phoneRegex = /^(\+?549|0)?[1-9][0-9]{9,10}$/;
        if (!value || !phoneRegex.test(value.replace(/[\s\-\(\)]/g, ''))) {
            throw new Error(`${fieldName} debe ser un teléfono válido`);
        }
        return true;
    }

    /**
     * Valida formato de CUIT/CUIL argentino
     */
    static cuit(value, fieldName) {
        if (!value) return true; // Opcional

        const cuitStr = value.toString().replace(/[^\d]/g, '');
        if (cuitStr.length !== 11) {
            throw new Error(`${fieldName} debe tener 11 dígitos`);
        }

        // Algoritmo de validación de CUIT
        const multipliers = [5, 4, 3, 2, 7, 6, 5, 4, 3, 2];
        let sum = 0;
        for (let i = 0; i < 10; i++) {
            sum += parseInt(cuitStr[i]) * multipliers[i];
        }
        const remainder = sum % 11;
        const checkDigit = remainder === 0 ? 0 : remainder === 1 ? 9 : 11 - remainder;

        if (parseInt(cuitStr[10]) !== checkDigit) {
            throw new Error(`${fieldName} no es válido`);
        }
        return true;
    }

    /**
     * Valida que un array no esté vacío
     */
    static notEmptyArray(value, fieldName) {
        if (!Array.isArray(value) || value.length === 0) {
            throw new Error(`${fieldName} debe contener al menos un elemento`);
        }
        return true;
    }

    /**
     * Valida formato de precio (acepta comas y puntos)
     */
    static price(value, fieldName) {
        if (!value) {
            throw new Error(`${fieldName} es requerido`);
        }

        const priceStr = value.toString().replace(/[$,\s]/g, '');
        const priceRegex = /^\d+(\.\d{1,2})?$/;

        if (!priceRegex.test(priceStr)) {
            throw new Error(`${fieldName} debe ser un precio válido (ej: 123.45)`);
        }

        const num = parseFloat(priceStr);
        if (num <= 0) {
            throw new Error(`${fieldName} debe ser mayor a 0`);
        }

        return true;
    }

    /**
     * Valida que una fecha no sea futura
     */
    static notFutureDate(value, fieldName) {
        const date = new Date(value);
        const now = new Date();

        if (date > now) {
            throw new Error(`${fieldName} no puede ser una fecha futura`);
        }
        return true;
    }

    /**
     * Valida que un ID exista en una lista
     */
    static existsInList(value, list, fieldName) {
        if (!list || !list.includes(value)) {
            throw new Error(`${fieldName} seleccionado no es válido`);
        }
        return true;
    }

    /**
     * Valida datos de producto
     */
    static validateProducto(productoData) {
        this.required(productoData.tipo_pizza, 'Tipo de pizza');
        this.minLength(productoData.tipo_pizza, 2, 'Tipo de pizza');
        this.maxLength(productoData.tipo_pizza, 100, 'Tipo de pizza');

        this.required(productoData.precio, 'Precio');
        this.price(productoData.precio, 'Precio');

        if (productoData.descripcion) {
            this.maxLength(productoData.descripcion, 500, 'Descripción');
        }
    }

    /**
     * Valida datos de vendedor
     */
    static validateVendedor(vendedorData) {
        this.required(vendedorData.nombre, 'Nombre del vendedor');
        this.minLength(vendedorData.nombre, 2, 'Nombre del vendedor');
        this.maxLength(vendedorData.nombre, 100, 'Nombre del vendedor');

        if (vendedorData.email) {
            this.email(vendedorData.email, 'Email');
        }

        if (vendedorData.telefono) {
            this.phone(vendedorData.telefono, 'Teléfono');
        }

        if (vendedorData.cuit) {
            this.cuit(vendedorData.cuit, 'CUIT');
        }
    }

    /**
     * Valida datos de venta
     */
    static validateVenta(ventaData) {
        this.required(ventaData.vendedor, 'Vendedor');
        this.required(ventaData.cliente, 'Cliente');
        this.notEmptyArray(ventaData.items, 'Productos');

        if (ventaData.items) {
            ventaData.items.forEach((item, index) => {
                if (!item.product_id) {
                    throw new Error(`Producto ${index + 1}: ID requerido`);
                }
                if (!item.cantidad || item.cantidad <= 0) {
                    throw new Error(`Producto ${index + 1}: Cantidad debe ser mayor a 0`);
                }
            });
        }

        if (ventaData.payment_method) {
            const validMethods = ['efectivo', 'tarjeta', 'transferencia'];
            if (!validMethods.includes(ventaData.payment_method)) {
                throw new Error('Método de pago no válido');
            }
        }

        if (ventaData.tipo_entrega) {
            const validTypes = ['retiro', 'delivery'];
            if (!validTypes.includes(ventaData.tipo_entrega)) {
                throw new Error('Tipo de entrega no válido');
            }
        }
    }

    /**
     * Valida credenciales de login
     */
    static validateLogin(loginData) {
        this.required(loginData.username, 'Usuario');
        this.minLength(loginData.username, 3, 'Usuario');

        this.required(loginData.password, 'Contraseña');
        this.minLength(loginData.password, 4, 'Contraseña');
    }

    /**
     * Valida datos de producto
     */
    static validateProducto(productoData) {
        this.required(productoData.tipo_pizza, 'Tipo de pizza');
        this.minLength(productoData.tipo_pizza, 2, 'Tipo de pizza');
        this.maxLength(productoData.tipo_pizza, 100, 'Tipo de pizza');

        this.required(productoData.precio, 'Precio');
        this.price(productoData.precio, 'Precio');
    }

    /**
     * Valida datos de vendedor
     */
    static validateVendedor(vendedorData) {
        this.required(vendedorData.nombre, 'Nombre del vendedor');
        this.minLength(vendedorData.nombre, 2, 'Nombre del vendedor');
        this.maxLength(vendedorData.nombre, 100, 'Nombre del vendedor');

        if (vendedorData.cuit) {
            this.cuit(vendedorData.cuit, 'CUIT');
        }

        if (vendedorData.email) {
            this.email(vendedorData.email, 'Email');
        }

        if (vendedorData.telefono) {
            this.phone(vendedorData.telefono, 'Teléfono');
        }
    }
}

// Export para testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { Validators };
}