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

