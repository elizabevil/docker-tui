#!/usr/bin/env bash
// # dtui UI Preview — Theme Switcher
// # Reads ?theme= URL param and sets data-theme attribute on <html>

document.addEventListener('DOMContentLoaded', () => {
    const select = document.getElementById('theme-select');
    if (!select) return;

    const params = new URLSearchParams(window.location.search);
    const saved = localStorage.getItem('dtui-preview-theme') || 'default';
    const theme = params.get('theme') || saved;

    document.documentElement.setAttribute('data-theme', theme);
    select.value = theme;

    select.addEventListener('change', (e) => {
        const t = e.target.value;
        document.documentElement.setAttribute('data-theme', t);
        localStorage.setItem('dtui-preview-theme', t);
    });
});
