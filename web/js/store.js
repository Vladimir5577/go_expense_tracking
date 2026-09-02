// Состояние приложения: один объект и единственная точка его изменения.
//
// Подписок и диффинга DOM здесь нет — вьюхи перерисовывают свою секцию целиком
// после того, как контейнер экрана обновил состояние. На объёмах этого
// приложения это быстро, а главное — не превращается в самописный мини-React,
// который потом пришлось бы выпиливать.

const initialState = {
    user: null,
    categories: [],
    expenses: [],
    summary: null,

    // Сколько записей всего подходит под фильтр. Список ограничен limit,
    // и это единственный способ узнать, что показана не вся выборка.
    totalItems: 0,

    // Параметры выборки. Границы периода считает сервер — клиент шлёт period
    // и опорную дату anchor (null — «сегодня»).
    period: 'week',
    anchor: null,
    categoryFilter: [],
};

const state = { ...initialState };

export function getState() {
    return state;
}

export function setState(patch) {
    Object.assign(state, patch);
}

export function resetState() {
    setState({ ...initialState, categoryFilter: [] });
}
