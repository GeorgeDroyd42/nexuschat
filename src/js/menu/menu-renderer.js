class MenuRenderer {
static renderItems(container, items, options = {}) {
        const {
            itemClass = 'menu-item',
            buttonClass = 'btn btn-secondary',
            separatorClass = 'menu-separator'
        } = options;

        items.forEach(item => {
            if (item.type === 'separator') {
                const separator = document.createElement('div');
                separator.className = separatorClass;
                container.appendChild(separator);
            } else if (item.type === 'button') {
                const element = buttonClass.includes('context-menu-item') 
                    ? document.createElement('div')
                    : document.createElement('button');
                element.className = buttonClass;
                element.textContent = item.text;
                if (item.color) element.style.color = item.color;
                element.onclick = item.action;
                container.appendChild(element);
            } else if (item.type === 'input' || item.type === 'textarea') {
                const formGroup = document.createElement('div');
                formGroup.className = `form-group ${itemClass}`;
                
                const label = document.createElement('label');
                label.textContent = item.label;
                label.setAttribute('for', item.id);
                formGroup.appendChild(label);
                
                const input = document.createElement(item.type === 'input' ? 'input' : 'textarea');
                input.id = item.id;
                input.value = item.value;
                input.placeholder = item.placeholder;
                if (item.type === 'input') {
                    input.type = 'text';
                } else {
                    input.rows = item.rows;
                }
                formGroup.appendChild(input);
                
                container.appendChild(formGroup);
            }
        });
    }

    static clearItems(container, itemClass = 'menu-item') {
        const existingItems = container.querySelectorAll(`.${itemClass}`);
        existingItems.forEach(item => item.remove());
    }
}

window.MenuRenderer = MenuRenderer;