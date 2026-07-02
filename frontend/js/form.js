/**
 * form.js - Mejorado con arquitectura MVC
 * Maneja la interacción del formulario de ventas usando los controllers
 */

let productosEnVenta = [];
let datosNegocio = {};
let API_BASE = null;  // Se inicializa en DOMContentLoaded

// Logger condicional - solo en desarrollo
const Logger = {
    isDev: window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1',
    log: (msg, data) => {
        if (Logger.isDev) console.log(msg, data || '');
    }
};

// ============= ACTUALIZAR SELECTS =============
function actualizarSelectVendedores() {
    const select = document.getElementById('vendedor');
    if (!select) return;
    
    select.innerHTML = '<option value="">Selecciona un vendedor</option>';
    const vends = datosNegocio.vendedores || Object.keys(datosNegocio.clientesPorVendedor || {});
    
    vends.forEach(v => {
        const opt = document.createElement('option');
        const vendedorNombre = typeof v === 'string' ? v : (v.nombre || v);
        opt.value = vendedorNombre;
        opt.textContent = vendedorNombre;
        select.appendChild(opt);
    });
    
    // Autoseleccionar vendedor guardado
    const ultimoVendedor = localStorage.getItem('ultimoVendedor');
    if (ultimoVendedor) {
        select.value = ultimoVendedor;
        // Disparar change para cargar clientes de ese vendedor
        setTimeout(() => select.dispatchEvent(new Event('change')), 100);
    }
    
    // Ocultar spinner cuando se cargan los vendedores
    const loader = document.getElementById('vendedor-loader');
    if (loader) loader.classList.add('hidden');
}

function actualizarSelectProductos() {
    const grid = document.getElementById('posProductsGrid');
    if (!grid) return;
    
    grid.innerHTML = '';
    const prods = datosNegocio.productos || [];
    
    if (prods.length === 0) {
        grid.innerHTML = '<div class="loader-pos">No hay productos disponibles</div>';
        return;
    }
    
    prods.forEach(p => {
        const card = document.createElement('div');
        card.className = 'pos-product-card';
        
        const itemEnCarrito = productosEnVenta.find(item => item.producto_id === p.id);
        const qty = itemEnCarrito ? itemEnCarrito.cantidad : 0;
        
        card.innerHTML = `
            <div class="pos-product-name">${p.tipo_pizza}</div>
            <div class="pos-product-price">$${p.precio}</div>
            <div class="pos-product-controls">
                <button type="button" class="pos-btn-control btn-minus" data-id="${p.id}">-</button>
                <span class="pos-product-qty" id="qty-${p.id}">${qty}</span>
                <button type="button" class="pos-btn-control btn-plus" data-id="${p.id}">+</button>
            </div>
        `;
        
        grid.appendChild(card);
    });

    document.querySelectorAll('.btn-plus').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            modificarCantidadPOS(parseInt(btn.getAttribute('data-id')), 1);
        });
    });

    document.querySelectorAll('.btn-minus').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            modificarCantidadPOS(parseInt(btn.getAttribute('data-id')), -1);
        });
    });
}

function modificarCantidadPOS(producto_id, delta) {
    const prods = datosNegocio.productos || [];
    const prod = prods.find(p => p.id === producto_id);
    if (!prod) return;

    let itemIndex = productosEnVenta.findIndex(item => item.producto_id === producto_id);
    
    if (itemIndex > -1) {
        productosEnVenta[itemIndex].cantidad += delta;
        if (productosEnVenta[itemIndex].cantidad <= 0) {
            productosEnVenta.splice(itemIndex, 1);
        } else {
            productosEnVenta[itemIndex].total = productosEnVenta[itemIndex].cantidad * prod.precio;
        }
    } else if (delta > 0) {
        productosEnVenta.push({
            producto_id: prod.id,
            nombre: prod.tipo_pizza, // Fijado
            cantidad: delta,
            precio: prod.precio,
            total: delta * prod.precio
        });
    }

    const qtyElement = document.getElementById(`qty-${producto_id}`);
    if (qtyElement) {
        const itemActual = productosEnVenta.find(item => item.producto_id === producto_id);
        qtyElement.textContent = itemActual ? itemActual.cantidad : 0;
    }

    actualizarResumen();
    renderizarPedido();
}

// ============= EVENT LISTENERS PRINCIPAL =============
document.addEventListener('DOMContentLoaded', async () => {
    Logger.log('🚀 Inicializando formulario de ventas...');
    
    // Timeout de seguridad para detener el spinner
    const spinnerTimeout = setTimeout(() => {
        console.warn('⚠️ Timeout cargando datos');
        UIUtils.showSpinner(false);
    }, 10000);
    
    try {
        // Esperar a que env.js haya establecido window.BACKEND_URL
        let retries = 0;
        while (!window.BACKEND_URL && retries < 50) {
            await new Promise(resolve => setTimeout(resolve, 10));
            retries++;
        }
        
        if (!window.BACKEND_URL) {
            throw new Error('BACKEND_URL no fue establecida por env.js');
        }
        
        // Usar la instancia global 'api' que ya existe
        API_BASE = api.baseURL;
        Logger.log('✅ API_BASE:', API_BASE);
        
        // Cargar datos iniciales
        UIUtils.showSpinner(true);
        Logger.log('📡 Fetching data via APIService...');
        const jsonResp = await api.getData();
        // El backend retorna {status, data, message}, extraer data
        datosNegocio = (jsonResp && jsonResp.data) ? jsonResp.data : jsonResp;
        Logger.log('✅ Datos cargados:', datosNegocio);
        
        // Actualizar selects
        actualizarSelectVendedores();
        actualizarSelectProductos();
        Logger.log('✅ Selects actualizados');

        // Precargar estadísticas en background
        setTimeout(() => {
            Logger.log('📡 Precargando estadísticas en background...');
            api.obtenerEstadisticas(10, 1).catch(e => Logger.log('Error precarga stats:', e));
        }, 500);
    } catch (error) {
        console.error('❌ Error inicializando:', error);
        UIUtils.showMessage('Error cargando datos iniciales: ' + error.message, 'error');
    } finally {
        clearTimeout(spinnerTimeout);
        UIUtils.showSpinner(false);
    }
    
    // Setup eventos
    const btnVentas = document.getElementById('btnVerVentas');
    if (btnVentas) btnVentas.addEventListener('click', () => window.location.href = 'estadisticas.html');
    
    const btnAdmin = document.getElementById('btnAdminPanel');
    if (btnAdmin) {
        btnAdmin.addEventListener('click', () => {
            if (localStorage.getItem('authToken')) {
                window.location.href = 'admin.html';
            } else {
                window.location.href = 'login.html';
            }
        });
    }
    
    const vend = document.getElementById('vendedor');
    if (vend) {
        vend.addEventListener('change', (e) => {
            if (e.target.value) localStorage.setItem('ultimoVendedor', e.target.value);
            const clientes = (datosNegocio.clientesPorVendedor && datosNegocio.clientesPorVendedor[e.target.value]) || [];
            const drop = document.getElementById('clientes-dropdown');
            const lista = document.getElementById('clientes-list');
            if (e.target.value && clientes.length > 0) {
                drop.classList.remove('hidden');
                lista.innerHTML = '';
                    clientes.forEach(c => {
                        const div = document.createElement('div');
                        div.className = 'cliente-item';
                        // c puede ser string (legacy) o objeto {id, nombre, telefono}
                        const nombre = (typeof c === 'string') ? c : c.nombre;
                        const telefono = (typeof c === 'string') ? null : c.telefono;
                        div.textContent = nombre;
                        div.addEventListener('click', () => {
                            document.getElementById('cliente').value = nombre;
                            // Si tenemos teléfono, cargarlo en el input (editable)
                            const telInput = document.getElementById('telefono_cliente');
                            if (telInput) {
                                if (telefono) telInput.value = telefono;
                                else telInput.value = '';
                            }
                            drop.classList.add('hidden');
                        });
                        lista.appendChild(div);
                    });
            } else {
                drop.classList.add('hidden');
            }
        });
    }
    
    const btnClose = document.querySelector('.btn-close-dropdown');
    if (btnClose) btnClose.addEventListener('click', (e) => {
        e.preventDefault();
        document.getElementById('clientes-dropdown').classList.add('hidden');
    });
    
    const cliente = document.getElementById('cliente');
    if (cliente) {
        cliente.addEventListener('focus', () => {
            if (document.getElementById('vendedor').value) {
                document.getElementById('clientes-dropdown').classList.remove('hidden');
            }
        });
        cliente.addEventListener('click', () => {
            if (document.getElementById('vendedor').value) {
                document.getElementById('clientes-dropdown').classList.remove('hidden');
            }
        });
        cliente.addEventListener('input', (e) => {
            const term = e.target.value.toLowerCase();
            const items = document.querySelectorAll('.cliente-item');
            let hasVisible = false;
            items.forEach(item => {
                if (item.textContent.toLowerCase().includes(term)) {
                    item.style.display = 'block';
                    hasVisible = true;
                } else {
                    item.style.display = 'none';
                }
            });
            if (document.getElementById('vendedor').value && hasVisible) {
                document.getElementById('clientes-dropdown').classList.remove('hidden');
            }
        });
        cliente.addEventListener('blur', () => {
            setTimeout(() => document.getElementById('clientes-dropdown').classList.add('hidden'), 200);
        });
    }
    
    // Los listeners de los viejos botones de cantidad han sido reemplazados por la cuadrícula POS
    
    // Configurar Chips Selectors
    const setupChips = (containerId, inputId) => {
        const container = document.getElementById(containerId);
        const input = document.getElementById(inputId);
        if (container && input) {
            const btns = container.querySelectorAll('.chip-btn');
            btns.forEach(btn => {
                btn.addEventListener('click', () => {
                    btns.forEach(b => b.classList.remove('selected'));
                    btn.classList.add('selected');
                    input.value = btn.getAttribute('data-value');
                });
            });
        }
    };
    
    setupChips('chips-pago', 'payment_method');
    setupChips('chips-estado', 'estado');
    
    const form = document.getElementById('ventaForm');
    if (form) {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (!productosEnVenta.length) {
                UIUtils.showMessage('Agrega al menos un producto', 'error');
                return;
            }
            const vend = document.getElementById('vendedor').value.trim();
            const cliente = document.getElementById('cliente').value.trim();
            const pago = document.getElementById('payment_method').value;
            const est = document.getElementById('estado').value;
            const tip = document.querySelector('input[name="tipo_entrega"]:checked')?.value;
            if (!vend || !cliente || !pago || !est || !tip) {
                UIUtils.showMessage('Completa todos los campos (vendedor y cliente requeridos)', 'error');
                return;
            }
            if (cliente.length === 0) {
                UIUtils.showMessage('El cliente no puede estar vacío', 'error');
                return;
            }
            const combos = productosEnVenta.map(p => ({
                tipo: 'producto',
                product_id: p.producto_id,
                cantidad: p.cantidad,
                precio: p.precio,
                total: p.total
            }));
            const data = {
                vendedor: vend,
                cliente: cliente,
                telefono_cliente: (function(){
                    const v = document.getElementById('telefono_cliente').value || '';
                    const trimmed = v.replace(/\D/g, ''); // Extraer solo dígitos
                    if (trimmed === '') return 0;
                    return parseInt(trimmed, 10) || 0;
                })(),
                items: combos,
                payment_method: pago,
                estado: est,
                tipo_entrega: tip
            };
            const btn = document.querySelector('.btn-submit');
            const btnText = document.querySelector('.btn-text');
            const btnSpinner = document.querySelector('.btn-spinner');
            if (btn && btnSpinner) {
                btn.disabled = true;
                btnText.style.display = 'none';
                btnSpinner.style.display = 'inline-block';
            }
            UIUtils.showSpinner(true);
            try {
                const resp = await fetch(`${API_BASE}/submit`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                if (!resp.ok) {
                    const err = await resp.json();
                    UIUtils.showMessage(err.message || 'Error al guardar', 'error');
                    return;
                }
                UIUtils.showMessage('✅ Venta registrada', 'success');
                form.reset();
                productosEnVenta = [];
                document.getElementById('pedidoItems').innerHTML = '<div class="pedido-vacio">📋 Agrega productos a tu pedido</div>';
                actualizarResumen();
                setTimeout(() => window.location.reload(), 1000);
            } catch (err) {
                console.error('Error:', err);
                UIUtils.showMessage('Error al guardar', 'error');
            } finally {
                UIUtils.showSpinner(false);
                if (btn && btnSpinner) {
                    btn.disabled = false;
                    btnText.style.display = 'inline';
                    btnSpinner.style.display = 'none';
                }
            }
        });
    }
});

// ============= FUNCIONES AUXILIARES =============
function actualizarResumen() {
    const total = productosEnVenta.reduce((sum, p) => sum + p.total, 0);
    const el = document.getElementById('totalVenta');
    if (el) el.textContent = UIUtils.formatCurrency(total);
}

function renderizarPedido() {
    const container = document.getElementById('pedidoItems');
    if (!container) return;
    
    if (!productosEnVenta.length) {
        container.innerHTML = '<div class="pedido-vacio">📋 Agrega productos a tu pedido</div>';
        return;
    }
    
    container.innerHTML = productosEnVenta.map((p, i) => `
        <div class="pedido-item">
            <div class="item-info">
                <strong>${p.nombre}</strong>
                <span>${p.cantidad} x ${UIUtils.formatCurrency(p.precio)}</span>
            </div>
            <div class="item-total">
                <strong>${UIUtils.formatCurrency(p.total)}</strong>
                <button type="button" class="btn-remove-pedido-text" onclick="removerProducto(${i})">✕ Quitar</button>
            </div>
        </div>
    `).join('');
}

function removerProducto(index) {
    productosEnVenta.splice(index, 1);
    actualizarResumen();
    renderizarPedido();
}
