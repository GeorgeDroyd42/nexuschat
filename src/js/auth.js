let isSubmitting = false;

async function submitAuthForm(form, endpoint) {
    if (isSubmitting) return;
    isSubmitting = true;
    
    const messageContainer = document.getElementById('error-container');
    messageContainer.textContent = '';
    messageContainer.className = 'error-container'; 
    messageContainer.style.display = 'none';
    
    try {
        // Get form data including the file for registration
        const formData = new FormData();
        formData.append('username', form.username.value);
        formData.append('password', form.password.value);

        // Include profile picture in registration request
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
}

setupImageUpload('profile_picture', 'profile-preview', 'select-profile-btn');

document.addEventListener('DOMContentLoaded', () => {
    const sessionMessage = sessionStorage.getItem('sessionMessage');
    const banMessage = sessionStorage.getItem('banMessage');
    
    if (sessionMessage) {
        displayErrorMessage(sessionMessage);
        sessionStorage.removeItem('sessionMessage');
    } else if (banMessage) {
        displayErrorMessage(banMessage);
        sessionStorage.removeItem('banMessage');
    }

    if (document.getElementById('loginForm')) {
        const loginForm = document.getElementById('loginForm');
        
        loginForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const submitBtn = loginForm.querySelector('button[type="submit"]');
            toggleLoading(submitBtn.id || 'loginButton', true);
            try {
                await submitAuthForm(loginForm, '/api/auth/login');
            } finally {
                toggleLoading(submitBtn.id || 'loginButton', false);
            }
        });
    }
    
    if (document.getElementById('registerForm')) {
        const registerForm = document.getElementById('registerForm');
        
        registerForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const submitBtn = registerForm.querySelector('button[type="submit"]');
            toggleLoading(submitBtn.id || 'registerButton', true);
            await submitAuthForm(registerForm, '/api/auth/register');
            toggleLoading(submitBtn.id || 'registerButton', false);
        });

    }
});