// form.js
let datosNegocio = {};
let productosEnVenta = [];
let productosDisponibles = [];

// Loading Spinner Functions
function showLoadingSpinner(show = true) {
    const overlay = document.getElementById('loadingOverlay');
    if (overlay) {
        if (show) {
            overlay.classList.remove('hidden');
        } else {
            overlay.classList.add('hidden');
        }
    }
}

function hideLoadingSpinner() {
    showLoadingSpinner(false);
}

const getAPIBase = () => {
    // 1. Si hay variable de entorno (Netlify, Render, etc)
    const envUrl = window.__ENV?.REACT_APP_API_URL || window.REACT_APP_API_URL;
    if (envUrl) {
        const url = envUrl.endsWith('/api') ? envUrl : envUrl + '/api';
        console.log('✅ API Base from env:', url);
        return url;
    }
    
    // 2. Si está en localhost, usar localhost:8080
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
        return 'http://localhost:8080/api';
    }
    
    // 3. En producción, asumir backend en mismo dominio
    const protocol = window.location.protocol;
    const host = window.location.hostname;
    const url = `${protocol}//${host}/api`;
    console.log('ℹ️  Using same-server API:', url);
    return url;
};

const API_BASE = getAPIBase();
console.log('API Base URL:', API_BASE);

function parseArgentinoFloat(value) {
    if (typeof value === 'number') return value;
    if (!value) return 0;
    let str = String(value).trim().replace('$', '').replace(/\./g, '').replace(',', '.');
    return parseFloat(str) || 0;
}

function actualizarSeleccionProducto() {
    const select = document.getElementById('comboTipo');
    if (!select || !productosDisponibles.length) return;
    select.innerHTML = '<option value="">Selecciona un producto...</option>';
    productosDisponibles.forEach(p => {
        const opt = document.createElement('option');
        opt.value = p.id;
        opt.textContent = `${p.tipo_pizza} - $${(p.precio || 0).toFixed(2)}`;
        select.appendChild(opt);
    });
}

function actualizarPrecio() {
    const id = parseInt(document.getElementById('comboTipo').value);
    const cant = parseInt(document.getElementById('comboCantidad').value) || 1;
    const precio = document.getElementById('comboPrice');
    if (!id) { precio.textContent = '$0.00'; return; }
    const prod = productosDisponibles.find(p => p.id === id);
    if (!prod) { precio.textContent = '$0.00'; return; }
    precio.textContent = `$${((prod.precio || 0) * cant).toFixed(2)}`;
}

function agregarProductoAlPedido() {
    const id = parseInt(document.getElementById('comboTipo').value);
    const cant = parseInt(document.getElementById('comboCantidad').value) || 1;
    if (!id) { showMessage('Selecciona un producto', 'error'); return; }
    const prod = productosDisponibles.find(p => p.id === id);
    if (!prod) { showMessage('Producto no encontrado', 'error'); return; }
    productosEnVenta.push({
        id: Date.now(),
        producto_id: id,
        tipo: prod.tipo_pizza,
        cantidad: cant,
        precio: prod.precio,
        total: prod.precio * cant
    });
    renderizarPedido();
    actualizarResumen();
    resetearSelectorProductos();
}

function verificarBtnAgregarAlPedido() {
    const btn = document.getElementById('btnAddToPedido');
    const sel = document.getElementById('comboTipo');
    if (!btn || !sel) return;
    const ok = sel.value !== '';
    btn.disabled = !ok;
    btn.style.opacity = ok ? '1' : '0.5';
    btn.style.cursor = ok ? 'pointer' : 'not-allowed';
}

function resetearSelectorProductos() {
    document.getElementById('comboTipo').value = '';
    document.getElementById('comboCantidad').value = 1;
    actualizarPrecio();
}

function renderizarPedido() {
    const container = document.getElementById('pedidoItems');
    if (!productosEnVenta.length) {
        container.innerHTML = '<div class="pedido-vacio">📋 Agrega productos a tu pedido</div>';
        return;
    }
    container.innerHTML = '';
    productosEnVenta.forEach((item, idx) => {
        const div = document.createElement('div');
        div.className = 'pedido-item';
        div.innerHTML = `
            <div class="pedido-item-info">
                <div class="pedido-item-name">🍕 ${item.tipo}</div>
                <div class="pedido-item-cantidad">Cantidad: ${item.cantidad}</div>
            </div>
            <div class="pedido-item-price">$${item.total.toFixed(2)}</div>
            <button type="button" class="btn-remove-pedido" data-index="${idx}">✕ Quitar</button>
        `;
        container.appendChild(div);
    });
    document.querySelectorAll('.btn-remove-pedido').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            const idx = parseInt(e.target.dataset.index);
            productosEnVenta.splice(idx, 1);
            renderizarPedido();
            actualizarResumen();
        });
    });
}

function actualizarResumen() {
    const span = document.getElementById('totalVenta');
    let total = 0;
    productosEnVenta.forEach(item => total += item.total);
    if (span) span.textContent = `$${total.toFixed(2)}`;
}

function showMessage(text, type) {
    const msg = document.getElementById('mensaje');
    msg.textContent = text;
    msg.className = type === 'error' ? 'error' : 'success';
    if (type !== 'error') {
        setTimeout(() => { msg.className = ''; msg.textContent = ''; }, 5000);
    }
}

async function init() {
    try {
        showLoadingSpinner(true);
        
        const vend = document.getElementById('vendedor');
        const loader = document.getElementById('vendedor-loader');
        if (loader) loader.style.display = 'inline-block';
        
        console.log('Iniciando carga de datos...');
        console.log('API_BASE:', API_BASE);
        
        const url = `${API_BASE}/api/data`;
        console.log('Fetching from:', url);
        
        const resp = await fetch(url);
        console.log('Response status:', resp.status);
        
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        datosNegocio = await resp.json();
        console.log('Datos cargados:', datosNegocio);
        
        vend.innerHTML = '<option value="">Selecciona un vendedor</option>';
        const vends = datosNegocio.vendedores || Object.keys(datosNegocio.clientesPorVendedor || {});
        console.log('Vendedores:', vends);
        
        vends.forEach(v => {
            const opt = document.createElement('option');
            // Manejar tanto objetos {id, nombre} como strings
            const vendedorNombre = typeof v === 'string' ? v : (v.nombre || v);
            opt.value = vendedorNombre;
            opt.textContent = vendedorNombre;
            vend.appendChild(opt);
        });
        
        if (datosNegocio.productos && Array.isArray(datosNegocio.productos)) {
            productosDisponibles = datosNegocio.productos;
            console.log('Productos cargados:', productosDisponibles);
            actualizarSeleccionProducto();
        } else {
            console.warn('No hay productos en la respuesta:', datosNegocio);
        }
        
        if (loader) loader.style.display = 'none';
        hideLoadingSpinner();
    } catch (err) {
        console.error('Error en init():', err);
        showMessage('Error al cargar datos: ' + err.message, 'error');
        hideLoadingSpinner();
    }
}

document.addEventListener('DOMContentLoaded', () => {
    init();
    
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
            const clientes = (datosNegocio.clientesPorVendedor && datosNegocio.clientesPorVendedor[e.target.value]) || [];
            const drop = document.getElementById('clientes-dropdown');
            const lista = document.getElementById('clientes-list');
            if (e.target.value && clientes.length > 0) {
                drop.classList.remove('hidden');
                lista.innerHTML = '';
                clientes.forEach(c => {
                    const div = document.createElement('div');
                    div.className = 'cliente-item';
                    div.textContent = c;
                    div.addEventListener('click', () => {
                        document.getElementById('cliente').value = c;
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
        cliente.addEventListener('blur', () => {
            setTimeout(() => document.getElementById('clientes-dropdown').classList.add('hidden'), 200);
        });
    }
    
    const comboTipo = document.getElementById('comboTipo');
    if (comboTipo) {
        comboTipo.addEventListener('change', () => {
            actualizarPrecio();
            verificarBtnAgregarAlPedido();
        });
    }
    
    const comboCant = document.getElementById('comboCantidad');
    if (comboCant) comboCant.addEventListener('change', actualizarPrecio);
    
    const btnMas = document.getElementById('btnCantidadMas');
    if (btnMas) {
        btnMas.addEventListener('click', (e) => {
            e.preventDefault();
            comboCant.value = parseInt(comboCant.value) + 1;
            actualizarPrecio();
        });
    }
    
    const btnMenos = document.getElementById('btnCantidadMenos');
    if (btnMenos) {
        btnMenos.addEventListener('click', (e) => {
            e.preventDefault();
            if (parseInt(comboCant.value) > 1) {
                comboCant.value = parseInt(comboCant.value) - 1;
                actualizarPrecio();
            }
        });
    }
    
    const btnAdd = document.getElementById('btnAddToPedido');
    if (btnAdd) {
        btnAdd.addEventListener('click', (e) => {
            e.preventDefault();
            agregarProductoAlPedido();
        });
        verificarBtnAgregarAlPedido();
    }
    
    const btnAnother = document.getElementById('btnAddAnother');
    if (btnAnother) {
        btnAnother.addEventListener('click', (e) => {
            e.preventDefault();
            resetearSelectorProductos();
        });
    }
    
    const form = document.getElementById('ventaForm');
    if (form) {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (!productosEnVenta.length) {
                showMessage('Agrega al menos un producto', 'error');
                return;
            }
            const vend = document.getElementById('vendedor').value;
            const cliente = document.getElementById('cliente').value;
            const pago = document.getElementById('payment_method').value;
            const est = document.getElementById('estado').value;
            const tip = document.querySelector('input[name="tipo_entrega"]:checked')?.value;
            if (!vend || !cliente || !pago || !est || !tip) {
                showMessage('Completa todos los campos', 'error');
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
            showLoadingSpinner(true);
            try {
                const resp = await fetch(`${API_BASE}/api/submit`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                if (!resp.ok) {
                    const err = await resp.json();
                    showMessage(err.message || 'Error al guardar', 'error');
                    hideLoadingSpinner();
                    if (btn && btnSpinner) {
                        btn.disabled = false;
                        btnText.style.display = 'inline';
                        btnSpinner.style.display = 'none';
                    }
                    return;
                }
                showMessage('✅ Venta registrada', 'success');
                form.reset();
                productosEnVenta = [];
                document.getElementById('pedidoItems').innerHTML = '<div class="pedido-vacio">📋 Agrega productos a tu pedido</div>';
                actualizarResumen();
                hideLoadingSpinner();
                setTimeout(() => window.location.reload(), 1000);
            } catch (err) {
                console.error('Error:', err);
                hideLoadingSpinner();
                showMessage('Error al guardar', 'error');
            } finally {
                if (btn && btnSpinner) {
                    btn.disabled = false;
                    btnText.style.display = 'inline';
                    btnSpinner.style.display = 'none';
                }
            }
        });
    }
});
