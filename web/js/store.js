// Состояние приложения: один объект и подписки на его изменение.
//
// Здесь нет реактивности и диффинга DOM — вьюхи перерисовывают свою секцию
// целиком. На объёмах этого приложения это быстро, а главное — не превращается
// в самописный мини-React, который потом пришлось бы выпиливать.

const state = {
    user: null,
    categories: [],
    expenses: [],
    summary: null,

    // Параметры выборки. Границы периода считает сервер — клиент шлёт period.
    period: 'week',
    categoryFilter: [],

    loading: false,
    error: null,
};

const listeners = new Set();

export function getState() {
    return state;
}

export function subscribe(listener) {
    listeners.add(listener);
    return () => listeners.delete(listener);
}

export function setState(patch) {
    Object.assign(state, patch);
    for (const listener of listeners) {
        listener(state);
    }
}

export function resetState() {
    setState({
        user: null,
        categories: [],
        expenses: [],
        summary: null,
        period: 'week',
        categoryFilter: [],
        loading: false,
        error: null,
    });
}
