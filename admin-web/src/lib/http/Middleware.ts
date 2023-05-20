import { Request, Response } from "./types";

type MiddlewareContext = {
    readonly request: Request;
    readonly response: Response<any>;
    readonly next: () => Promise<void>;
};

type Middleware = (ctx: MiddlewareContext) => Promise<void>;

export const runMiddlewares = async (middlewares: Middleware[], request: Request): Promise<Response<any>> => {
    const response = {
        id: request.id,
        headers: {},
        code: 0,
        isSuccess: false,
    } as Response<any>;
    let nextFuncHolder = () => Promise.resolve();
    const context = {
        request: request,
        response: response,
        next: nextFuncHolder,
    };
    let index = 0;
    const nextFunc = async () => {
        if (index >= middlewares.length) {
            return;
        }
        return middlewares[index++](context);
    };
    nextFuncHolder = nextFunc;
    await nextFunc();
    return context.response;
};

export type {
    Middleware,
    MiddlewareContext,
}