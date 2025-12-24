import { Request, Response, NextFunction } from 'express';
export declare class AppError extends Error {
    message: string;
    statusCode: number;
    code?: (string | Record<string, unknown>) | undefined;
    details?: Record<string, unknown>;
    constructor(message: string, statusCode?: number, code?: (string | Record<string, unknown>) | undefined);
}
export declare const errorHandler: (err: Error, req: Request, res: Response, _next: NextFunction) => Response<any, Record<string, any>>;
//# sourceMappingURL=errorHandler.d.ts.map