// Переключатель день/неделя/месяц. Только разметка и колбэк — какой период
// сейчас выбран, знает контейнер экрана.

const PERIODS = [
    { key: 'day', title: 'День' },
    { key: 'week', title: 'Неделя' },
    { key: 'month', title: 'Месяц' },
];

export function renderPeriod(box, period, onChange) {
    box.innerHTML = `
        <div role="group" aria-label="Период"
             class="grid grid-cols-3 gap-1 rounded-xl bg-neutral-200 p-1 dark:bg-neutral-800">
            ${PERIODS.map(
                (p) => `
                <button type="button" data-period="${p.key}" aria-pressed="${p.key === period}"
                        class="min-h-11 rounded-lg text-sm font-medium transition-colors ${
                            p.key === period
                                ? 'bg-white text-neutral-900 shadow-sm dark:bg-neutral-700 dark:text-neutral-100'
                                : 'text-neutral-600 dark:text-neutral-300'
                        }">${p.title}</button>
            `,
            ).join('')}
        </div>
    `;

    for (const button of box.querySelectorAll('[data-period]')) {
        button.addEventListener('click', () => onChange(button.dataset.period));
    }
}
