// Сводка за период: листание периодов, итог, средний расход в день и доли
// категорий.
//
// Стрелки живут именно здесь, а не отдельной строкой над карточкой: они листают
// ровно тот диапазон, который карточка и показывает.

import { colorForCategory } from '../../colors.js';
import { daysInRange, formatDateRange, formatMoney, formatPercent, todayIso } from '../../format.js';
import { escapeHtml } from '../layout.js';

const navButton = (action, label, path) => `
    <button type="button" data-${action} aria-label="${label}"
            class="flex size-11 shrink-0 items-center justify-center rounded-lg text-neutral-500
                   hover:bg-neutral-100 disabled:pointer-events-none disabled:opacity-30
                   dark:text-neutral-400 dark:hover:bg-neutral-800">
        <svg viewBox="0 0 20 20" fill="currentColor" class="size-5" aria-hidden="true">
            <path d="${path}"/>
        </svg>
    </button>
`;

const PATH_LEFT =
    'M12.7 5.2a.75.75 0 0 1 0 1.06L8.96 10l3.74 3.74a.75.75 0 1 1-1.06 1.06L7.37 10.53a.75.75 0 0 1 0-1.06l4.27-4.27a.75.75 0 0 1 1.06 0Z';
const PATH_RIGHT =
    'M7.3 5.2a.75.75 0 0 0 0 1.06L11.04 10 7.3 13.74a.75.75 0 1 0 1.06 1.06l4.27-4.27a.75.75 0 0 0 0-1.06L8.36 5.2a.75.75 0 0 0-1.06 0Z';

export function renderSummary(box, { summary, categories }, { onPrev, onNext, onToday, onPickDate }) {
    if (!summary) {
        box.innerHTML = `
            <div class="card p-4">
                <div class="mx-auto h-5 w-40 animate-pulse rounded bg-neutral-200 dark:bg-neutral-800"></div>
                <div class="mx-auto mt-3 h-9 w-48 animate-pulse rounded bg-neutral-200 dark:bg-neutral-800"></div>
            </div>
        `;
        return;
    }

    const today = todayIso();
    // Вперёд идти некуда, если текущий диапазон уже дотянулся до сегодня.
    const atLatest = !summary.to || summary.to >= today;
    const showToday = Boolean(summary.from) && (today < summary.from || today > summary.to);

    const days = daysInRange(summary.from, summary.to);
    const average = days > 1 && summary.total > 0 ? summary.total / days : null;
    const rangeTitle = formatDateRange(summary.from, summary.to);

    box.innerHTML = `
        <div class="card p-4">
            <div class="flex items-center gap-1">
                ${navButton('prev', 'Предыдущий период', PATH_LEFT)}

                <!-- Подпись диапазона открывает нативный календарь: он умеет то,
                     чего стрелки не умеют, — прыжок на произвольную дату.
                     Инпут лежит поверх, прозрачный и не ловит клики; кнопке он
                     нужен только как источник showPicker(). -->
                <div class="relative min-w-0 flex-1">
                    <button type="button" data-pick
                            aria-label="Выбрать дату. Сейчас: ${escapeHtml(rangeTitle)}"
                            class="flex min-h-11 w-full cursor-pointer items-center justify-center
                                   gap-1 text-sm text-neutral-500 hover:text-neutral-900
                                   dark:text-neutral-400 dark:hover:text-neutral-100">
                        <span class="truncate underline decoration-dotted underline-offset-4">
                            ${escapeHtml(rangeTitle)}
                        </span>
                        <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"
                             class="size-4 shrink-0 opacity-60">
                            <path d="M5.2 7.3a.75.75 0 0 1 1.1 0L10 11l3.7-3.7a.75.75 0 1 1 1.1 1.06l-4.25 4.25a.75.75 0 0 1-1.1 0L5.2 8.36a.75.75 0 0 1 0-1.06Z"/>
                        </svg>
                    </button>

                    <input type="date" data-date tabindex="-1" aria-hidden="true"
                           value="${summary.from || today}" max="${today}"
                           class="pointer-events-none absolute inset-0 h-full w-full opacity-0">
                </div>

                ${navButton('next', 'Следующий период', PATH_RIGHT)}
            </div>

            <p class="amount mt-1 text-center text-3xl font-semibold">${formatMoney(summary.total)}</p>

            ${
                average === null
                    ? ''
                    : `<p class="mt-1 text-center text-sm text-neutral-500 dark:text-neutral-400">
                           в среднем ${formatMoney(average)} в день
                       </p>`
            }

            ${
                showToday
                    ? `<div class="text-center">
                           <button type="button" data-today
                                   class="inline-flex min-h-11 items-center px-3 text-sm text-blue-600
                                          dark:text-blue-400">Вернуться к сегодня</button>
                       </div>`
                    : ''
            }

            ${
                summary.byCategory.length > 0
                    ? `
                <div class="mt-4 flex h-2 overflow-hidden rounded-full bg-neutral-200 dark:bg-neutral-800"
                     aria-hidden="true">
                    ${summary.byCategory
                        .map((item) => {
                            const color = colorForCategory(findCategory(categories, item.categoryId));
                            return `<div class="${color.bar}" style="width: ${item.share}%"></div>`;
                        })
                        .join('')}
                </div>

                <ul class="mt-3 space-y-2">
                    ${summary.byCategory
                        .map((item) => {
                            const color = colorForCategory(findCategory(categories, item.categoryId));
                            return `
                            <li class="flex items-center gap-2 text-sm">
                                <span class="size-2.5 shrink-0 rounded-full ${color.bar}" aria-hidden="true"></span>
                                <span class="truncate">${escapeHtml(item.name)}</span>
                                <span class="flex-1"></span>
                                <span class="amount text-neutral-500 dark:text-neutral-400">
                                    ${formatPercent(item.share)}
                                </span>
                                <span class="amount w-28 text-right font-medium">${formatMoney(item.total)}</span>
                            </li>
                        `;
                        })
                        .join('')}
                </ul>
            `
                    : ''
            }
        </div>
    `;

    box.querySelector('[data-prev]').addEventListener('click', onPrev);

    const date = box.querySelector('[data-date]');
    date.addEventListener('change', () => {
        if (date.value) onPickDate(date.value);
    });

    box.querySelector('[data-pick]').addEventListener('click', () => {
        // showPicker есть во всех актуальных браузерах, но требует «живого»
        // клика и может бросить — тогда открываем инпут обычным кликом.
        try {
            date.showPicker();
        } catch {
            date.click();
        }
    });

    const next = box.querySelector('[data-next]');
    next.disabled = atLatest;
    next.addEventListener('click', onNext);

    box.querySelector('[data-today]')?.addEventListener('click', onToday);
}

function findCategory(categories, id) {
    return categories.find((category) => category.id === id);
}
