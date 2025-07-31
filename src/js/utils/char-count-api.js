class CharCountAPI {
    static counters = new Map();
    static defaultOptions = {
        maxLength: 2000,
        warningThreshold: 1800,
        errorThreshold: 1950,
        className: 'char-count',
        position: 'after'
    };

    static add(textareaId, options = {}) {
        const textarea = document.getElementById(textareaId);
        if (!textarea) {
            console.warn(`CharCountAPI: Textarea ${textareaId} not found`);
            return false;
        }

        this.remove(textareaId);

        const config = { ...this.defaultOptions, ...options };
        
        const counter = document.createElement('span');
        counter.id = `${textareaId}-char-counter`;
        counter.className = config.className;
        counter.textContent = `0/${config.maxLength}`;
        
        if (textarea.nextSibling) {
            textarea.parentNode.insertBefore(counter, textarea.nextSibling);
        } else {
            textarea.parentNode.appendChild(counter);
        }
        
        const updateCounter = () => {
            const length = textarea.value.length;
            counter.textContent = `${length}/${config.maxLength}`;
            
            if (length >= config.errorThreshold) {
                counter.style.color = 'var(--error-color)';
            } else if (length >= config.warningThreshold) {
                counter.style.color = 'var(--warning-color)';
            } else {
                counter.style.color = 'var(--text-muted)';
            }
        };
        
        textarea.addEventListener('input', updateCounter);
        updateCounter();
        
        this.counters.set(textareaId, { counter, textarea, updateCounter, config });
        return true;
    }

    static remove(textareaId) {
        const counterInfo = this.counters.get(textareaId);
        if (counterInfo) {
            counterInfo.textarea.removeEventListener('input', counterInfo.updateCounter);
            if (counterInfo.counter.parentNode) {
                counterInfo.counter.parentNode.removeChild(counterInfo.counter);
            }
            this.counters.delete(textareaId);
        }
    }

    static update(textareaId) {
        const counterInfo = this.counters.get(textareaId);
        if (counterInfo) {
            counterInfo.updateCounter();
        }
    }

    static addMultiple(textareas) {
        textareas.forEach(({ id, options = {} }) => {
            this.add(id, options);
        });
    }
}

window.CharCountAPI = CharCountAPI;