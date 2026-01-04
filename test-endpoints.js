#!/usr/bin/env node

/**
 * Script de pruebas de integración para endpoints del backend
 * Ejecuta pruebas contra el servidor backend para verificar funcionalidad
 */

const https = require('https');
const http = require('http');

// Configuración
const BASE_URL = process.env.BACKEND_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;

// Colores para output
const colors = {
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
};

let testResults = {
  passed: 0,
  failed: 0,
  total: 0
};

let authToken = null;

/**
 * Función helper para hacer requests HTTP
 */
function makeRequest(url, options = {}) {
  return new Promise((resolve, reject) => {
    const protocol = url.startsWith('https:') ? https : http;
    const urlObj = new URL(url);

    const defaultOptions = {
      hostname: urlObj.hostname,
      port: urlObj.port,
      path: urlObj.pathname + urlObj.search,
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      }
    };

    const reqOptions = { ...defaultOptions, ...options };

    // Agregar token si existe
    if (authToken && !reqOptions.headers.Authorization) {
      reqOptions.headers.Authorization = `Bearer ${authToken}`;
    }

    const req = protocol.request(reqOptions, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        try {
          const jsonData = data ? JSON.parse(data) : null;
          resolve({
            status: res.statusCode,
            headers: res.headers,
            data: jsonData,
            raw: data
          });
        } catch (e) {
          resolve({
            status: res.statusCode,
            headers: res.headers,
            data: null,
            raw: data,
            parseError: e.message
          });
        }
      });
    });

    req.on('error', (err) => {
      reject(err);
    });

    if (options.body) {
      req.write(typeof options.body === 'string' ? options.body : JSON.stringify(options.body));
    }

    req.end();
  });
}

/**
 * Función para ejecutar una prueba
 */
async function runTest(name, testFn) {
  testResults.total++;
  process.stdout.write(`${colors.blue}→${colors.reset} ${name}... `);

  try {
    const result = await testFn();
    if (result === true || result === undefined) {
      testResults.passed++;
      console.log(`${colors.green}✓ PASSED${colors.reset}`);
      return true;
    } else {
      testResults.failed++;
      console.log(`${colors.red}✗ FAILED${colors.reset}`);
      if (typeof result === 'string') {
        console.log(`   ${colors.red}Error: ${result}${colors.reset}`);
      }
      return false;
    }
  } catch (error) {
    testResults.failed++;
    console.log(`${colors.red}✗ ERROR${colors.reset}`);
    console.log(`   ${colors.red}Exception: ${error.message}${colors.reset}`);
    return false;
  }
}

/**
 * Función para validar respuesta
 */
function validateResponse(response, expectedStatus = 200, expectedData = null) {
  if (response.status !== expectedStatus) {
    return `Expected status ${expectedStatus}, got ${response.status}`;
  }

  if (expectedData && response.data) {
    if (typeof expectedData === 'function') {
      if (!expectedData(response.data)) {
        return 'Data validation failed';
      }
    } else if (JSON.stringify(response.data) !== JSON.stringify(expectedData)) {
      return `Data mismatch: expected ${JSON.stringify(expectedData)}, got ${JSON.stringify(response.data)}`;
    }
  }

  return true;
}

/**
 * Pruebas de Health Check
 */
async function testHealthCheck() {
  const response = await makeRequest(`${API_BASE}/health`);
  return validateResponse(response, 200, (data) => data && data.data && data.data.status === 'ok');
}

/**
 * Pruebas de Autenticación
 */
async function testAuthLogin() {
  const loginData = {
    username: 'admin',
    password: 'admin123'
  };

  const response = await makeRequest(`${API_BASE}/auth/login`, {
    method: 'POST',
    body: loginData
  });

  if (response.status === 200 && response.data && response.data.data && response.data.data.token) {
    authToken = response.data.data.token;
    return true;
  }

  return `Login failed: ${response.status} - ${response.raw}`;
}

async function testAuthMe() {
  const response = await makeRequest(`${API_BASE}/auth/me`);
  return validateResponse(response, 200, (data) => data && data.user);
}

/**
 * Pruebas de Productos
 */
async function testGetProductos() {
  const response = await makeRequest(`${API_BASE}/productos`);
  return validateResponse(response, 200, (data) => data && data.data && Array.isArray(data.data));
}

async function testCreateProducto() {
  const productoData = {
    tipo_pizza: 'Test Pizza',
    descripcion: 'Pizza para testing',
    precio: 1500
  };

  const response = await makeRequest(`${API_BASE}/productos`, {
    method: 'POST',
    body: productoData
  });

  return validateResponse(response, 201, (data) => data && data.id);
}

/**
 * Pruebas de Vendedores
 */
async function testGetVendedores() {
  const response = await makeRequest(`${API_BASE}/vendedores`);
  return validateResponse(response, 200, (data) => data && data.data && Array.isArray(data.data));
}

/**
 * Pruebas de Ventas
 */
async function testGetVentas() {
  const response = await makeRequest(`${API_BASE}/estadisticas`);
  return validateResponse(response, 200);
}

async function testGetEstadisticas() {
  const response = await makeRequest(`${API_BASE}/estadisticas-sheet`);
  return validateResponse(response, 200, (data) => data && data.data && typeof data.data === 'object');
}

async function testCreateVenta() {
  // Primero obtener productos y vendedores
  const productosResp = await makeRequest(`${API_BASE}/productos`);
  const vendedoresResp = await makeRequest(`${API_BASE}/vendedores`);

  if (productosResp.data && productosResp.data.data && productosResp.data.data.length > 0 &&
      vendedoresResp.data && vendedoresResp.data.data && vendedoresResp.data.data.length > 0) {

    const ventaData = {
      vendedor: vendedoresResp.data.data[0].nombre,
      cliente: 'Cliente Test',
      telefono_cliente: 123456789,
      items: [{
        product_id: productosResp.data.data[0].id,
        cantidad: 1
      }],
      payment_method: 'efectivo',
      tipo_entrega: 'retiro',
      estado: 'pendiente'
    };

    const response = await makeRequest(`${API_BASE}/ventas`, {
      method: 'POST',
      body: ventaData
    });

    return validateResponse(response, 201, (data) => data && data.id);
  }

  return 'No hay productos o vendedores disponibles para test';
}

/**
 * Pruebas de Datos Iniciales
 */
async function testGetDataInicial() {
  const response = await makeRequest(`${API_BASE}/data`);
  return validateResponse(response, 200, (data) =>
    data && data.data &&
    typeof data.data === 'object' &&
    data.data.vendedores && Array.isArray(data.data.vendedores) &&
    data.data.productos && Array.isArray(data.data.productos)
  );
}

/**
 * Función principal
 */
async function main() {
  console.log(`${colors.bold}${colors.blue}🚀 Iniciando pruebas de integración - Pizzas ECOS${colors.reset}`);
  console.log(`📍 Backend URL: ${BASE_URL}`);
  console.log(`🔗 API Base: ${API_BASE}`);
  console.log('');

  // Pruebas de health check (sin auth)
  console.log(`${colors.bold}🏥 Health Check${colors.reset}`);
  await runTest('Health check endpoint', testHealthCheck);

  // Pruebas de autenticación
  console.log(`\n${colors.bold}🔐 Autenticación${colors.reset}`);
  const loginSuccess = await runTest('Login de usuario', testAuthLogin);

  if (loginSuccess) {
    await runTest('Obtener usuario actual', testAuthMe);
  } else {
    console.log(`${colors.yellow}⚠️  Saltando pruebas que requieren autenticación${colors.reset}`);
  }

  // Pruebas de datos básicos
  console.log(`\n${colors.bold}📊 Datos Básicos${colors.reset}`);
  await runTest('Obtener productos', testGetProductos);
  await runTest('Obtener vendedores', testGetVendedores);
  await runTest('Obtener datos iniciales', testGetDataInicial);

  // Pruebas de ventas
  console.log(`\n${colors.bold}🛒 Ventas${colors.reset}`);
  await runTest('Obtener ventas', testGetVentas);
  await runTest('Obtener estadísticas', testGetEstadisticas);

  if (loginSuccess) {
    await runTest('Crear nueva venta', testCreateVenta);
    await runTest('Crear nuevo producto', testCreateProducto);
  }

  // Resultados finales
  console.log(`\n${colors.bold}📊 Resultados Finales${colors.reset}`);
  console.log(`Total de pruebas: ${testResults.total}`);
  console.log(`${colors.green}Pasaron: ${testResults.passed}${colors.reset}`);
  console.log(`${colors.red}Fallaron: ${testResults.failed}${colors.reset}`);

  const successRate = (testResults.passed / testResults.total * 100).toFixed(1);
  console.log(`Tasa de éxito: ${successRate}%`);

  if (testResults.failed === 0) {
    console.log(`\n${colors.green}🎉 ¡Todas las pruebas pasaron!${colors.reset}`);
    process.exit(0);
  } else {
    console.log(`\n${colors.red}❌ Algunas pruebas fallaron${colors.reset}`);
    process.exit(1);
  }
}

// Ejecutar pruebas
if (require.main === module) {
  main().catch((error) => {
    console.error(`${colors.red}Error fatal:${colors.reset}`, error);
    process.exit(1);
  });
}

module.exports = {
  makeRequest,
  validateResponse,
  runTest
};