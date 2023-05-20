type NotificationListener = (payload: any) => Promise<void>;

type Disposable = () => boolean;

export class NotificationEmitter {
    private _listnerMap: Map<string, NotificationListener[]>;

    constructor() {
        this._listnerMap = new Map();
    }

    emit(key: string, payload: any): void {
        this.emitAwaitable(key, payload);
    }

    emitAwaitable(key: string, payload: any): Promise<void> {
        const listeners = this._listnerMap.get(key);
        if (listeners && listeners.length) {
            return Promise.all(listeners.map(l => l(payload))).then(() => Promise.resolve());
        }
        return Promise.resolve();
    }

    on(key: string, listener: NotificationListener): Disposable {
        let listeners = this._listnerMap.get(key);
        !listeners && (listeners = []);
        listeners.push(listener);
        this._listnerMap.set(key, listeners);
        return () => this.off(key, listener);
    }

    once(key: string, listener: NotificationListener): Disposable {
        let wrappedListenerRef: NotificationListener;
        const wrappedListener = async (payload: any) => {
            await listener(payload);
            this.off(key, wrappedListenerRef);
        };
        wrappedListenerRef = wrappedListener;
        return () => this.off(key, wrappedListenerRef);
    }

    off(key: string, listener: NotificationListener): boolean {
        const listners = this._listnerMap.get(key);
        if (listners && listners.length) {
            this._listnerMap.set(key, listners.filter(l => l !== listener))
            return true;
        }
        return false;
    }
}

export const emitter = new NotificationEmitter();