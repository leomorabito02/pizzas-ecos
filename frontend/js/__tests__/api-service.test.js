/**
 * Pruebas unitarias para APIService
 * Cubre la lógica de comunicación con el backend
 * Patrón AAA: Arrange, Act, Assert
 */

const { APIService } = require('../api-service.js');

describe('APIService', () => {
  let apiService;

  beforeEach(() => {
    // Arrange: Resetear mocks y crear instancia fresca
    resetAllMocks();
    apiService = new APIService();
    // Limpiar cualquier token persistente
    sessionStorage.getItem.mockReturnValue(null);
    // Spy on setToken method
    jest.spyOn(apiService, 'setToken');
  });

  describe('Inicialización', () => {
    test('debe inicializarse correctamente', () => {
      // Arrange & Act
      const service = new APIService();

      // Assert
      expect(service).toBeInstanceOf(APIService);
      expect(typeof service.baseURL).toBe('string');
    });

    test('debe determinar URL correcta para desarrollo', () => {
      // Arrange
      window.location.hostname = 'localhost';

      // Act
      const service = new APIService();

      // Assert
      expect(service.baseURL).toContain('localhost:8080');
    });
  });

  describe('Gestión de Tokens', () => {
    test('debe guardar token en sessionStorage', () => {
      // Arrange
      const token = 'test-jwt-token';

      // Act
      apiService.setToken(token);

      // Assert
      expect(sessionStorage.setItem).toHaveBeenCalledWith('authToken', token);
    });

    test('debe obtener token de sessionStorage', () => {
      // Arrange
      const token = 'test-jwt-token';
      sessionStorage.getItem.mockReturnValue(token);

      // Act
      const result = apiService.getStoredToken();

      // Assert
      expect(result).toBe(token);
      expect(sessionStorage.getItem).toHaveBeenCalledWith('authToken');
    });

    test('debe retornar null si no hay token', () => {
      // Arrange
      sessionStorage.getItem.mockReturnValue(null);

      // Act
      const result = apiService.getStoredToken();

      // Assert
      expect(result).toBeNull();
    });
  });

  describe('Método request', () => {
    test('debe hacer request GET exitosamente', async () => {
      // Arrange
      sessionStorage.getItem.mockReturnValue(null); // Asegurar no hay token
      const testData = { success: true };
      const mockResponse = {
        ok: true,
        json: jest.fn().mockResolvedValue(testData)
      };
      global.fetch.mockResolvedValue(mockResponse);

      // Act
      const result = await apiService.request('/test-endpoint');

      // Assert
      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/test-endpoint',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      );
      expect(result).toEqual(testData);
    });

    test('debe manejar errores de red', async () => {
      // Arrange
      const networkError = new Error('Network Error');
      global.fetch.mockRejectedValue(networkError);

      // Act & Assert
      await expect(apiService.request('/test-endpoint')).rejects.toThrow('Network Error');
    });

    test('debe manejar respuestas de error HTTP', async () => {
      // Arrange
      const mockResponse = {
        ok: false,
        status: 404,
        json: jest.fn().mockResolvedValue({ error: 'Not found' })
      };
      global.fetch.mockResolvedValue(mockResponse);

      // Act & Assert
      await expect(apiService.request('/test-endpoint')).rejects.toThrow();
    });
  });

  describe('Autenticación', () => {
    test('login debe enviar credenciales y guardar token', async () => {
      // Arrange
      const credentials = { username: 'admin', password: '123' };
      const mockResponse = {
        token: 'jwt-token',
        user: { id: 1, username: 'admin' }
      };
      const apiResponse = {
        ok: true,
        json: jest.fn().mockResolvedValue(mockResponse)
      };
      global.fetch.mockResolvedValue(apiResponse);

      // Act
      const result = await apiService.login(credentials.username, credentials.password);

      // Assert
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/auth/login'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(credentials)
        })
      );
      expect(sessionStorage.setItem).toHaveBeenCalledWith('authToken', 'jwt-token');
      expect(result).toEqual(mockResponse);
    });

    test('logout debe limpiar token', () => {
      // Arrange & Act
      apiService.logout();

      // Assert
      expect(apiService.setToken).toHaveBeenCalledWith(null);
    });
  });

  describe('Operaciones de Datos', () => {
    test('obtenerProductos debe hacer request correcto', async () => {
      // Arrange
      sessionStorage.getItem.mockReturnValue(null); // Asegurar no hay token
      const mockProductos = [
        { id: 1, tipo_pizza: 'Margarita' },
        { id: 2, tipo_pizza: 'Pepperoni' }
      ];
      const mockResponse = {
        ok: true,
        json: jest.fn().mockResolvedValue(mockProductos)
      };
      global.fetch.mockResolvedValue(mockResponse);

      // Act
      const result = await apiService.obtenerProductos();

      // Assert
      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/productos',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      );
      expect(result).toEqual(mockProductos);
    });

    test('obtenerVentas debe hacer request correcto', async () => {
      // Arrange
      sessionStorage.getItem.mockReturnValue(null); // Asegurar no hay token
      const mockVentas = [
        { id: 1, total: 1000 },
        { id: 2, total: 2000 }
      ];
      const mockResponse = {
        ok: true,
        json: jest.fn().mockResolvedValue(mockVentas)
      };
      global.fetch.mockResolvedValue(mockResponse);

      // Act
      const result = await apiService.obtenerVentas();

      // Assert
      expect(global.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/estadisticas',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      );
      expect(result).toEqual(mockVentas);
    });
  });
});