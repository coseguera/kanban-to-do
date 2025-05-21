/**
 * Copyright (c) 2025 Carlos Oseguera (@coseguera)
 * This code is licensed under a dual-license model.
 * See LICENSE.md for more information.
 */

// Theme toggler functionality
document.addEventListener('DOMContentLoaded', function() {
    const toggleSwitch = document.querySelector('#checkbox');
    const currentTheme = localStorage.getItem('theme');

    // If theme was previously set, apply it
    if (currentTheme) {
        document.documentElement.setAttribute('data-theme', currentTheme);
        
        // Update toggle position if theme is dark
        if (currentTheme === 'dark') {
            toggleSwitch.checked = true;
        }
    } else {
        // Check for system preference on first visit
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
            document.documentElement.setAttribute('data-theme', 'dark');
            toggleSwitch.checked = true;
            localStorage.setItem('theme', 'dark');
        }
    }

    // Listen for toggle switch changes
    toggleSwitch.addEventListener('change', function(e) {
        if (e.target.checked) {
            switchTheme('dark');
        } else {
            switchTheme('light');
        }
    });

    // Function to switch themes
    function switchTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
        
        // Announce theme change for screen readers
        const announcement = document.createElement('div');
        announcement.setAttribute('aria-live', 'polite');
        announcement.classList.add('sr-only');
        announcement.textContent = `Switched to ${theme} mode`;
        document.body.appendChild(announcement);
        
        // Remove announcement after it's been read
        setTimeout(() => {
            document.body.removeChild(announcement);
        }, 3000);
    }
});
