import { IdEntity, IdEntityDiffer } from './IdEntity';

const textDecoder = new TextDecoder('utf8');

export const sleep = (n: number): Promise<void> => new Promise<void>((r) => setTimeout(() => r(), n));

export const doRetry = async <T>(action: () => Promise<T>, retryCount: number, retryInterval: number = 0): Promise<T> => {
    try { 
        return action();
    } catch (e) {
        if (retryCount > 0) {
            await sleep(retryInterval);
            return doRetry(action, retryCount - 1);
        }
        throw e;
    }
};

export const doWithTimeout = <T>(action: () => Promise<T>, timeout: number, timeoutErr: any): Promise<T> => {
    const timeoutId = setTimeout(() => {
        throw timeoutErr;
    }, timeout);
    return action().then(data => {
        clearTimeout(timeoutId);
        return data;
    }).catch(err => {
        clearTimeout(timeoutId);
        throw err;
    });
};

type DiffSet<T> = {
    added: IdEntity<T>[];
    deleted: IdEntity<T>[];
    updated: IdEntity<T>[];
};

export const cmoputeDiff = <K, T extends IdEntity<K> = IdEntity<K>>(original: T[], updated: T[], comparator: IdEntityDiffer<T>): DiffSet<K> => {
    const result: DiffSet<K> = {
        added: [],
        deleted: [],
        updated: [],
    };
    const originalIdMap = new Map(original.map(ori => [ori.id, ori]));
    // get added and overlapping ones
    updated.forEach(e => {
        if (originalIdMap.has(e.id)) {
            // overlapping
            if (!!comparator(originalIdMap.get(e.id)!, e)) {
                result.updated.push(e);
            }
            originalIdMap.delete(e.id);
        } else {
            result.added.push(e);
        }
    });
    // remaining ones are deleted
    originalIdMap.forEach((v) => {
        result.deleted.push(v);
    });
    return result;
};

export const getCurrentQueryParams = (): Record<string, string> => {
    const urlString = document.URL;
    const paramString = urlString.split('?')[1];
    const queryString = new URLSearchParams(paramString);
    const result: Record<string, string> = {};
    queryString.forEach((v, k) => {
        result[k] = v;
    });
    return result;
};

export const uint8ToBase64 = (u8: Uint8Array): string => btoa(textDecoder.decode(u8));