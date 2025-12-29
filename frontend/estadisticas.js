// estadisticas.js
let datosVentas = {};
let ventaEnEdicion = null;

function getAPIBase() {
    // Si existe variable de entorno, usarla (para Vercel/Netlify)
    if (typeof window !== 'undefined' && window.ENV?.API_URL) {
        return window.ENV.API_URL;
    }
    
    // Si está en localhost, usar localhost:8080
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
        return 'http://localhost:8080/api';
    }
    
    // Si está en producción, usar la misma IP/dominio que el frontend
    const protocol = window.location.protocol; // http: o https:
    const host = window.location.hostname;
    const port = 8080;
    return `${protocol}//${host}:${port}/api`;
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
        // Obtener datos del backend desde el sheet "estadisticas"
        const response1 = await fetch(`${API_BASE}/estadisticas-sheet`);
        if (!response1.ok) throw new Error('No se pudieron cargar las estadísticas');
        datosVentas = await response1.json();
        
        // Obtener detalle de ventas para la tabla "Todas las Ventas"
        const response2 = await fetch(`${API_BASE}/estadisticas`);
        if (response2.ok) {
            const ventasData = await response2.json();
            datosVentas.ventas = ventasData.ventas; // Agregar las ventas al objeto
        }
        
        console.log('Datos de estadísticas:', datosVentas);
        
        // Renderizar tabs
        renderizarResumen();
        renderizarVendedores();
        renderizarVentas();
    } catch (error) {
        console.error('Error:', error);
        showMessage('Error al cargar estadísticas: ' + error.message, 'error');
    }
}

function renderizarResumen() {
    if (!datosVentas.resumen) return;

    const resumen = datosVentas.resumen;

    // Actualizar UI con datos del sheet
    document.getElementById('totalMuzzas').textContent = Math.round(resumen.total_muzzas);
    document.getElementById('totalJamones').textContent = Math.round(resumen.total_jamones);
    document.getElementById('totalDelivery').textContent = Math.round(resumen.total_delivery);
    document.getElementById('totalRetiro').textContent = Math.round(resumen.total_retiro);
    document.getElementById('pendienteCobro').textContent = `$${resumen.pendiente_cobro.toFixed(2)}`;
    document.getElementById('efectivoCobrado').textContent = `$${resumen.efectivo_cobrado.toFixed(2)}`;
    document.getElementById('transferenciaCobrada').textContent = `$${resumen.transferencia_cobrada.toFixed(2)}`;
    document.getElementById('totalCobrado').textContent = `$${resumen.total_cobrado.toFixed(2)}`;
    document.getElementById('ventasSinPagar').textContent = Math.round(resumen.ventas_sin_pagar);
    document.getElementById('ventasPagadas').textContent = Math.round(resumen.ventas_pagadas);
    document.getElementById('ventasEntregadas').textContent = Math.round(resumen.ventas_entregadas);
    document.getElementById('totalVentas').textContent = Math.round(resumen.ventas_totales);
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

        const card = document.createElement('div');
        card.className = 'vendedor-card';
        card.innerHTML = `
            <h3>👤 ${vendedor.nombre}</h3>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">📊 Cantidad de ventas:</span>
                <span class="vendedor-stat-value">${Math.round(vendedor.cantidad_ventas)}</span>
            </div>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">🧀 Muzzas vendidas:</span>
                <span class="vendedor-stat-value">${Math.round(vendedor.muzzas)}</span>
            </div>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">🍖 Jamones vendidos:</span>
                <span class="vendedor-stat-value">${Math.round(vendedor.jamones)}</span>
            </div>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">⏳ Sin pagar:</span>
                <span class="vendedor-stat-value">$${vendedor.sin_pagar.toFixed(2)}</span>
            </div>
            <div class="vendedor-stat">
                <span class="vendedor-stat-label">✓ Pagado:</span>
                <span class="vendedor-stat-value">$${vendedor.pagado.toFixed(2)}</span>
            </div>
            <div class="vendedor-stat" style="background: #f0f0f0; padding: 8px; border-radius: 4px; margin-top: 10px;">
                <span class="vendedor-stat-label" style="font-weight: 600;">💰 Total vendedor:</span>
                <span class="vendedor-stat-value" style="font-size: 24px; color: #ff6b35;">$${vendedor.total.toFixed(2)}</span>
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

function renderizarVentas() {
    if (!datosVentas.ventas) return;

    const tbody = document.getElementById('ventasTableBody');
    tbody.innerHTML = '';

    datosVentas.ventas.forEach(venta => {
        const combosResumen = venta.combos.map(c => 
            `${c.cantidad}x ${c.tipo.toUpperCase()} C${c.combo + 1}`
        ).join(', ');

        const estadoClass = venta.estado.replace(' ', '-');
        const totalParseado = parseArgentinoFloat(venta.total);
        
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${venta.id}</td>
            <td>${venta.vendedor}</td>
            <td>${venta.cliente}</td>
            <td style="font-size: 12px;">${combosResumen}</td>
            <td><strong>$${totalParseado.toFixed(2)}</strong></td>
            <td><span class="estado-badge ${estadoClass}">${venta.estado}</span></td>
            <td>${venta.payment_method === 'efectivo' ? '💵' : '🏦'}</td>
            <td>${venta.tipo_entrega === 'envio' ? '🚚' : '🏪'}</td>
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
    document.getElementById('editarEstado').value = venta.estado;
    document.getElementById('editarPago').value = venta.payment_method;
    
    // Cargar combos
    // Inicializar todos en 0
    document.getElementById('editMuzzaC1').value = 0;
    document.getElementById('editMuzzaC2').value = 0;
    document.getElementById('editMuzzaC3').value = 0;
    document.getElementById('editJamonC1').value = 0;
    document.getElementById('editJamonC2').value = 0;
    document.getElementById('editJamonC3').value = 0;
    
    // Cargar valores desde los combos de la venta
    if (venta.combos) {
        venta.combos.forEach(combo => {
            const fieldId = `edit${combo.tipo === 'muzza' ? 'Muzza' : 'Jamon'}C${combo.combo + 1}`;
            document.getElementById(fieldId).value = combo.cantidad;
        });
    }
    
    document.getElementById('modalEditarVenta').classList.remove('hidden');
}

function cerrarModal() {
    document.getElementById('modalEditarVenta').classList.add('hidden');
    ventaEnEdicion = null;
}

async function guardarCambios() {
    if (!ventaEnEdicion) return;

    const estado = document.getElementById('editarEstado').value;
    const pago = document.getElementById('editarPago').value;
    
    // Recopilar combos editados
    const combosEditados = [];
    
    // Muzzas
    const muzzaC1 = parseInt(document.getElementById('editMuzzaC1').value) || 0;
    const muzzaC2 = parseInt(document.getElementById('editMuzzaC2').value) || 0;
    const muzzaC3 = parseInt(document.getElementById('editMuzzaC3').value) || 0;
    
    if (muzzaC1 > 0) combosEditados.push({ tipo: 'muzza', combo: 0, cantidad: muzzaC1 });
    if (muzzaC2 > 0) combosEditados.push({ tipo: 'muzza', combo: 1, cantidad: muzzaC2 });
    if (muzzaC3 > 0) combosEditados.push({ tipo: 'muzza', combo: 2, cantidad: muzzaC3 });
    
    // Jamones
    const jamonC1 = parseInt(document.getElementById('editJamonC1').value) || 0;
    const jamonC2 = parseInt(document.getElementById('editJamonC2').value) || 0;
    const jamonC3 = parseInt(document.getElementById('editJamonC3').value) || 0;
    
    if (jamonC1 > 0) combosEditados.push({ tipo: 'jamon', combo: 0, cantidad: jamonC1 });
    if (jamonC2 > 0) combosEditados.push({ tipo: 'jamon', combo: 1, cantidad: jamonC2 });
    if (jamonC3 > 0) combosEditados.push({ tipo: 'jamon', combo: 2, cantidad: jamonC3 });

    try {
        const response = await fetch(`${API_BASE}/actualizar-venta`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                id: ventaEnEdicion.id,
                estado: estado,
                payment_method: pago,
                combos: combosEditados
            })
        });

        if (response.ok) {
            showMessage('✓ Venta actualizada correctamente', 'success');
            cerrarModal();
            cargarDatos();
        } else {
            showMessage('✗ Error al actualizar', 'error');
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

    // Volver al home
    document.getElementById('btnVolver').addEventListener('click', () => {
        window.location.href = 'index.html';
    });

    // Modal
    document.querySelector('.btn-close-modal').addEventListener('click', cerrarModal);
    document.querySelector('.btn-cancelar-modal').addEventListener('click', cerrarModal);
    document.querySelector('.btn-guardar-cambios').addEventListener('click', guardarCambios);

    document.getElementById('modalEditarVenta').addEventListener('click', (e) => {
        if (e.target === document.getElementById('modalEditarVenta')) {
            cerrarModal();
        }
    });
});
