// Фильтр по категориям.
//
// Чипы, а не multi-select: на телефоне выбор нескольких значений в нативном
// select неудобен. Состояние передаётся не только цветом — у каждого чипа есть
// aria-pressed, иначе для скринридера все они выглядят одинаково.

import { colorForCategory } from '../../colors.js';
import { escapeHtml } from '../layout.js';

export function renderFilter(box, { categories, categoryFilter }, onToggle) {
    if (categories.length === 0) {
        box.innerHTML = '';
        return;
    }

    box.innerHTML = `
        <div role="group" aria-label="Фильтр по категориям"
             class="-mx-3 flex gap-2 overflow-x-auto px-3 pb-1">
            ${categories
                .map((category) => {
                    const color = colorForCategory(category);
                    const active = categoryFilter.includes(category.id);
                    return `
                    <button type="button" data-category="${category.id}" aria-pressed="${active}"
                            class="flex min-h-11 shrink-0 items-center gap-2 rounded-full border px-3 text-sm transition-colors ${
                                active
                                    ? 'border-transparent bg-neutral-900 text-white dark:bg-neutral-100 dark:text-neutral-900'
                                    : 'border-neutral-300 text-neutral-600 dark:border-neutral-700 dark:text-neutral-300'
                            }">
                        <span class="size-2.5 rounded-full ${color.dot}" aria-hidden="true"></span>
                        ${escapeHtml(category.name)}
                    </button>
                `;
                })
                .join('')}
        </div>
    `;

    for (const button of box.querySelectorAll('[data-category]')) {
        button.addEventListener('click', () => onToggle(Number(button.dataset.category)));
    }
}
