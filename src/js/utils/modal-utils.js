class ModalManager {
    constructor() {
        this.openModals = new Set();
        this.setupGlobalHandlers();
    }

    setupGlobalHandlers() {
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeTopModal();
            }
        });

        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('modal-overlay') || 
                e.target.classList.contains('profile-menu-overlay')) {
                this.closeModal(e.target.id);
            }
        });
    }
openModal(modalId) {
    const modal = $(modalId);
    if (modal) {
        modal.style.display = 'flex';
        this.openModals.add(modalId);
        
        requestAnimationFrame(() => {
            modal.classList.add('active');
        });
    }
}

closeModal(modalId) {
    const modal = $(modalId);
    if (modal) {
        modal.classList.remove('active');
        
        setTimeout(() => {
            modal.style.display = 'none';
        }, 300);
        
        this.openModals.delete(modalId);
    }
}

    closeTopModal() {
        if (this.openModals.size > 0) {
            const modalIds = Array.from(this.openModals);
            const topModal = modalIds[modalIds.length - 1];
            this.closeModal(topModal);
        }
    }

    closeAllModals() {
        this.openModals.forEach(modalId => this.closeModal(modalId));
    }

    setupModal(modalId, openTriggerId, closeTriggerId) {
        const modal = $(modalId);
        const openTrigger = $(openTriggerId);
        const closeTrigger = $(closeTriggerId);
                
        if (!modal) return;
        
        if (openTrigger) {
            openTrigger.addEventListener('click', () => {
                this.openModal(modalId);
            });
        }
        
        if (closeTrigger) {
            closeTrigger.addEventListener('click', () => {
                this.closeModal(modalId);
            });
        }
        
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeModal(modalId);
            }
        });
    }
}

window.modalManager = new ModalManager();

