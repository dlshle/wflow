import { HTTPClient, Request, Response } from "../http";
import {AdminActivitiesResponse, AdminWorkersResponse, Worker, Job, JobLog, WrappedLogs as JobLogs, Activity, JobReport, AdminActivitiesWithJobIDsResponse, ActivityWithJobIDs, WrappedLogs} from '../types';

export class AdminClient {
    private _httpClient: HTTPClient;

    constructor(private _host: string) {
        this._httpClient = new HTTPClient();
    }

    async getActiveActivities(): Promise<Activity[]> {
        const response = await this._httpClient.request<AdminActivitiesResponse>({
            path: `${this._host}/activeActivities`,
            method: 'GET',
        } as Request);
        return [...this._handleResponse(response)?.activities?.values()];
    };

    async getActivities(): Promise<ActivityWithJobIDs[]> {
        const response = await this._httpClient.request<AdminActivitiesWithJobIDsResponse>({
            path: `${this._host}/activities`,
            method: 'GET',
        } as Request);
        return [...Object.values(this._handleResponse(response)?.activities?? {})];
    };

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

    async getWorkersByActivityID(activityID: string): Promise<Worker[]> {
        const response = await this._httpClient.request<AdminWorkersResponse>({
            path: `${this._host}/workers/${activityID}`,
            method: 'GET',
        } as Request);
        const workers = this._handleResponse(response).workers;
        return [...Object.values(workers?? {})];
    }

    async getJob(jobID: string): Promise<JobReport> {
        const response = await this._httpClient.request<JobReport>({
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

    async dispatchJob(activityID: string, base64EncodedParam: string, workerID?: string): Promise<JobReport> {
        const requestBody = {
            activityId: activityID,
            payload: base64EncodedParam,
        };
        // @ts-ignore
        !!workerID && (requestBody.workerId = workerID);
        const response = await this._httpClient.request<JobReport>({
            path: `${this._host}/jobs`,
            method: 'POST',
            body: requestBody,
        } as Request);
        return this._handleResponse(response);
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