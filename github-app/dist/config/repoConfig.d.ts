import { z } from 'zod';
import { Octokit } from '@octokit/rest';
declare const repoConfigSchema: z.ZodDefault<z.ZodObject<{
    version: z.ZodOptional<z.ZodString>;
    review: z.ZodDefault<z.ZodObject<{
        enabled: z.ZodDefault<z.ZodBoolean>;
        auto_review: z.ZodDefault<z.ZodBoolean>;
        max_files: z.ZodDefault<z.ZodNumber>;
        ignore_patterns: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
        languages: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
    }, "strip", z.ZodTypeAny, {
        enabled: boolean;
        auto_review: boolean;
        max_files: number;
        ignore_patterns: string[];
        languages?: string[] | undefined;
    }, {
        enabled?: boolean | undefined;
        auto_review?: boolean | undefined;
        max_files?: number | undefined;
        ignore_patterns?: string[] | undefined;
        languages?: string[] | undefined;
    }>>;
    rules: z.ZodDefault<z.ZodObject<{
        preset: z.ZodDefault<z.ZodEnum<["minimal", "standard", "strict"]>>;
        enable: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
        disable: z.ZodDefault<z.ZodArray<z.ZodString, "many">>;
    }, "strip", z.ZodTypeAny, {
        preset: "minimal" | "standard" | "strict";
        enable: string[];
        disable: string[];
    }, {
        preset?: "minimal" | "standard" | "strict" | undefined;
        enable?: string[] | undefined;
        disable?: string[] | undefined;
    }>>;
    comments: z.ZodDefault<z.ZodObject<{
        inline: z.ZodDefault<z.ZodBoolean>;
        summary: z.ZodDefault<z.ZodBoolean>;
        request_changes: z.ZodDefault<z.ZodBoolean>;
        min_severity: z.ZodDefault<z.ZodEnum<["info", "warning", "error", "critical"]>>;
    }, "strip", z.ZodTypeAny, {
        inline: boolean;
        summary: boolean;
        request_changes: boolean;
        min_severity: "info" | "error" | "warning" | "critical";
    }, {
        inline?: boolean | undefined;
        summary?: boolean | undefined;
        request_changes?: boolean | undefined;
        min_severity?: "info" | "error" | "warning" | "critical" | undefined;
    }>>;
    labels: z.ZodDefault<z.ZodObject<{
        add_on_issues: z.ZodDefault<z.ZodBoolean>;
        critical: z.ZodDefault<z.ZodString>;
        reviewed: z.ZodDefault<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        critical: string;
        add_on_issues: boolean;
        reviewed: string;
    }, {
        critical?: string | undefined;
        add_on_issues?: boolean | undefined;
        reviewed?: string | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    review: {
        enabled: boolean;
        auto_review: boolean;
        max_files: number;
        ignore_patterns: string[];
        languages?: string[] | undefined;
    };
    rules: {
        preset: "minimal" | "standard" | "strict";
        enable: string[];
        disable: string[];
    };
    comments: {
        inline: boolean;
        summary: boolean;
        request_changes: boolean;
        min_severity: "info" | "error" | "warning" | "critical";
    };
    labels: {
        critical: string;
        add_on_issues: boolean;
        reviewed: string;
    };
    version?: string | undefined;
}, {
    version?: string | undefined;
    review?: {
        enabled?: boolean | undefined;
        auto_review?: boolean | undefined;
        max_files?: number | undefined;
        ignore_patterns?: string[] | undefined;
        languages?: string[] | undefined;
    } | undefined;
    rules?: {
        preset?: "minimal" | "standard" | "strict" | undefined;
        enable?: string[] | undefined;
        disable?: string[] | undefined;
    } | undefined;
    comments?: {
        inline?: boolean | undefined;
        summary?: boolean | undefined;
        request_changes?: boolean | undefined;
        min_severity?: "info" | "error" | "warning" | "critical" | undefined;
    } | undefined;
    labels?: {
        critical?: string | undefined;
        add_on_issues?: boolean | undefined;
        reviewed?: string | undefined;
    } | undefined;
}>>;
export type RepoConfig = z.infer<typeof repoConfigSchema>;
/**
 * Load configuration from a repository.
 */
export declare function loadRepoConfig(octokit: Octokit, owner: string, repo: string, ref?: string): Promise<RepoConfig>;
/**
 * Clear configuration cache for a repository.
 */
export declare function clearRepoConfigCache(owner: string, repo: string): void;
export {};
//# sourceMappingURL=repoConfig.d.ts.map