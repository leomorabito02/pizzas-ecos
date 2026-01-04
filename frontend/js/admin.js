// admin.js - Admin Panel Controller
// Maneja toda la lógica de administración de productos y vendedores

// Loading Spinner Functions
let loadingTimeout = null;  // Para timeout de pantalla de carga

function showLoadingSpinner(show = true) {
    const overlay = document.getElementById('loadingOverlay');
    if (overlay) {
        if (show) {
            overlay.classList.remove('hidden');
            
            // Timeout: ocultar automáticamente después de 10 segundos
            if (loadingTimeout) clearTimeout(loadingTimeout);
            loadingTimeout = setTimeout(() => {
                hideLoadingSpinner();
                console.log('Loading timeout - se ocultó después de 10 segundos');
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

// API Base URL - usando APIService centralizado (api ya está disponible globalmente)
const API_BASE = api.baseURL;

// Caches for data
let usuariosCache = [];

// Check authentication on load
window.addEventListener('load', () => {
    showLoadingSpinner(true);
    const token = sessionStorage.getItem('authToken');
    const userId = sessionStorage.getItem('userId');
    const username = sessionStorage.getItem('username');

    if (!token || !userId) {
        window.location.href = 'login.html';
        return;
    }

    // Mostrar nombre de usuario en el hamburguesa
    document.getElementById('currentUser').textContent = username || 'Usuario';

    // Setup form listeners after DOM is ready
    setupEditProductForm();
    setupProductForm();
    setupVendedorForm();
    setupEditVendedorForm();
    setupClearDatabaseBtn();

    loadDashboard();
    hideLoadingSpinner();

    // Header scroll animation for visual feedback
    const header = document.querySelector('.header');
    const mainContent = document.querySelector('.main-content');
    
    if (mainContent && header) {
        let scrollTimeout;
        mainContent.addEventListener('scroll', () => {
            clearTimeout(scrollTimeout);
            if (mainContent.scrollTop > 10) {
                if (!header.classList.contains('scrolled')) {
                    header.classList.add('scrolled');
                }
            } else {
                header.classList.remove('scrolled');
            }
        }, { passive: true });
    }
});

// Menu navigation
document.querySelectorAll('.menu-link').forEach(link => {
    link.addEventListener('click', (e) => {
        e.preventDefault();
        const section = link.getAttribute('data-section');
        showSection(section);

        // Update active menu item
        document.querySelectorAll('.menu-link').forEach(l => l.classList.remove('active'));
        link.classList.add('active');
    });
});

function showSection(section) {
    document.querySelectorAll('.content-section').forEach(s => s.classList.remove('active'));
    document.getElementById(section).classList.add('active');

    // Update title
    const titles = {
        dashboard: 'Dashboard',
        productos: 'Gestión de Productos',
        vendedores: 'Gestión de Vendedores',
        usuarios: 'Gestión de Usuarios'
    };
    document.getElementById('sectionTitle').textContent = titles[section] || 'Dashboard';

    // Load data for specific sections
    if (section === 'productos') loadProductos();
    if (section === 'vendedores') loadVendedores();
    if (section === 'usuarios') {
        loadUsuarios();
        setupUsuarioForm();
        setupEditUsuarioForm();
    }
}

// Mobile sidebar toggle
function toggleSidebar() {
    const sidebar = document.querySelector('.sidebar');
    const overlay = document.getElementById('sidebarOverlay');
    sidebar.classList.toggle('open');
    if (overlay) {
        overlay.classList.toggle('open');
    }
}

// Close sidebar when clicking on a menu link
document.querySelectorAll('.sidebar-menu a').forEach(link => {
    link.addEventListener('click', () => {
        const sidebar = document.querySelector('.sidebar');
        const overlay = document.getElementById('sidebarOverlay');
        sidebar.classList.remove('open');
        if (overlay) {
            overlay.classList.remove('open');
        }
    });
});

// Close sidebar when clicking on overlay
const overlay = document.getElementById('sidebarOverlay');
if (overlay) {
    overlay.addEventListener('click', () => {
        const sidebar = document.querySelector('.sidebar');
        sidebar.classList.remove('open');
        overlay.classList.remove('open');
    });
}

// Close sidebar when clicking outside
document.addEventListener('click', (e) => {
    const sidebar = document.querySelector('.sidebar');
    const hamburger = document.getElementById('hamburgerBtn');
    const overlay = document.getElementById('sidebarOverlay');
    if (sidebar && hamburger && !sidebar.contains(e.target) && !hamburger.contains(e.target)) {
        sidebar.classList.remove('open');
        if (overlay) {
            overlay.classList.remove('open');
        }
    }
});

// Logout function
function logout() {
    const modal = document.getElementById('logoutModal');
    if (modal) {
        modal.classList.remove('hidden');
    }
}

// Modal handlers
const logoutModal = document.getElementById('logoutModal');
const closeLogoutModal = document.getElementById('closeLogoutModal');
const cancelLogout = document.getElementById('cancelLogout');
const confirmLogout = document.getElementById('confirmLogout');

function closeModal() {
    if (logoutModal) {
        logoutModal.classList.add('hidden');
    }
}

if (closeLogoutModal) {
    closeLogoutModal.addEventListener('click', closeModal);
}

if (cancelLogout) {
    cancelLogout.addEventListener('click', closeModal);
}

if (confirmLogout) {
    confirmLogout.addEventListener('click', () => {
        sessionStorage.removeItem('authToken');
        sessionStorage.removeItem('userId');
        window.location.href = 'login.html';
    });
}

// Close modal when clicking outside
if (logoutModal) {
    logoutModal.addEventListener('click', (e) => {
        if (e.target === logoutModal) {
            closeModal();
        }
    });
}

// Keep old logout button handler for backward compatibility
const logoutBtn = document.getElementById('logoutBtn');
if (logoutBtn) {
    logoutBtn.addEventListener('click', logout);
}

// Delete Product Modal Handlers
const deleteProductModal = document.getElementById('deleteProductModal');
const closeDeleteProductModal = document.getElementById('closeDeleteProductModal');
const cancelDeleteProduct = document.getElementById('cancelDeleteProduct');
const confirmDeleteProduct = document.getElementById('confirmDeleteProduct');

function closeDeleteProductModalDialog() {
    if (deleteProductModal) {
        deleteProductModal.classList.add('hidden');
    }
    productIdToDelete = null;
}

if (closeDeleteProductModal) {
    closeDeleteProductModal.addEventListener('click', closeDeleteProductModalDialog);
}

if (cancelDeleteProduct) {
    cancelDeleteProduct.addEventListener('click', closeDeleteProductModalDialog);
}

if (confirmDeleteProduct) {
    confirmDeleteProduct.addEventListener('click', async () => {
        if (!productIdToDelete) return;
        
        try {
            await api.eliminarProducto(productIdToDelete);

            showSuccess('Producto eliminado exitosamente');
            closeDeleteProductModalDialog();
            await loadProductos();

        } catch (error) {
            console.error('Error:', error);
            showError('Error eliminando producto');
        }
    });
}

if (deleteProductModal) {
    deleteProductModal.addEventListener('click', (e) => {
        if (e.target === deleteProductModal) {
            closeDeleteProductModalDialog();
        }
    });
}

// Delete Vendedor Modal Handlers
const deleteVendedorModal = document.getElementById('deleteVendedorModal');
const closeDeleteVendedorModal = document.getElementById('closeDeleteVendedorModal');
const cancelDeleteVendedor = document.getElementById('cancelDeleteVendedor');
const confirmDeleteVendedor = document.getElementById('confirmDeleteVendedor');

function closeDeleteVendedorModalDialog() {
    if (deleteVendedorModal) {
        deleteVendedorModal.classList.add('hidden');
    }
    vendedorIdToDelete = null;
}

if (closeDeleteVendedorModal) {
    closeDeleteVendedorModal.addEventListener('click', closeDeleteVendedorModalDialog);
}

if (cancelDeleteVendedor) {
    cancelDeleteVendedor.addEventListener('click', closeDeleteVendedorModalDialog);
}

if (confirmDeleteVendedor) {
    confirmDeleteVendedor.addEventListener('click', async () => {
        if (!vendedorIdToDelete) return;
        
        try {
            await api.eliminarVendedor(vendedorIdToDelete);

            showSuccess('Vendedor eliminado exitosamente');
            closeDeleteVendedorModalDialog();
            await loadVendedores();

        } catch (error) {
            console.error('Error:', error);
            showError('Error eliminando vendedor');
        }
    });
}

if (deleteVendedorModal) {
    deleteVendedorModal.addEventListener('click', (e) => {
        if (e.target === deleteVendedorModal) {
            closeDeleteVendedorModalDialog();
        }
    });
}

// Load dashboard data
async function loadDashboard() {
    try {
        showLoadingSpinner(true);
        
        // Obtener datos de estadísticas (mismo endpoint que estadísticas.js)
        const statsResponse = await api.request('/estadisticas-sheet');
        console.log('Dashboard - statsResponse:', statsResponse);
        
        // Extraer resumen (backend retorna {status, data: {resumen, vendedores, ventas}})
        const datosStats = (statsResponse && statsResponse.data) ? statsResponse.data : statsResponse || {};
        const resumen = datosStats.resumen || {};
        
        console.log('Dashboard - resumen:', resumen);
        
        // Mostrar total de ventas (total_ventas) y monto cobrado (total_cobrado) - mismo que estadísticas
        document.getElementById('totalVentas').textContent = resumen.ventas_totales || 0;
        document.getElementById('totalMonto').textContent = `$${(resumen.total_cobrado || 0).toFixed(2)}`;

        // Load additional data
        const dataInfo = await api.getData();
        console.log('Dashboard getData() response:', dataInfo);
        // Handle response format: could be {status, data: {...}} or direct object
        const vendedoresData = dataInfo?.data || dataInfo || {};
        const vendedores = vendedoresData.vendedores || [];
        console.log('Dashboard vendedores count:', vendedores.length);
        document.getElementById('totalVendedores').textContent = vendedores.length || 0;

        // Mostrar ventas recientes del resumen
        const ventasRecientes = (datosStats.ventas || []).slice(0, 10);
        displayRecentSales(ventasRecientes);
        
        hideLoadingSpinner();

    } catch (error) {
        console.error('Error:', error);
        showError('Error cargando el dashboard');
        hideLoadingSpinner();
    }
}

function displayRecentSales(sales) {
    const container = document.getElementById('recentSalesContainer');

    if (!sales || sales.length === 0) {
        container.innerHTML = '<p class="no-data">No hay ventas registradas</p>';
        return;
    }

    let html = `
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Vendedor</th>
                    <th>Cliente</th>
                    <th>Total</th>
                    <th>Estado</th>
                    <th>Fecha</th>
                </tr>
            </thead>
            <tbody>
    `;

    sales.forEach(venta => {
        const rowClass = venta.estado === 'cancelada' ? 'class="cancelada"' : '';
        const totalMonto = typeof venta.total === 'string' ? parseFloat(venta.total) : venta.total;
        html += `
            <tr ${rowClass}>
                <td data-label="ID">#${venta.id}</td>
                <td data-label="Vendedor">${venta.vendedor}</td>
                <td data-label="Cliente">${venta.cliente}</td>
                <td data-label="Total">$${totalMonto.toFixed(2)}</td>
                <td data-label="Estado"><strong>${venta.estado}</strong></td>
                <td data-label="Fecha">${new Date(venta.created_at).toLocaleDateString()}</td>
            </tr>
        `;
    });

    html += '</tbody></table>';
    container.innerHTML = html;
}

// Productos functionality - setup form listener when DOM is ready
function setupProductForm() {
    const form = document.getElementById('productForm');
    if (!form) return;
    
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const nombre = document.getElementById('newProductName').value.trim();
        const descripcion = document.getElementById('newProductDesc').value.trim();
        const precio = parseFloat(document.getElementById('newProductPrice').value);

        if (!nombre || precio <= 0) {
            showError('El nombre y precio del producto son requeridos');
            return;
        }

        try {
            await api.crearProducto({
                tipo_pizza: nombre,
                descripcion: descripcion,
                precio: precio
            });

            showSuccess('Producto creado exitosamente');
            document.getElementById('productForm').reset();
            await loadProductos();

        } catch (error) {
            console.error('Error:', error);
            showError('Error creando producto: ' + error.message);
        }
    });
}

async function loadProductos() {
    try {
        showLoadingSpinner(true);
        const response = await api.obtenerProductos();
        // Handle both response formats: {status, data, message} or direct array
        const productos = Array.isArray(response) ? response : (response?.data || []);
        const container = document.getElementById('productosTableContainer');

        if (!productos || productos.length === 0) {
            container.innerHTML = '<p class="no-data">No hay productos registrados</p>';
            return;
        }

        let html = `
            <table>
                <thead>
                    <tr>
                        <th>Tipo de Pizza</th>
                        <th>Descripción</th>
                        <th>Precio</th>
                        <th>Acciones</th>
                    </tr>
                </thead>
                <tbody>
        `;

        productos.forEach(producto => {
            html += `
                <tr>
                    <td data-label="Tipo de Pizza"><strong>${producto.tipo_pizza}</strong></td>
                    <td data-label="Descripción">${producto.descripcion || '-'}</td>
                    <td data-label="Precio">$${(producto.precio || 0).toFixed(2)}</td>
                    <td data-label="Acciones">
                        <button class="btn-small btn-edit" data-action="edit" data-id="${producto.id}" data-name="${producto.tipo_pizza.replace(/"/g, '&quot;')}" data-desc="${(producto.descripcion || '').replace(/"/g, '&quot;')}" data-price="${producto.precio}">Editar</button>
                        <button class="btn-small btn-delete" data-action="delete" data-id="${producto.id}">Eliminar</button>
                    </td>
                </tr>
            `;
        });

        html += '</tbody></table>';
        container.innerHTML = html;
        hideLoadingSpinner();

    } catch (error) {
        console.error('Error:', error);
        document.getElementById('productosTableContainer').innerHTML = '<p class="no-data">Error cargando productos</p>';
        hideLoadingSpinner();
    }
}

function abrirModalProducto(productoId, nombre, descripcion, precio) {
    document.getElementById('editProductId').value = productoId;
    document.getElementById('editProductName').value = nombre;
    document.getElementById('editProductDesc').value = descripcion || '';
    document.getElementById('editProductPrice').value = precio;
    document.getElementById('modalEditarProducto').classList.remove('hidden');
}

function cerrarModalProducto() {
    document.getElementById('modalEditarProducto').classList.add('hidden');
    document.getElementById('editProductForm').reset();
}

// Setup edit product form listener when DOM is ready
function setupEditProductForm() {
    const form = document.getElementById('editProductForm');
    if (!form) {
        console.error('editProductForm not found');
        return;
    }
    
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const productoId = document.getElementById('editProductId').value;
        const nombre = document.getElementById('editProductName').value.trim();
        const descripcion = document.getElementById('editProductDesc').value.trim();
        const precio = parseFloat(document.getElementById('editProductPrice').value);

        if (!nombre || precio <= 0) {
            showError('El nombre y precio son requeridos');
            return;
        }

        try {
            await api.actualizarProducto(productoId, { 
                tipo_pizza: nombre,
                descripcion: descripcion,
                precio: precio, 
                activo: true 
            });

            showSuccess('Producto actualizado exitosamente');
            cerrarModalProducto();
            await loadProductos();

        } catch (error) {
            console.error('Error:', error);
            showError('Error actualizando producto: ' + error.message);
        }
    });
}

let productIdToDelete = null;

async function eliminarProducto(productoId) {
    productIdToDelete = productoId;
    const modal = document.getElementById('deleteProductModal');
    if (modal) {
        modal.classList.remove('hidden');
    }
}

// Vendedores functionality - setup form listener when DOM is ready
function setupVendedorForm() {
    const form = document.getElementById('vendedorForm');
    if (!form) return;
    
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const nombre = document.getElementById('newVendedorName').value.trim();

        if (!nombre) {
            showError('El nombre del vendedor es requerido');
            return;
        }

        try {
            await api.crearVendedor({ nombre: nombre });

            showSuccess('Vendedor creado exitosamente');
            document.getElementById('vendedorForm').reset();
            await loadVendedores();

        } catch (error) {
            console.error('Error:', error);
            showError('Error creando vendedor: ' + error.message);
        }
    });
}

async function loadVendedores() {
    try {
        showLoadingSpinner(true);
        const data = await api.getData();
        console.log('getData() response:', data);
        // Handle response format: could be {status, data: {...}} or direct object
        const vendedoresData = data?.data || data || {};
        const vendedores = vendedoresData.vendedores || [];
        console.log('Vendedores extracted:', vendedores);
        const container = document.getElementById('vendedoresTableContainer');

        if (!vendedores || vendedores.length === 0) {
            container.innerHTML = '<p class="no-data">No hay vendedores registrados</p>';
            hideLoadingSpinner();
            return;
        }

        let html = `
            <table>
                <thead>
                    <tr>
                        <th>Nombre</th>
                        <th>Acciones</th>
                    </tr>
                </thead>
                <tbody>
        `;

        vendedores.forEach(vendedor => {
            const nombre = vendedor.nombre || 'Sin nombre';
            html += `
                <tr>
                    <td data-label="Nombre"><strong>${nombre}</strong></td>
                    <td data-label="Acciones">
                        <button class="btn-small btn-edit" data-action="edit" data-id="${vendedor.id}" data-name="${nombre.replace(/"/g, '&quot;')}">Editar</button>
                        <button class="btn-small btn-delete" data-action="delete" data-id="${vendedor.id}">Eliminar</button>
                    </td>
                </tr>
            `;
        });

        html += '</tbody></table>';
        container.innerHTML = html;
        hideLoadingSpinner();

    } catch (error) {
        console.error('Error:', error);
        document.getElementById('vendedoresTableContainer').innerHTML = '<p class="no-data">Error cargando vendedores</p>';
        hideLoadingSpinner();
    }
}

function abrirModalVendedor(vendedorId, nombre) {
    document.getElementById('editVendedorId').value = vendedorId;
    document.getElementById('editVendedorName').value = nombre;
    document.getElementById('modalEditarVendedor').classList.remove('hidden');
}

function cerrarModalVendedor() {
    document.getElementById('modalEditarVendedor').classList.add('hidden');
    document.getElementById('editVendedorForm').reset();
}

// Setup edit vendedor form listener when DOM is ready
function setupEditVendedorForm() {
    const form = document.getElementById('editVendedorForm');
    if (!form) {
        console.error('editVendedorForm not found');
        return;
    }
    
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const vendedorId = document.getElementById('editVendedorId').value;
        const nombre = document.getElementById('editVendedorName').value.trim();

        if (!nombre) {
            showError('El nombre del vendedor es requerido');
            return;
        }

        try {
            await api.actualizarVendedor(vendedorId, { nombre: nombre });

            showSuccess('Vendedor actualizado exitosamente');
            cerrarModalVendedor();
            await loadVendedores();

        } catch (error) {
            console.error('Error:', error);
            showError('Error actualizando vendedor: ' + error.message);
        }
    });
}

let vendedorIdToDelete = null;

async function eliminarVendedor(vendedorId) {
    vendedorIdToDelete = vendedorId;
    const modal = document.getElementById('deleteVendedorModal');
    if (modal) {
        modal.classList.remove('hidden');
    }
}

function showSuccess(message) {
    const msg = document.getElementById('successMessage');
    msg.textContent = message;
    msg.classList.add('show');
    setTimeout(() => msg.classList.remove('show'), 3000);
}

function showError(message) {
    const msg = document.getElementById('errorMessage');
    msg.textContent = message;
    msg.classList.add('show');
    setTimeout(() => msg.classList.remove('show'), 3000);
}

// Setup event listeners (antes usado con onclick=)
document.getElementById('hamburgerBtn')?.addEventListener('click', toggleSidebar);
document.getElementById('logoutBtnHeader')?.addEventListener('click', logout);
document.getElementById('closeModalProducto')?.addEventListener('click', cerrarModalProducto);
document.getElementById('cancelModalProducto')?.addEventListener('click', cerrarModalProducto);
document.getElementById('closeModalVendedor')?.addEventListener('click', cerrarModalVendedor);
document.getElementById('cancelModalVendedor')?.addEventListener('click', cerrarModalVendedor);
document.getElementById('closeEditUsuarioModal')?.addEventListener('click', cerrarModalUsuario);
document.getElementById('cancelEditUsuario')?.addEventListener('click', cerrarModalUsuario);

// Event delegation para botones dinámicos (Editar/Eliminar)
document.addEventListener('click', (e) => {
    const btn = e.target.closest('[data-action]');
    if (!btn) return;

    const action = btn.getAttribute('data-action');
    const id = btn.getAttribute('data-id');

    if (action === 'edit') {
        const name = btn.getAttribute('data-name');
        const desc = btn.getAttribute('data-desc');
        const price = btn.getAttribute('data-price');

        if (btn.closest('#productosTableContainer')) {
            abrirModalProducto(id, name, desc, price);
        } else if (btn.closest('#vendedoresTableContainer')) {
            abrirModalVendedor(id, name);
        } else if (btn.closest('#usuariosTableContainer')) {
            abrirModalUsuario(id);
        }
    } else if (action === 'delete') {
        if (btn.closest('#productosTableContainer')) {
            eliminarProducto(id);
        } else if (btn.closest('#vendedoresTableContainer')) {
            eliminarVendedor(id);
        } else if (btn.closest('#usuariosTableContainer')) {
            eliminarUsuario(id);
        }
    }
});

// Usuarios Management
function setupUsuarioForm() {
    const form = document.getElementById('usuarioForm');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const username = document.getElementById('newUsuarioUsername').value.trim();
        const password = document.getElementById('newUsuarioPassword').value.trim();
        const errorDiv = document.getElementById('usuarioFormError');

        // Limpiar errores previos
        if (errorDiv) {
            errorDiv.classList.remove('show');
            errorDiv.textContent = '';
        }

        // Validaciones
        if (!username) {
            const msg = 'El usuario es requerido';
            if (errorDiv) {
                errorDiv.textContent = msg;
                errorDiv.classList.add('show');
            }
            showError(msg);
            return;
        }

        if (username.length < 3) {
            const msg = 'El usuario debe tener al menos 3 caracteres';
            if (errorDiv) {
                errorDiv.textContent = msg;
                errorDiv.classList.add('show');
            }
            showError(msg);
            return;
        }

        if (!password) {
            const msg = 'La contraseña es requerida';
            if (errorDiv) {
                errorDiv.textContent = msg;
                errorDiv.classList.add('show');
            }
            showError(msg);
            return;
        }

        if (password.length < 4) {
            const msg = 'La contraseña debe tener al menos 4 caracteres';
            if (errorDiv) {
                errorDiv.textContent = msg;
                errorDiv.classList.add('show');
            }
            showError(msg);
            return;
        }

        try {
            showLoadingSpinner(true);
            const response = await fetch(`${API_BASE}/crear-usuario`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`
                },
                body: JSON.stringify({ username, password })
            });

            const responseData = await response.json();

            if (response.ok) {
                form.reset();
                if (errorDiv) {
                    errorDiv.classList.remove('show');
                    errorDiv.textContent = '';
                }
                await loadUsuarios();
                showSuccess('Admin creado exitosamente');
            } else {
                const errorMsg = responseData.message || 'Error al crear admin';
                if (errorDiv) {
                    errorDiv.textContent = errorMsg;
                    errorDiv.classList.add('show');
                }
                showError(errorMsg);
            }
        } catch (error) {
            console.error('Error:', error);
            const errorMsg = 'Error de conexión al crear admin';
            if (errorDiv) {
                errorDiv.textContent = errorMsg;
                errorDiv.classList.add('show');
            }
            showError(errorMsg);
        } finally {
            hideLoadingSpinner();
        }
    });
}

async function loadUsuarios() {
    try {
        showLoadingSpinner(true);
        const response = await fetch(`${API_BASE}/usuarios`, {
            headers: {
                'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`
            }
        });

        const responseData = await response.json();
        const usuarios = responseData.data || responseData;

        if (!Array.isArray(usuarios)) {
            console.error('Expected array of usuarios:', usuarios);
            return;
        }

        usuariosCache = usuarios;

        const container = document.getElementById('usuariosTableContainer');
        if (!container) return;

        if (usuarios.length === 0) {
            container.innerHTML = '<p class="no-data">No hay usuarios registrados</p>';
            hideLoadingSpinner();
            return;
        }

        container.innerHTML = `
            <table class="admin-table">
                <thead>
                    <tr>
                        <th>Usuario</th>
                        <th>Acciones</th>
                    </tr>
                </thead>
                <tbody>
                    ${usuarios.map(u => `
                        <tr>
                            <td data-label="Usuario"><strong>${u.username}</strong></td>
                            <td data-label="Acciones">
                                <button class="btn-small btn-edit" data-action="edit" data-id="${u.id}" data-username="${u.username.replace(/"/g, '&quot;')}">Editar</button>
                                <button class="btn-small btn-delete" data-action="delete" data-id="${u.id}">Eliminar</button>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;

        hideLoadingSpinner();
    } catch (error) {
        console.error('Error loading usuarios:', error);
        showError('Error al cargar usuarios');
        hideLoadingSpinner();
    }
}

function abrirModalUsuario(id) {
    const usuarioAEditar = usuariosCache.find(u => u.id == id);
    if (!usuarioAEditar) return;

    document.getElementById('editUsuarioId').value = usuarioAEditar.id;
    document.getElementById('editUsuarioUsername').value = usuarioAEditar.username;
    document.getElementById('editUsuarioPassword').value = '';
    // Store the current rol for the update
    document.getElementById('editUsuarioId').dataset.rol = usuarioAEditar.rol;

    document.getElementById('editUsuarioModal').classList.remove('hidden');
}

function cerrarModalUsuario() {
    document.getElementById('editUsuarioModal').classList.add('hidden');
    document.getElementById('editUsuarioForm').reset();
}

function setupEditUsuarioForm() {
    const form = document.getElementById('editUsuarioForm');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const id = document.getElementById('editUsuarioId').value;
        const username = document.getElementById('editUsuarioUsername').value.trim();
        const password = document.getElementById('editUsuarioPassword').value.trim();
        const errorDiv = document.getElementById('editUsuarioError');

        // Limpiar errores previos
        errorDiv.classList.remove('show');
        errorDiv.textContent = '';

        // Validaciones
        if (!username) {
            errorDiv.textContent = 'El usuario es requerido';
            errorDiv.classList.add('show');
            return;
        }

        if (username.length < 3) {
            errorDiv.textContent = 'El usuario debe tener al menos 3 caracteres';
            errorDiv.classList.add('show');
            return;
        }

        if (password && password.length < 4) {
            errorDiv.textContent = 'Si cambias la contraseña, debe tener al menos 4 caracteres';
            errorDiv.classList.add('show');
            return;
        }

        const body = { 
            username,
            rol: document.getElementById('editUsuarioId').dataset.rol
        };
        if (password) {
            body.password = password;
        }

        try {
            showLoadingSpinner(true);
            const response = await fetch(`${API_BASE}/actualizar-usuario/${id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`
                },
                body: JSON.stringify(body)
            });

            const responseData = await response.json();

            if (response.ok) {
                cerrarModalUsuario();
                await loadUsuarios();
                showSuccess('Admin actualizado exitosamente');
            } else {
                errorDiv.textContent = responseData.message || 'Error al actualizar admin';
                errorDiv.classList.add('show');
                showError(responseData.message || 'Error al actualizar admin');
            }
        } catch (error) {
            console.error('Error:', error);
            showError('Error de conexión al actualizar admin');
        } finally {
            hideLoadingSpinner();
        }
    });
}

async function eliminarUsuario(id) {
    if (!confirm('¿Estás seguro de que deseas eliminar este usuario?')) {
        return;
    }

    try {
        showLoadingSpinner(true);
        const response = await fetch(`${API_BASE}/eliminar-usuario/${id}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`
            }
        });

        const responseData = await response.json();

        if (response.ok) {
            await loadUsuarios();
            showSuccess('Usuario eliminado exitosamente');
        } else {
            showError(responseData.message || 'Error al eliminar usuario');
        }
    } catch (error) {
        console.error('Error:', error);
        showError('Error de conexión al eliminar usuario');
    } finally {
        hideLoadingSpinner();
    }
}

// Clear Database Function
function setupClearDatabaseBtn() {
    const btn = document.getElementById('clearDatabaseBtn');
    if (!btn) return;

    btn.addEventListener('click', async () => {
        const confirmation1 = confirm('⚠️ ADVERTENCIA: Esto eliminará TODOS los datos de la aplicación (ventas, detalles, clientes, vendedores y productos) pero mantendrá los usuarios.\n\n¿Estás seguro?');
        if (!confirmation1) return;

        const confirmation2 = confirm('Esta es la última advertencia. Esta acción NO SE PUEDE DESHACER.\n\n¿Realmente quieres continuar?');
        if (!confirmation2) return;

        try {
            showLoadingSpinner(true);
            const response = await fetch(`${API_BASE}/limpiar-base-datos`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${sessionStorage.getItem('authToken')}`
                }
            });

            const responseData = await response.json();

            if (response.ok) {
                showSuccess('✅ Base de datos limpiada exitosamente. Recargando...');
                setTimeout(() => {
                    window.location.reload();
                }, 2000);
            } else {
                showError(responseData.message || 'Error al limpiar la base de datos');
            }
        } catch (error) {
            console.error('Error:', error);
            showError('Error de conexión al limpiar la base de datos');
        } finally {
            hideLoadingSpinner();
        }
    });
}
