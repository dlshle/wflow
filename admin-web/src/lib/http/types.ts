enum RequestMethod {
    GET = 'GET',
    POST = 'POST',
    PATCH = 'PATCH',
    PUT = 'PUT',
    DELETE = 'DELETE',
    HEAD = 'HEAD',
    OPTIONS = 'OPTIONS',
};

type Request = {
    id?: string;
    headers?: Record<string, string>;
    body?: any;
    path: string;
    host?: string;
    params?: Record<string, string>;
    method: RequestMethod;
    timeout?: number;
    retry?: number;
    retryInterval?: number;
};

type Response<T> = {
    id: string;
    headers: Record<string, string>;
    code: number;
    isSuccess: boolean;
    body?: T;
};

export { 
    RequestMethod,
};

export type { 
    Request, 
    Response,
};