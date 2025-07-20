class EmbedUtils {
    static urlRegex = /(https?:\/\/[^\s]+)/g;
    static angleBracketUrlRegex = /<(https?:\/\/[^\s>]+)>/g;
    
    static linkifyURLs(text) {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/\"/g, '&quot;')
            .replace(/'/g, '&#39;')
            .replace(this.angleBracketUrlRegex, '<a href="$1" target="_blank" rel="noopener noreferrer" class="message-link">$1</a>')
            .replace(this.urlRegex, '<a href="$1" target="_blank" rel="noopener noreferrer" class="message-link">$1</a>')
            .replace(/\n/g, '<br>');
    }

    static createEmbedElement(content) {
        const angleBracketUrls = [...content.matchAll(this.angleBracketUrlRegex)].map(match => match[1]);
        const allUrls = content.match(this.urlRegex) || [];
        const embedUrls = allUrls.filter(url => !angleBracketUrls.includes(url));
        
        if (embedUrls.length === 0) return null;
        
        const container = document.createElement('div');
        container.className = 'message-embeds';
        
        embedUrls.slice(0, 3).forEach(url => {
            const embedContainer = document.createElement('div');
            embedContainer.className = 'message-embed embed-skeleton';
            this.loadEmbed(url, embedContainer);
            container.appendChild(embedContainer);
        });
        
        return container;
    }

    static async loadEmbed(url, container) {
        const cacheKey = `embed_${url.replace(/[^a-zA-Z0-9]/g, '_')}`;
        const cached = localStorage.getItem(cacheKey);
        
        if (cached) {
            try {
                const cacheData = JSON.parse(cached);
                if (cacheData.expires > Date.now() && cacheData.data.success) {
                    container.classList.remove('embed-skeleton');
                    container.innerHTML = this.renderEmbed(cacheData.data);
                    return;
                } else {
                    localStorage.removeItem(cacheKey);
                }
            } catch {
                localStorage.removeItem(cacheKey);
            }
        }
        
        try {
            const response = await fetch(`/api/embed?url=${encodeURIComponent(url)}`);
            const data = await response.json();
            
            if (data.success && (data.title || data.description || data.image)) {
                const cacheData = {
                    data: data,
                    expires: Date.now() + (24 * 60 * 60 * 1000)
                };
                localStorage.setItem(cacheKey, JSON.stringify(cacheData));
                container.classList.remove('embed-skeleton');
                container.innerHTML = this.renderEmbed(data);
            } else {
                container.style.display = 'none';
            }
        } catch (error) {
            container.style.display = 'none';
        }
    }

    static renderEmbed(data) {
        return `
            <div class="embed-content">
                ${data.image ? `<div class="embed-image"><img src="${data.image}" alt="Preview"></div>` : ''}
                <div class="embed-text">
                    ${data.title ? `<div class="embed-title">${data.title}</div>` : ''}
                    ${data.description ? `<div class="embed-description">${data.description}</div>` : ''}
                    ${data.site_name ? `<div class="embed-site">${data.site_name}</div>` : ''}
                </div>
            </div>
        `;
    }
}

window.EmbedUtils = EmbedUtils;