import { HTTPClient, Request, Response } from "../http";
import {AdminWorkersResponse, Worker, Job, JobLog, WrappedLogs as JobLogs} from '../types';

export class AdminClient {
    private _httpClient: HTTPClient;

    constructor(private _host: string) {
        this._httpClient = new HTTPClient();
    }

    async getActiveWorkers(): Promise<AdminWorkersResponse> {
        const response = await this._httpClient.request<AdminWorkersResponse>({
            path: `${this._host}/activeWorkers`,
            method: 'GET',
        } as Request);
        return this._handleResponse(response);
    }

    async getWorker(workerID: string): Promise<Worker> {
        const response = await this._httpClient.request<Worker>({
            path: `${this._host}/worker/${workerID}`,
            method: 'GET',
        } as Request);
        return this._handleResponse(response);    
    }

    async getJob(jobID: string): Promise<Job> {
        const response = await this._httpClient.request<Job>({
            path: `${this._host}/jobs/${jobID}`,
            method: 'GET',
        } as Request);
        return this._handleResponse(response);    
    }

    async getJobLogs(jobID: string): Promise<JobLog[]> {
        const response = await this._httpClient.request<JobLogs>({
            path: `${this._host}/jobs/${jobID}/logs`,
            method: 'GET',
        } as Request);
        const logs = this._handleResponse(response);
        return logs.logs;
    }

    private _handleResponse<T>(response: Response<T>): T {
        if (response.isSuccess && response.body) {
            return response.body;
        }
        if (!response.body) {
            throw 'No response body';
        }
        throw response.body;       
    }
}

export const adminClient = new AdminClient('http://localhost:8080');