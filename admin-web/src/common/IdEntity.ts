export type IdEntity<K> = {
    id: K;
};

export type IdEntityDiffer<T extends IdEntity<unknown>> = (a: T, b: T) => number;