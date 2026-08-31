// Хешевый роутер: #/login, #/expenses, #/categories.
//
// Хеш выбран сознательно — он не требует ни фолбэка на сервере, ни настройки
// nginx, а на React Router ложится один в один.

const routes = new Map();
let notFoundHandler = null;

export function route(path, handler) {
    routes.set(path, handler);
}

export function setNotFound(handler) {
    notFoundHandler = handler;
}

export function currentPath() {
    return location.hash.slice(1) || '/';
}

export function navigate(path) {
    if (currentPath() === path) {
        // hashchange не сработает — перерисовываем вручную.
        resolve();
        return;
    }
    location.hash = path;
}

function resolve() {
    const path = currentPath();
    const handler = routes.get(path) ?? notFoundHandler;
    handler?.(path);
}

export function startRouter() {
    window.addEventListener('hashchange', resolve);
    resolve();
}
