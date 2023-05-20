type AsyncFunction = () => Promise<any>;
export const runInAsync = async (functions: AsyncFunction[], concurrency: number = 4, stopOnException?: boolean): Promise<any[]> => {
    let index = 0;
    let shouldStop = false;
    const result = Array.from({ length: functions.length });
    try {
        await Promise.all(Array.from({ length: concurrency }, 
            async () => {
                while (index < functions.length && !shouldStop) {
                    const currIndex = index++;
                    const task = functions[currIndex];
                    try {
                        const result = await task();
                        result[currIndex] = result[currIndex];
                    } catch (e) {
                        if (stopOnException) {
                            shouldStop = true;
                            throw e;
                        }
                        result[currIndex] = undefined;
                    }
                }
            }
        ));
    } catch (e) {
        throw e;
    }
    return result;
};

export const runInSeries = async (functions: AsyncFunction[], stopOnException?:boolean): Promise<any[]> => {
    return functions.map(async task => {
        try {
            const result = await task();
            return result;
        } catch (e) {
            if (stopOnException) {
                throw e;
            }
            return undefined;
        }
    });
};