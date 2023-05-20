export class Observable<T> {
    private _value: T | undefined;
    private _subscribers: ((value?: T) => Promise<void>)[];

    constructor(value?: T) {
        this._value = value;
        this._subscribers = [];
    }

    subscribe(listener: (value?: T) => Promise<void>): () => void {
        if (!listener) {
            return () => {};
        }
        this._subscribers.push(listener);
        return () => {
            this._subscribers.filter(x => x !== listener);
        };
    }

    reset(): void {
        this._subscribers = [];
    }

    get value() {
        return this._value;
    }

    set value(value: T | undefined) {
        this._value = value;
        this._subscribers.forEach(s => s(value));
    }
}