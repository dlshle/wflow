const requestMap = new Map<string, any>();

export const SingleRequest = (target: any, propertyKey: string, descriptor: PropertyDescriptor) => {
    const originalMethod = descriptor.value;
    descriptor.value = (...args: any[]) => {
        return doSingleRequest(`${originalMethod.name}${args.reduce((accu, curr) => accu + curr, '')}`, originalMethod.apply(this, args));
    }
}

export const doSingleRequest = (key: string, invoker: () => any) => {
    // serialize args
    if (requestMap.has(key)) {
        return requestMap.get(key);
    }
    const result = invoker();
    requestMap.set(key, result);
    if (result instanceof Promise) {
        result.finally(() => {
            requestMap.delete(key);
        });
    }
    return result;
}