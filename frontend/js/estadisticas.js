// estadisticas.js - Refactorizado para usar APIService
// Ya no define su propia URL de API, usa APIService centralizado

let API_BASE = null;  // Se inicializa en DOMContentLoaded
let loadingTimeout = null;  // Para timeout de pantalla de carga

// Loading Spinner Functions
function showLoadingSpinner(show = true) {
    const overlay = document.getElementById('loadingOverlay');
    if (overlay) {
        if (show) {
            overlay.classList.remove('hidden');
            
            // Timeout: ocultar automáticamente después de 10 segundos
            if (loadingTimeout) clearTimeout(loadingTimeout);
            loadingTimeout = setTimeout(() => {
                hideLoadingSpinner();
                Logger.log('Loading timeout - se ocultó después de 10 segundos');
            }, 10000);
        } else {
            overlay.classList.add('hidden');
            
            // Limpiar timeout si se oculta manualmente
            if (loadingTimeout) {
                clearTimeout(loadingTimeout);
                loadingTimeout = null;
            }
        }
    }
}

function hideLoadingSpinner() {
    showLoadingSpinner(false);
}

// Logger condicional - solo en desarrollo
const Logger = {
    isDev: window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1',
    log: (msg, data) => {
        if (Logger.isDev) console.log(msg, data || '');
    }
};

// Función para parsear números en formato argentino ($1.000,50 -> 1000.50)
function parseArgentinoFloat(value) {
    if (typeof value === 'number') return value;
    if (!value) return 0;
    
    let str = String(value).trim();
    // Remover $ si existe
    str = str.replace('$', '');
    // Remover separadores de miles (.)
    str = str.replace(/\./g, '');
    // Reemplazar separador decimal (,) por punto
    str = str.replace(',', '.');
    
    return parseFloat(str) || 0;
}

// Función para inicializar filtros (llamada DESPUÉS de cargar datos)
function inicializarFiltros() {
    const filtroVendedor = document.getElementById('filtroVendedor');
    const filtroEntrega = document.getElementById('filtroEntrega');
    const filtroPago = document.getElementById('filtroPago');

    // Llenar lista de vendedores en el filtro
    if (filtroVendedor && datosVentas.ventas && datosVentas.ventas.length > 0) {
        // Limpiar opciones previas (excepto la primera que es "Ver todos")
        while (filtroVendedor.options.length > 1) {
            filtroVendedor.remove(1);
        }

        const vendedoresUnicos = [...new Set(datosVentas.ventas.map(v => v.vendedor))].filter(v => v).sort();
        Logger.log('Vendedores únicos encontrados:', vendedoresUnicos);
        
        vendedoresUnicos.forEach(vendedor => {
            const option = document.createElement('option');
            option.value = vendedor;
            option.textContent = vendedor;
            filtroVendedor.appendChild(option);
        });

        filtroVendedor.addEventListener('change', () => {
            renderizarVentas();
        });
    }

    if (filtroEntrega) {
        filtroEntrega.addEventListener('change', () => {
            renderizarVentas();
        });
    }

    if (filtroPago) {
        filtroPago.addEventListener('change', () => {
            renderizarVentas();
        });
    }

    // Filtros de Vendedores en Tab Vendedores
    const filtroEstadoVendedores = document.getElementById('filtroEstadoVendedores');
    const filtroVendedorEspecifico = document.getElementById('filtroVendedorEspecifico');
    const optgroupVendedores = document.getElementById('optgroupVendedores');
    
    // Función para actualizar dinámicamente el optgroup según el filtro de estado
    function actualizarOptgroupVendedores() {
        if (!optgroupVendedores || !datosVentas.vendedores) return;
        
        optgroupVendedores.innerHTML = '';
        
        // Obtener todos los vendedores registrados
        const todosVendedores = datosVentas.vendedores.sort((a, b) => a.nombre.localeCompare(b.nombre));
        
        // Contar ventas por vendedor
        const ventasPorVendedor = {};
        (datosVentas.ventas || []).forEach(v => {
            if (v.vendedor) {
                ventasPorVendedor[v.vendedor] = (ventasPorVendedor[v.vendedor] || 0) + 1;
            }
        });
        
        // Obtener estado actual del primer filtro
        const estadoActual = filtroEstadoVendedores?.value || '';
        
        // Filtrar vendedores según el estado
        let vendedoresAMostrar = todosVendedores;
        
        if (estadoActual === 'con-ventas') {
            vendedoresAMostrar = todosVendedores.filter(v => (ventasPorVendedor[v.nombre] || 0) > 0);
        } else if (estadoActual === 'sin-ventas') {
            vendedoresAMostrar = todosVendedores.filter(v => (ventasPorVendedor[v.nombre] || 0) === 0);
        }
        
        Logger.log('Vendedores a mostrar en optgroup:', vendedoresAMostrar);
        
        // Agregar opciones al optgroup
        vendedoresAMostrar.forEach(vendedor => {
            const option = document.createElement('option');
            option.value = vendedor.nombre;
            const cantidadVentas = ventasPorVendedor[vendedor.nombre] || 0;
            option.textContent = `${vendedor.nombre} (${cantidadVentas} ${cantidadVentas === 1 ? 'venta' : 'ventas'})`;
            optgroupVendedores.appendChild(option);
        });
    }
    
    // Inicializar el optgroup
    actualizarOptgroupVendedores();
    
    if (filtroEstadoVendedores) {
        filtroEstadoVendedores.addEventListener('change', () => {
            // Limpiar el segundo filtro cuando cambia el primero
            if (filtroVendedorEspecifico) {
                filtroVendedorEspecifico.value = '';
            }
            // Actualizar opciones del segundo filtro
            actualizarOptgroupVendedores();
            renderizarVendedores();
        });
    }
    
    if (filtroVendedorEspecifico) {
        filtroVendedorEspecifico.addEventListener('change', () => {
            renderizarVendedores();
        });
    }

    Logger.log('Filtros inicializados');
}

async function cargarDatos() {
    try {
        showLoadingSpinner(true);
        const api = new APIService(); // Usar APIService centralizado
        
        // 1. Obtener productos para cache
        try {
            const prodResp = await api.obtenerProductos();
            // El backend retorna {status, data, message}, extraer el array
            productosCache = (prodResp && prodResp.data) ? prodResp.data : prodResp || [];
            Logger.log('Productos cargados:', productosCache);
        } catch (e) {
            Logger.log('No se pudieron cargar productos');
            productosCache = [];
        }

        // 2. Obtener datos de estadísticas
        const statsResp = await api.request('/estadisticas-sheet');
        datosVentas = (statsResp && statsResp.data) ? statsResp.data : statsResp || {};
        Logger.log('estadisticas - statsResp:', statsResp);
        Logger.log('estadisticas - datosVentas.resumen:', datosVentas?.resumen);
        
        // 3. Obtener detalle de ventas para la tabla
        try {
            const ventasData = await api.obtenerVentas();
            Logger.log('estadisticas - ventasData:', ventasData);
            // El backend retorna {status, data, message}, extraer el array de data
            const ventasArray = Array.isArray(ventasData) ? ventasData : (ventasData?.data || []);
            Logger.log('estadisticas - ventasArray:', ventasArray);
            datosVentas.ventas = Array.isArray(ventasArray) ? ventasArray : [];
        } catch (e) {
            Logger.log('No se pudieron cargar ventas');
            datosVentas.ventas = [];
        }

        Logger.log('Datos de estadísticas:', datosVentas);
        
        // 4. Renderizar tabs
        renderizarResumen();
        renderizarVendedores();
        renderizarVentas();
        renderizarComboCounters();
        renderizarProductosPorTipo();
        
        // 5. Inicializar filtros (DESPUÉS de cargar datos)
        inicializarFiltros();
        
        hideLoadingSpinner();
    } catch (error) {
        console.error('Error:', error);
        showMessage('Error al cargar estadísticas: ' + error.message, 'error');
        hideLoadingSpinner();
    }
}

// Nueva función para renderizar contadores de productos dinámicamente
// Mapeo de combos a cantidades de pizzas por tipo
const COMBO_PIZZA_MAP = {
    'Muzza': { 'Muzza': 1, 'Muzza y Jamón': 0 },
    'Muzza y Jamón': { 'Muzza': 0, 'Muzza y Jamón': 1 },
    'La dupla | 1 Muzza + 1 Muzza y Jamón': { 'Muzza': 1, 'Muzza y Jamón': 1 },
    'Mix Familia grande | 2 Muzza + 1 Muzza y Jamón': { 'Muzza': 2, 'Muzza y Jamón': 1 },
    'Mix Juntada amigos | 3 Muzza + 2 Muzza y jamón': { 'Muzza': 3, 'Muzza y Jamón': 2 }
};

function renderizarComboCounters() {
    if (!datosVentas.ventas || !productosCache) return;

    const container = document.getElementById('productosCounters');
    container.innerHTML = '';

    // Contar ventas por producto desde detalle_ventas (sin canceladas)
    const ventasPorProducto = {};
    
    productosCache.forEach(producto => {
        ventasPorProducto[producto.id] = 0;
    });

    // Sumar cantidades de cada producto (excluyendo canceladas)
    if (datosVentas.ventas && Array.isArray(datosVentas.ventas)) {
        datosVentas.ventas.forEach(venta => {
            if (venta.estado === 'cancelada') return;
            // Las ventas llegadas del backend deben tener información de productos
            // Por ahora contamos desde el array de items si existen
            // Si no, hacemos un conteo genérico
        });
    }

    // Renderizar tarjetas para cada producto
    productosCache.forEach(producto => {
        // Calcular total vendido para este producto (excluyendo canceladas)
        let totalVendido = 0;
        if (datosVentas.ventas && Array.isArray(datosVentas.ventas)) {
            datosVentas.ventas.forEach(venta => {
                // Excluir ventas canceladas
                if (venta.estado === 'cancelada') return;
                // Si la venta tiene array de items con product_id
                if (venta.items && Array.isArray(venta.items)) {
                    venta.items.forEach(item => {
                        if (item.product_id === producto.id) {
                            totalVendido += item.cantidad || 0;
                        }
                    });
                }
            });
        }

        const card = document.createElement('div');
        card.className = 'stat-card';
        card.innerHTML = `
            <div class="stat-label">${producto.tipo_pizza}</div>
            <div class="stat-value">${totalVendido}</div>
            <div style="font-size: 12px; color: #666; margin-top: 5px;">$${producto.precio.toFixed(2)} c/u</div>
        `;
        container.appendChild(card);
    });
}

// Nueva función para calcular y renderizar pizzas por tipo (Muzza y Muzza y Jamón)
function renderizarProductosPorTipo() {
    if (!datosVentas.ventas) return;

    const container = document.getElementById('productosPorTipoCounters');
    if (!container) return; // El contenedor debe existir en el HTML

    container.innerHTML = '';

    // Inicializar contadores
    const pizzasPorTipo = {
        'Muzza': 0,
        'Muzza y Jamón': 0
    };

    // Procesar cada venta
    if (Array.isArray(datosVentas.ventas)) {
        datosVentas.ventas.forEach(venta => {
            // Excluir ventas canceladas
            if (venta.estado === 'cancelada') return;

            // Procesar items de la venta
            if (venta.items && Array.isArray(venta.items)) {
                venta.items.forEach(item => {
                    const cantidad = parseInt(item.cantidad) || 0;
                    if (cantidad <= 0) return; // Ignorar items con cantidad 0 o negativa
                    
                    const comboNombre = item.tipo || item.tipo_pizza;
                    
                    // Primero, verificar si es un combo mapeado
                    if (COMBO_PIZZA_MAP[comboNombre]) {
                        const pizzasDelCombo = COMBO_PIZZA_MAP[comboNombre];
                        
                        pizzasPorTipo['Muzza'] += (pizzasDelCombo['Muzza'] || 0) * cantidad;
                        pizzasPorTipo['Muzza y Jamón'] += (pizzasDelCombo['Muzza y Jamón'] || 0) * cantidad;
                    } else if (pizzasPorTipo.hasOwnProperty(comboNombre)) {
                        // Si no es un combo, pero es un tipo de pizza directo, contar como producto individual
                        pizzasPorTipo[comboNombre] = (pizzasPorTipo[comboNombre] || 0) + cantidad;
                    }
                });
            }
        });
    }

    // Renderizar tarjetas para cada tipo de pizza
    Object.entries(pizzasPorTipo).forEach(([tipo, cantidad]) => {
        const card = document.createElement('div');
        card.className = 'stat-card';
        card.innerHTML = `
            <div class="stat-label">${tipo}</div>
            <div class="stat-value">${cantidad}</div>
            <div style="font-size: 12px; color: #666; margin-top: 5px;">pizzas vendidas</div>
        `;
        container.appendChild(card);
    });
}

function renderizarProductosCounters() {
    if (!datosVentas.ventas || !productosCache) return;

    const container = document.getElementById('productosCounters');
    container.innerHTML = '';

    // Contar ventas por producto desde detalle_ventas (sin canceladas)
    const ventasPorProducto = {};
    
    productosCache.forEach(producto => {
        ventasPorProducto[producto.id] = 0;
    });

    // Sumar cantidades de cada producto (excluyendo canceladas)
    if (datosVentas.ventas && Array.isArray(datosVentas.ventas)) {
        datosVentas.ventas.forEach(venta => {
            if (venta.estado === 'cancelada') return;
            // Las ventas llegadas del backend deben tener información de productos
            // Por ahora contamos desde el array de items si existen
            // Si no, hacemos un conteo genérico
        });
    }

    // Renderizar tarjetas para cada producto
    productosCache.forEach(producto => {
        // Calcular total vendido para este producto (excluyendo canceladas)
        let totalVendido = 0;
        if (datosVentas.ventas && Array.isArray(datosVentas.ventas)) {
            datosVentas.ventas.forEach(venta => {
                // Excluir ventas canceladas
                if (venta.estado === 'cancelada') return;
                // Si la venta tiene array de items con product_id
                if (venta.items && Array.isArray(venta.items)) {
                    venta.items.forEach(item => {
                        if (item.product_id === producto.id) {
                            totalVendido += item.cantidad || 0;
                        }
                    });
                }
            });
        }

        const card = document.createElement('div');
        card.className = 'stat-card';
        card.innerHTML = `
            <div class="stat-label">${producto.tipo_pizza}</div>
            <div class="stat-value">${totalVendido}</div>
            <div style="font-size: 12px; color: #666; margin-top: 5px;">$${producto.precio.toFixed(2)} c/u</div>
        `;
        container.appendChild(card);
    });
}

function renderizarResumen() {
    if (!datosVentas.resumen) return;

    const resumen = datosVentas.resumen;

    // Renderizar contadores de productos dinámicamente
    renderizarProductosCounters();

    // Actualizar entregas
    document.getElementById('totalDelivery').textContent = Math.round(resumen.total_delivery || 0);
    document.getElementById('totalRetiro').textContent = Math.round(resumen.total_retiro || 0);

    // Actualizar dinero y pagos
    document.getElementById('pendienteCobro').textContent = `$${(resumen.pendiente_cobro || 0).toFixed(2)}`;
    document.getElementById('efectivoCobrado').textContent = `$${(resumen.efectivo_cobrado || 0).toFixed(2)}`;
    document.getElementById('transferenciaCobrada').textContent = `$${(resumen.transferencia_cobrada || 0).toFixed(2)}`;
    
    // Total cobrado (incluye pagadas + entregadas)
    document.getElementById('totalCobrado').textContent = `$${(resumen.total_cobrado || 0).toFixed(2)}`;

    // Actualizar estados de ventas
    document.getElementById('ventasSinPagar').textContent = Math.round(resumen.ventas_sin_pagar || 0);
    document.getElementById('ventasPagadas').textContent = Math.round(resumen.ventas_pagadas || 0);
    document.getElementById('ventasEntregadas').textContent = Math.round(resumen.ventas_entregadas || 0);
    document.getElementById('totalVentas').textContent = Math.round(resumen.ventas_totales || 0);
}

function renderizarVendedores() {
    if (!datosVentas.vendedores) return;

    const vendedores = datosVentas.vendedores;
    const ventas = datosVentas.ventas || [];

    // Renderizar tarjetas
    const container = document.getElementById('vendedoresDetail');
    container.innerHTML = '';

    // Mensaje si no hay vendedores
    if (!vendedores || vendedores.length === 0) {
        container.innerHTML = '<div style="padding: 20px; text-align: center; color: #999;">❌ No se encontraron vendedores</div>';
        return;
    }

    // Obtener valores de los filtros
    const filtroEstado = document.getElementById('filtroEstadoVendedores')?.value || '';
    const filtroVendedor = document.getElementById('filtroVendedorEspecifico')?.value || '';

    // Contar ventas por vendedor
    const ventasPorVendedor = {};
    ventas.forEach(v => {
        if (v.vendedor) {
            ventasPorVendedor[v.vendedor] = (ventasPorVendedor[v.vendedor] || 0) + 1;
        }
    });

    // Aplicar Filtro 1: Estado de Ventas
    let vendedoresFiltrados = vendedores.filter(v => {
        const tieneVentas = (ventasPorVendedor[v.nombre] || 0) > 0;
        
        if (filtroEstado === 'con-ventas') {
            return tieneVentas;
        } else if (filtroEstado === 'sin-ventas') {
            return !tieneVentas;
        }
        // Si es "" (Ver todos), no filtrar por estado
        return true;
    });

    // Aplicar Filtro 2: Vendedor Específico
    if (filtroVendedor) {
        vendedoresFiltrados = vendedoresFiltrados.filter(v => v.nombre === filtroVendedor);
    }

    // Mensaje si no hay resultados
    if (vendedoresFiltrados.length === 0) {
        let mensaje = '❌ No hay vendedores';
        if (filtroEstado === 'con-ventas') {
            mensaje = '❌ No hay vendedores con ventas';
        } else if (filtroEstado === 'sin-ventas') {
            mensaje = '❌ No hay vendedores sin ventas';
        } else if (filtroVendedor) {
            mensaje = `❌ No se encontró el vendedor "${filtroVendedor}"`;
        }
        container.innerHTML = `<div style="padding: 20px; text-align: center; color: #999;">${mensaje}</div>`;
        return;
    }

    vendedoresFiltrados.forEach(vendedor => {
        // Filtrar ventas sin pagar de este vendedor
        const ventasSinPagar = ventas.filter(v => 
            v.vendedor === vendedor.nombre && v.estado === 'sin pagar'
        );

        // Calcular pizzas vendidas por tipo (Muzza y Muzza y Jamón)
        let pizzasMuzza = 0;
        let pizzasMuzzaJamon = 0;
        
        ventas.forEach(v => {
            if (v.vendedor === vendedor.nombre && v.estado !== 'cancelada') {
                if (v.items && Array.isArray(v.items)) {
                    v.items.forEach(item => {
                        const comboNombre = item.tipo || item.tipo_pizza;
                        const cantidad = parseInt(item.cantidad) || 0;
                        
                        // Buscar en COMBO_PIZZA_MAP
                        if (COMBO_PIZZA_MAP[comboNombre]) {
                            const pizzasDelCombo = COMBO_PIZZA_MAP[comboNombre];
                            pizzasMuzza += (pizzasDelCombo['Muzza'] || 0) * cantidad;
                            pizzasMuzzaJamon += (pizzasDelCombo['Muzza y Jamón'] || 0) * cantidad;
                        } else if (comboNombre === 'Muzza') {
                            pizzasMuzza += cantidad;
                        } else if (comboNombre === 'Muzza y Jamón') {
                            pizzasMuzzaJamon += cantidad;
                        }
                    });
                }
            }
        });

        // Calcular desgloses por método de pago para DEUDAS (sin pagar)
        let deudaEfectivo = 0, deudaTransferencia = 0;
        ventas.forEach(v => {
            if (v.vendedor === vendedor.nombre && v.estado === 'sin pagar') {
                const monto = parseArgentinoFloat(v.total);
                if (v.payment_method === 'efectivo') {
                    deudaEfectivo += monto;
                } else if (v.payment_method === 'transferencia') {
                    deudaTransferencia += monto;
                }
            }
        });

        // Calcular desgloses por método de pago para PAGOS (pagada o entregada)
        let pagadoEfectivo = 0, pagadoTransferencia = 0;
        ventas.forEach(v => {
            if (v.vendedor === vendedor.nombre && (v.estado === 'pagada' || v.estado === 'entregada')) {
                const monto = parseArgentinoFloat(v.total);
                if (v.payment_method === 'efectivo') {
                    pagadoEfectivo += monto;
                } else if (v.payment_method === 'transferencia') {
                    pagadoTransferencia += monto;
                }
            }
        });

        const card = document.createElement('div');
        card.className = 'vendedor-card';
        card.innerHTML = `
            <h3 style="margin: 0 0 12px 0; font-size: 18px;">👤 ${vendedor.nombre}</h3>
            
            <!-- ESTADÍSTICAS PRINCIPALES -->
            <div class="vendedor-stats-grid">
                <div class="vendedor-stat">
                    <span class="vendedor-stat-label">📊 Ventas:</span>
                    <span class="vendedor-stat-value">${Math.round(vendedor.cantidad || 0)}</span>
                </div>
                <div class="vendedor-stat">
                    <span class="vendedor-stat-label">🍕 Muzzas:</span>
                    <span class="vendedor-stat-value">${pizzasMuzza}</span>
                </div>
                <div class="vendedor-stat">
                    <span class="vendedor-stat-label">🍕 Muzza y Jamón:</span>
                    <span class="vendedor-stat-value">${pizzasMuzzaJamon}</span>
                </div>
            </div>
            
            <!-- DESGLOSE DE DEUDAS -->
            <div class="vendedor-desglose vendedor-desglose-deuda">
                <div class="desglose-header">
                    <span class="desglose-title">⏳ Monto sin pagar</span>
                    <span class="desglose-total">$${(vendedor.deuda || 0).toFixed(2)}</span>
                </div>
                <div class="desglose-items">
                    <div class="desglose-item">
                        <span>💵 Efectivo:</span>
                        <strong>$${deudaEfectivo.toFixed(2)}</strong>
                    </div>
                    <div class="desglose-item">
                        <span>🏦 Transferencia:</span>
                        <strong>$${deudaTransferencia.toFixed(2)}</strong>
                    </div>
                </div>
            </div>

            <!-- DESGLOSE DE PAGOS -->
            <div class="vendedor-desglose vendedor-desglose-pago">
                <div class="desglose-header">
                    <span class="desglose-title">✓ Monto pagado</span>
                    <span class="desglose-total">$${(vendedor.pagado || 0).toFixed(2)}</span>
                </div>
                <div class="desglose-items">
                    <div class="desglose-item">
                        <span>💵 Efectivo:</span>
                        <strong>$${pagadoEfectivo.toFixed(2)}</strong>
                    </div>
                    <div class="desglose-item">
                        <span>🏦 Transferencia:</span>
                        <strong>$${pagadoTransferencia.toFixed(2)}</strong>
                    </div>
                </div>
            </div>

            <!-- TOTAL VENDEDOR -->
            <div class="vendedor-total">
                <span class="vendedor-stat-label">💰 Total vendedor:</span>
                <span class="vendedor-stat-value">$${(vendedor.total || 0).toFixed(2)}</span>
            </div>
            
            <!-- DEUDORES -->
            ${ventasSinPagar.length > 0 ? `
                <div class="vendedor-deudores">
                    <h4 style="margin: 0 0 8px 0; font-size: 14px;">⚠️ Clientes que no pagaron (${ventasSinPagar.length})</h4>
                    <div class="deudores-list">
                        ${ventasSinPagar.map(venta => {
                            const metodo = venta.payment_method === 'efectivo' ? '💵' : 
                                          venta.payment_method === 'transferencia' ? '🏦' : 
                                          '❓';
                            const metodoText = venta.payment_method === 'efectivo' ? 'Efectivo' : 
                                             venta.payment_method === 'transferencia' ? 'Transferencia' : 
                                             'Otro';
                            return `<div class="deuda-item-mobile">
                                <div class="deuda-info">
                                    <strong>${venta.cliente}</strong>
                                    <span class="deuda-metodo">${metodo} ${metodoText}</span>
                                </div>
                                <div class="deuda-monto">$${parseArgentinoFloat(venta.total).toFixed(2)}</div>
                            </div>`;
                        }).join('')}
                    </div>
                </div>
            ` : '<div class="vendedor-pagado-completo">✓ Todos los clientes pagaron</div>'}
        `;
        container.appendChild(card);
    });
}

function renderizarVentas() {
    if (!datosVentas.ventas) return;

    const tbody = document.getElementById('ventasTableBody');
    tbody.innerHTML = '';

    // Obtener valores de todos los filtros
    const filtroVendedor = document.getElementById('filtroVendedor')?.value || '';
    const filtroEntrega = document.getElementById('filtroEntrega')?.value || '';
    const filtroPago = document.getElementById('filtroPago')?.value || '';

    // Filtrar ventas según TODOS los filtros activos
    let ventasFiltradas = datosVentas.ventas.filter(venta => {
        // Filtro por vendedor
        if (filtroVendedor && venta.vendedor !== filtroVendedor) {
            return false;
        }

        // Filtro por tipo de entrega
        if (filtroEntrega) {
            if (filtroEntrega === 'delivery') {
                if (venta.tipo_entrega !== 'delivery' && venta.tipo_entrega !== 'envio') {
                    return false;
                }
            } else if (filtroEntrega === 'retiro') {
                if (venta.tipo_entrega !== 'retiro') {
                    return false;
                }
            }
        }

        // Filtro por estado de pago
        if (filtroPago) {
            if (filtroPago === 'sin-pagar') {
                if (venta.estado !== 'sin pagar' && venta.estado !== undefined) {
                    return false;
                }
            } else if (filtroPago === 'pagada') {
                if (venta.estado !== 'pagada') {
                    return false;
                }
            } else if (filtroPago === 'entregada') {
                if (venta.estado !== 'entregada') {
                    return false;
                }
            }
        }

        return true;
    });

    // Mensaje si no hay ventas después de filtrar
    if (ventasFiltradas.length === 0) {
        const tr = document.createElement('tr');
        tr.innerHTML = '<td colspan="10" style="text-align: center; padding: 20px; color: #999;">❌ No se encontraron ventas</td>';
        tbody.appendChild(tr);
        return;
    }

    ventasFiltradas.forEach(venta => {
        // Crear resumen de items (productos)
        let itemsResumen = 'Sin items';
        if (venta.items && Array.isArray(venta.items) && venta.items.length > 0) {
            itemsResumen = venta.items.map(item => {
                const producto = productosCache.find(p => p.id === item.product_id);
                const nombreProducto = producto ? producto.tipo_pizza : `Producto #${item.product_id}`;
                return `${item.cantidad}x ${nombreProducto}`;
            }).join(', ');
        }

        const estadoClass = venta.estado ? venta.estado.replace(' ', '-') : 'sin-pagar';
        const totalParseado = parseArgentinoFloat(venta.total);
        
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${venta.id}</td>
            <td>${venta.vendedor}</td>
            <td>${venta.cliente}</td>
            <td>${venta.telefono_cliente || '-'}</td>
            <td style="font-size: 12px;">${itemsResumen}</td>
            <td><strong>$${totalParseado.toFixed(2)}</strong></td>
            <td><span class="estado-badge ${estadoClass}">${venta.estado || 'sin pagar'}</span></td>
            <td>${venta.payment_method === 'efectivo' ? '💵' : '🏦'}</td>
            <td>${venta.tipo_entrega === 'envio' || venta.tipo_entrega === 'delivery' ? '🚚' : '🏪'}</td>
            <td><button class="btn-editar" data-id="${venta.id}">Editar</button></td>
        `;
        tbody.appendChild(tr);
    });

    // Event listeners para botones editar
    document.querySelectorAll('.btn-editar').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = parseInt(e.target.dataset.id);
            abrirModalEditar(id);
        });
    });
}

function abrirModalEditar(id) {
    const venta = datosVentas.ventas.find(v => v.id === id);
    if (!venta) return;

    ventaEnEdicion = venta;
    document.getElementById('editarEstado').value = venta.estado || 'sin pagar';
    document.getElementById('editarPago').value = venta.payment_method || 'efectivo';
    document.getElementById('editarEntrega').value = venta.tipo_entrega || 'delivery';
    // Cargar teléfono en modal si existe
    const editarTel = document.getElementById('editarTelefono');
    if (editarTel) {
        editarTel.value = venta.telefono_cliente || '';
    }
    
    // Actualizar previsualización del tipo de entrega
    actualizarPreviaEntrega(venta.tipo_entrega || 'delivery');
    
    // Llenar selector de productos nuevos
    const selectNuevo = document.getElementById('nuevoProductoSelect');
    selectNuevo.innerHTML = '<option value="">Selecciona producto para agregar...</option>';
    productosCache.forEach(p => {
        selectNuevo.innerHTML += `<option value="${p.id}">${p.tipo_pizza} - $${p.precio}</option>`;
    });
    
    // Renderizar productos existentes
    renderizarProductosEnEdicion(venta);
    
    document.getElementById('modalEditarVenta').classList.remove('hidden');
}

function renderizarProductosEnEdicion(venta) {
    const container = document.getElementById('productosEditables');
    if (!venta.items || venta.items.length === 0) {
        container.innerHTML = '<p style="color: #999; text-align: center; padding: 20px;">Sin productos</p>';
        return;
    }

    let html = '';
    venta.items.forEach((item, index) => {
        // No mostrar items marcados para eliminación
        if (item._eliminar) return;
        
        const nombreProducto = item.tipo || item.tipo_pizza || 'Producto';
        
        html += `
            <div class="producto-editable">
                <div class="nombre">${nombreProducto}</div>
                <div class="cantidad">
                    <input type="number" 
                           id="cant-${index}" 
                           min="1" 
                           value="${item.cantidad}">
                </div>
                <button type="button" class="btn-eliminar" onclick="eliminarProductoEnEdicion(${index})">✕ Quitar</button>
            </div>
        `;
    });
    container.innerHTML = html || '<p style="color: #999; text-align: center; padding: 20px;">Sin productos</p>';
}

function cerrarModal() {
    document.getElementById('modalEditarVenta').classList.add('hidden');
    ventaEnEdicion = null;
}

async function guardarCambios() {
    if (!ventaEnEdicion) return;

    const estado = document.getElementById('editarEstado').value;
    const pago = document.getElementById('editarPago').value;
    const entrega = document.getElementById('editarEntrega').value;
    
    // Recopilar cambios en productos
    const productosActualizados = [];
    const productosAEliminar = [];
    
    ventaEnEdicion.items.forEach((item, index) => {
        if (item._eliminar) {
            // Marcar para eliminación
            if (item.detalle_id) {
                productosAEliminar.push(item.detalle_id);
            }
        } else {
            const cantInput = document.getElementById(`cant-${index}`);
            if (cantInput) {
                const nuevaCantidad = parseInt(cantInput.value) || 0;
                if (nuevaCantidad > 0) {
                    productosActualizados.push({
                        detalle_id: item.detalle_id || null,
                        producto_id: item.product_id || item.ProductID || item.id,
                        cantidad: nuevaCantidad
                    });
                }
            }
        }
    });

    try {
        showLoadingSpinner(true);
        const payload = {
            estado: estado,
            payment_method: pago,
            tipo_entrega: entrega,
                productos: productosActualizados,
                cliente: ventaEnEdicion.cliente,
                telefono_cliente: (function(){
                    const v = document.getElementById('editarTelefono').value || '';
                    return v.trim() === '' ? null : parseInt(v);
                })()
        };
        
        if (productosAEliminar.length > 0) {
            payload.productos_eliminar = productosAEliminar;
        }

        const response = await fetch(`${API_BASE}/actualizar-venta/${ventaEnEdicion.id}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (response.ok) {
            showMessage('✓ Venta actualizada correctamente', 'success');
            cerrarModal();
            await cargarDatos();
        } else {
            const err = await response.text();
            showMessage('✗ Error al actualizar: ' + err, 'error');
        }
    } catch (error) {
        showMessage('Error: ' + error.message, 'error');
    } finally {
        hideLoadingSpinner();
    }
}

function showMessage(text, type) {
    const mensaje = document.getElementById('mensaje');
    mensaje.textContent = text;
    mensaje.classList.remove('hidden', 'success', 'error');
    mensaje.classList.add(type === 'error' ? 'error' : 'success');

    setTimeout(() => {
        mensaje.classList.add('hidden');
    }, 5000);
}

function eliminarProductoEnEdicion(index) {
    if (!ventaEnEdicion || !ventaEnEdicion.items) return;
    
    // Marcar para eliminación (enviaremos esto al backend)
    ventaEnEdicion.items[index]._eliminar = true;
    renderizarProductosEnEdicion(ventaEnEdicion);
}

function agregarProductoEnEdicion() {
    if (!ventaEnEdicion) return;

    const selectProducto = document.getElementById('nuevoProductoSelect');
    const cantidadInput = document.getElementById('nuevoProductoCantidad');
    
    const productoId = parseInt(selectProducto.value);
    const cantidad = parseInt(cantidadInput.value) || 1;
    
    if (!productoId || cantidad <= 0) {
        showMessage('Selecciona un producto y cantidad válida', 'error');
        return;
    }
    
    const producto = productosCache.find(p => p.id === productoId);
    if (!producto) return;
    
    // Agregar a items
    if (!ventaEnEdicion.items) ventaEnEdicion.items = [];
    
    ventaEnEdicion.items.push({
        id: producto.id,
        tipo_pizza: producto.tipo_pizza,
        cantidad: cantidad,
        detalle_id: null // Indica que es nuevo
    });
    
    // Resetear form
    selectProducto.value = '';
    cantidadInput.value = '1';
    
    // Renderizar
    renderizarProductosEnEdicion(ventaEnEdicion);
}

function actualizarPreviaEntrega(tipo) {
    const textos = {
        'delivery': '🚗 Delivery',
        'envio': '🚗 Delivery',
        'retiro': '🏪 Retiro'
    };
    const span = document.getElementById('entregaActual');
    if (span) {
        span.textContent = textos[tipo] || '🏪 Retiro';
    }
}

function incrementarCantidadProducto() {
    const input = document.getElementById('nuevoProductoCantidad');
    input.value = parseInt(input.value) + 1;
}

function decrementarCantidadProducto() {
    const input = document.getElementById('nuevoProductoCantidad');
    if (parseInt(input.value) > 1) {
        input.value = parseInt(input.value) - 1;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    // Inicializar API_BASE desde APIService
    const api = new APIService();
    API_BASE = api.baseURL;
    Logger.log('API Base URL:', API_BASE);
    
    cargarDatos();

    // Botones de tab
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const tab = e.target.dataset.tab;
            
            // Actualizar botones
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            e.target.classList.add('active');

            // Actualizar contenido
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            document.getElementById(`tab-${tab}`).classList.add('active');
        });
    });

    // Volver al home
    document.getElementById('btnVolver').addEventListener('click', () => {
        window.location.href = 'index.html';
    });

    // Panel Admin
    const btnAdminPanel = document.getElementById('btnAdminPanel');
    if (btnAdminPanel) {
        btnAdminPanel.addEventListener('click', () => {
            window.location.href = 'login.html';
        });
    }

    // Modal
    document.querySelector('.btn-close-modal').addEventListener('click', cerrarModal);
    document.querySelector('.btn-cancelar-modal').addEventListener('click', cerrarModal);
    document.querySelector('.btn-guardar-cambios').addEventListener('click', guardarCambios);

    // Botón agregar producto
    document.getElementById('btnAgregarProducto').addEventListener('click', agregarProductoEnEdicion);

    // Botones de cantidad en modal
    const btnMasProducto = document.getElementById('btnMasProducto');
    if (btnMasProducto) {
        btnMasProducto.addEventListener('click', (e) => {
            e.preventDefault();
            incrementarCantidadProducto();
        });
    }

    const btnMenosProducto = document.getElementById('btnMenosProducto');
    if (btnMenosProducto) {
        btnMenosProducto.addEventListener('click', (e) => {
            e.preventDefault();
            decrementarCantidadProducto();
        });
    }

    // Select de tipo de entrega
    const selectEntrega = document.getElementById('editarEntrega');
    if (selectEntrega) {
        selectEntrega.addEventListener('change', (e) => {
            actualizarPreviaEntrega(e.target.value);
        });
    }

    document.getElementById('modalEditarVenta').addEventListener('click', (e) => {
        if (e.target === document.getElementById('modalEditarVenta')) {
            cerrarModal();
        }
    });
});
