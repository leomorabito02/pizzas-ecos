/**
 * Pruebas unitarias para Models
 * Cubre las clases de datos y su lógica
 */

const {
  ProductoItem,
  Producto,
  Vendedor,
  Cliente,
  Venta,
  AppState
} = require('../models.js');

describe('ProductoItem', () => {
  test('debe crear instancia correctamente', () => {
    const item = new ProductoItem(1, 'Muzza', 2, 1000);

    expect(item.product_id).toBe(1);
    expect(item.tipo).toBe('Muzza');
    expect(item.cantidad).toBe(2);
    expect(item.precio).toBe(1000);
    expect(item.total).toBe(2000);
  });

  test('debe calcular total correctamente', () => {
    const item = new ProductoItem(1, 'Muzza', 3, 500);
    expect(item.total).toBe(1500);
  });

  test('debe manejar precio cero', () => {
    const item = new ProductoItem(1, 'Muzza', 1, 0);
    expect(item.total).toBe(0);
  });
});

describe('Producto', () => {
  test('debe crear instancia con valores por defecto', () => {
    const producto = new Producto(1, 'Muzza', 'Pizza muzza', 1000);

    expect(producto.id).toBe(1);
    expect(producto.tipo_pizza).toBe('Muzza');
    expect(producto.descripcion).toBe('Pizza muzza');
    expect(producto.precio).toBe(1000);
    expect(producto.activo).toBe(true);
  });

  test('debe permitir desactivar producto', () => {
    const producto = new Producto(1, 'Muzza', 'Pizza muzza', 1000, false);

    expect(producto.activo).toBe(false);
  });
});

describe('Vendedor', () => {
  test('debe crear instancia correctamente', () => {
    const vendedor = new Vendedor(1, 'Juan Pérez');

    expect(vendedor.id).toBe(1);
    expect(vendedor.nombre).toBe('Juan Pérez');
  });
});

describe('Cliente', () => {
  test('debe crear instancia correctamente', () => {
    const cliente = new Cliente(1, 'María García');

    expect(cliente.id).toBe(1);
    expect(cliente.nombre).toBe('María García');
  });
});

describe('Venta', () => {
  let venta;

  beforeEach(() => {
    venta = new Venta('Juan', 'María', [], 'efectivo', 'delivery');
  });

  test('debe crear instancia con valores por defecto', () => {
    expect(venta.vendedor).toBe('Juan');
    expect(venta.cliente).toBe('María');
    expect(venta.items).toEqual([]);
    expect(venta.payment_method).toBe('efectivo');
    expect(venta.tipo_entrega).toBe('delivery');
    expect(venta.estado).toBe('pendiente');
    expect(venta.total).toBe(0);
  });

  test('debe agregar items correctamente', () => {
    const item = new ProductoItem(1, 'Muzza', 2, 1000);
    venta.agregarItem(item);

    expect(venta.items).toHaveLength(1);
    expect(venta.items[0]).toBe(item);
    expect(venta.total).toBe(2000);
  });

  test('debe calcular total acumulado', () => {
    const item1 = new ProductoItem(1, 'Muzza', 1, 1000);
    const item2 = new ProductoItem(2, 'Napo', 2, 1200);

    venta.agregarItem(item1);
    venta.agregarItem(item2);

    expect(venta.total).toBe(3400); // 1000 + (1200 * 2)
  });

  test('debe eliminar items correctamente', () => {
    const item1 = new ProductoItem(1, 'Muzza', 1, 1000);
    const item2 = new ProductoItem(2, 'Napo', 1, 1200);

    venta.agregarItem(item1);
    venta.agregarItem(item2);
    venta.eliminarItem(0);

    expect(venta.items).toHaveLength(1);
    expect(venta.items[0].tipo).toBe('Napo');
    expect(venta.total).toBe(1200);
  });

  test('debe manejar eliminar item fuera de rango', () => {
    const item = new ProductoItem(1, 'Muzza', 1, 1000);
    venta.agregarItem(item);

    venta.eliminarItem(5); // Índice inválido

    expect(venta.items).toHaveLength(1); // No se eliminó nada porque el índice no existe
  });

  test('debe convertir a JSON correctamente', () => {
    const item = new ProductoItem(1, 'Muzza', 1, 1000);
    venta.agregarItem(item);

    const json = venta.toJSON();

    expect(json.vendedor).toBe('Juan');
    expect(json.cliente).toBe('María');
    expect(json.items).toHaveLength(1);
    expect(json.payment_method).toBe('efectivo');
    expect(json.tipo_entrega).toBe('delivery');
    expect(json.estado).toBe('pendiente');
  });
});

describe('AppState', () => {
  let appState;

  beforeEach(() => {
    appState = new AppState();
  });

  test('debe inicializar con valores por defecto', () => {
    expect(appState.productos).toEqual([]);
    expect(appState.vendedores).toEqual([]);
    expect(appState.clientesPorVendedor).toEqual({});
    expect(appState.ventaActual).toBeNull();
    expect(appState.ventasListado).toEqual([]);
  });

  test('debe cargar datos correctamente', () => {
    const data = {
      productos: [{ id: 1, tipo_pizza: 'Muzza', descripcion: 'Pizza', precio: 1000, activo: true }],
      vendedores: [{ id: 1, nombre: 'Juan' }],
      clientesPorVendedor: { 'Juan': [{ id: 1, nombre: 'María' }] }
    };

    appState.cargarDatos(data);

    expect(appState.productos).toHaveLength(1);
    expect(appState.productos[0]).toBeInstanceOf(Producto);
    expect(appState.vendedores).toHaveLength(1);
    expect(appState.vendedores[0]).toBeInstanceOf(Vendedor);
    expect(appState.clientesPorVendedor).toEqual(data.clientesPorVendedor);
  });

  test('debe crear nueva venta', () => {
    const venta = appState.crearVenta();
    expect(venta).toBeInstanceOf(Venta);
    expect(appState.ventaActual).toBe(venta);
  });

  test('debe limpiar venta actual', () => {
    appState.crearVenta();
    appState.limpiarVenta();

    expect(appState.ventaActual).toBeNull();
  });

  test('debe actualizar ventas listado', () => {
    const ventas = [{ id: 1, cliente: 'Test' }];
    appState.setVentas(ventas);

    expect(appState.ventasListado).toEqual(ventas);
  });
});