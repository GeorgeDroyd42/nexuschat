let isSubmitting = false;

const AuthForms = {
    async submitAuthForm(form, endpoint) {
        if (isSubmitting) return;
        isSubmitting = true;
        
        const messageContainer = document.getElementById('error-container');
        messageContainer.textContent = '';
        messageContainer.className = 'error-container'; 
        messageContainer.style.display = 'none';
        
        try {
            const formData = new FormData();
            formData.append('username', form.username.value);
            formData.append('password', form.password.value);

            const profilePicture = form.profile_picture ? form.profile_picture.files[0] : null;
            if (profilePicture) {
                formData.append('profile_picture', profilePicture);
            }

            const result = await (endpoint.includes('login') ? 
                AuthAPI.login(formData) : 
                AuthAPI.register(formData));
            
            if (result.error) {
                isSubmitting = false;
                displayErrorMessage(result.error);
                return { error: result.error };
            }

            isSubmitting = false;
            NavigationUtils.redirectToMain();
            return result;
            
        } catch (error) {
            isSubmitting = false;
            displayErrorMessage('An error occurred');
            return { error };
        }
    },

    setupAuthForm(formId, endpoint, defaultButtonId) {
        const form = document.getElementById(formId);
        if (form) {
            form.addEventListener('submit', async (event) => {
                event.preventDefault();
                const submitBtn = form.querySelector('button[type="submit"]');
                toggleLoading(submitBtn.id || defaultButtonId, true);
                try {
                    await this.submitAuthForm(form, endpoint);
                } finally {
                    toggleLoading(submitBtn.id || defaultButtonId, false);
                }
            });
        }
    }
};


window.AuthForms = AuthForms;

document.addEventListener('DOMContentLoaded', () => {
    AuthForms.setupAuthForm('loginForm', '/api/auth/login', 'loginButton');
    AuthForms.setupAuthForm('registerForm', '/api/auth/register', 'registerButton');
});