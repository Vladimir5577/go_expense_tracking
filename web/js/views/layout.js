// Общие куски разметки: каркас экрана, шапка, строки списков, тосты.
// Ничего не грузит и ничего не знает про API — только `данные → разметка`.

import { logout } from '../session.js';
import { getState } from '../store.js';

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

// Каркас защищённого экрана: шапка сверху, переданная разметка внутри main.
// Возвращает main — контейнер, внутри которого вьюха ищет свои секции.
//
// activeTab: 'expenses' | 'categories' | 'form'. На экране формы вкладка
// «Траты» остаётся подсвеченной, а кнопка добавления убирается — вести с формы
// на форму незачем.
export function renderShell(root, activeTab, bodyHtml) {
    root.innerHTML = `
        ${appHeader(activeTab)}
        <main class="mx-auto max-w-screen-sm px-3 pb-8"
              style="padding-bottom: calc(2rem + env(safe-area-inset-bottom))">
            ${bodyHtml}
        </main>
    `;

    root.querySelector('#logout').addEventListener('click', logout);

    return root.querySelector('main');
}

// Шапка одинакова на всех экранах, кроме логина.
// activeTab подсвечивает текущий раздел.
function appHeader(activeTab) {
    const state = getState();
    const onForm = activeTab === 'form';

    const tab = (path, title, isActive) => `
        <a href="#${path}"
           ${isActive ? 'aria-current="page"' : ''}
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
                <nav class="flex gap-1" aria-label="Разделы">
                    ${tab('/expenses', 'Траты', activeTab === 'expenses' || onForm)}
                    ${tab('/categories', 'Категории', activeTab === 'categories')}
                </nav>
                ${
                    onForm
                        ? ''
                        : // Короткая подпись на узких экранах: с полной шапка
                          // перестаёт помещаться в 360 px.
                          `<a href="#/expense/new" class="btn-add ml-1">
                               <span class="sm:hidden">+ Трата</span>
                               <span class="hidden sm:inline">+ Добавить трату</span>
                           </a>`
                }
                <span class="flex-1"></span>
                <span class="hidden text-sm text-neutral-500 md:inline dark:text-neutral-400">
                    ${escapeHtml(state.user?.name || state.user?.login || '')}
                </span>
                <button id="logout" type="button"
                        class="min-h-11 rounded-lg px-3 text-sm text-neutral-500 hover:text-red-600
                               dark:text-neutral-400 dark:hover:text-red-400">Выйти</button>
            </div>
        </header>
    `;
}

export function skeleton(lines = 3) {
    return Array.from({ length: lines })
        .map(() => `<div class="h-12 animate-pulse rounded-lg bg-neutral-200 dark:bg-neutral-800"></div>`)
        .join('');
}

export function emptyState(text) {
    return `
        <p class="px-4 py-10 text-center text-sm text-neutral-500 dark:text-neutral-400">
            ${escapeHtml(text)}
        </p>
    `;
}

const ICON_MORE = `
    <svg viewBox="0 0 20 20" fill="currentColor" class="size-5" aria-hidden="true">
        <path d="M10 6a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Zm0 5.5a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Zm0 5.5a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3Z"/>
    </svg>
`;

// Строка списка: цветная полоса слева, текст, справа кнопка «⋮» с действиями.
// Сама строка не кликается намеренно: раньше тап по ней открывал форму
// изменения, и это срабатывало неожиданно — по строке чаще попадают, чтобы
// прочитать, а не чтобы править.
export function listRow({ id, accent, title, subtitle = '', trailing = '' }) {
    return `
        <li data-row="${id}" class="flex items-stretch">
            <span class="w-1 shrink-0 ${accent}" aria-hidden="true"></span>

            <div class="flex min-h-14 flex-1 items-center gap-3 px-3 py-2">
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
                ${trailing}
            </div>

            <button type="button" data-menu aria-label="Действия: ${escapeHtml(title)}"
                    class="flex min-h-14 w-11 shrink-0 cursor-pointer items-center justify-center
                           text-neutral-400 hover:text-neutral-900 dark:hover:text-neutral-100">
                ${ICON_MORE}
            </button>
        </li>
    `;
}

// Обработчик вешается один раз на контейнер, а не на каждую строку после каждой
// перерисовки: содержимое контейнера меняется через innerHTML, и подписки на
// сами строки пришлось бы восстанавливать заново.
export function onRowMenu(container, handler) {
    container.addEventListener('click', (event) => {
        const row = event.target.closest('[data-row]');
        if (!row || !container.contains(row)) return;
        if (!event.target.closest('[data-menu]')) return;

        handler(Number(row.dataset.row));
    });
}

// Пункты меню — ровные строки, а не залитые кнопки: залитая красная «Удалить»
// перетягивает на себя внимание с «Изменить», которым пользуются чаще.
const MENU_ITEM =
    'min-h-11 w-full cursor-pointer rounded-lg px-3 text-sm hover:bg-neutral-100 dark:hover:bg-neutral-800';

// Меню действий над строкой.
//
// Нативный <dialog> с showModal(): он рисуется в верхнем слое, поэтому не
// обрезается overflow-hidden у карточки списка, и бесплатно даёт закрытие по
// Esc, возврат фокуса и затемнение фона. На телефоне выезжает снизу, на широком
// экране показывается по центру.
export function actionSheet({ title, subtitle = '', actions }) {
    const dialog = document.createElement('dialog');
    dialog.setAttribute('aria-label', title);
    // Не во всю ширину: меню из двух пунктов, растянутое от края до края, весит
    // больше, чем то, что оно делает. Небольшая карточка, приподнятая над нижним
    // краем, с полями по 12px на узких экранах.
    dialog.className =
        'mx-auto mt-auto mb-3 w-[min(20rem,calc(100%_-_1.5rem))] rounded-2xl border-0 bg-white ' +
        'p-0 text-neutral-900 shadow-xl backdrop:bg-black/40 sm:my-auto ' +
        'dark:bg-neutral-900 dark:text-neutral-100';

    dialog.innerHTML = `
        <div class="px-3 pt-3 pb-2">
            <p class="truncate text-sm font-medium">${escapeHtml(title)}</p>
            ${
                subtitle
                    ? `<p class="mt-0.5 truncate text-xs text-neutral-500 dark:text-neutral-400">
                           ${escapeHtml(subtitle)}
                       </p>`
                    : ''
            }
        </div>

        <div class="space-y-1 border-t border-neutral-200 p-2 dark:border-neutral-800"
             style="padding-bottom: calc(0.5rem + env(safe-area-inset-bottom))">
            ${actions
                .map(
                    (action, index) => `
                <button type="button" data-index="${index}" class="${MENU_ITEM} ${
                    action.danger ? 'text-red-600 dark:text-red-400' : ''
                }">
                    ${escapeHtml(action.title)}
                </button>
            `,
                )
                .join('')}
            <button type="button" data-close class="${MENU_ITEM} text-neutral-500 dark:text-neutral-400">
                Отмена
            </button>
        </div>
    `;

    // Клик по затемнению приходит на сам dialog — содержимое лежит во вложенных
    // блоках, поэтому такой клик означает «мимо меню».
    dialog.addEventListener('click', (event) => {
        if (event.target === dialog) dialog.close();
    });

    dialog.addEventListener('close', () => dialog.remove());

    for (const button of dialog.querySelectorAll('[data-index]')) {
        button.addEventListener('click', () => {
            dialog.close();
            actions[Number(button.dataset.index)].onClick();
        });
    }

    dialog.querySelector('[data-close]').addEventListener('click', () => dialog.close());

    document.body.append(dialog);
    dialog.showModal();
}

// Приглушает секцию на время фонового запроса. Скелетон здесь не годится:
// он меняет высоту блока, и список прыгает при каждой смене фильтра.
export function setBusy(element, busy) {
    element.setAttribute('aria-busy', String(busy));
    element.classList.toggle('opacity-50', busy);
    element.classList.toggle('transition-opacity', true);
}

// Всплывающее сообщение. Держим одно на всё приложение, чтобы они не
// накапливались стопкой. action — необязательная кнопка вроде «Отменить».
export function toast(text, { action, onAction, variant = 'error' } = {}) {
    document.querySelector('#toast')?.remove();

    const element = document.createElement('div');
    element.id = 'toast';
    element.setAttribute('role', 'status');
    element.setAttribute('aria-live', 'polite');
    element.className =
        'fixed inset-x-3 bottom-4 z-50 mx-auto flex max-w-screen-sm items-center gap-3 ' +
        'rounded-lg px-4 py-3 text-sm text-white shadow-lg ' +
        (variant === 'error' ? 'bg-red-600' : 'bg-neutral-800 dark:bg-neutral-700');
    element.style.paddingBottom = 'calc(0.75rem + env(safe-area-inset-bottom))';

    const message = document.createElement('span');
    message.className = 'flex-1';
    message.textContent = text;
    element.append(message);

    if (action) {
        const button = document.createElement('button');
        button.type = 'button';
        button.className = 'shrink-0 font-medium underline underline-offset-2';
        button.textContent = action;
        button.addEventListener('click', () => {
            element.remove();
            onAction?.();
        });
        element.append(button);
    }

    document.body.append(element);

    // С кнопкой тост живёт дольше: сообщение нужно успеть не только прочитать,
    // но и нажать.
    setTimeout(() => element.remove(), action ? 7000 : 4000);
}
