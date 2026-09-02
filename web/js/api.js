// Тонкая обёртка над fetch: подставляет токен, разбирает наш формат ошибок
// и уводит на логин при 401.
//
// Здесь нет ни DOM, ни разметки — при переезде на React файл переносится
// без единой правки.

const TOKEN_KEY = 'expenses.token';

export function getToken() {
    return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
    localStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
    constructor(status, code) {
        super(code);
        this.name = 'ApiError';
        this.status = status;
        this.code = code;
    }
}

// Приложение регистрирует сюда единственный обработчик протухшей сессии,
// чтобы каждое место вызова не проверяло 401 самостоятельно.
let unauthorizedHandler = null;

export function onUnauthorized(handler) {
    unauthorizedHandler = handler;
}

async function request(method, path, { body, params } = {}) {
    const url = new URL(path, location.origin);

    for (const [key, value] of Object.entries(params ?? {})) {
        if (value !== undefined && value !== null && value !== '') {
            url.searchParams.set(key, value);
        }
    }

    const headers = {};
    const token = getToken();
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    if (body !== undefined) {
        headers['Content-Type'] = 'application/json';
    }

    let response;
    try {
        response = await fetch(url, {
            method,
            headers,
            body: body === undefined ? undefined : JSON.stringify(body),
        });
    } catch {
        // fetch падает только на сетевом уровне — HTTP-ошибки сюда не попадают.
        throw new ApiError(0, 'network_error');
    }

    if (response.status === 401) {
        clearToken();
        unauthorizedHandler?.();
        throw new ApiError(401, 'unauthorized');
    }

    if (response.status === 204) {
        return null;
    }

    const data = await response.json().catch(() => null);

    if (!response.ok) {
        throw new ApiError(response.status, data?.code ?? 'internal_error');
    }

    return data;
}

export const api = {
    login: (login, password) => request('POST', '/api/auth/login', { body: { login, password } }),
    me: () => request('GET', '/api/me'),

    listCategories: () => request('GET', '/api/categories'),
    createCategory: (body) => request('POST', '/api/categories', { body }),
    updateCategory: (id, body) => request('PATCH', `/api/categories/${id}`, { body }),
    deleteCategory: (id) => request('DELETE', `/api/categories/${id}`),

    listExpenses: (params) => request('GET', '/api/expenses', { params }),
    getExpense: (id) => request('GET', `/api/expenses/${id}`),
    createExpense: (body) => request('POST', '/api/expenses', { body }),
    updateExpense: (id, body) => request('PATCH', `/api/expenses/${id}`, { body }),
    deleteExpense: (id) => request('DELETE', `/api/expenses/${id}`),

    summary: (params) => request('GET', '/api/reports/summary', { params }),
};
