export const getItem = (key: string) => window.localStorage.getItem(key);
export const setItem = (key: string, value: string) => window.localStorage.setItem(key, value);
export const deleteItem = (key: string) => window.localStorage.clear();
export const clear = () => window.localStorage.clear();