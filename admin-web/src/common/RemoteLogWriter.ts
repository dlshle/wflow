import { getItem, setItem, deleteItem, clear } from './LocalStorageUtil';
import { LogEntity, LogWriter, Logger } from './Logger';
import { HTTPClient, RequestMethod, Response } from '../lib/http';

const LOG_DATA_ID_INDEX_KEY = 'log_data_id_index';
export class LocalLogDataManager {
    private _logDataIDIndexList: number[];
    private _idCounter: number;
    constructor() {
        const idIndex = getItem(LOG_DATA_ID_INDEX_KEY)?? '[]';
        this._logDataIDIndexList = JSON.parse(idIndex);
        this._idCounter = this._logDataIDIndexList[this._logDataIDIndexList.length - 1]?? 0;
    }

    forEachLogDataSet(cb: (id: number, serialisedLogDataSet: string) => Promise<void>) {
        this._logDataIDIndexList.forEach((id: number) => {
            const sid = id + '';
            const data = getItem(sid);
            data? cb(id, data): deleteItem(sid);
        });
    }

    deleteLogDataByID(id: number) {
        const sid = id + '';
        deleteItem(sid);
        this._logDataIDIndexList = this._logDataIDIndexList.filter(k => k !== id);
    }

    saveLogData(data: string) {
        this._idCounter++;
        const nextID = this._idCounter + '';
        setItem(nextID, data);
        this._logDataIDIndexList.push(this._idCounter);
        this._updateLoadDataIDIndexList();
    }

    getLatestLog(): string {
        return getItem(this._logDataIDIndexList[this._logDataIDIndexList.length - 1] + '')?? '';
    }

    private _updateLoadDataIDIndexList() {
        setItem(LOG_DATA_ID_INDEX_KEY, JSON.stringify(this._logDataIDIndexList));
    }

    clearAll() {
        clear();
    }
}

const logDataManager: LocalLogDataManager = new LocalLogDataManager();

const LOG_UPLOAD_MAX_COUNT = 500;
const LOG_UPLOAD_JOB_INTERVAL = 1000 * 60;
export class LogCollector implements LogWriter {
    private _logDataList: Record<string, string>[] = [];
    private _currLogCount: number = 0;
    private _dataUploadJobID: number = -1;

    constructor(private _appId: string, private _logger: Logger, private _httpClient: HTTPClient, private _logServerBaseURL: string) {
        this._dataUploadJobID = setInterval(this._dataUploadIntervalJob, LOG_UPLOAD_JOB_INTERVAL);
    }

    write(entity: LogEntity): void {
        this._writeData(entity);
    }

    private async _writeData(logEntity: LogEntity) {
        // @ts-ignore
        const data = this._flattenLogEntity(logEntity);
        this._currLogCount++;
        this._logDataList.push(data);
        if (this._currLogCount >= LOG_UPLOAD_MAX_COUNT) {
            const toUploadData = [...this._logDataList];
            this._currLogCount = 0;
            this._logDataList = [];
            // upload in async
            this._uploadLogData(toUploadData);
        }
    }

    private _flattenLogEntity(logEntity: LogEntity): Record<string, string> {
        const context = logEntity.context;
        return {
            ...context,
            level: logEntity.level as string,
            timestamp: logEntity.timestamp.toISOString(),
            message: logEntity.message,
        }
    }

    private async _uploadLogData(data: Record<string, string>[]): Promise<void> {
        const serializedData = JSON.stringify(data);
        return this._doUploadLogData(serializedData).catch(err => {
            this._logger.error('unable to upload log data due to ', JSON.stringify(err), ' try caching data locally');
            return this._cacheLogLocally(serializedData);
        });
    }

    private async _doUploadLogData(data: string): Promise<void> {
        const response = await this._httpClient.request<void>({retry: 3, 
            body: data, 
            headers: {'X-App-ID': this._appId},
            host: this._logServerBaseURL, 
            path: '/logs', 
            method: RequestMethod.POST}).catch(err => ({isSuccess: false, headers: {}, code: 0, id: '', data: err} as Response<void>));
        if (!response.isSuccess) {
            throw response.body;
        }
        return Promise.resolve();
    }

    private async _cacheLogLocally(data: string): Promise<void> {
        logDataManager.saveLogData(data);
    }

    private _dataUploadIntervalJob = async () => {
        logDataManager.forEachLogDataSet(async (id: number, logData: string) => {
            if (logData == "") {
                logDataManager.deleteLogDataByID(id);
                return;
            }
            this._doUploadLogData(logData).then(() => logDataManager.deleteLogDataByID(id)).catch(err => {
                this._logger.error('unable to upload log data from local cache due to ', JSON.stringify(err));
            });
        });
    }
}