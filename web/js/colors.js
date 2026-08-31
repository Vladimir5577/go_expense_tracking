// Палитра категорий.
//
// Классы перечислены literal-строками намеренно: Tailwind собирает CSS,
// сканируя исходники на вхождения имён классов, и динамически собранное
// `bg-cat-${key}` в сборку не попадёт.

export const CATEGORY_COLORS = [
    { key: 'red', title: 'Красный', dot: 'bg-cat-red', bar: 'bg-cat-red', ring: 'ring-cat-red' },
    { key: 'orange', title: 'Оранжевый', dot: 'bg-cat-orange', bar: 'bg-cat-orange', ring: 'ring-cat-orange' },
    { key: 'amber', title: 'Янтарный', dot: 'bg-cat-amber', bar: 'bg-cat-amber', ring: 'ring-cat-amber' },
    { key: 'green', title: 'Зелёный', dot: 'bg-cat-green', bar: 'bg-cat-green', ring: 'ring-cat-green' },
    { key: 'teal', title: 'Бирюзовый', dot: 'bg-cat-teal', bar: 'bg-cat-teal', ring: 'ring-cat-teal' },
    { key: 'blue', title: 'Синий', dot: 'bg-cat-blue', bar: 'bg-cat-blue', ring: 'ring-cat-blue' },
    { key: 'indigo', title: 'Индиго', dot: 'bg-cat-indigo', bar: 'bg-cat-indigo', ring: 'ring-cat-indigo' },
    { key: 'violet', title: 'Фиолетовый', dot: 'bg-cat-violet', bar: 'bg-cat-violet', ring: 'ring-cat-violet' },
    { key: 'pink', title: 'Розовый', dot: 'bg-cat-pink', bar: 'bg-cat-pink', ring: 'ring-cat-pink' },
    { key: 'gray', title: 'Серый', dot: 'bg-cat-gray', bar: 'bg-cat-gray', ring: 'ring-cat-gray' },
];

const BY_KEY = new Map(CATEGORY_COLORS.map((c) => [c.key, c]));

const FALLBACK = BY_KEY.get('gray');

// Цвет по ключу из БД. Поле color на бэкенде — свободная строка,
// поэтому неизвестное значение не должно ломать отрисовку.
export function colorOf(key) {
    return BY_KEY.get(key) ?? FALLBACK;
}

// Детерминированный цвет по id — чтобы категории без выбранного цвета
// всё равно различались в списке, а не были все серыми.
export function colorForCategory(category) {
    if (category?.color && BY_KEY.has(category.color)) {
        return BY_KEY.get(category.color);
    }
    if (category?.id) {
        return CATEGORY_COLORS[category.id % CATEGORY_COLORS.length];
    }
    return FALLBACK;
}
