import { getState } from '../store.js';

// Шапка одинакова на всех экранах, кроме логина.
// activeTab подсвечивает текущий раздел.
export function appHeader(activeTab) {
    const state = getState();
    const tab = (path, title, isActive) => `
        <a href="#${path}"
           class="rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
               isActive
                   ? 'bg-neutral-200 text-neutral-900 dark:bg-neutral-800 dark:text-neutral-100'
                   : 'text-neutral-500 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100'
           }">${title}</a>
    `;

    return `
        <header class="sticky top-0 z-10 border-b border-neutral-200 bg-neutral-100/90 backdrop-blur
                       dark:border-neutral-800 dark:bg-neutral-950/90">
            <div class="mx-auto flex max-w-screen-sm items-center gap-1 px-3 py-2">
                ${tab('/expenses', 'Траты', activeTab === 'expenses')}
                ${tab('/categories', 'Категории', activeTab === 'categories')}
                <span class="flex-1"></span>
                <span class="hidden text-sm text-neutral-500 sm:inline dark:text-neutral-400">
                    ${escapeHtml(state.user?.name || state.user?.login || '')}
                </span>
                <button id="logout" type="button"
                        class="min-h-11 rounded-lg px-3 text-sm text-neutral-500 hover:text-red-600
                               dark:text-neutral-400 dark:hover:text-red-400">Выйти</button>
            </div>
        </header>
    `;
}

// Экранирование обязательно: описания трат и названия категорий вводит
// пользователь, а мы вставляем их через innerHTML.
export function escapeHtml(value) {
    return String(value ?? '')
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
}

export function skeleton(lines = 3) {
    return Array.from({ length: lines })
        .map(
            () => `<div class="h-12 animate-pulse rounded-lg bg-neutral-200 dark:bg-neutral-800"></div>`,
        )
        .join('');
}

export function emptyState(text) {
    return `
        <p class="px-4 py-10 text-center text-sm text-neutral-500 dark:text-neutral-400">
            ${escapeHtml(text)}
        </p>
    `;
}

// Всплывающее сообщение об ошибке. Держим одно на всё приложение,
// чтобы они не накапливались стопкой.
export function toast(text) {
    document.querySelector('#toast')?.remove();

    const element = document.createElement('div');
    element.id = 'toast';
    element.className =
        'fixed inset-x-3 bottom-4 z-50 mx-auto max-w-screen-sm rounded-lg bg-red-600 px-4 py-3 ' +
        'text-sm text-white shadow-lg';
    element.style.paddingBottom = 'calc(0.75rem + env(safe-area-inset-bottom))';
    element.textContent = text;

    document.body.append(element);
    setTimeout(() => element.remove(), 4000);
}
