import Dexie from "dexie";

export type TableConfig = {
    name: string;
    indecies: string[];
    autoIncrID?: boolean;
};

export class Database extends Dexie {
    constructor(private _name: string, private _version: number, tablesConfigs: TableConfig[]) {
        super(_name);
        this._init(tablesConfigs);
    }

    private _init(configs: TableConfig[]) {
        if (!(configs && configs.length)) {
            throw 'invalid table config for db ' + this._name;
        }
        const schema = {} as Record<string, string>;
        configs.forEach(cfg => {
            let tblSchema = '' + cfg.indecies;
            cfg.autoIncrID && (tblSchema = 'id++,' + tblSchema);
            schema[cfg.name] = tblSchema;
        });
        this.version(this._version).stores(schema);
    }

    async openDB(): Promise<void> {
        if (!this.isOpen()) {
            await super.open();
        }
    }

    async closeDB() {
        if (this.isOpen()) {
            super.close();
        }
    }

    async withTransaction<T, K>(tblName: string, callback: (tbl: Dexie.Table<T, K>) => Promise<void>): Promise<void> {
        return super.transaction('rw', super.table(tblName), tx => {
            return callback(tx.table<T, K>(tblName));
        });
    }

    getByID<K, T>(tbl: string, id: K): Promise<T | undefined> {
        return super.table<T, K>(tbl).get(id);
    }

    bulkGet<K, T>(tbl: string, keys: K[]): Promise<(T | undefined)[]> {
        return super.table<T, K>(tbl).bulkGet(keys);
    }

    // put returns key of th entity
    put<K, T>(tbl: string, entity: T, id?: K): Promise<K> {
        return super.table<T, K>(tbl).put(entity, id);
    }

    bulkPut<K, T = string | number>(tbl: string, entities: readonly T[]): Promise<K[]> {
        return super.table<T, K>(tbl).bulkPut<true>(entities, {allKeys: true});
    }

    deleteByID<K>(tbl: string, key: K): Promise<void> {
        return super.table<unknown, K>(tbl).delete(key);
    }

    deleteByIDs<K>(tbl: string, keys: K[]): Promise<void> {
        return super.table<unknown, K>(tbl).bulkDelete(keys);
    }

    filter<T>(tbl: string, filterFunc: (entity: T) => boolean): Promise<T[]> {
        let resolverFunc: (res: T[]) => void;
        let rejecterFunc: (err: any) => void;
        const results = [] as T[];
        const promise = new Promise<T[]>((resolver, rejecter) => {
            resolverFunc = resolver;
            rejecterFunc = rejecter;
        });
        super.table<T, unknown>(tbl).filter(filterFunc).each((obj: T) => results.push(obj)).then(() => resolverFunc(results)).catch(err => rejecterFunc(err));
        return promise;
    }

    dropTable(tbl: string): Promise<void> {
        return super.table(tbl).clear();
    }

    withRead<T, K>(tbl: string, cb: (table: Dexie.Table<T, K>) => Promise<void>) {
        return super.transaction('r', super.table(tbl), tx => {
            return cb(tx.table<T, K>(tbl));
        });
    }
}