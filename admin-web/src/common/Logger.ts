import { HTTPClient, RequestMethod, Response } from "../lib/http";
import { runInAsync } from "./Concurrency";
import { clear, deleteItem, getItem, setItem } from "./LocalStorageUtil";

export enum LogLevel {
    TRACE = "trace",
    DEBUG = "debug",
    INFO = "info",
    WARN = "warn",
    ERROR = "error",
    FATAL = "fatal"
};


export interface ILogger {
    trace(...data: any[]): void;
    debug(...data: any[]): void;
    info(...data: any[]): void;
    warn(...data: any[]): void;
    error(...data: any[]): void;
    fatal(...data: any[]): void;
    withPrefix(prefix: string, override?: boolean): ILogger;
    withContext(context: Record<string, string>): ILogger;
}

export type LogEntity = {
    level: LogLevel;
    message: string;
    timestamp: Date;
    context: Record<string, string>;
};

export interface LogWriter {
    write(entity: LogEntity): void;
}

export type LoggerConfig = {writer: LogWriter, prefix?: string, context?: Record<string, string>};

export class Logger implements ILogger {
    private _writer: LogWriter;
    private _prefix: string;
    private _context: Record<string, string>;

    constructor(config: LoggerConfig) {
        this._writer = config.writer;
        this._prefix = config.prefix?? '';
        this._context = config.context?? {};
    }

    private _toEntity(level: LogLevel, data: any[]): LogEntity {
        return {
            level,
            message: `${this._prefix?? ''}${data?.reduce((accu, curr) => `${accu}${curr}`?? '')}`,
            timestamp: new Date(),
            context: this._context,
        };
    }

    trace(...data: any[]): void {
        this._writer.write(this._toEntity(LogLevel.TRACE, data));
    }

    debug(...data: any[]): void {
        this._writer.write(this._toEntity(LogLevel.DEBUG, data));
    }

    info(...data: any[]): void {
        this._writer.write(this._toEntity(LogLevel.INFO, data));
    }

    warn(...data: any[]): void {
        this._writer.write(this._toEntity(LogLevel.WARN, data));
    }

    error(...data: any[]): void {
        this._writer.write(this._toEntity(LogLevel.ERROR, data));
    }

    fatal(...data: any[]): void {
        this._writer.write(this._toEntity(LogLevel.FATAL, data));
    }

    withPrefix(prefix: string, override?: boolean): ILogger {
        return new Logger({...this.config, prefix: (override? prefix?? '': '') + prefix});
    }

    withContext(context: Record<string, string>): ILogger {
        return new Logger({...this.config, context});
    }

    get config() {
        return {writer: this._writer, prefix: this._prefix, context: this._context};
    }
}

export class TeeLogWriter implements LogWriter {
    constructor(private _writers: LogWriter[]) {}

    write(entity: LogEntity): void {
        this._writers.forEach(writer => writer.write(entity));
    }
}

const CONSOLE_LOG_LEVELS = [LogLevel.DEBUG, LogLevel.INFO, LogLevel.WARN, LogLevel.ERROR];
export class ConsoleWriter implements LogWriter {
    private _outputMap: {
        [key: string]: (...data: any[]) => void;
    };

    constructor() {
        this._outputMap = CONSOLE_LOG_LEVELS.reduce((accu, curr) => {
            // @ts-ignore
            accu[curr] = console[curr];
            return accu;
        }, {});
        this._outputMap[LogLevel.TRACE] = console.log;
        this._outputMap[LogLevel.FATAL] = console.error;
    }

    write(entity: LogEntity): void {
        this._outputMap[entity.level](entity.timestamp.toISOString(), `[${entity.level}]`, entity.message);
    }
}

export class ConsoleLogger extends Logger {
    constructor(config: Omit<LoggerConfig, 'writer'>) {
        super({...config, writer: new ConsoleWriter()});
    }
}
