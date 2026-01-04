// jest.setup.js - Configuración global para pruebas Jest
// Configura el entorno de pruebas para el frontend

// Mock de fetch global para pruebas
global.fetch = jest.fn();

// Mock de localStorage y sessionStorage
const createMockStorage = () => {
  let storage = {};
  return {
    getItem: jest.fn(key => storage[key] || null),
    setItem: jest.fn((key, value) => {
      storage[key] = value.toString();
    }),
    removeItem: jest.fn(key => {
      delete storage[key];
    }),
    clear: jest.fn(() => {
      storage = {};
    }),
    key: jest.fn(index => Object.keys(storage)[index] || null),
    get length() {
      return Object.keys(storage).length;
    }
  };
};

Object.defineProperty(window, 'localStorage', {
  value: createMockStorage(),
  writable: true
});

Object.defineProperty(window, 'sessionStorage', {
  value: createMockStorage(),
  writable: true
});

// Mock de console para reducir ruido en pruebas
global.console = {
  ...console,
  // Mantener logs de error pero silenciar otros
  log: jest.fn(),
  warn: jest.fn(),
  info: jest.fn(),
  debug: jest.fn(),
};

// Mock de window.location
delete window.location;
window.location = {
  hostname: 'localhost',
  protocol: 'http:',
  port: '5000',
  href: 'http://localhost:5000',
  pathname: '/',
  search: '',
  hash: ''
};

// Mock de document para pruebas DOM
Object.defineProperty(document, 'querySelector', {
  writable: true,
  value: jest.fn()
});

Object.defineProperty(document, 'querySelectorAll', {
  writable: true,
  value: jest.fn(() => [])
});

Object.defineProperty(document, 'getElementById', {
  writable: true,
  value: jest.fn()
});

Object.defineProperty(document, 'createElement', {
  writable: true,
  value: jest.fn(() => ({
    classList: {
      add: jest.fn(),
      remove: jest.fn(),
      contains: jest.fn(),
      toggle: jest.fn()
    },
    style: {},
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    setAttribute: jest.fn(),
    getAttribute: jest.fn(),
    appendChild: jest.fn(),
    removeChild: jest.fn(),
    innerHTML: '',
    textContent: '',
    value: ''
  }))
});

Object.defineProperty(document, 'addEventListener', {
  writable: true,
  value: jest.fn()
});

Object.defineProperty(document, 'removeEventListener', {
  writable: true,
  value: jest.fn()
});

// Mock de DOMParser para pruebas XML/HTML
global.DOMParser = class {
  parseFromString(str, contentType) {
    return {
      documentElement: {
        textContent: str
      }
    };
  }
};

// Helper para resetear mocks entre pruebas
global.resetAllMocks = () => {
  jest.clearAllMocks();
  fetch.mockClear();
  localStorage.clear();
  sessionStorage.clear();
  console.log.mockClear();
  console.warn.mockClear();
  console.info.mockClear();
  console.debug.mockClear();
};

// Helper para crear respuesta fetch mock
global.createMockResponse = (data, status = 200, statusText = 'OK') => ({
  ok: status >= 200 && status < 300,
  status,
  statusText,
  json: jest.fn().mockResolvedValue(data),
  text: jest.fn().mockResolvedValue(JSON.stringify(data)),
  headers: new Map([['content-type', 'application/json']])
});

// Helper para crear error de red
global.createNetworkError = () => {
  const error = new Error('Network Error');
  error.name = 'TypeError';
  return error;
};

// Cargar Validators globalmente para pruebas
const { Validators } = require('./js/validators.js');
global.Validators = Validators;