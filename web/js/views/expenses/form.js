// Экран добавления и изменения траты.
//
// Отдельная страница, а не форма внутри списка: форма посреди экрана занимала
// место у того, ради чего сюда заходят — сводки и списка, — и вносить траты
// с неё было неудобно. Сюда ведёт кнопка в шапке и тап по строке списка.

import { api } from '../../api.js';
import { parseAmount, todayIso } from '../../format.js';
import { errorMessage } from '../../messages.js';
import { navigate } from '../../router.js';
import { getState, setState } from '../../store.js';
import { escapeHtml, renderShell, skeleton, toast } from '../layout.js';

// id === null — новая трата.
export async function renderExpenseForm(root, id) {
    const main = renderShell(root, 'form', `<div id="form" class="pt-3">${skeleton(1)}</div>`);
    const box = main.querySelector('#form');

    let categories = getState().categories;
    let expense = null;

    try {
        // При заходе по прямой ссылке ни категорий, ни самой траты в состоянии
        // ещё нет — грузим параллельно, чтобы не ждать две задержки сети.
        const [loadedCategories, loadedExpense] = await Promise.all([
            categories.length > 0 ? categories : api.listCategories(),
            id === null ? null : api.getExpense(id),
        ]);

        categories = loadedCategories;
        expense = loadedExpense;
        setState({ categories });
    } catch (error) {
        toast(errorMessage(error));
        navigate('/expenses');
        return;
    }

    if (categories.length === 0) {
        box.innerHTML = `
            <div class="card p-4 text-sm text-neutral-500 dark:text-neutral-400">
                Сначала заведите категорию — <a class="text-blue-600 underline dark:text-blue-400"
                href="#/categories">перейти к категориям</a>.
            </div>
        `;
        return;
    }

    render(box, categories, expense, id);
}

function render(box, categories, expense, id) {
    const submitTitle = id === null ? 'Добавить' : 'Сохранить';

    box.innerHTML = `
        <h1 class="text-xl font-semibold">${id === null ? 'Новая трата' : 'Изменить трату'}</h1>

        <form id="expense-form" class="card mt-3 space-y-3 p-4" novalidate>
            <div>
                <label class="label" for="amount">Сумма</label>
                <input class="field amount" id="amount" name="amount" required
                       inputmode="decimal" autocomplete="off" placeholder="0,00"
                       value="${expense ? String(expense.amount).replace('.', ',') : ''}">
            </div>

            <div>
                <label class="label" for="categoryId">Категория</label>
                <select class="field" id="categoryId" name="categoryId" required>
                    ${categories
                        .map(
                            (c) =>
                                `<option value="${c.id}" ${
                                    expense && expense.categoryId === c.id ? 'selected' : ''
                                }>${escapeHtml(c.name)}</option>`,
                        )
                        .join('')}
                </select>
            </div>

            <div>
                <label class="label" for="description">Описание</label>
                <input class="field" id="description" name="description" maxlength="500"
                       autocomplete="off" placeholder="Необязательно"
                       value="${escapeHtml(expense?.description ?? '')}">
            </div>

            <div>
                <label class="label" for="spentAt">Дата</label>
                <input class="field" id="spentAt" name="spentAt" type="date"
                       value="${expense?.spentAt ?? todayIso()}">
            </div>

            <div class="flex gap-2 pt-1">
                <button class="btn-primary flex-1" type="submit" id="expense-submit">
                    ${submitTitle}
                </button>
                <a class="btn-secondary" href="#/expenses">Отмена</a>
            </div>
        </form>
    `;

    const form = box.querySelector('#expense-form');
    const submit = box.querySelector('#expense-submit');

    form.addEventListener('submit', async (event) => {
        event.preventDefault();

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
            if (id === null) {
                await api.createExpense(body);
            } else {
                await api.updateExpense(id, body);
            }
        } catch (error) {
            toast(errorMessage(error));
            submit.disabled = false;
            submit.textContent = submitTitle;
            return;
        }

        // Возвращаемся к периоду, в котором лежит сохранённая трата. Без этого
        // трата, внесённая задним числом (или добавленная, пока список показывал
        // прошлый месяц), после сохранения просто не видна.
        setState({ anchor: body.spentAt });
        navigate('/expenses');
        toast(id === null ? 'Трата добавлена' : 'Трата сохранена', { variant: 'neutral' });
    });

    form.amount.focus();
}
