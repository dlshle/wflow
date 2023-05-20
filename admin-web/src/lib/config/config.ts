import defaultConfig from "../../default_config.json";
import { HTTPClient, RequestMethod } from "../http";
import { emitter } from "../../common/NotificationEmitter";
import { Immutable } from "../../common";

export const EVENT_CONFIG_CHANGE = 'cfg_change';

export type AppConfig = Immutable<{
    host: string;
}>;

export class Config {
    private _config: AppConfig;
    private _httpClient: HTTPClient;

    constructor() {
        this._config = defaultConfig;
        this._httpClient = new HTTPClient();
        console.log("default config loaded: ", JSON.stringify(this._config));
    }

    async updateConfig(sourcePath: string): Promise<void> {
        const newConfig = await this._httpClient.request<AppConfig>({path: sourcePath, method: RequestMethod.GET}).catch(err => {
            console.error('unable to fetch new config from ', sourcePath, ' due to ', JSON.stringify(err));
            throw err;
        });
        if (!newConfig.body) {
            throw 'config body is missing from response';
        }
        this._config = newConfig.body;
        emitter.emit(EVENT_CONFIG_CHANGE, this._config);
    }

    get config() {
        return this._config;
    }
}

export const config = new Config();