// form.js
let datosNegocio = {}; // Se llena al cargar la página
let combosEnVenta = []; // Array de combos agregados a la venta

// URL del backend API - configurable según el ambiente
// En desarrollo: http://localhost:8080
// En producción: usar variable de entorno en Netlify
const getAPIBase = () => {
    // 1. Si está en localhost, usar localhost:8080
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
        console.log('ℹ️  Using localhost API: http://localhost:8080/api');
        return 'http://localhost:8080/api';
    }
    
    // 2. En producción, debe estar en variable de entorno
    const apiUrl = window.REACT_APP_API_URL || window.env?.REACT_APP_API_URL;
    if (apiUrl) {
        console.log('✅ API URL from environment:', apiUrl);
        return apiUrl;
    }
    
    console.error('❌ REACT_APP_API_URL no está configurada. Configúrala en Netlify.');
    return null;
};

const API_BASE = getAPIBase();
console.log('API Base URL:', API_BASE);

async function init() {
    try {
        // Mostrar loader mientras carga
        const vendedorSelect = document.getElementById('vendedor');
        const loader = document.getElementById('vendedor-loader');
        if (loader) loader.style.display = 'inline-block';
        
        // Pedimos al Go los datos procesados (Vendedores, clientes históricos y precios)
        const response = await fetch(`${API_BASE}/data`); 
        if (!response.ok) throw new Error('No se pudieron cargar los datos');
        
        datosNegocio = await response.json(); 
        
        // Debug: mostrar los datos cargados - INFORMACIÓN DETALLADA
        console.log('=== DATOS DEL BACKEND ===');
        console.log('Vendedores:', datosNegocio.vendedores);
        console.log('ClientesPorVendedor:', datosNegocio.clientesPorVendedor);
        
        // Llenar select de vendedores
        vendedorSelect.innerHTML = '<option value="">Selecciona un vendedor</option>';
        const vendedores = datosNegocio.vendedores || Object.keys(datosNegocio.clientesPorVendedor || {});
        
        vendedores.forEach(v => {
            const option = document.createElement('option');
            option.value = v;
            option.textContent = v;
            vendedorSelect.appendChild(option);
        });
        
        // Ocultar loader después de cargar
        if (loader) loader.style.display = 'none';
        
        console.log('Datos cargados:', datosNegocio);
    } catch (error) {
        console.error('Error en init:', error);
        showMessage('Error al cargar vendedores', 'error');
    }
}

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

// Función auxiliar para obtener el precio de un combo
function obtenerPrecio(tipo, comboIndex) {
    const pizza = datosNegocio.pizzas?.[tipo];
    if (pizza && pizza.precios && comboIndex >= 0 && comboIndex < pizza.precios.length) {
        return parseArgentinoFloat(pizza.precios[comboIndex]);
    }
    return 0;
}

document.addEventListener('DOMContentLoaded', () => {
    init();
    
    // Botón Ver Estadísticas
    const btnVerVentas = document.getElementById('btnVerVentas');
    if (btnVerVentas) {
        btnVerVentas.addEventListener('click', () => {
            window.location.href = 'estadisticas.html';
        });
    }
    
    const vendedorSelect = document.getElementById('vendedor');
    if (vendedorSelect) {
        vendedorSelect.addEventListener('change', (e) => {
            const vendedor = e.target.value;
            const clientes = (datosNegocio.clientesPorVendedor && datosNegocio.clientesPorVendedor[vendedor]) || [];
            
            // Mostrar/ocultar dropdown de clientes
            const dropdown = document.getElementById('clientes-dropdown');
            const clientesList = document.getElementById('clientes-list');
            
            if (vendedor && clientes.length > 0) {
                // Mostrar dropdown con clientes
                dropdown.classList.remove('hidden');
                clientesList.innerHTML = '';
                
                clientes.forEach(cliente => {
                    const div = document.createElement('div');
                    div.className = 'cliente-item';
                    div.textContent = cliente;
                    div.addEventListener('click', () => {
                        document.getElementById('cliente').value = cliente;
                        dropdown.classList.add('hidden');
                    });
                    clientesList.appendChild(div);
                });
            } else {
                dropdown.classList.add('hidden');
            }
        });
    }
    
    // Botón para cerrar dropdown de clientes
    const btnCloseDropdown = document.querySelector('.btn-close-dropdown');
    if (btnCloseDropdown) {
        btnCloseDropdown.addEventListener('click', (e) => {
            e.preventDefault();
            document.getElementById('clientes-dropdown').classList.add('hidden');
        });
    }
    
    // Campo de cliente: mostrar/ocultar dropdown al escribir
    const clienteInput = document.getElementById('cliente');
    if (clienteInput) {
        clienteInput.addEventListener('focus', () => {
            const vendedor = document.getElementById('vendedor').value;
            if (vendedor) {
                document.getElementById('clientes-dropdown').classList.remove('hidden');
            }
        });
        
        clienteInput.addEventListener('blur', () => {
            // Ocultar después de un pequeño delay para permitir click
            setTimeout(() => {
                document.getElementById('clientes-dropdown').classList.add('hidden');
            }, 200);
        });
    }
    
    // Botón para agregar combos
    const btnAddCombo = document.getElementById('btnAddCombo');
    if (btnAddCombo) {
        btnAddCombo.addEventListener('click', (e) => {
            e.preventDefault();
            agregarCombo();
        });
    }
    
    // Manejar el envío del formulario
    document.getElementById('ventaForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        
        // Validar que haya al menos un combo
        if (combosEnVenta.length === 0) {
            showMessage('Debes agregar al menos un combo', 'error');
            return;
        }
        
        // Validar que todos los combos tengan un tipo seleccionado (no "Selecciona")
        for (let combo of combosEnVenta) {
            if (combo.combo < 0) {
                showMessage('⚠️ Todos los combos deben tener un combo seleccionado', 'error');
                return;
            }
        }
        
        // Validar datos básicos
        const vendedor = document.getElementById('vendedor').value;
        const cliente = document.getElementById('cliente').value;
        const paymentMethod = document.getElementById('payment_method').value;
        const estado = document.getElementById('estado').value;
        const tipoEntrega = document.querySelector('input[name="tipo_entrega"]:checked')?.value;
        
        if (!vendedor || !cliente || !paymentMethod || !estado || !tipoEntrega) {
            showMessage('Por favor completa todos los campos requeridos', 'error');
            return;
        }
        
        const data = {
            vendedor: vendedor,
            cliente: cliente,
            combos: combosEnVenta,
            payment_method: paymentMethod,
            estado: estado,
            tipo_entrega: tipoEntrega
        };
        
        // Bloquear botón y mostrar spinner
        const btn = document.querySelector('.btn-submit');
        const btnText = document.querySelector('.btn-text');
        const btnSpinner = document.querySelector('.btn-spinner');
        
        btn.disabled = true;
        btnText.classList.add('hidden');
        btnSpinner.classList.remove('hidden');
        
        try {
            const response = await fetch(`${API_BASE}/submit`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            
            if (response.ok) {
                showMessage('✓ Venta guardada correctamente', 'success');
                document.getElementById('ventaForm').reset();
                combosEnVenta = [];
                document.getElementById('combosList').innerHTML = '';
                actualizarResumen();
            } else {
                const error = await response.text();
                showMessage('✗ Error al guardar: ' + error, 'error');
            }
        } catch (error) {
            showMessage('✗ Error de conexión: ' + error.message, 'error');
            console.error('Error:', error);
        } finally {
            // Restaurar botón
            btn.disabled = false;
            btnText.classList.remove('hidden');
            btnSpinner.classList.add('hidden');
        }
    });
});

function agregarCombo() {
    const id = Date.now(); // ID único para el combo
    
    combosEnVenta.push({
        id: id,
        tipo: 'muzza',
        combo: -1, // Cambiar a -1 para indicar que no hay combo seleccionado
        cantidad: 1,
        precio: 0,
        total: 0
    });
    
    // Deshabilitar botón para agregar más combos
    const btnAddCombo = document.getElementById('btnAddCombo');
    if (btnAddCombo) {
        btnAddCombo.disabled = true;
    }
    
    renderizarCombos();
}

function verificarYHabilitarBotomAgregar() {
    const btnAddCombo = document.getElementById('btnAddCombo');
    if (!btnAddCombo) return;
    
    // Verificar si hay algún combo sin completar (combo < 0)
    const hayComboIncompleto = combosEnVenta.some(combo => combo.combo < 0);
    
    btnAddCombo.disabled = hayComboIncompleto;
}

function renderizarCombos() {
    const combosList = document.getElementById('combosList');
    combosList.innerHTML = '';
    
    combosEnVenta.forEach((combo, index) => {
        const div = document.createElement('div');
        div.className = 'combo-item';
        
        const pizza = datosNegocio.pizzas?.[combo.tipo];
        
        // Obtener precio del backend usando la función auxiliar
        const precioActual = obtenerPrecio(combo.tipo, combo.combo);
        
        // Actualizar el combo con el precio correcto
        combo.precio = precioActual;
        combo.total = precioActual * combo.cantidad;
        
        // Debug: mostrar datos en consola
        console.log(`Combo ${index}:`, {tipo: combo.tipo, combo: combo.combo, precio: precioActual, cantidad: combo.cantidad, total: combo.total});
        
        div.innerHTML = `
            <div>
                <div class="combo-item-label">Tipo</div>
                <select data-index="${index}" class="combo-tipo">
                    <option value="muzza" ${combo.tipo === 'muzza' ? 'selected' : ''}>🧀 Muzza</option>
                    <option value="jamon" ${combo.tipo === 'jamon' ? 'selected' : ''}>🍖 Jamón</option>
                </select>
            </div>
            
            <div>
                <div class="combo-item-label">Combo</div>
                <select data-index="${index}" class="combo-combo">
                    <option value="" ${combo.combo < 0 ? 'selected' : ''}>Selecciona combo...</option>
                    ${pizza ? pizza.combos.map((c, i) => {
                        const precio = parseArgentinoFloat(pizza.precios[i]);
                        return `
                            <option value="${i}" ${combo.combo === i ? 'selected' : ''}>
                                ${c} - $${precio.toFixed(2)}
                            </option>
                        `;
                    }).join('') : ''}
                </select>
            </div>
            
            <div class="combo-cantidad-container">
                <div class="combo-item-label">Cantidad</div>
                <div class="cantidad-controls">
                    <button type="button" class="btn-cantidad-menos" data-index="${index}">−</button>
                    <span class="cantidad-display" data-index="${index}">${combo.cantidad}</span>
                    <button type="button" class="btn-cantidad-mas" data-index="${index}">+</button>
                </div>
            </div>
            
            <div>
                <div class="combo-item-price">$${combo.total.toFixed(2)}</div>
                <button type="button" class="btn-remove-combo" data-index="${index}">
                    ✕ Quitar
                </button>
            </div>
        `;
        
        combosList.appendChild(div);
    });
    
    // Agregar event listeners
    document.querySelectorAll('.combo-tipo').forEach(select => {
        select.addEventListener('change', (e) => {
            const index = parseInt(e.target.dataset.index);
            combosEnVenta[index].tipo = e.target.value;
            combosEnVenta[index].combo = -1; // Reset combo (sin seleccionar)
            renderizarCombos();
            actualizarResumen();
            verificarYHabilitarBotomAgregar();
        });
    });
    
    document.querySelectorAll('.combo-combo').forEach(select => {
        select.addEventListener('change', (e) => {
            const index = parseInt(e.target.dataset.index);
            const valor = e.target.value;
            // Si el valor está vacío, no actualizar combo (seguir mostrando el anterior)
            // Si tiene valor, actualizar al índice del combo seleccionado
            if (valor !== '') {
                combosEnVenta[index].combo = parseInt(valor);
            }
            renderizarCombos();
            actualizarResumen();
            verificarYHabilitarBotomAgregar();
        });
    });
    
    // Botones para aumentar cantidad
    document.querySelectorAll('.btn-cantidad-mas').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            const index = parseInt(e.target.dataset.index);
            combosEnVenta[index].cantidad++;
            renderizarCombos();
            actualizarResumen();
            verificarYHabilitarBotomAgregar();
        });
    });
    
    // Botones para disminuir cantidad
    document.querySelectorAll('.btn-cantidad-menos').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            const index = parseInt(e.target.dataset.index);
            if (combosEnVenta[index].cantidad > 1) {
                combosEnVenta[index].cantidad--;
            }
            renderizarCombos();
            actualizarResumen();
            verificarYHabilitarBotomAgregar();
        });
    });
    
    document.querySelectorAll('.btn-remove-combo').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.preventDefault();
            const index = parseInt(e.target.dataset.index);
            combosEnVenta.splice(index, 1);
            renderizarCombos();
            actualizarResumen();
            verificarYHabilitarBotomAgregar();
        });
    });
    
    actualizarResumen();
    verificarYHabilitarBotomAgregar();

function actualizarResumen() {
    const resumenDiv = document.getElementById('resumenCombos');
    const totalVentaSpan = document.getElementById('totalVenta');
    
    // Agrupar por tipo y combo
    const resumen = {
        'muzza-c1': 0,
        'muzza-c2': 0,
        'muzza-c3': 0,
        'jamon-c1': 0,
        'jamon-c2': 0,
        'jamon-c3': 0
    };
    
    let totalVenta = 0;
    
    combosEnVenta.forEach(combo => {
        if (combo.combo >= 0) {
            const key = `${combo.tipo}-c${combo.combo + 1}`;
            resumen[key] += combo.cantidad;
            totalVenta += combo.total;
        }
    });
    
    // Renderizar resumen
    resumenDiv.innerHTML = '';
    
    // Muzza
    if (resumen['muzza-c1'] > 0 || resumen['muzza-c2'] > 0 || resumen['muzza-c3'] > 0) {
        const div = document.createElement('div');
        div.className = 'resumen-item';
        div.innerHTML = `
            <div class="resumen-item-label">🧀 Muzza</div>
            <div class="resumen-item-value">
                ${resumen['muzza-c1'] > 0 ? `<span class="resumen-combo">C1: ${resumen['muzza-c1']}</span>` : ''}
                ${resumen['muzza-c2'] > 0 ? `<span class="resumen-combo">C2: ${resumen['muzza-c2']}</span>` : ''}
                ${resumen['muzza-c3'] > 0 ? `<span class="resumen-combo">C3: ${resumen['muzza-c3']}</span>` : ''}
            </div>
        `;
        resumenDiv.appendChild(div);
    }
    
    // Jamón
    if (resumen['jamon-c1'] > 0 || resumen['jamon-c2'] > 0 || resumen['jamon-c3'] > 0) {
        const div = document.createElement('div');
        div.className = 'resumen-item';
        div.innerHTML = `
            <div class="resumen-item-label">🍖 Jamón</div>
            <div class="resumen-item-value">
                ${resumen['jamon-c1'] > 0 ? `<span class="resumen-combo">C1: ${resumen['jamon-c1']}</span>` : ''}
                ${resumen['jamon-c2'] > 0 ? `<span class="resumen-combo">C2: ${resumen['jamon-c2']}</span>` : ''}
                ${resumen['jamon-c3'] > 0 ? `<span class="resumen-combo">C3: ${resumen['jamon-c3']}</span>` : ''}
            </div>
        `;
        resumenDiv.appendChild(div);
    }
    
    // Mostrar mensaje si no hay combos
    if (resumenDiv.innerHTML === '') {
        resumenDiv.innerHTML = '<div class="resumen-vacio">📦 Agrega combos a la venta</div>';
    }
    
    totalVentaSpan.textContent = `$${totalVenta.toFixed(2)}`;
}

function showMessage(text, type) {
    const mensaje = document.getElementById('mensaje');
    mensaje.textContent = text;
    mensaje.classList.remove('hidden', 'success', 'error');
    mensaje.classList.add(type === 'error' ? 'error' : 'success');
    
    // Auto-limpiar después de 5 segundos si es success
    if (type !== 'error') {
        setTimeout(() => {
            mensaje.classList.add('hidden');
            mensaje.textContent = '';
        }, 5000);
    }
}
}