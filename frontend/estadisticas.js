// estadisticas.js
let datosVentas = {};
let ventaEnEdicion = null;
let productosCache = []; // Cache de productos para generar contadores

function getAPIBase() {
    // Si está en localhost, usar localhost:8080 para desarrollo
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
        return 'http://localhost:8080/api';
    }
    // En producción, usar BACKEND_URL definido en config.js
    return BACKEND_URL;
}

const API_BASE = getAPIBase();
console.log('API Base URL:', API_BASE);

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

async function cargarDatos() {
    try {
        // 1. Obtener productos para cache
        const productosResponse = await fetch(`${API_BASE}/productos`);
        if (productosResponse.ok) {
            productosCache = await productosResponse.json();
            console.log('Productos cargados:', productosCache);
        }

        // 2. Obtener datos de estadísticas del sheet
        const response1 = await fetch(`${API_BASE}/estadisticas-sheet`);
        if (!response1.ok) throw new Error('No se pudieron cargar las estadísticas');
        datosVentas = await response1.json();
        
        // 3. Obtener detalle de ventas para la tabla
        const response2 = await fetch(`${API_BASE}/estadisticas`);
        if (response2.ok) {
            const ventasData = await response2.json();
            datosVentas.ventas = ventasData;
        }
        
        console.log('Datos de estadísticas:', datosVentas);
        
        // 4. Renderizar tabs
        renderizarResumen();
        renderizarVendedores();
        renderizarVentas();
    } catch (error) {
        console.error('Error:', error);
        showMessage('Error al cargar estadísticas: ' + error.message, 'error');
    }
}

// Nueva función para renderizar contadores de productos dinámicamente
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
    datosVentas.ventas.forEach(venta => {
        if (venta.estado === 'cancelada') return;
        // Las ventas llegadas del backend deben tener información de productos
        // Por ahora contamos desde el array de items si existen
        // Si no, hacemos un conteo genérico
    });

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

    vendedores.forEach(vendedor => {
        // Filtrar ventas sin pagar de este vendedor
        const ventasSinPagar = ventas.filter(v => 
            v.vendedor === vendedor.nombre && v.estado === 'sin pagar'
        );

        // Calcular total de items vendidos por este vendedor
        let totalItems = 0;
        ventas.forEach(v => {
            if (v.vendedor === vendedor.nombre) {
                if (v.items && Array.isArray(v.items)) {
                    v.items.forEach(item => {
                        totalItems += item.cantidad || 0;
                    });
                }
            }
        });

        const card = document.createElement('div');
        card.className = 'vendedor-card';
        card.innerHTML = `
            <h3>👤 ${vendedor.nombre}</h3>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">📊 Cantidad de ventas:</span>
                <span class="vendedor-stat-value">${Math.round(vendedor.cantidad || 0)}</span>
            </div>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">📦 Total de productos vendidos:</span>
                <span class="vendedor-stat-value">${Math.round(totalItems)}</span>
            </div>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">⏳ Monto sin pagar:</span>
                <span class="vendedor-stat-value">$${(vendedor.deuda || 0).toFixed(2)}</span>
            </div>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">✓ Monto pagado:</span>
                <span class="vendedor-stat-value">$${(vendedor.pagado || 0).toFixed(2)}</span>
            </div>
            <div class="vendedor-stat" style="background: #f0f0f0; padding: 8px; border-radius: 4px; margin-top: 10px;">
                <span class="vendedor-stat-label" style="font-weight: 600;">💰 Total vendedor:</span>
                <span class="vendedor-stat-value" style="font-size: 24px; color: #ff6b35;">$${(vendedor.total || 0).toFixed(2)}</span>
            </div>
            ${ventasSinPagar.length > 0 ? `
                <div class="vendedor-deudas" style="margin-top: 15px; padding: 10px; background: #fff3cd; border-left: 4px solid #ff6b35; border-radius: 4px;">
                    <h4 style="margin: 0 0 8px 0; color: #ff6b35;">⚠️ Clientes que no pagaron (${ventasSinPagar.length})</h4>
                    ${ventasSinPagar.map(venta => `
                        <div class="deuda-item" style="margin: 5px 0; font-size: 14px;">
                            <strong>${venta.cliente}:</strong> $${parseArgentinoFloat(venta.total).toFixed(2)}
                        </div>
                    `).join('')}
                </div>
            ` : '<div style="color: #28a745; padding: 10px; text-align: center; font-weight: 600; margin-top: 10px;">✓ Todos los clientes pagaron</div>'}
        `;
        container.appendChild(card);
    });
}

function renderizarVentas(filtro = '') {
    if (!datosVentas.ventas) return;

    const tbody = document.getElementById('ventasTableBody');
    tbody.innerHTML = '';

    // Filtrar ventas según el filtro seleccionado
    let ventasFiltradas = datosVentas.ventas;
    
    if (filtro === 'no-entregada') {
        // Mostrar solo ventas no entregadas (estado != 'entregada')
        ventasFiltradas = datosVentas.ventas.filter(v => v.estado !== 'entregada' && v.estado !== 'cancelada');
    } else if (filtro === 'entregada') {
        // Mostrar solo ventas entregadas
        ventasFiltradas = datosVentas.ventas.filter(v => v.estado === 'entregada');
    } else if (filtro === 'delivery') {
        // Solo delivery
        ventasFiltradas = datosVentas.ventas.filter(v => v.tipo_entrega === 'delivery' || v.tipo_entrega === 'envio');
    } else if (filtro === 'retiro') {
        // Solo retiro
        ventasFiltradas = datosVentas.ventas.filter(v => v.tipo_entrega === 'retiro');
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
        const payload = {
            id: ventaEnEdicion.id,
            estado: estado,
            payment_method: pago,
            tipo_entrega: entrega,
            productos: productosActualizados
        };
        
        if (productosAEliminar.length > 0) {
            payload.productos_eliminar = productosAEliminar;
        }

        const response = await fetch(`${API_BASE}/actualizar-venta`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (response.ok) {
            showMessage('✓ Venta actualizada correctamente', 'success');
            cerrarModal();
            cargarDatos();
        } else {
            const err = await response.text();
            showMessage('✗ Error al actualizar: ' + err, 'error');
        }
    } catch (error) {
        showMessage('Error: ' + error.message, 'error');
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

    // Filtro de entregas en la tabla de ventas
    const filtroEntrega = document.getElementById('filtroEntrega');
    if (filtroEntrega) {
        filtroEntrega.addEventListener('change', (e) => {
            renderizarVentas(e.target.value);
        });
    }

    // Volver al home
    document.getElementById('btnVolver').addEventListener('click', () => {
        window.location.href = 'index.html';
    });

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
