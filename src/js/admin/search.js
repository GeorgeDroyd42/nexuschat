const AdminSearch = {
    setupSearch() {
        const searchInput = document.getElementById('user-search');
        const clearBtn = document.getElementById('clear-search');
        
        if (!searchInput) return;
        
        searchInput.addEventListener('input', (e) => {
            const term = e.target.value.toLowerCase().trim();
            const rows = document.querySelectorAll('#user-list-body tr');
            
            rows.forEach(row => {
                const username = row.querySelector('.member-name')?.textContent?.toLowerCase() || '';
                row.style.display = !term || username.startsWith(term) ? '' : 'none';
            });
        });
        
        if (clearBtn) {
            clearBtn.addEventListener('click', () => {
                searchInput.value = '';
                document.querySelectorAll('#user-list-body tr').forEach(row => {
                    row.style.display = '';
                });
            });
        }
    }
};

window.AdminSearch = AdminSearch;