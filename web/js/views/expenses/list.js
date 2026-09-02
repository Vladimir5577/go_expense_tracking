// Список трат, сгруппированный по датам.

import { colorForCategory } from '../../colors.js';
import { formatDate, formatMoney, formatWeekday } from '../../format.js';
import { emptyState, escapeHtml, listRow } from '../layout.js';

export function renderList(box, { expenses, totalItems, categories }) {
    if (expenses.length === 0) {
        // Пустой список у нового пользователя чаще всего означает не «трат нет»,
        // а «не с чего начать»: без категорий форма всё равно ничего не примет.
        box.innerHTML =
            categories.length === 0
                ? `<div class="card p-4 text-center text-sm text-neutral-500 dark:text-neutral-400">
                       Сначала заведите категорию — <a class="text-blue-600 underline dark:text-blue-400"
                       href="#/categories">перейти к категориям</a>.
                   </div>`
                : emptyState('За этот период трат нет.');
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

    const byId = new Map(categories.map((category) => [category.id, category]));

    box.innerHTML = `
        ${groups
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
                    ${group.items.map((expense) => expenseRow(expense, byId)).join('')}
                </ul>
            </section>
        `,
            )
            .join('')}

        ${
            // Сервер отдаёт не больше limit записей. Молча показывать часть выборки
            // нельзя: итог в сводке считается по всему периоду и со списком тогда
            // не сходится.
            totalItems > expenses.length
                ? `<p class="px-1 text-center text-xs text-neutral-500 dark:text-neutral-400">
                       Показаны последние ${expenses.length} трат из ${totalItems} —
                       сузьте период или фильтр, чтобы увидеть остальные.
                   </p>`
                : ''
        }
    `;
}

function expenseRow(expense, byId) {
    const color = colorForCategory(byId.get(expense.categoryId));

    return listRow({
        id: expense.id,
        accent: color.bar,
        title: expense.description || expense.categoryName,
        subtitle: expense.description ? expense.categoryName : '',
        trailing: `<span class="amount shrink-0 font-medium">${formatMoney(expense.amount)}</span>`,
    });
}
