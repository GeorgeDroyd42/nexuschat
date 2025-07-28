class ChannelMenuAPI {
    constructor() {
        this.tabs = new Map();
        this.activeTab = 'overview';
        this.currentChannelId = null;
        this.currentChannelName = null;
    }

    reset() {
        this.tabs.clear();
        this.activeTab = 'overview';
        return this;
    }

    addTab(id, label) {
        this.tabs.set(id, { label, items: [] });
        if (this.tabs.size === 1) this.activeTab = id;
        return this;
    }

    addToTab(tabId, item) {
        if (this.tabs.has(tabId)) {
            this.tabs.get(tabId).items.push(item);
        }
        return this;
    }

    renderButtons(modalId, channelId, channelName) {
        this.currentChannelId = channelId;
        this.currentChannelName = channelName;
        
        const modal = document.getElementById(modalId);
        if (!modal) return;
        
        const content = modal.querySelector('.profile-menu-content');
        if (!content) return;
        
        this.createSidebarLayout(content);
        this.setupTabSwitching(content);
        this.populateTabContent();
    }

    createSidebarLayout(content) {
        content.innerHTML = '';
        
        const layout = this.createElement('div', 'settings-layout');
        const sidebar = this.createElement('div', 'settings-sidebar');
        const contentArea = this.createElement('div', 'settings-content');
        
        this.tabs.forEach((tab, id) => {
            const tabBtn = this.createElement('button', `settings-tab ${id === this.activeTab ? 'active' : ''}`, tab.label);
            tabBtn.dataset.tab = id;
            sidebar.appendChild(tabBtn);
        });
        
        this.tabs.forEach((tab, id) => {
            const panel = this.createElement('div', `settings-panel ${id === this.activeTab ? 'active' : ''}`);
            panel.id = `${id}-panel`;
            contentArea.appendChild(panel);
        });
        
        layout.appendChild(sidebar);
        layout.appendChild(contentArea);
        content.appendChild(layout);
    }

    setupTabSwitching(content) {
        const tabs = content.querySelectorAll('.settings-tab');
        tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const tabId = tab.dataset.tab;
                this.switchTab(tabId, content);
            });
        });
    }

    switchTab(tabId, content) {
        this.activeTab = tabId;
        
        content.querySelectorAll('.settings-tab').forEach(t => t.classList.remove('active'));
        content.querySelectorAll('.settings-panel').forEach(p => p.classList.remove('active'));
        
        content.querySelector(`[data-tab="${tabId}"]`).classList.add('active');
        content.querySelector(`#${tabId}-panel`).classList.add('active');
    }

populateTabContent() {
    this.tabs.forEach((tab, tabId) => {
        const panel = document.getElementById(`${tabId}-panel`);
        if (panel) {
            panel.innerHTML = '';
            tab.items.forEach(item => this.renderItem(panel, item));
        }
    });
    
    // Set current bio as default value for profile menu
    if (this.currentUser) {
        const bioEdit = document.getElementById('bio-edit');
        if (bioEdit && this.currentUser.bio) {
            bioEdit.value = this.currentUser.bio;
        }
    }
}
    renderItem(panel, item) {
        switch (item.type) {
            case 'channel-info':
                this.renderChannelInfo(panel);
                break;
            case 'button':
                this.renderButton(panel, item);
                break;
            case 'textarea':
                this.renderTextarea(panel, item);
                break;
            case 'separator':
                this.renderSeparator(panel);
                break;
            case 'webhook-list':
                this.renderWebhookList(panel);
                break;
        }
    }

renderInfoSection(panel, fields, sectionClass = 'info-section') {
    const infoSection = this.createElement('div', sectionClass);
    
    fields.forEach(field => {
        const group = this.createElement('div', 'form-group');
        const label = this.createElement('label', '', field.label);
        const display = this.createElement('div', `info-display ${field.class || ''}`, field.value);
        display.id = field.id;
        
        group.appendChild(label);
        group.appendChild(display);
        infoSection.appendChild(group);
    });
    
    panel.appendChild(infoSection);
}

renderChannelInfo(panel) {
    const fields = [
        { label: 'CHANNEL NAME', id: 'channel-info-name', value: `#${this.currentChannelName}` },
        { label: 'DESCRIPTION', id: 'channel-info-description', value: 'Loading...' },
        { label: 'CHANNEL ID', id: 'channel-info-id', value: this.currentChannelId, class: 'channel-id' }
    ];
    
    this.renderInfoSection(panel, fields, 'channel-info-section');
}

    renderButton(panel, item) {
        const btn = this.createElement('button', 'btn btn-secondary channel-menu-btn', item.text);
        if (item.color) btn.style.backgroundColor = item.color;
        btn.addEventListener('click', item.action);
        panel.appendChild(btn);
    }

renderTextarea(panel, item) {
    const group = this.createElement('div', 'form-group');
    const label = this.createElement('label', '', item.label);
    const textarea = this.createElement('textarea', '', '');
    
    Object.assign(textarea, {
        id: item.id,
        placeholder: item.placeholder,
        rows: item.rows,
        value: item.value
    });
    
    group.appendChild(label);
    group.appendChild(textarea);
    panel.appendChild(group);
}

    renderSeparator(panel) {
        const sep = this.createElement('div', 'separator');
        sep.style.cssText = 'height: 1px; background: var(--border-light); margin: 16px 0;';
        panel.appendChild(sep);
    }

renderWebhookList(panel) {
        const container = this.createElement('div', 'webhook-container');
        const list = this.createElement('div', 'webhook-list');
        const createSection = this.createElement('div', 'webhook-create-section');
        
        const profileGroup = this.createElement('div', 'form-group profile-picture-group');
        const profileContainer = this.createElement('div', 'profile-preview-container');
        const profilePreview = this.createElement('img');
        profilePreview.id = 'webhook-profile-preview';
        profilePreview.src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23cccccc'%3E%3Cpath d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z'/%3E%3C/svg%3E";
        profilePreview.alt = 'Webhook Profile Preview';
        profilePreview.className = 'image-preview';
        
        const profileInput = this.createElement('input');
        profileInput.type = 'file';
        profileInput.id = 'webhook-profile-picture';
        profileInput.name = 'profile_picture';
        profileInput.accept = 'image/*';
        profileInput.className = 'profile-input';
        
        const profileBtn = this.createElement('button', 'btn', 'Select Image');
        profileBtn.type = 'button';
        profileBtn.id = 'select-webhook-profile-btn';
        
        profileContainer.appendChild(profilePreview);
        profileContainer.appendChild(profileInput);
        profileContainer.appendChild(profileBtn);
        profileGroup.appendChild(profileContainer);
        
        const input = this.createElement('input');
        input.type = 'text';
        input.id = 'webhook-name';
        input.placeholder = 'Webhook name';
        input.className = 'form-group';
        
        const createBtn = this.createElement('button', 'btn btn-primary', 'Create Webhook');
        createBtn.onclick = () => this.createWebhook();
        
        createSection.appendChild(profileGroup);
        createSection.appendChild(input);
        createSection.appendChild(createBtn);
        container.appendChild(createSection);
        container.appendChild(list);
        panel.appendChild(container);
        
        setupImageUpload('webhook-profile-picture', 'webhook-profile-preview', 'select-webhook-profile-btn');
        
        this.loadWebhooks(list);
    }

async loadWebhooks(container) {
    try {
        const data = await BaseAPI.get(`/api/webhook/list/${this.currentChannelId}`);
        this.displayWebhooks(container, data.webhooks || []);
    } catch (error) {
        console.error('Error loading webhooks:', error);
    }
}

displayWebhooks(container, webhooks) {
    container.innerHTML = '';
    
    if (webhooks.length === 0) {
        const empty = this.createElement('div', 'empty-state', 'No webhooks created yet');
        container.appendChild(empty);
        return;
    }
    
    webhooks.forEach(webhook => {
        const item = this.createElement('div', 'webhook-item');
        const header = this.createElement('div', 'webhook-header');
        
        const webhookInfo = this.createElement('div', 'webhook-info-section');
        const avatarContainer = this.createElement('div', 'webhook-avatar-container');
        const avatar = this.createElement('img', 'webhook-avatar');
        
        if (webhook.profile_picture && webhook.profile_picture !== '') {
            avatar.src = webhook.profile_picture;
        } else {
            avatar.src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23cccccc'%3E%3Cpath d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z'/%3E%3C/svg%3E";
        }
        avatar.alt = webhook.name + ' avatar';
        
        avatarContainer.appendChild(avatar);
        
        const nameContainer = this.createElement('div', 'webhook-name-container');
        const name = this.createElement('strong', 'webhook-name', webhook.name);
        nameContainer.appendChild(name);
        
        webhookInfo.appendChild(avatarContainer);
        webhookInfo.appendChild(nameContainer);
        
        const buttonGroup = this.createElement('div', 'webhook-button-group');
        const copyBtn = this.createElement('button', 'webhook-copy-btn', '📋');
        const deleteBtn = this.createElement('button', 'webhook-delete-btn', '🗑️');
        
        const info = this.createElement('div', 'webhook-details');
        const created = this.createElement('small', 'webhook-date', `Created: ${new Date(webhook.created_at).toLocaleDateString()}`);
        const url = this.createElement('div', 'webhook-url');
        const maskedToken = webhook.token ? '•'.repeat(16) : 'TOKEN_HIDDEN';
        const urlText = this.createElement('code', 'webhook-url-text', `POST /api/webhook/${webhook.webhook_id}/${maskedToken}`);
        
        copyBtn.onclick = () => {
            const webhookUrl = `${window.location.origin}/api/webhook/${webhook.webhook_id}/${webhook.token}`;
            navigator.clipboard.writeText(webhookUrl);
            displayErrorMessage('Webhook URL copied!', 'channel-settings-error-container', 'success');
        };
        deleteBtn.onclick = () => this.deleteWebhook(webhook.webhook_id);
        
        buttonGroup.appendChild(copyBtn);
        buttonGroup.appendChild(deleteBtn);
        header.appendChild(webhookInfo);
        header.appendChild(buttonGroup);
        info.appendChild(created);
        url.appendChild(urlText);
        item.appendChild(header);
        item.appendChild(info);
        item.appendChild(url);
        container.appendChild(item);
    });
}
refreshWebhookList() {
    const list = document.querySelector('.webhook-list');
    if (list) this.loadWebhooks(list);
}
async deleteWebhook(webhookId) {
        try {
            await BaseAPI.delete(`/api/webhook/delete/${webhookId}`);
            this.refreshWebhookList();
        } catch (error) {
            displayErrorMessage('Failed to delete webhook', 'channel-settings-error-container', 'error');
        }
    }

async createWebhook() {
            const name = document.getElementById('webhook-name').value;
            if (!name) return;
            
            try {
                const formData = new FormData();
                formData.append('name', name);
                
                const profilePicture = document.getElementById('webhook-profile-picture').files[0];
                if (profilePicture) {
                    formData.append('profile_picture', profilePicture);
                }
                
                const data = await BaseAPI.post(`/api/webhook/create/${this.currentChannelId}`, formData, true);
                displayErrorMessage(`Webhook created! ID: ${data.webhook_id} Token: ${data.token}`, 'channel-settings-error-container', 'success');
                const list = document.querySelector('.webhook-list');
                if (list) this.loadWebhooks(list);
                document.getElementById('webhook-name').value = '';
                
                const profilePreview = document.getElementById('webhook-profile-preview');
                const profileInput = document.getElementById('webhook-profile-picture');
                if (profilePreview && profileInput) {
                    profilePreview.src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%23cccccc'%3E%3Cpath d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z'/%3E%3C/svg%3E";
                    profileInput.value = '';
                }
            } catch (error) {
                displayErrorMessage('Failed to create webhook', 'channel-settings-error-container', 'error');
            }
        }

    createElement(tag, className = '', textContent = '') {
        const element = document.createElement(tag);
        if (className) element.className = className;
        if (textContent) element.textContent = textContent;
        return element;
    }
}

window.channelMenuAPI = new ChannelMenuAPI();

channelMenuAPI
.addTab('overview', 'Overview')
.addTab('webhooks', 'Webhooks')

.addToTab('overview', { type: 'channel-info' })
.addToTab('overview', { type: 'separator' })
.addToTab('overview', { type: 'textarea', label: 'DESCRIPTION', id: 'channel-desc-edit', value: '', placeholder: 'Enter channel description...', rows: 3 })
.addToTab('overview', { type: 'button', text: 'Save Changes', action: async () => {
    const newDesc = document.getElementById('channel-desc-edit').value;
    const currentName = channelMenuAPI.currentChannelName;
    
    try {
        const result = await ChannelAPI.edit(channelMenuAPI.currentChannelId, currentName, newDesc);
        if (result.status === 'success') {
            const descElement = document.getElementById('channel-info-description');
            if (descElement) {
                descElement.textContent = newDesc || 'No description provided';
            }
            console.log('Channel updated successfully');
        }
    } catch (error) {
        console.error('Error updating channel:', error);
    }
}, color: '#4CAF50' })
.addToTab('webhooks', { type: 'webhook-list' })