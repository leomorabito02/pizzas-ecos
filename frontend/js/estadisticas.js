// estadisticas.js - Refactorizado para usar APIService
// Ya no define su propia URL de API, usa APIService centralizado

let API_BASE = null;  // Se inicializa en DOMContentLoaded
let loadingTimeout = null;  // Para timeout de pantalla de carga
let datosVentas = { ventas: [], clientesPorVendedor: {} };  // Cache de datos
let productosCache = [];  // Cache de productos
let ventaEnEdicion = null;  // Venta en edición en modal
let currentPage = 1;
let currentLimit = 10;

// Función helper para formatear estado visualmente
function formatEstado(estado) {
    if (!estado) return 'Sin Pagar';
    // Normalizar: convertir espacio a underscore para consistencia
    const estadoNormalizado = estado.replace(' ', '_');
    const estadoMap = {
        'sin_pagar': 'Sin Pagar',
        'pagada': 'Pagada',
        'entregada': 'Entregada',
        'cancelada': 'Cancelada'
    };
    return estadoMap[estadoNormalizado] || 'Sin Pagar';
}

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

async function cargarDatos(reloadGeneralStats = true) {
    try {
        showLoadingSpinner(true);
        const api = new APIService(); // Usar APIService centralizado
        
        // 0. Obtener datos iniciales (vendedores, clientes, productos)
        if (reloadGeneralStats) {
            const dataResp = await api.request('/data');
            const initialData = (dataResp && dataResp.data) ? dataResp.data : dataResp || {};
            Logger.log('Datos iniciales cargados:', initialData);
            datosVentas.clientesPorVendedor = initialData.clientesPorVendedor || {};
            
            // 1. Obtener productos para cache
            try {
                const prodResp = await api.obtenerProductos();
                productosCache = (prodResp && prodResp.data) ? prodResp.data : prodResp || [];
            } catch (e) {
                productosCache = [];
            }

            // 2. Obtener datos de estadísticas resumidas
            const statsResp = await api.obtenerEstadisticas(currentLimit, currentPage);
            datosVentas = Object.assign(datosVentas, (statsResp && statsResp.data) ? statsResp.data : statsResp || {});
        } else {
            // Si no recargamos estadísticas generales, al menos actualizamos la tabla llamando a obtenerVentas
            const ventasData = await api.obtenerVentas(currentLimit, currentPage);
            const ventasArray = Array.isArray(ventasData) ? ventasData : (ventasData?.data || []);
            datosVentas.ventas = Array.isArray(ventasArray) ? ventasArray : [];
        }
        
        // 4. Renderizar tabs
        if (reloadGeneralStats) {
            renderizarResumen();
            renderizarVendedores();
        }
        renderizarVentas();
        
        // 5. Inicializar filtros (DESPUÉS de cargar datos)
        if (reloadGeneralStats) {
            inicializarFiltros();
            inicializarPaginacion();
        }
        
    } catch (error) {
        console.error('Error:', error);
        showMessage('Error al cargar estadísticas: ' + error.message, 'error');
    } finally {
        hideLoadingSpinner();
    }
}

function inicializarPaginacion() {
    const btnPrev = document.getElementById('btnPrevPage');
    const btnNext = document.getElementById('btnNextPage');
    const selectLimit = document.getElementById('filtroLimitVentas');
    
    if (btnPrev) {
        btnPrev.addEventListener('click', () => {
            if (currentPage > 1) {
                currentPage--;
                actualizarPaginacionDisplay();
                cargarDatos(false);
            }
        });
    }

    if (btnNext) {
        btnNext.addEventListener('click', () => {
            // Asumimos que si vienen limit elementos, puede haber más
            if (datosVentas.ventas.length === currentLimit) {
                currentPage++;
                actualizarPaginacionDisplay();
                cargarDatos(false);
            }
        });
    }

    if (selectLimit) {
        selectLimit.addEventListener('change', (e) => {
            currentLimit = parseInt(e.target.value) || 10;
            currentPage = 1;
            actualizarPaginacionDisplay();
            cargarDatos(false);
        });
    }
    
    actualizarPaginacionDisplay();
}

function actualizarPaginacionDisplay() {
    const display = document.getElementById('currentPageDisplay');
    const btnPrev = document.getElementById('btnPrevPage');
    const btnNext = document.getElementById('btnNextPage');
    
    if (display) display.textContent = `Página ${currentPage}`;
    if (btnPrev) btnPrev.disabled = currentPage === 1;
    if (btnNext) btnNext.disabled = datosVentas.ventas.length < currentLimit;
}

function renderizarProductosCounters() {
    const container = document.getElementById('productosCounters');
    const containerDesglose = document.getElementById('productosDesglosadosCounters');
    
    if (container) container.innerHTML = '';
    if (containerDesglose) containerDesglose.innerHTML = '';
    
    const resumen = datosVentas.resumen;
    if (!resumen) return;

    if (container && resumen.productos_vendidos) {
        resumen.productos_vendidos.forEach(producto => {
            const card = document.createElement('div');
            card.className = 'stat-card';
            card.innerHTML = `
                <div class="stat-label">${producto.nombre}</div>
                <div class="stat-value">${producto.cantidad}</div>
                <div style="font-size: 12px; color: #666; margin-top: 5px;">$${(producto.precio || 0).toFixed(2)} c/u</div>
            `;
            container.appendChild(card);
        });
    }

    if (containerDesglose && resumen.productos_desglosados) {
        resumen.productos_desglosados.forEach(producto => {
            const card = document.createElement('div');
            card.className = 'stat-card';
            card.innerHTML = `
                <div class="stat-label">${producto.nombre}</div>
                <div class="stat-value">${producto.cantidad}</div>
                <div style="font-size: 12px; color: #666; margin-top: 5px;">Unidades individuales</div>
            `;
            containerDesglose.appendChild(card);
        });
    }
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
        container.innerHTML = '<div style="padding: 20px; text-align: center; color: #999;">No se encontraron vendedores</div>';
        return;
    }

    // Obtener valores de los filtros
    const filtroEstado = document.getElementById('filtroEstadoVendedores')?.value || '';
    const filtroVendedor = document.getElementById('filtroVendedorEspecifico')?.value || '';

    // Contar ventas por vendedor (ahora usando el valor ya calculado)
    const ventasPorVendedor = {};
    vendedores.forEach(v => {
        ventasPorVendedor[v.nombre] = v.cantidad || 0;
    });

    // Aplicar Filtro 1: Estado de Ventas
    let vendedoresFiltrados = vendedores.filter(v => {
        const tieneVentas = (v.cantidad || 0) > 0;
        
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
        let mensaje = 'No hay vendedores';
        if (filtroEstado === 'con-ventas') {
            mensaje = 'No hay vendedores con ventas';
        } else if (filtroEstado === 'sin-ventas') {
            mensaje = 'No hay vendedores sin ventas';
        } else if (filtroVendedor) {
            mensaje = `No se encontró el vendedor "${filtroVendedor}"`;
        }
        container.innerHTML = `<div style="padding: 20px; text-align: center; color: #999;">${mensaje}</div>`;
        return;
    }

    vendedoresFiltrados.forEach(vendedor => {
        // Extraer estadísticas precálculadas por backend
        const cantidadVentas = Math.round(vendedor.cantidad || 0);
        const deudaEfectivo = vendedor.deuda_efectivo || 0;
        const deudaTransferencia = vendedor.deuda_transferencia || 0;
        const pagadoEfectivo = vendedor.pagado_efectivo || 0;
        const pagadoTransferencia = vendedor.pagado_transferencia || 0;
        const ventasSinPagar = vendedor.deudores || [];
        const productosVendidos = vendedor.productos_vendidos || [];

        // Generar HTML de productos dinámicos para este vendedor
        let productosHTML = '';
        if (productosVendidos.length > 0) {
            productosHTML = productosVendidos.map(p => `
                <div class="vendedor-stat">
                    <span class="vendedor-stat-label"><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: middle; margin-right: 4px;">local_pizza</span> ${p.nombre}:</span>
                    <span class="vendedor-stat-value">${p.cantidad}</span>
                </div>
            `).join('');
        } else {
            productosHTML = `
                <div class="vendedor-stat" style="grid-column: span 3; text-align: center; color: #999;">
                    Sin productos vendidos
                </div>
            `;
        }

        const card = document.createElement('div');
        card.className = 'vendedor-card';
        card.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; border-bottom: 1px solid var(--md-sys-color-outline-variant); padding-bottom: 12px;">
                <h3 style="margin: 0; font-size: 18px; display: flex; align-items: center; gap: 8px;">
                    <span class="material-symbols-outlined">person</span> ${vendedor.nombre}
                </h3>
                <div style="background: var(--md-sys-color-primary-container); color: var(--md-sys-color-on-primary-container); padding: 4px 12px; border-radius: 16px; font-weight: 600; font-size: 14px; display: flex; align-items: center; gap: 6px;">
                    <span class="material-symbols-outlined" style="font-size: 18px;">receipt_long</span>
                    ${cantidadVentas} ${cantidadVentas === 1 ? 'Venta' : 'Ventas'}
                </div>
            </div>
            
            <!-- PRODUCTOS VENDIDOS -->
            <div class="vendedor-stats-grid" style="grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));">
                ${productosHTML}
            </div>
            
            <!-- DESGLOSE DE DEUDAS -->
            <div class="vendedor-desglose vendedor-desglose-deuda">
                <div class="desglose-header">
                    <span class="desglose-title"><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: middle; margin-right: 4px;">schedule</span> Monto sin pagar</span>
                    <span class="desglose-total">$${(vendedor.deuda || 0).toFixed(2)}</span>
                </div>
                <div class="desglose-items">
                    <div class="desglose-item">
                        <span><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: middle; margin-right: 4px;">payments</span> Efectivo:</span>
                        <strong>$${deudaEfectivo.toFixed(2)}</strong>
                    </div>
                    <div class="desglose-item">
                        <span><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: middle; margin-right: 4px;">account_balance</span> Transferencia:</span>
                        <strong>$${deudaTransferencia.toFixed(2)}</strong>
                    </div>
                </div>
            </div>

            <!-- DESGLOSE DE PAGOS -->
            <div class="vendedor-desglose vendedor-desglose-pago">
                <div class="desglose-header">
                    <span class="desglose-title"><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: middle; margin-right: 4px;">check_circle</span> Monto pagado</span>
                    <span class="desglose-total">$${(vendedor.pagado || 0).toFixed(2)}</span>
                </div>
                <div class="desglose-items">
                    <div class="desglose-item">
                        <span><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: middle; margin-right: 4px;">payments</span> Efectivo:</span>
                        <strong>$${pagadoEfectivo.toFixed(2)}</strong>
                    </div>
                    <div class="desglose-item">
                        <span><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: middle; margin-right: 4px;">account_balance</span> Transferencia:</span>
                        <strong>$${pagadoTransferencia.toFixed(2)}</strong>
                    </div>
                </div>
            </div>

            <!-- TOTAL VENDEDOR -->
            <div class="vendedor-total">
                <span class="vendedor-stat-label"><span class="material-symbols-outlined" style="font-size: 16px; vertical-align: middle; margin-right: 4px;">monetization_on</span> Total vendedor:</span>
                <span class="vendedor-stat-value">$${(vendedor.total || 0).toFixed(2)}</span>
            </div>
            
            <!-- DEUDORES -->
            ${ventasSinPagar.length > 0 ? `
                <div class="vendedor-deudores">
                    <h4 style="margin: 0 0 8px 0; font-size: 14px; display: flex; align-items: center; gap: 4px;"><span class="material-symbols-outlined" style="color: var(--md-sys-color-error);">warning</span> Clientes que no pagaron (${ventasSinPagar.length})</h4>
                    <div class="deudores-list">
                        ${ventasSinPagar.map(venta => {
                            const metodoIcon = venta.payment_method === 'efectivo' ? 'payments' : 
                                              venta.payment_method === 'transferencia' ? 'account_balance' : 
                                              'help_outline';
                            const metodoText = venta.payment_method === 'efectivo' ? 'Efectivo' : 
                                             venta.payment_method === 'transferencia' ? 'Transferencia' : 
                                             'Otro';
                            return `<div class="deuda-item-mobile">
                                <div class="deuda-info">
                                    <strong>${venta.cliente}</strong>
                                    <span class="deuda-metodo" style="display: flex; align-items: center; gap: 4px;"><span class="material-symbols-outlined" style="font-size: 16px;">${metodoIcon}</span> ${metodoText}</span>
                                </div>
                                <div class="deuda-monto">$${parseArgentinoFloat(venta.total).toFixed(2)}</div>
                            </div>`;
                        }).join('')}
                    </div>
                </div>
            ` : '<div class="vendedor-pagado-completo" style="display: flex; align-items: center; gap: 4px;"><span class="material-symbols-outlined" style="color: #4CAF50;">check_circle</span> Todos los clientes pagaron</div>'}
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
                if (venta.estado !== 'sin_pagar' && venta.estado !== undefined) {
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
        tr.innerHTML = '<td colspan="10" style="text-align: center; padding: 20px; color: #999;">No se encontraron ventas</td>';
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

        // Normalizar estado y crear clase CSS: sin_pagar o 'sin pagar' (con espacio) → sin-pagar
        let estadoClass = 'sin-pagar';
        if (venta.estado) {
            if (venta.estado === 'sin pagar' || venta.estado === 'sin_pagar') {
                estadoClass = 'sin-pagar';
            } else {
                estadoClass = venta.estado.replace('_', '-');
            }
        }
        const totalParseado = parseArgentinoFloat(venta.total);
        
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${venta.id}</td>
            <td>${venta.vendedor}</td>
            <td>${venta.cliente}</td>
            <td>${venta.telefono_cliente || '-'}</td>
            <td style="font-size: 12px;">${itemsResumen}</td>
            <td><strong>$${totalParseado.toFixed(2)}</strong></td>
            <td><span class="estado-badge ${estadoClass}">${formatEstado(venta.estado) || 'Sin Pagar'}</span></td>
            <td>${venta.payment_method === 'efectivo' ? '<span class="material-symbols-outlined" title="Efectivo">payments</span>' : '<span class="material-symbols-outlined" title="Transferencia">account_balance</span>'}</td>
            <td>${venta.tipo_entrega === 'envio' || venta.tipo_entrega === 'delivery' ? '<span class="material-symbols-outlined" title="Delivery">local_shipping</span>' : '<span class="material-symbols-outlined" title="Retiro">storefront</span>'}</td>
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

    ventaEnEdicion = JSON.parse(JSON.stringify(venta));
    
    // Actualizar título con el nombre del vendedor
    const modalTitle = document.getElementById('modalEditarTitle');
    if (modalTitle) {
        modalTitle.textContent = `Editar venta de ${venta.vendedor || 'Vendedor'}`;
    }

    // Cargar cliente
    const editarCliente = document.getElementById('editarCliente');
    if (editarCliente) {
        editarCliente.value = venta.cliente || '';
    }
    // Cargar teléfono en modal si existe
    const editarTel = document.getElementById('editarTelefono');
    if (editarTel) {
        editarTel.value = venta.telefono_cliente || '';
    }
    document.getElementById('editarEstado').value = venta.estado || 'sin_pagar';
    document.getElementById('editarPago').value = venta.payment_method || 'efectivo';
    let entregaVal = venta.tipo_entrega || 'envio';
    if (entregaVal === 'delivery') {
        entregaVal = 'envio';
    }
    document.getElementById('editarEntrega').value = entregaVal;
    
    // Cargar dropdown de clientes del vendedor
    cargarClientesDropdownEditar(venta.vendedor);
    
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

function cargarClientesDropdownEditar(vendedor) {
    // Obtener clientes del vendedor desde datosVentas
    const clientesVendedor = datosVentas.clientesPorVendedor && datosVentas.clientesPorVendedor[vendedor] ? datosVentas.clientesPorVendedor[vendedor] : [];
    
    const drop = document.getElementById('clientesDropdownEditar');
    const lista = document.getElementById('clientesListEditar');
    
    if (clientesVendedor.length > 0) {
        lista.innerHTML = '';
        clientesVendedor.forEach(c => {
            const div = document.createElement('div');
            div.className = 'cliente-item';
            // c puede ser string (legacy) u objeto {id, nombre, telefono}
            const nombre = (typeof c === 'string') ? c : c.nombre;
            const telefono = (typeof c === 'string') ? null : c.telefono;
            div.textContent = nombre;
            div.addEventListener('click', () => {
                document.getElementById('editarCliente').value = nombre;
                // Si tenemos teléfono, cargarlo en el input (editable)
                const telInput = document.getElementById('editarTelefono');
                if (telInput) {
                    if (telefono) telInput.value = telefono;
                    else telInput.value = '';
                }
                drop.classList.add('hidden');
            });
            lista.appendChild(div);
        });
        drop.classList.remove('hidden');
    } else {
        lista.innerHTML = '';
        drop.classList.add('hidden');
    }
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
            <div class="producto-editable-row">
                <div class="nombre">${nombreProducto}</div>
                <div class="cantidad">
                    <input type="number" 
                           id="cant-${index}" 
                           min="1" 
                           value="${item.cantidad}">
                </div>
                <button type="button" class="btn-eliminar-item-modal" onclick="eliminarProductoEnEdicion(${index})">✕ Quitar</button>
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
    const cliente = (document.getElementById('editarCliente').value || '').trim();
    
    Logger.log('guardarCambios - cliente:', cliente);
    
    // Validar que cliente no esté vacío
    if (!cliente) {
        showMessage('El cliente no puede estar vacío', 'error');
        return;
    }
    
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
            cliente: cliente,
            telefono_cliente: (function(){
                const v = document.getElementById('editarTelefono').value || '';
                const trimmed = v.replace(/\D/g, ''); // Extraer solo dígitos
                if (trimmed === '') return 0;
                return parseInt(trimmed, 10) || 0;
            })(),
            productos: productosActualizados
        };
        
        if (productosAEliminar.length > 0) {
            payload.productos_eliminar = productosAEliminar;
        }

        try {
            const response = await api.request(`/actualizar-venta/${ventaEnEdicion.id}`, {
                method: 'POST',
                body: JSON.stringify(payload)
            });

            Logger.log('guardarCambios - payload enviado:', payload);
            showMessage('Venta actualizada correctamente', 'success');
            cerrarModal();
            await cargarDatos();
        } catch (error) {
            showMessage('✗ Error al actualizar: ' + error.message, 'error');
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

    // Precargar datos principales en background
    setTimeout(() => {
        Logger.log('📡 Precargando datos principales en background...');
        api.getData().catch(e => Logger.log('Error precarga index:', e));
    }, 500);

    // Leer tab desde la URL
    const urlParams = new URLSearchParams(window.location.search);
    const initialTab = urlParams.get('tab');
    if (initialTab) {
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
        const tabEl = document.getElementById(`tab-${initialTab}`);
        if (tabEl) tabEl.classList.add('active');
        
        document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
        const activeNavBtn = document.querySelector(`.nav-btn[data-tab="${initialTab}"]`);
        if (activeNavBtn) activeNavBtn.classList.add('active');
    }

    // Botones de tab de la navbar
    document.querySelectorAll('.nav-btn[data-tab]').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            const button = e.currentTarget;
            const tab = button.dataset.tab;
            
            // Actualizar URL sin recargar
            const newUrl = new URL(window.location);
            newUrl.searchParams.set('tab', tab);
            window.history.pushState({}, '', newUrl);
            
            // Actualizar botones
            document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
            button.classList.add('active');

            // Actualizar contenido
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            document.getElementById(`tab-${tab}`).classList.add('active');
        });
    });

    // Modal
    document.querySelector('.btn-close-modal').addEventListener('click', cerrarModal);
    document.querySelector('.btn-cancelar-modal').addEventListener('click', cerrarModal);
    document.querySelector('.btn-guardar-cambios').addEventListener('click', guardarCambios);

    // Dropdown de clientes en modal
    const editarClienteInput = document.getElementById('editarCliente');
    if (editarClienteInput) {
        editarClienteInput.addEventListener('focus', () => {
            const drop = document.getElementById('clientesDropdownEditar');
            if (drop && !drop.classList.contains('hidden')) {
                drop.classList.remove('hidden');
            }
        });
        editarClienteInput.addEventListener('blur', () => {
            setTimeout(() => {
                document.getElementById('clientesDropdownEditar').classList.add('hidden');
            }, 200);
        });
    }
    
    const btnCloseDropdownEditar = document.querySelector('#clientesDropdownEditar .btn-close-dropdown');
    if (btnCloseDropdownEditar) {
        btnCloseDropdownEditar.addEventListener('click', (e) => {
            e.preventDefault();
            document.getElementById('clientesDropdownEditar').classList.add('hidden');
        });
    }

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



    document.getElementById('modalEditarVenta').addEventListener('click', (e) => {
        if (e.target === document.getElementById('modalEditarVenta')) {
            cerrarModal();
        }
    });
});
