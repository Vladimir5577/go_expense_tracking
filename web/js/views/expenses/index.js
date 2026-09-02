// Главный экран: период, фильтр, сводка и список трат.
// Добавление и правка живут на отдельном экране — views/expenses/form.js.
//
// Это контейнер: он один ходит в API и держит состояние, а секции ниже —
// чистая разметка плюс колбэки. При переезде на React секции становятся
// компонентами, а этот файл — страницей с хуками.

import { api } from '../../api.js';
import { addDays, formatDate, formatMoney } from '../../format.js';
import { errorMessage } from '../../messages.js';
import { navigate } from '../../router.js';
import { getState, setState } from '../../store.js';
import { actionSheet, onRowMenu, renderShell, setBusy, skeleton, toast } from '../layout.js';
import { renderFilter } from './filter.js';
import { renderList } from './list.js';
import { renderPeriod } from './period.js';
import { renderSummary } from './summary.js';

// Потолок выборки на сервере — 200 записей на страницу. Берём его целиком:
// пагинация на экране, где сверху и так есть период и фильтр, только мешала бы,
// а о том, что показана не вся выборка, список честно пишет отдельной строкой.
const LIST_LIMIT = 200;

let sections = null;

// Номер последнего отправленного запроса. Пользователь успевает переключить
// период раньше, чем придёт ответ на предыдущий, и без этой проверки более
// медленный ответ затирает более свежий.
let requestId = 0;

export function renderExpenses(root) {
    const main = renderShell(
        root,
        'expenses',
        `
        <div id="period" class="pt-3"></div>
        <div id="filter" class="pt-3"></div>
        <div id="summary" class="pt-3"></div>
        <div id="list" class="space-y-4 pt-4">${skeleton(4)}</div>
    `,
    );

    sections = {
        period: main.querySelector('#period'),
        filter: main.querySelector('#filter'),
        summary: main.querySelector('#summary'),
        list: main.querySelector('#list'),
    };

    onRowMenu(sections.list, openMenu);

    drawPeriod();
    drawSummary();
    void loadAll();
}

// --- загрузка данных ---

async function loadAll() {
    try {
        setState({ categories: await api.listCategories() });
        drawFilter();
    } catch (error) {
        toast(errorMessage(error));
    }
    await loadExpenses();
}

async function loadExpenses() {
    const { period, anchor, categoryFilter } = getState();
    const params = {
        period,
        date: anchor ?? undefined,
        categoryIds: categoryFilter.length ? categoryFilter.join(',') : undefined,
        limit: LIST_LIMIT,
    };

    const id = ++requestId;
    setBusy(sections.list, true);
    setBusy(sections.summary, true);

    try {
        // Список и отчёт независимы — запрашиваем параллельно, чтобы экран
        // не ждал две последовательные задержки сети.
        const [list, summary] = await Promise.all([api.listExpenses(params), api.summary(params)]);
        if (id !== requestId) return;

        setState({ expenses: list.items, totalItems: list.totalItems, summary });
        drawSummary();
        drawList();
    } catch (error) {
        if (id !== requestId) return;
        toast(errorMessage(error));
    } finally {
        if (id === requestId) {
            setBusy(sections.list, false);
            setBusy(sections.summary, false);
        }
    }
}

// --- отрисовка секций ---

function drawPeriod() {
    renderPeriod(sections.period, getState().period, changePeriod);
}

function drawFilter() {
    renderFilter(sections.filter, getState(), toggleCategory);
}

function drawSummary() {
    renderSummary(sections.summary, getState(), {
        onPrev: goPrev,
        onNext: goNext,
        onToday: goToday,
        onPickDate: pickDate,
    });
}

function drawList() {
    renderList(sections.list, getState());
}

// --- выбор периода ---

function changePeriod(period) {
    // Опорную дату не сбрасываем: смотрели июль месяцем, переключились на
    // неделю — логично остаться в июле, а не прыгать в текущую неделю.
    setState({ period });
    drawPeriod();
    void loadExpenses();
}

// Листаем от границ, которые посчитал сервер: день до начала диапазона — это
// гарантированно предыдущий период, каким бы он ни был. Клиенту не нужно знать
// ни длину месяца, ни с какого дня начинается неделя.
function goPrev() {
    const { summary } = getState();
    if (!summary?.from) return;

    setState({ anchor: addDays(summary.from, -1) });
    void loadExpenses();
}

function goNext() {
    const { summary } = getState();
    if (!summary?.to) return;

    setState({ anchor: addDays(summary.to, 1) });
    void loadExpenses();
}

function goToday() {
    setState({ anchor: null });
    void loadExpenses();
}

// Дата из календаря — та же опорная дата, что и у стрелок: сервер сам развернёт
// её в день, неделю или месяц, смотря какой период выбран.
function pickDate(iso) {
    setState({ anchor: iso });
    void loadExpenses();
}

function toggleCategory(id) {
    const current = getState().categoryFilter;
    const next = current.includes(id) ? current.filter((x) => x !== id) : [...current, id];

    setState({ categoryFilter: next });
    drawFilter();
    void loadExpenses();
}

// --- изменение трат ---

// Действия над тратой собраны в меню за кнопкой «⋮»: сама строка не кликается,
// чтобы случайный тап не уводил на форму изменения.
function openMenu(id) {
    const expense = findExpense(id);
    if (!expense) return;

    actionSheet({
        title: expense.description || expense.categoryName,
        subtitle: `${expense.categoryName} · ${formatMoney(expense.amount)} · ${formatDate(expense.spentAt)}`,
        actions: [
            { title: 'Изменить', onClick: () => navigate(`/expense/${id}`) },
            { title: 'Удалить', danger: true, onClick: () => void removeExpense(id) },
        ],
    });
}

async function removeExpense(id) {
    const expense = findExpense(id);
    if (!expense) return;

    try {
        await api.deleteExpense(id);
        await loadExpenses();
    } catch (error) {
        toast(errorMessage(error));
        return;
    }

    toast(`Трата на ${formatMoney(expense.amount)} удалена`, {
        variant: 'neutral',
        action: 'Отменить',
        onAction: () => void restoreExpense(expense),
    });
}

// Отмена удаления — это создание заново: мягкого удаления на сервере нет,
// поэтому у восстановленной траты будет новый id. Для пользователя разницы нет,
// а альтернатива — откладывать удаление на несколько секунд — теряется, если
// закрыть вкладку раньше.
async function restoreExpense(expense) {
    try {
        await api.createExpense({
            categoryId: expense.categoryId,
            amount: expense.amount,
            description: expense.description,
            spentAt: expense.spentAt,
        });
        await loadExpenses();
    } catch (error) {
        toast(errorMessage(error));
    }
}

function findExpense(id) {
    return getState().expenses.find((expense) => expense.id === id);
}
