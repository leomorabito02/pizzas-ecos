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
    
    // Ocultar spinner cuando se cargan los vendedores
    const loader = document.getElementById('vendedor-loader');
    if (loader) loader.classList.add('hidden');
}

function actualizarSelectProductos() {
    const select = document.getElementById('comboTipo');
    if (!select) return;
    
    select.innerHTML = '<option value="">Selecciona un producto</option>';
    const prods = datosNegocio.productos || [];
    
    prods.forEach(p => {
        const opt = document.createElement('option');
        opt.value = p.id;
        opt.textContent = `${p.tipo_pizza} - $${p.precio}`;
        select.appendChild(opt);
    });
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
        const url = `${API_BASE}/data`;
        Logger.log('📡 Fetching from:', url);
        const resp = await fetch(url);
        
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const jsonResp = await resp.json();
        // El backend retorna {status, data, message}, extraer data
        datosNegocio = (jsonResp && jsonResp.data) ? jsonResp.data : jsonResp;
        Logger.log('✅ Datos cargados:', datosNegocio);
        
        // Actualizar selects
        actualizarSelectVendedores();
        actualizarSelectProductos();
        Logger.log('✅ Selects actualizados');
        
        clearTimeout(spinnerTimeout);
        UIUtils.showSpinner(false);
    } catch (error) {
        console.error('❌ Error inicializando:', error);
        UIUtils.showMessage('Error cargando datos iniciales: ' + error.message, 'error');
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
                    return v.trim() === '' ? 0 : parseInt(v);
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
                    UIUtils.showSpinner(false);
                    if (btn && btnSpinner) {
                        btn.disabled = false;
                        btnText.style.display = 'inline';
                        btnSpinner.style.display = 'none';
                    }
                    return;
                }
                UIUtils.showMessage('✅ Venta registrada', 'success');
                form.reset();
                productosEnVenta = [];
                document.getElementById('pedidoItems').innerHTML = '<div class="pedido-vacio">📋 Agrega productos a tu pedido</div>';
                actualizarResumen();
                UIUtils.showSpinner(false);
                setTimeout(() => window.location.reload(), 1000);
            } catch (err) {
                console.error('Error:', err);
                UIUtils.showSpinner(false);
                UIUtils.showMessage('Error al guardar', 'error');
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

// ============= FUNCIONES AUXILIARES =============
function agregarProductoAlPedido() {
    const tipo = document.getElementById('comboTipo').value;
    const cantidad = parseInt(document.getElementById('comboCantidad').value) || 1;
    
    if (!tipo || !cantidad) {
        UIUtils.showMessage('Completa los datos del producto', 'error');
        return;
    }
    
    const producto = datosNegocio.productos?.find(p => p.id == tipo);
    if (!producto) return;
    
    productosEnVenta.push({
        producto_id: producto.id,
        nombre: producto.tipo_pizza,
        cantidad: cantidad,
        precio: producto.precio,
        total: producto.precio * cantidad
    });
    
    // Reset del formulario - limpiar ambos campos
    document.getElementById('comboTipo').value = '';
    document.getElementById('comboCantidad').value = '1';
    document.getElementById('comboPrice').textContent = '$0.00';
    
    actualizarResumen();
    renderizarPedido();
    verificarBtnAgregarAlPedido();
    UIUtils.showMessage('Producto agregado al pedido', 'success');
}

function actualizarResumen() {
    const total = productosEnVenta.reduce((sum, p) => sum + p.total, 0);
    const el = document.getElementById('totalVenta');
    if (el) el.textContent = UIUtils.formatCurrency(total);
}

function actualizarPrecio() {
    const tipo = document.getElementById('comboTipo').value;
    const cantidad = parseInt(document.getElementById('comboCantidad').value) || 1;
    
    if (!tipo) {
        document.getElementById('comboPrice').textContent = '$0.00';
        return;
    }
    
    const producto = datosNegocio.productos?.find(p => String(p.id) === String(tipo));
    if (!producto) {
        Logger.log('Producto no encontrado:', tipo);
        document.getElementById('comboPrice').textContent = '$0.00';
        return;
    }
    
    const total = producto.precio * cantidad;
    document.getElementById('comboPrice').textContent = UIUtils.formatCurrency(total);
    verificarBtnAgregarAlPedido();
}

function verificarBtnAgregarAlPedido() {
    const btn = document.getElementById('btnAddToPedido');
    if (!btn) return;
    
    const tipo = document.getElementById('comboTipo').value;
    btn.disabled = !tipo;
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
                <button type="button" class="btn-remove" onclick="removerProducto(${i})">✕</button>
            </div>
        </div>
    `).join('');
}

function removerProducto(index) {
    productosEnVenta.splice(index, 1);
    actualizarResumen();
    renderizarPedido();
}
