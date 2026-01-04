/**
 * Pruebas unitarias para Validadores
 * Cubre todas las funciones de validación de datos
 * Patrón AAA: Arrange, Act, Assert
 */

const { Validators } = require('../validators.js');

describe('Validators', () => {
  describe('required', () => {
    test('debe pasar con valor válido', () => {
      // Arrange & Act & Assert
      expect(() => Validators.required('test', 'Campo')).not.toThrow();
    });

    test('debe fallar con valor vacío', () => {
      // Arrange & Act & Assert
      expect(() => Validators.required('', 'Campo')).toThrow('Campo es requerido');
    });

    test('debe fallar con valor null', () => {
      // Arrange & Act & Assert
      expect(() => Validators.required(null, 'Campo')).toThrow('Campo es requerido');
    });

    test('debe fallar con valor undefined', () => {
      // Arrange & Act & Assert
      expect(() => Validators.required(undefined, 'Campo')).toThrow('Campo es requerido');
    });
  });

  describe('positive', () => {
    test('debe pasar con número positivo', () => {
      // Arrange & Act & Assert
      expect(() => Validators.positive(10, 'Precio')).not.toThrow();
    });

    test('debe pasar con string numérico positivo', () => {
      // Arrange & Act & Assert
      expect(() => Validators.positive('15.5', 'Precio')).not.toThrow();
    });

    test('debe fallar con cero', () => {
      // Arrange & Act & Assert
      expect(() => Validators.positive(0, 'Precio')).toThrow('Precio debe ser mayor a 0');
    });

    test('debe fallar con número negativo', () => {
      // Arrange & Act & Assert
      expect(() => Validators.positive(-5, 'Precio')).toThrow('Precio debe ser mayor a 0');
    });

    test('debe fallar con NaN', () => {
      // Arrange & Act & Assert
      expect(() => Validators.positive('abc', 'Precio')).toThrow('Precio debe ser mayor a 0');
    });
  });

  describe('range', () => {
    test('debe pasar con número dentro del rango', () => {
      // Arrange & Act & Assert
      expect(() => Validators.range(5, 1, 10, 'Edad')).not.toThrow();
    });

    test('debe fallar con número menor al mínimo', () => {
      // Arrange & Act & Assert
      expect(() => Validators.range(0, 1, 10, 'Edad')).toThrow('Edad debe estar entre 1 y 10');
    });

    test('debe fallar con número mayor al máximo', () => {
      // Arrange & Act & Assert
      expect(() => Validators.range(15, 1, 10, 'Edad')).toThrow('Edad debe estar entre 1 y 10');
    });
  });

  describe('minLength', () => {
    test('debe pasar con longitud suficiente', () => {
      // Arrange & Act & Assert
      expect(() => Validators.minLength('hello', 3, 'Nombre')).not.toThrow();
    });

    test('debe fallar con longitud insuficiente', () => {
      // Arrange & Act & Assert
      expect(() => Validators.minLength('hi', 3, 'Nombre')).toThrow('Nombre debe tener al menos 3 caracteres');
    });
  });

  describe('maxLength', () => {
    test('debe pasar con longitud aceptable', () => {
      // Arrange & Act & Assert
      expect(() => Validators.maxLength('hello', 10, 'Nombre')).not.toThrow();
    });

    test('debe fallar con longitud excesiva', () => {
      // Arrange & Act & Assert
      expect(() => Validators.maxLength('this is a very long string', 10, 'Nombre')).toThrow('Nombre no puede tener más de 10 caracteres');
    });
  });

  describe('email', () => {
    test('debe pasar con email válido', () => {
      // Arrange & Act & Assert
      expect(() => Validators.email('test@example.com', 'Email')).not.toThrow();
    });

    test('debe fallar con email inválido', () => {
      // Arrange & Act & Assert
      expect(() => Validators.email('invalid-email', 'Email')).toThrow('Email debe ser un email válido');
    });

    test('debe fallar con email vacío', () => {
      // Arrange & Act & Assert
      expect(() => Validators.email('', 'Email')).toThrow('Email debe ser un email válido');
    });
  });

  describe('phone', () => {
    test('debe pasar con teléfono válido (11 dígitos)', () => {
      // Arrange & Act & Assert
      expect(() => Validators.phone('1123456789', 'Teléfono')).not.toThrow();
    });

    test('debe pasar con teléfono válido (+549)', () => {
      // Arrange & Act & Assert
      expect(() => Validators.phone('+5491123456789', 'Teléfono')).not.toThrow();
    });

    test('debe fallar con teléfono inválido', () => {
      // Arrange & Act & Assert
      expect(() => Validators.phone('123', 'Teléfono')).toThrow('Teléfono debe ser un teléfono válido');
    });
  });

  describe('cuit', () => {
    test('debe pasar con CUIT válido', () => {
      // Arrange & Act & Assert
      expect(() => Validators.cuit('20267565393', 'CUIT')).not.toThrow();
    });

    test('debe fallar con CUIT inválido', () => {
      // Arrange & Act & Assert
      expect(() => Validators.cuit('11111111111', 'CUIT')).toThrow('CUIT no es válido');
    });

    test('debe fallar con CUIT de longitud incorrecta', () => {
      // Arrange & Act & Assert
      expect(() => Validators.cuit('12345678', 'CUIT')).toThrow('CUIT debe tener 11 dígitos');
    });

    test('debe pasar con CUIT vacío (opcional)', () => {
      // Arrange & Act & Assert
      expect(() => Validators.cuit('', 'CUIT')).not.toThrow();
    });
  });

  describe('notEmptyArray', () => {
    test('debe pasar con array no vacío', () => {
      // Arrange & Act & Assert
      expect(() => Validators.notEmptyArray([1, 2, 3], 'Items')).not.toThrow();
    });

    test('debe fallar con array vacío', () => {
      // Arrange & Act & Assert
      expect(() => Validators.notEmptyArray([], 'Items')).toThrow('Items debe contener al menos un elemento');
    });

    test('debe fallar con no array', () => {
      // Arrange & Act & Assert
      expect(() => Validators.notEmptyArray('not array', 'Items')).toThrow('Items debe contener al menos un elemento');
    });
  });

  describe('price', () => {
    test('debe pasar con precio válido', () => {
      // Arrange & Act & Assert
      expect(() => Validators.price('123.45', 'Precio')).not.toThrow();
    });

    test('debe pasar con precio entero', () => {
      // Arrange & Act & Assert
      expect(() => Validators.price('100', 'Precio')).not.toThrow();
    });

    test('debe fallar con precio cero', () => {
      // Arrange & Act & Assert
      expect(() => Validators.price('0', 'Precio')).toThrow('Precio debe ser mayor a 0');
    });

    test('debe fallar con precio negativo', () => {
      // Arrange & Act & Assert
      expect(() => Validators.price('-10', 'Precio')).toThrow('Precio debe ser un precio válido (ej: 123.45)');
    });

    test('debe fallar con formato inválido', () => {
      // Arrange & Act & Assert
      expect(() => Validators.price('abc', 'Precio')).toThrow('Precio debe ser un precio válido');
    });
  });

  describe('notFutureDate', () => {
    test('debe pasar con fecha pasada', () => {
      // Arrange
      const pastDate = new Date();
      pastDate.setDate(pastDate.getDate() - 1);

      // Act & Assert
      expect(() => Validators.notFutureDate(pastDate.toISOString(), 'Fecha')).not.toThrow();
    });

    test('debe fallar con fecha futura', () => {
      // Arrange
      const futureDate = new Date();
      futureDate.setDate(futureDate.getDate() + 1);

      // Act & Assert
      expect(() => Validators.notFutureDate(futureDate.toISOString(), 'Fecha')).toThrow('Fecha no puede ser una fecha futura');
    });
  });

  describe('existsInList', () => {
    test('debe pasar con valor existente en lista', () => {
      // Arrange
      const list = ['opcion1', 'opcion2', 'opcion3'];

      // Act & Assert
      expect(() => Validators.existsInList('opcion2', list, 'Opción')).not.toThrow();
    });

    test('debe fallar con valor no existente en lista', () => {
      // Arrange
      const list = ['opcion1', 'opcion2', 'opcion3'];

      // Act & Assert
      expect(() => Validators.existsInList('opcion4', list, 'Opción')).toThrow('Opción seleccionado no es válido');
    });
  });

  describe('validateProducto', () => {
    test('debe pasar con datos válidos', () => {
      // Arrange
      const productoData = {
        tipo_pizza: 'Margarita',
        precio: '150.50',
        descripcion: 'Pizza deliciosa'
      };

      // Act & Assert
      expect(() => Validators.validateProducto(productoData)).not.toThrow();
    });

    test('debe fallar sin tipo_pizza', () => {
      // Arrange
      const productoData = {
        precio: '150.50'
      };

      // Act & Assert
      expect(() => Validators.validateProducto(productoData)).toThrow('Tipo de pizza es requerido');
    });

    test('debe fallar con precio inválido', () => {
      // Arrange
      const productoData = {
        tipo_pizza: 'Margarita',
        precio: '0'
      };

      // Act & Assert
      expect(() => Validators.validateProducto(productoData)).toThrow('Precio debe ser mayor a 0');
    });
  });

  describe('validateVendedor', () => {
    test('debe pasar con datos válidos', () => {
      // Arrange
      const vendedorData = {
        nombre: 'Juan Pérez',
        email: 'juan@example.com',
        telefono: '1123456789',
        cuit: '20267565393'
      };

      // Act & Assert
      expect(() => Validators.validateVendedor(vendedorData)).not.toThrow();
    });

    test('debe fallar sin nombre', () => {
      // Arrange
      const vendedorData = {
        email: 'juan@example.com'
      };

      // Act & Assert
      expect(() => Validators.validateVendedor(vendedorData)).toThrow('Nombre del vendedor es requerido');
    });

    test('debe fallar con email inválido', () => {
      // Arrange
      const vendedorData = {
        nombre: 'Juan Pérez',
        email: 'invalid-email'
      };

      // Act & Assert
      expect(() => Validators.validateVendedor(vendedorData)).toThrow('Email debe ser un email válido');
    });
  });

  describe('validateVenta', () => {
    test('debe pasar con datos válidos', () => {
      // Arrange
      const ventaData = {
        vendedor: 'Juan',
        cliente: 'Maria',
        items: [
          { product_id: 1, cantidad: 2 },
          { product_id: 2, cantidad: 1 }
        ],
        payment_method: 'efectivo',
        tipo_entrega: 'retiro'
      };

      // Act & Assert
      expect(() => Validators.validateVenta(ventaData)).not.toThrow();
    });

    test('debe fallar sin items', () => {
      // Arrange
      const ventaData = {
        vendedor: 'Juan',
        cliente: 'Maria',
        items: []
      };

      // Act & Assert
      expect(() => Validators.validateVenta(ventaData)).toThrow('Productos debe contener al menos un elemento');
    });

    test('debe fallar con método de pago inválido', () => {
      // Arrange
      const ventaData = {
        vendedor: 'Juan',
        cliente: 'Maria',
        items: [{ product_id: 1, cantidad: 2 }],
        payment_method: 'bitcoin'
      };

      // Act & Assert
      expect(() => Validators.validateVenta(ventaData)).toThrow('Método de pago no válido');
    });
  });

  describe('validateLogin', () => {
    test('debe pasar con credenciales válidas', () => {
      // Arrange & Act & Assert
      expect(() => Validators.validateLogin({ username: 'admin', password: 'password123' })).not.toThrow();
    });

    test('debe fallar con usuario corto', () => {
      // Arrange & Act & Assert
      expect(() => Validators.validateLogin({ username: 'ad', password: 'password123' })).toThrow('Usuario debe tener al menos 3 caracteres');
    });

    test('debe fallar con contraseña corta', () => {
      // Arrange & Act & Assert
      expect(() => Validators.validateLogin({ username: 'admin', password: '123' })).toThrow('Contraseña debe tener al menos 4 caracteres');
    });
  });
});