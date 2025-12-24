export declare const config: {
    readonly nodeEnv: "development" | "production" | "test";
    readonly port: number;
    readonly logLevel: "debug" | "info" | "warn" | "error";
    readonly isDevelopment: boolean;
    readonly isProduction: boolean;
    readonly github: {
        readonly appId: number;
        readonly privateKey: string;
        readonly webhookSecret: string;
        readonly clientId: string | undefined;
        readonly clientSecret: string | undefined;
    };
    readonly ai: {
        readonly provider: "ollama" | "openai" | "gemini" | "groq" | "auto";
        readonly model: string;
        readonly ollamaBaseUrl: string;
        readonly openaiApiKey: string | undefined;
        readonly geminiApiKey: string | undefined;
        readonly groqApiKey: string | undefined;
    };
    readonly rateLimit: {
        readonly rps: number;
        readonly burst: number;
    };
    readonly cache: {
        readonly ttl: number;
        readonly maxEntries: number;
    };
    readonly review: {
        readonly maxFiles: number;
        readonly maxDiffSize: number;
        readonly timeout: number;
    };
};
export type Config = typeof config;
//# sourceMappingURL=index.d.ts.map