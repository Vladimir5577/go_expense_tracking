// Форматирование денег, дат и процентов. Чистые функции без DOM —
// при переезде на React файл переносится как есть.

// Валюта задана здесь одним местом: символ ₽ больше нигде в коде не пишется
// руками, иначе он рано или поздно разойдётся между списком и сводкой.
const moneyFormat = new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
});

const monthDayFormat = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long' });
const monthDayShortFormat = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'short' });
const weekdayFormat = new Intl.DateTimeFormat('ru-RU', { weekday: 'short' });

const MS_IN_DAY = 24 * 60 * 60 * 1000;

// 1543.2 → "1 543,20 ₽"
export function formatMoney(value) {
    return moneyFormat.format(value ?? 0);
}

// 41.53 → "41,5%"
export function formatPercent(value) {
    return `${(value ?? 0).toFixed(1).replace('.', ',')}%`;
}

// Разбирает 'YYYY-MM-DD' в локальную дату.
//
// Через new Date('2026-08-30') нельзя: строка без времени трактуется как UTC,
// и в западных таймзонах при выводе получится 29 августа.
export function parseDate(iso) {
    const [year, month, day] = iso.split('-').map(Number);
    return new Date(year, month - 1, day);
}

// '2026-08-30' → "30 августа"
export function formatDate(iso) {
    return monthDayFormat.format(parseDate(iso));
}

// '2026-08-30' → "30 авг"
export function formatDateShort(iso) {
    return monthDayShortFormat.format(parseDate(iso));
}

// '2026-08-30' → "вс"
export function formatWeekday(iso) {
    return weekdayFormat.format(parseDate(iso));
}

// '2026-08-01', '2026-08-31' → "1–31 августа"
export function formatDateRange(fromIso, toIso) {
    if (!fromIso && !toIso) return 'за всё время';
    if (!fromIso) return `по ${formatDate(toIso)}`;
    if (!toIso) return `с ${formatDate(fromIso)}`;
    if (fromIso === toIso) return formatDate(fromIso);

    const from = parseDate(fromIso);
    const to = parseDate(toIso);

    // Внутри одного месяца день начала пишем без названия месяца: «1–31 августа».
    if (from.getMonth() === to.getMonth() && from.getFullYear() === to.getFullYear()) {
        return `${from.getDate()}–${formatDate(toIso)}`;
    }
    return `${formatDate(fromIso)} – ${formatDate(toIso)}`;
}

function toIso(date) {
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${date.getFullYear()}-${month}-${day}`;
}

// Сегодняшняя дата в формате, который принимает API.
export function todayIso() {
    return toIso(new Date());
}

// Сдвиг даты на несколько дней. Единственная календарная арифметика на клиенте:
// длины месяцев и начало недели по-прежнему считает сервер, а листание периодов
// отталкивается от уже посчитанных им границ — «день до from», «день после to».
export function addDays(iso, days) {
    const date = parseDate(iso);
    date.setDate(date.getDate() + days);
    return toIso(date);
}

// Количество дней в диапазоне, включая обе границы.
export function daysInRange(fromIso, toIso) {
    if (!fromIso || !toIso) return 0;
    // Округление нужно из-за перехода на летнее время: между двумя полуночами
    // может оказаться 23 или 25 часов.
    return Math.round((parseDate(toIso) - parseDate(fromIso)) / MS_IN_DAY) + 1;
}

// Разбирает введённую пользователем сумму: принимает и запятую, и точку.
// Возвращает null, если введено не число.
export function parseAmount(raw) {
    const normalized = String(raw).trim().replace(',', '.').replace(/\s/g, '');
    if (normalized === '') return null;

    const value = Number(normalized);
    if (!Number.isFinite(value)) return null;

    return Math.round(value * 100) / 100;
}
