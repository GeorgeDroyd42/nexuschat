function $(id) {
    return document.getElementById(id);
}

function $$(selector) {
    return document.querySelector(selector);
}

function hide(elementOrId) {
    const el = typeof elementOrId === 'string' ? $(elementOrId) : elementOrId;
    if (el) el.style.display = 'none';
}

function show(elementOrId) {
    const el = typeof elementOrId === 'string' ? $(elementOrId) : elementOrId;
    if (el) el.style.display = 'block';
}

function setText(selector, text) {
    const el = $$(selector);
    if (el) el.textContent = text;
}

window.domUtils = { $, $$, hide, show, setText };