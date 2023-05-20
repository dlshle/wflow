import { doRetry, doWithTimeout } from '../../common/utils';
import { Request, Response } from './types';
import { Middleware, MiddlewareContext, runMiddlewares } from './Middleware';

const DEFAULT_REQUEST_TIMEOUT = 60000;
const TIMEOUT_RESPONSE: Omit<Response<string>, 'id'> = {
    code: -1,
    body: 'request timeout',
    headers: {},
    isSuccess: false,
};
const DEFAULT_RETRY_INTERVAL = 1000;

export class HTTPClient {
    private _middlewares: Middleware[];

    constructor() {
        this._middlewares = [];
    }

    public request<T>(request: Request): Promise<Response<T>> {
        return runMiddlewares([...this._middlewares, this._handleRequestMiddleware], request);
    }

    private _handleRequestMiddleware = async (ctx: MiddlewareContext): Promise<void> => {
        const response = await this._handleRequest(ctx.request);
        Object.keys(response).forEach(key =>{
            // @ts-ignore
            ctx.response[key] = response[key];
        })
        /*
        ctx.response.id = response.id;
        ctx.response.code = response.code;
        ctx.response.body = response.body;
        ctx.response.headers = response.headers;
        ctx.response.isSuccess = response.isSuccess;
        */
    }
    
    private _handleRequest = async (request: Request): Promise<Response<any>> => {
        const response: Response<any> = await doRetry(async () => {
            const response = await this._doFetch(request);
            if (response.code >= 400) {
                throw response;
            }
            return response;
        }, request.retry?? 0, request.retryInterval?? DEFAULT_RETRY_INTERVAL);
        return response;
    }

    private _getContentType = (body: any): string => {
        if (typeof body === 'string') {
            return body.startsWith('{')? 'application/json': 'text/plain;charset=UTF-8';
        }
        return 'application/json';
    }

    private _doFetch = async (request: Request) => {
        if (request.body) {
            request.headers = {...request.headers, 'Content-Type': this._getContentType(request.body)}
        }
        const response = await this._fetchWithTimeout(request.host?`${request.host}${request.path}`: request.path, {
            method: request.method,
            headers: request.headers,
            body: JSON.stringify(request.body),
        }, request.timeout?? DEFAULT_REQUEST_TIMEOUT);
        const body = await response.json().catch(err => undefined);
        const code = response.status;
        const headers: Record<string, string> = {};
        response.headers.forEach((v, k) => {
            headers[k] = v;
        })
        return {
            body,
            code,
            headers,
            id: request.id,
            isSuccess: code > 0 && code < 400,
        } as Response<any>
    }

    private _fetchWithTimeout = async (host: string, options: {}, timeout: number) => {
        const controller = new AbortController();
        const id = setTimeout(() => controller.abort(), timeout);
        const response = fetch(host, {
          ...options,
          signal: controller.signal  
        }).then(resp => {
            clearTimeout(id);
            if (!resp.ok) {
                resp.json = () => resp.text().then(text => {
                    console.error('oh no ', JSON.stringify(text));
                    return text;
                });
            }
            return resp;
        });
        clearTimeout(id);
        return response;
      }
}