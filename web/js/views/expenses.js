import { api } from '../api.js';
import { colorForCategory } from '../colors.js';
import {
    formatDate,
    formatDateRange,
    formatMoney,
    formatPercent,
    formatWeekday,
    parseAmount,
    todayIso,
} from '../format.js';
import { errorMessage } from '../messages.js';
import { getState, setState } from '../store.js';
import { appHeader, emptyState, escapeHtml, skeleton, toast } from './layout.js';

const PERIODS = [
    { key: 'day', title: 'День' },
    { key: 'week', title: 'Неделя' },
    { key: 'month', title: 'Месяц' },
];

let root = null;
// id редактируемой траты; null — форма работает в режиме добавления.
let editingId = null;

export function renderExpenses(container) {
    root = container;
    editingId = null;

    root.innerHTML = `
        ${appHeader('expenses')}
        <main class="mx-auto max-w-screen-sm px-3 pb-8" style="padding-bottom: calc(2rem + env(safe-area-inset-bottom))">
            <div id="period" class="pt-3"></div>
            <div id="filter" class="pt-3"></div>
            <div id="summary" class="pt-3"></div>
            <div id="form" class="pt-3"></div>
            <div id="list" class="pt-4 space-y-4">${skeleton(4)}</div>
        </main>
    `;

    root.querySelector('#logout').addEventListener('click', async () => {
        const { logout } = await import('../main.js');
        logout();
    });

    renderPeriod();
    renderForm();
    void loadAll();
}

// --- загрузка данных ---

async function loadAll() {
    try {
        const categories = await api.listCategories();
        setState({ categories });
        renderFilter();
        renderForm();
    } catch (error) {
        toast(errorMessage(error));
    }
    await loadExpenses();
}

async function loadExpenses() {
    const { period, categoryFilter } = getState();
    const params = {
        period,
        categoryIds: categoryFilter.length ? categoryFilter.join(',') : undefined,
        limit: 200,
    };

    try {
        // Список и отчёт независимы — запрашиваем параллельно, чтобы экран
        // не ждал две последовательные задержки сети.
        const [list, summary] = await Promise.all([api.listExpenses(params), api.summary(params)]);
        setState({ expenses: list.items, summary });
        renderSummary();
        renderList();
    } catch (error) {
        toast(errorMessage(error));
    }
}

// --- переключатель периода ---

function renderPeriod() {
    const { period } = getState();

    root.querySelector('#period').innerHTML = `
        <div class="grid grid-cols-3 gap-1 rounded-xl bg-neutral-200 p-1 dark:bg-neutral-800">
            ${PERIODS.map(
                (p) => `
                <button type="button" data-period="${p.key}"
                        class="min-h-11 rounded-lg text-sm font-medium transition-colors ${
                            p.key === period
                                ? 'bg-white text-neutral-900 shadow-sm dark:bg-neutral-700 dark:text-neutral-100'
                                : 'text-neutral-600 dark:text-neutral-300'
                        }">${p.title}</button>
            `,
            ).join('')}
        </div>
    `;

    for (const button of root.querySelectorAll('[data-period]')) {
        button.addEventListener('click', () => {
            setState({ period: button.dataset.period });
            renderPeriod();
            root.querySelector('#list').innerHTML = skeleton(4);
            void loadExpenses();
        });
    }
}

// --- фильтр по категориям ---

function renderFilter() {
    const { categories, categoryFilter } = getState();
    const box = root.querySelector('#filter');

    if (categories.length === 0) {
        box.innerHTML = '';
        return;
    }

    // Чипы, а не multi-select: на телефоне выбор нескольких значений
    // в нативном select неудобен.
    box.innerHTML = `
        <div class="-mx-3 flex gap-2 overflow-x-auto px-3 pb-1">
            ${categories
                .map((category) => {
                    const color = colorForCategory(category);
                    const active = categoryFilter.includes(category.id);
                    return `
                    <button type="button" data-category="${category.id}"
                            class="flex min-h-11 shrink-0 items-center gap-2 rounded-full border px-3 text-sm transition-colors ${
                                active
                                    ? 'border-transparent bg-neutral-900 text-white dark:bg-neutral-100 dark:text-neutral-900'
                                    : 'border-neutral-300 text-neutral-600 dark:border-neutral-700 dark:text-neutral-300'
                            }">
                        <span class="size-2.5 rounded-full ${color.dot}"></span>
                        ${escapeHtml(category.name)}
                    </button>
                `;
                })
                .join('')}
        </div>
    `;

    for (const button of box.querySelectorAll('[data-category]')) {
        button.addEventListener('click', () => {
            const id = Number(button.dataset.category);
            const current = getState().categoryFilter;
            const next = current.includes(id) ? current.filter((x) => x !== id) : [...current, id];

            setState({ categoryFilter: next });
            renderFilter();
            root.querySelector('#list').innerHTML = skeleton(4);
            void loadExpenses();
        });
    }
}

// --- сводка ---

function renderSummary() {
    const { summary } = getState();
    const box = root.querySelector('#summary');

    if (!summary) {
        box.innerHTML = '';
        return;
    }

    const hasData = summary.byCategory.length > 0;

    box.innerHTML = `
        <div class="card p-4">
            <p class="text-sm text-neutral-500 dark:text-neutral-400">
                ${escapeHtml(formatDateRange(summary.from, summary.to))}
            </p>
            <p class="amount mt-1 text-3xl font-semibold">${formatMoney(summary.total)} ₽</p>

            ${
                hasData
                    ? `
                <div class="mt-4 flex h-2 overflow-hidden rounded-full bg-neutral-200 dark:bg-neutral-800">
                    ${summary.byCategory
                        .map((item) => {
                            const color = colorForCategory(findCategory(item.categoryId));
                            return `<div class="${color.bar}" style="width: ${item.share}%"></div>`;
                        })
                        .join('')}
                </div>

                <ul class="mt-3 space-y-2">
                    ${summary.byCategory
                        .map((item) => {
                            const color = colorForCategory(findCategory(item.categoryId));
                            return `
                            <li class="flex items-center gap-2 text-sm">
                                <span class="size-2.5 shrink-0 rounded-full ${color.bar}"></span>
                                <span class="truncate">${escapeHtml(item.name)}</span>
                                <span class="flex-1"></span>
                                <span class="amount tabular-nums text-neutral-500 dark:text-neutral-400">
                                    ${formatPercent(item.share)}
                                </span>
                                <span class="amount w-24 text-right font-medium">${formatMoney(item.total)}</span>
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
}

function findCategory(id) {
    return getState().categories.find((category) => category.id === id);
}

// --- форма добавления и редактирования ---

function renderForm() {
    const { categories } = getState();
    const box = root.querySelector('#form');

    if (categories.length === 0) {
        box.innerHTML = `
            <div class="card p-4 text-sm text-neutral-500 dark:text-neutral-400">
                Сначала заведите категорию — <a class="text-blue-600 underline dark:text-blue-400"
                href="#/categories">перейти к категориям</a>.
            </div>
        `;
        return;
    }

    box.innerHTML = `
        <details id="form-details" class="card overflow-hidden">
            <summary class="flex min-h-11 cursor-pointer list-none items-center px-4 font-medium select-none">
                <span id="form-title">+ Добавить трату</span>
            </summary>

            <form id="expense-form" class="space-y-3 border-t border-neutral-200 p-4 dark:border-neutral-800" novalidate>
                <div>
                    <label class="label" for="amount">Сумма</label>
                    <input class="field amount" id="amount" name="amount" required
                           inputmode="decimal" autocomplete="off" placeholder="0,00">
                </div>

                <div>
                    <label class="label" for="categoryId">Категория</label>
                    <select class="field" id="categoryId" name="categoryId" required>
                        ${categories
                            .map((c) => `<option value="${c.id}">${escapeHtml(c.name)}</option>`)
                            .join('')}
                    </select>
                </div>

                <div>
                    <label class="label" for="description">Описание</label>
                    <input class="field" id="description" name="description"
                           autocomplete="off" placeholder="Необязательно">
                </div>

                <div>
                    <label class="label" for="spentAt">Дата</label>
                    <input class="field" id="spentAt" name="spentAt" type="date" value="${todayIso()}">
                </div>

                <div class="flex gap-2">
                    <button class="btn-primary flex-1" type="submit" id="expense-submit">Добавить</button>
                    <button class="btn-secondary hidden" type="button" id="expense-cancel">Отмена</button>
                </div>
            </form>
        </details>
    `;

    box.querySelector('#expense-form').addEventListener('submit', onSubmit);
    box.querySelector('#expense-cancel').addEventListener('click', resetForm);
}

async function onSubmit(event) {
    event.preventDefault();

    const form = event.currentTarget;
    const submit = form.querySelector('#expense-submit');

    const amount = parseAmount(form.amount.value);
    if (amount === null || amount <= 0) {
        toast('Введите сумму больше нуля.');
        form.amount.focus();
        return;
    }

    const body = {
        categoryId: Number(form.categoryId.value),
        amount,
        description: form.description.value.trim(),
        spentAt: form.spentAt.value,
    };

    submit.disabled = true;
    submit.textContent = 'Сохраняем…';

    try {
        if (editingId === null) {
            await api.createExpense(body);
        } else {
            await api.updateExpense(editingId, body);
        }

        // Категорию и дату оставляем: чаще всего следом вносят покупку
        // из того же похода в магазин.
        form.amount.value = '';
        form.description.value = '';
        resetForm();
        await loadExpenses();
    } catch (error) {
        toast(errorMessage(error));
    } finally {
        submit.disabled = false;
        submit.textContent = editingId === null ? 'Добавить' : 'Сохранить';
    }
}

function startEditing(expense) {
    editingId = expense.id;

    const details = root.querySelector('#form-details');
    const form = root.querySelector('#expense-form');

    details.open = true;
    root.querySelector('#form-title').textContent = 'Изменить трату';
    form.amount.value = String(expense.amount).replace('.', ',');
    form.categoryId.value = String(expense.categoryId);
    form.description.value = expense.description;
    form.spentAt.value = expense.spentAt;

    root.querySelector('#expense-submit').textContent = 'Сохранить';
    root.querySelector('#expense-cancel').classList.remove('hidden');

    details.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    form.amount.focus();
}

function resetForm() {
    editingId = null;

    const form = root.querySelector('#expense-form');
    if (!form) return;

    root.querySelector('#form-title').textContent = '+ Добавить трату';
    root.querySelector('#expense-submit').textContent = 'Добавить';
    root.querySelector('#expense-cancel').classList.add('hidden');
    form.amount.value = '';
    form.description.value = '';
}

// --- список трат ---

function renderList() {
    const { expenses } = getState();
    const box = root.querySelector('#list');

    if (expenses.length === 0) {
        box.innerHTML = emptyState('За этот период трат нет.');
        return;
    }

    // Список приходит отсортированным по дате убыванию — группируем подряд
    // идущие записи, без сортировки на клиенте.
    const groups = [];
    for (const expense of expenses) {
        if (groups.at(-1)?.date !== expense.spentAt) {
            groups.push({ date: expense.spentAt, items: [], total: 0 });
        }
        groups.at(-1).items.push(expense);
        groups.at(-1).total += expense.amount;
    }

    box.innerHTML = groups
        .map(
            (group) => `
        <section>
            <div class="flex items-baseline gap-2 px-1 pb-1">
                <h2 class="text-sm font-medium text-neutral-500 dark:text-neutral-400">
                    ${escapeHtml(formatDate(group.date))}
                </h2>
                <span class="text-xs text-neutral-400 dark:text-neutral-500">
                    ${escapeHtml(formatWeekday(group.date))}
                </span>
                <span class="flex-1"></span>
                <span class="amount text-sm text-neutral-500 dark:text-neutral-400">
                    ${formatMoney(Math.round(group.total * 100) / 100)}
                </span>
            </div>

            <ul class="card divide-y divide-neutral-200 overflow-hidden dark:divide-neutral-800">
                ${group.items.map(expenseRow).join('')}
            </ul>
        </section>
    `,
        )
        .join('');

    for (const row of box.querySelectorAll('[data-expense]')) {
        const expense = expenses.find((item) => item.id === Number(row.dataset.expense));

        row.querySelector('[data-edit]').addEventListener('click', () => startEditing(expense));
        row.querySelector('[data-delete]').addEventListener('click', async () => {
            if (!confirm(`Удалить трату на ${formatMoney(expense.amount)} ₽?`)) return;

            try {
                await api.deleteExpense(expense.id);
                await loadExpenses();
            } catch (error) {
                toast(errorMessage(error));
            }
        });
    }
}

function expenseRow(expense) {
    const color = colorForCategory(findCategory(expense.categoryId));
    const title = expense.description || expense.categoryName;
    const subtitle = expense.description ? expense.categoryName : '';

    return `
        <li data-expense="${expense.id}" class="flex items-stretch">
            <span class="w-1 shrink-0 ${color.bar}"></span>

            <button type="button" data-edit
                    class="flex min-h-14 flex-1 items-center gap-3 px-3 py-2 text-left">
                <span class="min-w-0 flex-1">
                    <span class="block truncate">${escapeHtml(title)}</span>
                    ${
                        subtitle
                            ? `<span class="block truncate text-xs text-neutral-500 dark:text-neutral-400">
                                   ${escapeHtml(subtitle)}
                               </span>`
                            : ''
                    }
                </span>
                <span class="amount shrink-0 font-medium">${formatMoney(expense.amount)}</span>
            </button>

            <button type="button" data-delete aria-label="Удалить"
                    class="flex min-h-14 w-11 shrink-0 items-center justify-center text-neutral-400
                           hover:text-red-600 dark:hover:text-red-400">
                <svg viewBox="0 0 20 20" fill="currentColor" class="size-5">
                    <path d="M8.75 3h2.5a.75.75 0 0 1 .75.75V4.5h3.25a.75.75 0 0 1 0 1.5h-.6l-.7 9.1A2.25 2.25 0 0 1 11.7 17H8.3a2.25 2.25 0 0 1-2.25-1.9L5.35 6h-.6a.75.75 0 0 1 0-1.5H8v-.75A.75.75 0 0 1 8.75 3Z"/>
                </svg>
            </button>
        </li>
    `;
}
