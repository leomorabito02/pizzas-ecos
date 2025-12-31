/**
 * UI Utilities - Funciones compartidas para la UI
 */

class UIUtils {
    /**
     * Muestra/oculta spinner de carga
     */
    static showSpinner(show = true) {
        const overlay = document.getElementById('loadingOverlay');
        if (overlay) {
            if (show) {
                overlay.classList.remove('hidden');
            } else {
                overlay.classList.add('hidden');
            }
        }
    }

    /**
     * Muestra mensaje temporal (toast)
     */
    static showMessage(text, type = 'info', duration = 3000) {
        const container = document.getElementById('messageContainer') || this.createMessageContainer();
        
        const msg = document.createElement('div');
        msg.className = `message message-${type}`;
        msg.textContent = text;
        msg.style.cssText = `
            padding: 15px 20px;
            margin-bottom: 10px;
            border-radius: 5px;
            color: white;
            font-weight: 500;
            animation: slideIn 0.3s ease;
        `;

        if (type === 'success') {
            msg.style.backgroundColor = '#28a745';
        } else if (type === 'error') {
            msg.style.backgroundColor = '#dc3545';
        } else if (type === 'warning') {
            msg.style.backgroundColor = '#ffc107';
            msg.style.color = '#333';
        } else {
            msg.style.backgroundColor = '#17a2b8';
        }

        container.appendChild(msg);

        setTimeout(() => {
            msg.remove();
        }, duration);
    }

    /**
     * Crea contenedor de mensajes si no existe
     */
    static createMessageContainer() {
        const container = document.createElement('div');
        container.id = 'messageContainer';
        container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 10000;
            max-width: 400px;
        `;
        document.body.appendChild(container);
        return container;
    }

    /**
     * Formatea número a moneda
     */
    static formatCurrency(amount) {
        return new Intl.NumberFormat('es-AR', {
            style: 'currency',
            currency: 'ARS'
        }).format(amount);
    }

    /**
     * Formatea fecha
     */
    static formatDate(date) {
        return new Intl.DateTimeFormat('es-AR', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        }).format(new Date(date));
    }

    /**
     * Confirma acción con modal
     */
    static async confirmAction(message) {
        return new Promise((resolve) => {
            const confirmed = confirm(message);
            resolve(confirmed);
        });
    }

    /**
     * Desabilita/habilita elemento
     */
    static setDisabled(selector, disabled = true) {
        const el = document.querySelector(selector);
        if (el) {
            el.disabled = disabled;
            el.style.opacity = disabled ? '0.5' : '1';
            el.style.cursor = disabled ? 'not-allowed' : 'pointer';
        }
    }

    /**
     * Valida que un campo no esté vacío
     */
    static validateRequired(value, fieldName) {
        if (!value || value.toString().trim() === '') {
            throw new Error(`${fieldName} es requerido`);
        }
        return true;
    }

    /**
     * Valida que un número sea positivo
     */
    static validatePositive(value, fieldName) {
        const num = parseFloat(value);
        if (isNaN(num) || num <= 0) {
            throw new Error(`${fieldName} debe ser mayor a 0`);
        }
        return true;
    }

    /**
     * Parsea precio argentino (soporta comas y puntos)
     */
    static parsePrice(value) {
        if (typeof value === 'number') return value;
        if (!value) return 0;
        let str = String(value).trim()
            .replace('$', '')
            .replace(/\./g, '')
            .replace(',', '.');
        return parseFloat(str) || 0;
    }
}

// Estilos CSS inyectados
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(400px);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    .hidden {
        display: none !important;
    }

    #messageContainer {
        pointer-events: all;
    }
`;
document.head.appendChild(style);
