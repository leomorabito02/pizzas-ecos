/**
 * Pruebas unitarias para Controllers
 * Cubre la lógica de negocio de los controladores
 */

// Mock de dependencias globales
global.api = {
    getData: jest.fn(),
    obterProdutos: jest.fn(),
    obterVendedores: jest.fn(),
    obterClientes: jest.fn(),
    crearProducto: jest.fn(),
    actualizarProducto: jest.fn(),
    eliminarProducto: jest.fn(),
    crearVendedor: jest.fn(),
    actualizarVendedor: jest.fn(),
    eliminarVendedor: jest.fn(),
    login: jest.fn(),
    logout: jest.fn(),
    getStoredToken: jest.fn(),
    crearVenta: jest.fn(),
    obtenerVentas: jest.fn(),
    obtenerEstadisticas: jest.fn(),
    actualizarVenta: jest.fn(),
    obtenerProductos: jest.fn(),
    obtenerVendedores: jest.fn()
};

global.appState = {
    cargarDatos: jest.fn(),
    productos: [],
    vendedores: [],
    clientesPorVendedor: {}
};

global.UIUtils = {
    showSpinner: jest.fn(),
    showMessage: jest.fn()
};

// Validators se carga globalmente en jest.setup.js

const {
    VentaController,
    ProductoController,
    VendedorController,
    AuthController
} = require('../controllers.js');

// No importar Validators aquí, usar el global configurado

/**
 * Pruebas unitarias para Controllers
 * Cubre la lógica de negocio de los controladores
 * Patrón AAA: Arrange, Act, Assert
 */

// Mock de dependencias globales
global.api = {
    getData: jest.fn(),
    obtenerProductos: jest.fn(),
    obtenerVendedores: jest.fn(),
    obtenerClientes: jest.fn(),
    crearProducto: jest.fn(),
    actualizarProducto: jest.fn(),
    eliminarProducto: jest.fn(),
    crearVendedor: jest.fn(),
    actualizarVendedor: jest.fn(),
    eliminarVendedor: jest.fn(),
    login: jest.fn(),
    logout: jest.fn(),
    criarVenta: jest.fn(),
    obtenerVentas: jest.fn(),
    obtenerEstadisticas: jest.fn(),
    actualizarVenta: jest.fn(),
    getStoredToken: jest.fn()
};

global.appState = {
    cargarDatos: jest.fn(),
    setVentas: jest.fn(),
    productos: [],
    vendedores: [],
    clientesPorVendedor: {}
};

global.UIUtils = {
    showSpinner: jest.fn(),
    showMessage: jest.fn(),
    validateRequired: jest.fn(),
    validatePositive: jest.fn(),
    confirmAction: jest.fn()
};

describe('VentaController', () => {
  let controller;

  beforeEach(() => {
    // Arrange: Resetear mocks y crear instancia fresca
    resetAllMocks();
    controller = new VentaController();
  });

  describe('cargarDatos', () => {
    test('debe cargar datos exitosamente', async () => {
      // Arrange
      const mockData = {
        productos: [{ id: 1, tipo_pizza: 'Muzza' }],
        vendedores: [{ id: 1, nombre: 'Juan' }],
        clientesPorVendedor: {}
      };
      global.api.getData.mockResolvedValue(mockData);

      // Act
      const result = await controller.cargarDatos();

      // Assert
      expect(global.api.getData).toHaveBeenCalled();
      expect(global.appState.cargarDatos).toHaveBeenCalledWith(mockData);
      expect(global.UIUtils.showSpinner).toHaveBeenCalledWith(true);
      expect(global.UIUtils.showSpinner).toHaveBeenCalledWith(false);
      expect(result).toEqual(mockData);
    });

    test('debe manejar errores de carga', async () => {
      // Arrange
      const error = new Error('Network error');
      global.api.getData.mockRejectedValue(error);

      // Act & Assert
      await expect(controller.cargarDatos()).rejects.toThrow('Network error');
      expect(global.UIUtils.showMessage).toHaveBeenCalledWith('Error cargando datos: Network error', 'error');
      expect(global.UIUtils.showSpinner).toHaveBeenCalledWith(false);
    });
  });

  describe('crearVenta', () => {
    test('debe rechazar venta sin items', async () => {
      // Arrange
      const ventaData = {
        vendedor: 'Juan',
        cliente: 'Maria',
        items: []
      };

      // Act & Assert
      await expect(controller.crearVenta(ventaData)).rejects.toThrow('Productos debe contener al menos un elemento');
    });
  });

  describe('obtenerVentas', () => {
    test('debe obtener ventas exitosamente', async () => {
      // Arrange
      const mockVentas = [
        { id: 1, total: 1000 },
        { id: 2, total: 2000 }
      ];
      global.api.obtenerVentas.mockResolvedValue(mockVentas);

      // Act
      const result = await controller.obtenerVentas();

      // Assert
      expect(global.api.obtenerVentas).toHaveBeenCalled();
      expect(global.appState.setVentas).toHaveBeenCalledWith(mockVentas);
      expect(result).toEqual(mockVentas);
    });
  });

  describe('getClientesPorVendedor', () => {
    test('debe retornar clientes para vendedor existente', () => {
      // Arrange
      const vendedorNombre = 'Juan';
      const clientes = [{ id: 1, nombre: 'Cliente 1' }];
      global.appState.clientesPorVendedor = { [vendedorNombre]: clientes };

      // Act
      const result = controller.getClientesPorVendedor(vendedorNombre);

      // Assert
      expect(result).toEqual(clientes);
    });

    test('debe retornar array vacío para vendedor inexistente', () => {
      // Arrange
      const vendedorNombre = 'Inexistente';

      // Act
      const result = controller.getClientesPorVendedor(vendedorNombre);

      // Assert
      expect(result).toEqual([]);
    });
  });
});

describe('ProductoController', () => {
  let controller;

  beforeEach(() => {
    // Arrange: Resetear mocks y crear instancia fresca
    resetAllMocks();
    controller = new ProductoController();
  });

  describe('crearProducto', () => {
    test('debe crear producto exitosamente con datos válidos', async () => {
      // Arrange
      const productoData = {
        tipo_pizza: 'Nueva Pizza',
        descripcion: 'Deliciosa pizza nueva',
        precio: 1500
      };
      const mockResponse = { id: 1, ...productoData };
      const mockProductos = [mockResponse];
      global.api.crearProducto.mockResolvedValue(mockResponse);
      global.api.obtenerProductos.mockResolvedValue(mockProductos);

      // Act
      const result = await controller.crearProducto(productoData);

      // Assert
      expect(global.api.crearProducto).toHaveBeenCalledWith(productoData);
      expect(global.api.obtenerProductos).toHaveBeenCalled(); // Recarga lista
      expect(global.UIUtils.showMessage).toHaveBeenCalledWith('Producto creado exitosamente', 'success');
      expect(result).toEqual(mockResponse);
    });

    test('debe validar datos requeridos', async () => {
      // Arrange
      const productoData = { precio: 1500 }; // Falta tipo_pizza

      // Act & Assert
      await expect(controller.crearProducto(productoData)).rejects.toThrow('Tipo de pizza es requerido');
      expect(global.api.crearProducto).not.toHaveBeenCalled();
    });

    test('debe manejar errores de creación', async () => {
      // Arrange
      const productoData = {
        tipo_pizza: 'Nueva Pizza',
        descripcion: 'Deliciosa pizza nueva',
        precio: 1500
      };
      const error = new Error('API Error');
      global.api.crearProducto.mockRejectedValue(error);

      // Act & Assert
      await expect(controller.crearProducto(productoData)).rejects.toThrow('API Error');
      expect(global.UIUtils.showMessage).toHaveBeenCalledWith('Error: API Error', 'error');
    });
  });

  describe('obtenerProductos', () => {
    test('debe obtener productos exitosamente', async () => {
      // Arrange
      const mockProductos = [
        { id: 1, tipo_pizza: 'Margarita', precio: 1000 },
        { id: 2, tipo_pizza: 'Pepperoni', precio: 1200 }
      ];
      global.api.obtenerProductos.mockResolvedValue(mockProductos);

      // Act
      const result = await controller.obtenerProductos();

      // Assert
      expect(global.api.obtenerProductos).toHaveBeenCalled();
      expect(global.appState.productos).toEqual(mockProductos);
      expect(result).toEqual(mockProductos);
    });
  });

  describe('eliminarProducto', () => {
    test('debe eliminar producto cuando usuario confirma', async () => {
      // Arrange
      const productId = 1;
      const mockResponse = { success: true };
      const mockProductos = [{ id: 2, tipo_pizza: 'Pepperoni' }];
      global.UIUtils.confirmAction.mockResolvedValue(true);
      global.api.eliminarProducto.mockResolvedValue(mockResponse);
      global.api.obtenerProductos.mockResolvedValue(mockProductos);

      // Act
      const result = await controller.eliminarProducto(productId);

      // Assert
      expect(global.UIUtils.confirmAction).toHaveBeenCalledWith('¿Estás seguro de que deseas eliminar este producto?');
      expect(global.api.eliminarProducto).toHaveBeenCalledWith(productId);
      expect(global.api.obtenerProductos).toHaveBeenCalled(); // Recarga lista
      expect(result).toEqual(mockResponse);
    });

    test('no debe eliminar producto cuando usuario cancela', async () => {
      // Arrange
      const productId = 1;
      global.UIUtils.confirmAction.mockResolvedValue(false);

      // Act
      const result = await controller.eliminarProducto(productId);

      // Assert
      expect(global.UIUtils.confirmAction).toHaveBeenCalled();
      expect(global.api.eliminarProducto).not.toHaveBeenCalled();
      expect(result).toBeUndefined();
    });
  });
});

describe('VendedorController', () => {
  let controller;

  beforeEach(() => {
    // Arrange: Resetear mocks y crear instancia fresca
    resetAllMocks();
    controller = new VendedorController();
  });

  describe('crearVendedor', () => {
    test('debe crear vendedor exitosamente con datos válidos', async () => {
      // Arrange
      const vendedorData = { nombre: 'Nuevo Vendedor' };
      const mockResponse = { id: 1, ...vendedorData };
      const mockVendedores = [mockResponse];
      global.api.crearVendedor.mockResolvedValue(mockResponse);
      global.api.obtenerVendedores.mockResolvedValue(mockVendedores);

      // Act
      const result = await controller.crearVendedor(vendedorData);

      // Assert
      expect(global.api.crearVendedor).toHaveBeenCalledWith(vendedorData);
      expect(global.api.obtenerVendedores).toHaveBeenCalled(); // Recarga lista
      expect(global.UIUtils.showMessage).toHaveBeenCalledWith('Vendedor creado exitosamente', 'success');
      expect(result).toEqual(mockResponse);
    });

    test('debe rechazar nombre demasiado corto', async () => {
      // Arrange
      const vendedorData = { nombre: 'A' };

      // Act & Assert
      await expect(controller.crearVendedor(vendedorData)).rejects.toThrow('Nombre del vendedor debe tener al menos 2 caracteres');
      expect(global.api.crearVendedor).not.toHaveBeenCalled();
    });

    test('debe validar nombre requerido', async () => {
      // Arrange
      const vendedorData = {};

      // Act & Assert
      await expect(controller.crearVendedor(vendedorData)).rejects.toThrow('Nombre del vendedor es requerido');
      expect(global.api.crearVendedor).not.toHaveBeenCalled();
    });
  });

  describe('obtenerVendedores', () => {
    test('debe obtener vendedores exitosamente', async () => {
      // Arrange
      const mockVendedores = [
        { id: 1, nombre: 'Juan Pérez' },
        { id: 2, nombre: 'María García' }
      ];
      global.api.obtenerVendedores.mockResolvedValue(mockVendedores);

      // Act
      const result = await controller.obtenerVendedores();

      // Assert
      expect(global.api.obtenerVendedores).toHaveBeenCalled();
      expect(global.appState.vendedores).toEqual(mockVendedores);
      expect(result).toEqual(mockVendedores);
    });
  });

  describe('eliminarVendedor', () => {
    test('debe eliminar vendedor cuando usuario confirma', async () => {
      // Arrange
      const vendedorId = 1;
      const mockResponse = { success: true };
      const mockVendedores = [{ id: 2, nombre: 'María García' }];
      global.UIUtils.confirmAction.mockResolvedValue(true);
      global.api.eliminarVendedor.mockResolvedValue(mockResponse);
      global.api.obtenerVendedores.mockResolvedValue(mockVendedores);

      // Act
      const result = await controller.eliminarVendedor(vendedorId);

      // Assert
      expect(global.UIUtils.confirmAction).toHaveBeenCalledWith('¿Estás seguro de que deseas eliminar este vendedor?');
      expect(global.api.eliminarVendedor).toHaveBeenCalledWith(vendedorId);
      expect(global.api.obtenerVendedores).toHaveBeenCalled(); // Recarga lista
      expect(result).toEqual(mockResponse);
    });

    test('no debe eliminar vendedor cuando usuario cancela', async () => {
      // Arrange
      const vendedorId = 1;
      global.UIUtils.confirmAction.mockResolvedValue(false);

      // Act
      const result = await controller.eliminarVendedor(vendedorId);

      // Assert
      expect(global.UIUtils.confirmAction).toHaveBeenCalled();
      expect(global.api.eliminarVendedor).not.toHaveBeenCalled();
      expect(result).toBeUndefined();
    });
  });
});

describe('AuthController', () => {
  let controller;

  beforeEach(() => {
    // Arrange: Resetear mocks y crear instancia fresca
    resetAllMocks();
    controller = new AuthController();
  });

  describe('login', () => {
    test('debe hacer login exitosamente con credenciales válidas', async () => {
      // Arrange
      const username = 'admin';
      const password = '1234';
      const mockResponse = {
        token: 'jwt-token',
        user: { id: 1, username: 'admin' }
      };
      global.api.login.mockResolvedValue(mockResponse);

      // Act
      const result = await controller.login(username, password);

      // Assert
      expect(global.api.login).toHaveBeenCalledWith('admin', '1234');
      expect(global.UIUtils.showMessage).toHaveBeenCalledWith('Login exitoso', 'success');
      expect(result).toEqual(mockResponse);
    });

    test('debe validar usuario requerido', async () => {
      // Arrange
      const password = '1234';

      // Act & Assert
      await expect(controller.login('', password)).rejects.toThrow('Usuario es requerido');
      expect(global.api.login).not.toHaveBeenCalled();
    });

    test('debe validar contraseña requerida', async () => {
      // Arrange
      const username = 'admin';

      // Act & Assert
      await expect(controller.login(username, '')).rejects.toThrow('Contraseña es requerido');
      expect(global.api.login).not.toHaveBeenCalled();
    });

    test('debe manejar respuesta inválida del servidor', async () => {
      // Arrange
      const username = 'admin';
      const password = '1234';
      const mockResponse = { error: 'Invalid credentials' };
      global.api.login.mockResolvedValue(mockResponse);

      // Act & Assert
      await expect(controller.login(username, password)).rejects.toThrow('Respuesta inválida del servidor');
      expect(global.UIUtils.showMessage).toHaveBeenCalledWith('Error de login: Respuesta inválida del servidor', 'error');
    });

    test('debe manejar errores de login', async () => {
      // Arrange
      const username = 'admin';
      const password = '1234';
      const error = new Error('Invalid credentials');
      global.api.login.mockRejectedValue(error);

      // Act & Assert
      await expect(controller.login(username, password)).rejects.toThrow('Invalid credentials');
      expect(global.UIUtils.showMessage).toHaveBeenCalledWith('Error de login: Invalid credentials', 'error');
    });
  });

  describe('logout', () => {
    test('debe hacer logout correctamente', () => {
      // Arrange
      const mockLocation = { href: '' };
      delete window.location;
      window.location = mockLocation;

      // Act
      controller.logout();

      // Assert
      expect(global.api.logout).toHaveBeenCalled();
      expect(global.UIUtils.showMessage).toHaveBeenCalledWith('Sesión cerrada', 'info');
      expect(window.location.href).toBe('/login.html');

      // Cleanup
      window.location = location;
    });
  });

  describe('isAuthenticated', () => {
    test('debe retornar true cuando hay token', () => {
      // Arrange
      global.api.getStoredToken.mockReturnValue('valid-token');

      // Act
      const result = controller.isAuthenticated();

      // Assert
      expect(global.api.getStoredToken).toHaveBeenCalled();
      expect(result).toBe(true);
    });

    test('debe retornar false cuando no hay token', () => {
      // Arrange
      global.api.getStoredToken.mockReturnValue(null);

      // Act
      const result = controller.isAuthenticated();

      // Assert
      expect(global.api.getStoredToken).toHaveBeenCalled();
      expect(result).toBe(false);
    });
  });
});