function setupImageUpload(inputId, previewId, buttonId) {
    const input = $(inputId);
    const preview = $(previewId);
    const button = $(buttonId);
    
    if (input && preview && button) {
        button.addEventListener('click', () => {
            input.click();
        });
        
        input.addEventListener('change', (event) => {
            if (event.target.files && event.target.files[0]) {
                const reader = new FileReader();
                
                reader.onload = (e) => {
                    preview.src = e.target.result;
                };
                
                reader.readAsDataURL(event.target.files[0]);
            }
        });
    }
}

function setupModal(modalId, openTriggerId, closeTriggerId) {
    const modal = $(modalId);
    const openTrigger = $(openTriggerId);
    const closeTrigger = $(closeTriggerId);
    
    if (!modal) return;
    
    if (openTrigger) {
        openTrigger.addEventListener('click', () => {
            modal.classList.add('active');
        });
    }
    
    if (closeTrigger) {
        closeTrigger.addEventListener('click', () => {
            modal.classList.remove('active');
        });
    }
    
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.classList.remove('active');
        }
    });
}