/**
 * Input validation middleware using Zod schemas
 * Provides comprehensive validation for webhook payloads and API requests
 */
import type { RequestHandler } from 'express';
import { z } from 'zod';
/**
 * GitHub user schema
 */
export declare const GitHubUserSchema: z.ZodObject<{
    id: z.ZodNumber;
    login: z.ZodString;
    node_id: z.ZodOptional<z.ZodString>;
    avatar_url: z.ZodOptional<z.ZodString>;
    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
}, "strip", z.ZodTypeAny, {
    id: number;
    login: string;
    type?: "Bot" | "User" | "Organization" | undefined;
    node_id?: string | undefined;
    avatar_url?: string | undefined;
}, {
    id: number;
    login: string;
    type?: "Bot" | "User" | "Organization" | undefined;
    node_id?: string | undefined;
    avatar_url?: string | undefined;
}>;
/**
 * GitHub repository schema
 */
export declare const GitHubRepositorySchema: z.ZodObject<{
    id: z.ZodNumber;
    node_id: z.ZodOptional<z.ZodString>;
    name: z.ZodString;
    full_name: z.ZodString;
    private: z.ZodBoolean;
    owner: z.ZodObject<{
        id: z.ZodNumber;
        login: z.ZodString;
        node_id: z.ZodOptional<z.ZodString>;
        avatar_url: z.ZodOptional<z.ZodString>;
        type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }>;
    default_branch: z.ZodOptional<z.ZodString>;
    language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
}, "strip", z.ZodTypeAny, {
    owner: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    };
    id: number;
    name: string;
    private: boolean;
    full_name: string;
    node_id?: string | undefined;
    default_branch?: string | undefined;
    language?: string | null | undefined;
}, {
    owner: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    };
    id: number;
    name: string;
    private: boolean;
    full_name: string;
    node_id?: string | undefined;
    default_branch?: string | undefined;
    language?: string | null | undefined;
}>;
/**
 * GitHub installation schema
 */
export declare const GitHubInstallationSchema: z.ZodObject<{
    id: z.ZodNumber;
    account: z.ZodOptional<z.ZodObject<{
        id: z.ZodNumber;
        login: z.ZodString;
        node_id: z.ZodOptional<z.ZodString>;
        avatar_url: z.ZodOptional<z.ZodString>;
        type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }>>;
    repository_selection: z.ZodOptional<z.ZodEnum<["all", "selected"]>>;
    permissions: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
}, "strip", z.ZodTypeAny, {
    id: number;
    account?: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    } | undefined;
    repository_selection?: "all" | "selected" | undefined;
    permissions?: Record<string, string> | undefined;
}, {
    id: number;
    account?: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    } | undefined;
    repository_selection?: "all" | "selected" | undefined;
    permissions?: Record<string, string> | undefined;
}>;
/**
 * Pull request webhook payload
 */
export declare const PullRequestPayloadSchema: z.ZodObject<{
    action: z.ZodEnum<["opened", "closed", "reopened", "synchronize", "edited", "ready_for_review", "review_requested", "labeled", "unlabeled", "assigned", "unassigned"]>;
    number: z.ZodNumber;
    pull_request: z.ZodObject<{
        id: z.ZodNumber;
        number: z.ZodNumber;
        state: z.ZodEnum<["open", "closed"]>;
        title: z.ZodString;
        body: z.ZodOptional<z.ZodNullable<z.ZodString>>;
        draft: z.ZodOptional<z.ZodBoolean>;
        head: z.ZodObject<{
            ref: z.ZodString;
            sha: z.ZodString;
            repo: z.ZodOptional<z.ZodNullable<z.ZodObject<{
                id: z.ZodNumber;
                node_id: z.ZodOptional<z.ZodString>;
                name: z.ZodString;
                full_name: z.ZodString;
                private: z.ZodBoolean;
                owner: z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>;
                default_branch: z.ZodOptional<z.ZodString>;
                language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
            }, "strip", z.ZodTypeAny, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }>>>;
        }, "strip", z.ZodTypeAny, {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        }, {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        }>;
        base: z.ZodObject<{
            ref: z.ZodString;
            sha: z.ZodString;
            repo: z.ZodOptional<z.ZodNullable<z.ZodObject<{
                id: z.ZodNumber;
                node_id: z.ZodOptional<z.ZodString>;
                name: z.ZodString;
                full_name: z.ZodString;
                private: z.ZodBoolean;
                owner: z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>;
                default_branch: z.ZodOptional<z.ZodString>;
                language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
            }, "strip", z.ZodTypeAny, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }>>>;
        }, "strip", z.ZodTypeAny, {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        }, {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        }>;
        user: z.ZodObject<{
            id: z.ZodNumber;
            login: z.ZodString;
            node_id: z.ZodOptional<z.ZodString>;
            avatar_url: z.ZodOptional<z.ZodString>;
            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }>;
        merged: z.ZodOptional<z.ZodBoolean>;
        mergeable: z.ZodOptional<z.ZodNullable<z.ZodBoolean>>;
        changed_files: z.ZodOptional<z.ZodNumber>;
        additions: z.ZodOptional<z.ZodNumber>;
        deletions: z.ZodOptional<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        number: number;
        base: {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        };
        user: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        title: string;
        state: "open" | "closed";
        head: {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        };
        additions?: number | undefined;
        deletions?: number | undefined;
        body?: string | null | undefined;
        draft?: boolean | undefined;
        merged?: boolean | undefined;
        mergeable?: boolean | null | undefined;
        changed_files?: number | undefined;
    }, {
        number: number;
        base: {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        };
        user: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        title: string;
        state: "open" | "closed";
        head: {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        };
        additions?: number | undefined;
        deletions?: number | undefined;
        body?: string | null | undefined;
        draft?: boolean | undefined;
        merged?: boolean | undefined;
        mergeable?: boolean | null | undefined;
        changed_files?: number | undefined;
    }>;
    repository: z.ZodObject<{
        id: z.ZodNumber;
        node_id: z.ZodOptional<z.ZodString>;
        name: z.ZodString;
        full_name: z.ZodString;
        private: z.ZodBoolean;
        owner: z.ZodObject<{
            id: z.ZodNumber;
            login: z.ZodString;
            node_id: z.ZodOptional<z.ZodString>;
            avatar_url: z.ZodOptional<z.ZodString>;
            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }>;
        default_branch: z.ZodOptional<z.ZodString>;
        language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
    }, "strip", z.ZodTypeAny, {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    }, {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    }>;
    installation: z.ZodOptional<z.ZodObject<{
        id: z.ZodNumber;
        account: z.ZodOptional<z.ZodObject<{
            id: z.ZodNumber;
            login: z.ZodString;
            node_id: z.ZodOptional<z.ZodString>;
            avatar_url: z.ZodOptional<z.ZodString>;
            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }>>;
        repository_selection: z.ZodOptional<z.ZodEnum<["all", "selected"]>>;
        permissions: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    }, {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    }>>;
    sender: z.ZodObject<{
        id: z.ZodNumber;
        login: z.ZodString;
        node_id: z.ZodOptional<z.ZodString>;
        avatar_url: z.ZodOptional<z.ZodString>;
        type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    number: number;
    action: "opened" | "synchronize" | "reopened" | "edited" | "closed" | "ready_for_review" | "review_requested" | "labeled" | "unlabeled" | "assigned" | "unassigned";
    repository: {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    };
    pull_request: {
        number: number;
        base: {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        };
        user: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        title: string;
        state: "open" | "closed";
        head: {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        };
        additions?: number | undefined;
        deletions?: number | undefined;
        body?: string | null | undefined;
        draft?: boolean | undefined;
        merged?: boolean | undefined;
        mergeable?: boolean | null | undefined;
        changed_files?: number | undefined;
    };
    sender: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    };
    installation?: {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    } | undefined;
}, {
    number: number;
    action: "opened" | "synchronize" | "reopened" | "edited" | "closed" | "ready_for_review" | "review_requested" | "labeled" | "unlabeled" | "assigned" | "unassigned";
    repository: {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    };
    pull_request: {
        number: number;
        base: {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        };
        user: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        title: string;
        state: "open" | "closed";
        head: {
            ref: string;
            sha: string;
            repo?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | null | undefined;
        };
        additions?: number | undefined;
        deletions?: number | undefined;
        body?: string | null | undefined;
        draft?: boolean | undefined;
        merged?: boolean | undefined;
        mergeable?: boolean | null | undefined;
        changed_files?: number | undefined;
    };
    sender: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    };
    installation?: {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    } | undefined;
}>;
/**
 * Installation webhook payload
 */
export declare const InstallationPayloadSchema: z.ZodObject<{
    action: z.ZodEnum<["created", "deleted", "suspend", "unsuspend", "new_permissions_accepted"]>;
    installation: z.ZodObject<{
        id: z.ZodNumber;
        account: z.ZodOptional<z.ZodObject<{
            id: z.ZodNumber;
            login: z.ZodString;
            node_id: z.ZodOptional<z.ZodString>;
            avatar_url: z.ZodOptional<z.ZodString>;
            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }>>;
        repository_selection: z.ZodOptional<z.ZodEnum<["all", "selected"]>>;
        permissions: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    }, {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    }>;
    repositories: z.ZodOptional<z.ZodArray<z.ZodObject<{
        id: z.ZodNumber;
        name: z.ZodString;
        full_name: z.ZodString;
        private: z.ZodBoolean;
    }, "strip", z.ZodTypeAny, {
        id: number;
        name: string;
        private: boolean;
        full_name: string;
    }, {
        id: number;
        name: string;
        private: boolean;
        full_name: string;
    }>, "many">>;
    sender: z.ZodObject<{
        id: z.ZodNumber;
        login: z.ZodString;
        node_id: z.ZodOptional<z.ZodString>;
        avatar_url: z.ZodOptional<z.ZodString>;
        type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    action: "deleted" | "created" | "suspend" | "unsuspend" | "new_permissions_accepted";
    installation: {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    };
    sender: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    };
    repositories?: {
        id: number;
        name: string;
        private: boolean;
        full_name: string;
    }[] | undefined;
}, {
    action: "deleted" | "created" | "suspend" | "unsuspend" | "new_permissions_accepted";
    installation: {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    };
    sender: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    };
    repositories?: {
        id: number;
        name: string;
        private: boolean;
        full_name: string;
    }[] | undefined;
}>;
/**
 * Push webhook payload
 */
export declare const PushPayloadSchema: z.ZodObject<{
    ref: z.ZodString;
    before: z.ZodString;
    after: z.ZodString;
    created: z.ZodOptional<z.ZodBoolean>;
    deleted: z.ZodOptional<z.ZodBoolean>;
    forced: z.ZodOptional<z.ZodBoolean>;
    commits: z.ZodOptional<z.ZodArray<z.ZodObject<{
        id: z.ZodString;
        message: z.ZodString;
        author: z.ZodObject<{
            name: z.ZodString;
            email: z.ZodString;
        }, "strip", z.ZodTypeAny, {
            name: string;
            email: string;
        }, {
            name: string;
            email: string;
        }>;
        added: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
        removed: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
        modified: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
    }, "strip", z.ZodTypeAny, {
        message: string;
        id: string;
        author: {
            name: string;
            email: string;
        };
        added?: string[] | undefined;
        modified?: string[] | undefined;
        removed?: string[] | undefined;
    }, {
        message: string;
        id: string;
        author: {
            name: string;
            email: string;
        };
        added?: string[] | undefined;
        modified?: string[] | undefined;
        removed?: string[] | undefined;
    }>, "many">>;
    repository: z.ZodObject<{
        id: z.ZodNumber;
        node_id: z.ZodOptional<z.ZodString>;
        name: z.ZodString;
        full_name: z.ZodString;
        private: z.ZodBoolean;
        owner: z.ZodObject<{
            id: z.ZodNumber;
            login: z.ZodString;
            node_id: z.ZodOptional<z.ZodString>;
            avatar_url: z.ZodOptional<z.ZodString>;
            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }>;
        default_branch: z.ZodOptional<z.ZodString>;
        language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
    }, "strip", z.ZodTypeAny, {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    }, {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    }>;
    installation: z.ZodOptional<z.ZodObject<{
        id: z.ZodNumber;
        account: z.ZodOptional<z.ZodObject<{
            id: z.ZodNumber;
            login: z.ZodString;
            node_id: z.ZodOptional<z.ZodString>;
            avatar_url: z.ZodOptional<z.ZodString>;
            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }>>;
        repository_selection: z.ZodOptional<z.ZodEnum<["all", "selected"]>>;
        permissions: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    }, {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    }>>;
    sender: z.ZodObject<{
        id: z.ZodNumber;
        login: z.ZodString;
        node_id: z.ZodOptional<z.ZodString>;
        avatar_url: z.ZodOptional<z.ZodString>;
        type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    ref: string;
    repository: {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    };
    sender: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    };
    before: string;
    after: string;
    deleted?: boolean | undefined;
    installation?: {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    } | undefined;
    created?: boolean | undefined;
    forced?: boolean | undefined;
    commits?: {
        message: string;
        id: string;
        author: {
            name: string;
            email: string;
        };
        added?: string[] | undefined;
        modified?: string[] | undefined;
        removed?: string[] | undefined;
    }[] | undefined;
}, {
    ref: string;
    repository: {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    };
    sender: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    };
    before: string;
    after: string;
    deleted?: boolean | undefined;
    installation?: {
        id: number;
        account?: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        } | undefined;
        repository_selection?: "all" | "selected" | undefined;
        permissions?: Record<string, string> | undefined;
    } | undefined;
    created?: boolean | undefined;
    forced?: boolean | undefined;
    commits?: {
        message: string;
        id: string;
        author: {
            name: string;
            email: string;
        };
        added?: string[] | undefined;
        modified?: string[] | undefined;
        removed?: string[] | undefined;
    }[] | undefined;
}>;
/**
 * Ping webhook payload
 */
export declare const PingPayloadSchema: z.ZodObject<{
    zen: z.ZodString;
    hook_id: z.ZodNumber;
    hook: z.ZodOptional<z.ZodObject<{
        id: z.ZodNumber;
        type: z.ZodString;
        events: z.ZodArray<z.ZodString, "many">;
        active: z.ZodBoolean;
    }, "strip", z.ZodTypeAny, {
        type: string;
        id: number;
        events: string[];
        active: boolean;
    }, {
        type: string;
        id: number;
        events: string[];
        active: boolean;
    }>>;
    repository: z.ZodOptional<z.ZodObject<{
        id: z.ZodNumber;
        node_id: z.ZodOptional<z.ZodString>;
        name: z.ZodString;
        full_name: z.ZodString;
        private: z.ZodBoolean;
        owner: z.ZodObject<{
            id: z.ZodNumber;
            login: z.ZodString;
            node_id: z.ZodOptional<z.ZodString>;
            avatar_url: z.ZodOptional<z.ZodString>;
            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }>;
        default_branch: z.ZodOptional<z.ZodString>;
        language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
    }, "strip", z.ZodTypeAny, {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    }, {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    }>>;
    sender: z.ZodOptional<z.ZodObject<{
        id: z.ZodNumber;
        login: z.ZodString;
        node_id: z.ZodOptional<z.ZodString>;
        avatar_url: z.ZodOptional<z.ZodString>;
        type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
    }, "strip", z.ZodTypeAny, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }, {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    zen: string;
    hook_id: number;
    repository?: {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    } | undefined;
    sender?: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    } | undefined;
    hook?: {
        type: string;
        id: number;
        events: string[];
        active: boolean;
    } | undefined;
}, {
    zen: string;
    hook_id: number;
    repository?: {
        owner: {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        };
        id: number;
        name: string;
        private: boolean;
        full_name: string;
        node_id?: string | undefined;
        default_branch?: string | undefined;
        language?: string | null | undefined;
    } | undefined;
    sender?: {
        id: number;
        login: string;
        type?: "Bot" | "User" | "Organization" | undefined;
        node_id?: string | undefined;
        avatar_url?: string | undefined;
    } | undefined;
    hook?: {
        type: string;
        id: number;
        events: string[];
        active: boolean;
    } | undefined;
}>;
/**
 * Admin job query params
 */
export declare const JobQuerySchema: z.ZodObject<{
    status: z.ZodOptional<z.ZodEnum<["pending", "processing", "completed", "failed"]>>;
    limit: z.ZodDefault<z.ZodOptional<z.ZodPipeline<z.ZodEffects<z.ZodString, number, string>, z.ZodNumber>>>;
    offset: z.ZodDefault<z.ZodOptional<z.ZodPipeline<z.ZodEffects<z.ZodString, number, string>, z.ZodNumber>>>;
}, "strip", z.ZodTypeAny, {
    limit: number;
    offset: number;
    status?: "pending" | "processing" | "completed" | "failed" | undefined;
}, {
    status?: "pending" | "processing" | "completed" | "failed" | undefined;
    limit?: string | undefined;
    offset?: string | undefined;
}>;
/**
 * Job ID param
 */
export declare const JobIdParamSchema: z.ZodObject<{
    id: z.ZodString;
}, "strip", z.ZodTypeAny, {
    id: string;
}, {
    id: string;
}>;
/**
 * Sanitizes a string by removing potentially dangerous characters.
 * Note: For webhook payloads from GitHub, we trust the source.
 * This sanitization removes control characters that could cause issues in logging/display.
 */
export declare function sanitizeString(input: string): string;
/**
 * Sanitizes an object recursively
 */
export declare function sanitizeObject<T extends Record<string, unknown>>(obj: T): T;
interface ValidationOptions {
    /** Where to find the data to validate */
    source?: 'body' | 'query' | 'params';
    /** Whether to sanitize strings */
    sanitize?: boolean;
    /** Custom error message */
    errorMessage?: string;
    /** Whether to strip unknown keys */
    stripUnknown?: boolean;
}
/**
 * Creates a validation middleware for a Zod schema
 */
export declare function validateRequest<T extends z.ZodType>(schema: T, options?: ValidationOptions): RequestHandler;
/**
 * Validates request body
 */
export declare function validateBody<T extends z.ZodType>(schema: T, options?: Omit<ValidationOptions, 'source'>): RequestHandler;
/**
 * Validates query parameters
 */
export declare function validateQuery<T extends z.ZodType>(schema: T, options?: Omit<ValidationOptions, 'source'>): RequestHandler;
/**
 * Validates route parameters
 */
export declare function validateParams<T extends z.ZodType>(schema: T, options?: Omit<ValidationOptions, 'source'>): RequestHandler;
/**
 * Validates GitHub webhook signature
 */
export declare function validateWebhookSignature(payload: string | Buffer, signature: string | undefined, secret: string): boolean;
/**
 * Middleware to validate GitHub webhook signatures
 */
export declare function webhookSignatureMiddleware(secret: string): RequestHandler;
export declare const validators: {
    pullRequest: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    installation: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    push: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    ping: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    jobQuery: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    jobIdParam: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
};
export type PullRequestPayload = z.infer<typeof PullRequestPayloadSchema>;
export type InstallationPayload = z.infer<typeof InstallationPayloadSchema>;
export type PushPayload = z.infer<typeof PushPayloadSchema>;
export type PingPayload = z.infer<typeof PingPayloadSchema>;
export type JobQuery = z.infer<typeof JobQuerySchema>;
export type JobIdParam = z.infer<typeof JobIdParamSchema>;
export declare const validation: {
    validateRequest: typeof validateRequest;
    validateBody: typeof validateBody;
    validateQuery: typeof validateQuery;
    validateParams: typeof validateParams;
    validateWebhookSignature: typeof validateWebhookSignature;
    webhookSignatureMiddleware: typeof webhookSignatureMiddleware;
    sanitizeString: typeof sanitizeString;
    sanitizeObject: typeof sanitizeObject;
    validators: {
        pullRequest: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
        installation: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
        push: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
        ping: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
        jobQuery: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
        jobIdParam: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    };
    schemas: {
        GitHubUserSchema: z.ZodObject<{
            id: z.ZodNumber;
            login: z.ZodString;
            node_id: z.ZodOptional<z.ZodString>;
            avatar_url: z.ZodOptional<z.ZodString>;
            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }, {
            id: number;
            login: string;
            type?: "Bot" | "User" | "Organization" | undefined;
            node_id?: string | undefined;
            avatar_url?: string | undefined;
        }>;
        GitHubRepositorySchema: z.ZodObject<{
            id: z.ZodNumber;
            node_id: z.ZodOptional<z.ZodString>;
            name: z.ZodString;
            full_name: z.ZodString;
            private: z.ZodBoolean;
            owner: z.ZodObject<{
                id: z.ZodNumber;
                login: z.ZodString;
                node_id: z.ZodOptional<z.ZodString>;
                avatar_url: z.ZodOptional<z.ZodString>;
                type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }>;
            default_branch: z.ZodOptional<z.ZodString>;
            language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
        }, "strip", z.ZodTypeAny, {
            owner: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            };
            id: number;
            name: string;
            private: boolean;
            full_name: string;
            node_id?: string | undefined;
            default_branch?: string | undefined;
            language?: string | null | undefined;
        }, {
            owner: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            };
            id: number;
            name: string;
            private: boolean;
            full_name: string;
            node_id?: string | undefined;
            default_branch?: string | undefined;
            language?: string | null | undefined;
        }>;
        GitHubInstallationSchema: z.ZodObject<{
            id: z.ZodNumber;
            account: z.ZodOptional<z.ZodObject<{
                id: z.ZodNumber;
                login: z.ZodString;
                node_id: z.ZodOptional<z.ZodString>;
                avatar_url: z.ZodOptional<z.ZodString>;
                type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }>>;
            repository_selection: z.ZodOptional<z.ZodEnum<["all", "selected"]>>;
            permissions: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
        }, "strip", z.ZodTypeAny, {
            id: number;
            account?: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            } | undefined;
            repository_selection?: "all" | "selected" | undefined;
            permissions?: Record<string, string> | undefined;
        }, {
            id: number;
            account?: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            } | undefined;
            repository_selection?: "all" | "selected" | undefined;
            permissions?: Record<string, string> | undefined;
        }>;
        PullRequestPayloadSchema: z.ZodObject<{
            action: z.ZodEnum<["opened", "closed", "reopened", "synchronize", "edited", "ready_for_review", "review_requested", "labeled", "unlabeled", "assigned", "unassigned"]>;
            number: z.ZodNumber;
            pull_request: z.ZodObject<{
                id: z.ZodNumber;
                number: z.ZodNumber;
                state: z.ZodEnum<["open", "closed"]>;
                title: z.ZodString;
                body: z.ZodOptional<z.ZodNullable<z.ZodString>>;
                draft: z.ZodOptional<z.ZodBoolean>;
                head: z.ZodObject<{
                    ref: z.ZodString;
                    sha: z.ZodString;
                    repo: z.ZodOptional<z.ZodNullable<z.ZodObject<{
                        id: z.ZodNumber;
                        node_id: z.ZodOptional<z.ZodString>;
                        name: z.ZodString;
                        full_name: z.ZodString;
                        private: z.ZodBoolean;
                        owner: z.ZodObject<{
                            id: z.ZodNumber;
                            login: z.ZodString;
                            node_id: z.ZodOptional<z.ZodString>;
                            avatar_url: z.ZodOptional<z.ZodString>;
                            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                        }, "strip", z.ZodTypeAny, {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        }, {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        }>;
                        default_branch: z.ZodOptional<z.ZodString>;
                        language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
                    }, "strip", z.ZodTypeAny, {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    }, {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    }>>>;
                }, "strip", z.ZodTypeAny, {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                }, {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                }>;
                base: z.ZodObject<{
                    ref: z.ZodString;
                    sha: z.ZodString;
                    repo: z.ZodOptional<z.ZodNullable<z.ZodObject<{
                        id: z.ZodNumber;
                        node_id: z.ZodOptional<z.ZodString>;
                        name: z.ZodString;
                        full_name: z.ZodString;
                        private: z.ZodBoolean;
                        owner: z.ZodObject<{
                            id: z.ZodNumber;
                            login: z.ZodString;
                            node_id: z.ZodOptional<z.ZodString>;
                            avatar_url: z.ZodOptional<z.ZodString>;
                            type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                        }, "strip", z.ZodTypeAny, {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        }, {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        }>;
                        default_branch: z.ZodOptional<z.ZodString>;
                        language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
                    }, "strip", z.ZodTypeAny, {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    }, {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    }>>>;
                }, "strip", z.ZodTypeAny, {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                }, {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                }>;
                user: z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>;
                merged: z.ZodOptional<z.ZodBoolean>;
                mergeable: z.ZodOptional<z.ZodNullable<z.ZodBoolean>>;
                changed_files: z.ZodOptional<z.ZodNumber>;
                additions: z.ZodOptional<z.ZodNumber>;
                deletions: z.ZodOptional<z.ZodNumber>;
            }, "strip", z.ZodTypeAny, {
                number: number;
                base: {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                };
                user: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                title: string;
                state: "open" | "closed";
                head: {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                };
                additions?: number | undefined;
                deletions?: number | undefined;
                body?: string | null | undefined;
                draft?: boolean | undefined;
                merged?: boolean | undefined;
                mergeable?: boolean | null | undefined;
                changed_files?: number | undefined;
            }, {
                number: number;
                base: {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                };
                user: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                title: string;
                state: "open" | "closed";
                head: {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                };
                additions?: number | undefined;
                deletions?: number | undefined;
                body?: string | null | undefined;
                draft?: boolean | undefined;
                merged?: boolean | undefined;
                mergeable?: boolean | null | undefined;
                changed_files?: number | undefined;
            }>;
            repository: z.ZodObject<{
                id: z.ZodNumber;
                node_id: z.ZodOptional<z.ZodString>;
                name: z.ZodString;
                full_name: z.ZodString;
                private: z.ZodBoolean;
                owner: z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>;
                default_branch: z.ZodOptional<z.ZodString>;
                language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
            }, "strip", z.ZodTypeAny, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }>;
            installation: z.ZodOptional<z.ZodObject<{
                id: z.ZodNumber;
                account: z.ZodOptional<z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>>;
                repository_selection: z.ZodOptional<z.ZodEnum<["all", "selected"]>>;
                permissions: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            }, {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            }>>;
            sender: z.ZodObject<{
                id: z.ZodNumber;
                login: z.ZodString;
                node_id: z.ZodOptional<z.ZodString>;
                avatar_url: z.ZodOptional<z.ZodString>;
                type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }>;
        }, "strip", z.ZodTypeAny, {
            number: number;
            action: "opened" | "synchronize" | "reopened" | "edited" | "closed" | "ready_for_review" | "review_requested" | "labeled" | "unlabeled" | "assigned" | "unassigned";
            repository: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            };
            pull_request: {
                number: number;
                base: {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                };
                user: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                title: string;
                state: "open" | "closed";
                head: {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                };
                additions?: number | undefined;
                deletions?: number | undefined;
                body?: string | null | undefined;
                draft?: boolean | undefined;
                merged?: boolean | undefined;
                mergeable?: boolean | null | undefined;
                changed_files?: number | undefined;
            };
            sender: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            };
            installation?: {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            } | undefined;
        }, {
            number: number;
            action: "opened" | "synchronize" | "reopened" | "edited" | "closed" | "ready_for_review" | "review_requested" | "labeled" | "unlabeled" | "assigned" | "unassigned";
            repository: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            };
            pull_request: {
                number: number;
                base: {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                };
                user: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                title: string;
                state: "open" | "closed";
                head: {
                    ref: string;
                    sha: string;
                    repo?: {
                        owner: {
                            id: number;
                            login: string;
                            type?: "Bot" | "User" | "Organization" | undefined;
                            node_id?: string | undefined;
                            avatar_url?: string | undefined;
                        };
                        id: number;
                        name: string;
                        private: boolean;
                        full_name: string;
                        node_id?: string | undefined;
                        default_branch?: string | undefined;
                        language?: string | null | undefined;
                    } | null | undefined;
                };
                additions?: number | undefined;
                deletions?: number | undefined;
                body?: string | null | undefined;
                draft?: boolean | undefined;
                merged?: boolean | undefined;
                mergeable?: boolean | null | undefined;
                changed_files?: number | undefined;
            };
            sender: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            };
            installation?: {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            } | undefined;
        }>;
        InstallationPayloadSchema: z.ZodObject<{
            action: z.ZodEnum<["created", "deleted", "suspend", "unsuspend", "new_permissions_accepted"]>;
            installation: z.ZodObject<{
                id: z.ZodNumber;
                account: z.ZodOptional<z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>>;
                repository_selection: z.ZodOptional<z.ZodEnum<["all", "selected"]>>;
                permissions: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            }, {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            }>;
            repositories: z.ZodOptional<z.ZodArray<z.ZodObject<{
                id: z.ZodNumber;
                name: z.ZodString;
                full_name: z.ZodString;
                private: z.ZodBoolean;
            }, "strip", z.ZodTypeAny, {
                id: number;
                name: string;
                private: boolean;
                full_name: string;
            }, {
                id: number;
                name: string;
                private: boolean;
                full_name: string;
            }>, "many">>;
            sender: z.ZodObject<{
                id: z.ZodNumber;
                login: z.ZodString;
                node_id: z.ZodOptional<z.ZodString>;
                avatar_url: z.ZodOptional<z.ZodString>;
                type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }>;
        }, "strip", z.ZodTypeAny, {
            action: "deleted" | "created" | "suspend" | "unsuspend" | "new_permissions_accepted";
            installation: {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            };
            sender: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            };
            repositories?: {
                id: number;
                name: string;
                private: boolean;
                full_name: string;
            }[] | undefined;
        }, {
            action: "deleted" | "created" | "suspend" | "unsuspend" | "new_permissions_accepted";
            installation: {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            };
            sender: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            };
            repositories?: {
                id: number;
                name: string;
                private: boolean;
                full_name: string;
            }[] | undefined;
        }>;
        PushPayloadSchema: z.ZodObject<{
            ref: z.ZodString;
            before: z.ZodString;
            after: z.ZodString;
            created: z.ZodOptional<z.ZodBoolean>;
            deleted: z.ZodOptional<z.ZodBoolean>;
            forced: z.ZodOptional<z.ZodBoolean>;
            commits: z.ZodOptional<z.ZodArray<z.ZodObject<{
                id: z.ZodString;
                message: z.ZodString;
                author: z.ZodObject<{
                    name: z.ZodString;
                    email: z.ZodString;
                }, "strip", z.ZodTypeAny, {
                    name: string;
                    email: string;
                }, {
                    name: string;
                    email: string;
                }>;
                added: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
                removed: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
                modified: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
            }, "strip", z.ZodTypeAny, {
                message: string;
                id: string;
                author: {
                    name: string;
                    email: string;
                };
                added?: string[] | undefined;
                modified?: string[] | undefined;
                removed?: string[] | undefined;
            }, {
                message: string;
                id: string;
                author: {
                    name: string;
                    email: string;
                };
                added?: string[] | undefined;
                modified?: string[] | undefined;
                removed?: string[] | undefined;
            }>, "many">>;
            repository: z.ZodObject<{
                id: z.ZodNumber;
                node_id: z.ZodOptional<z.ZodString>;
                name: z.ZodString;
                full_name: z.ZodString;
                private: z.ZodBoolean;
                owner: z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>;
                default_branch: z.ZodOptional<z.ZodString>;
                language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
            }, "strip", z.ZodTypeAny, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }>;
            installation: z.ZodOptional<z.ZodObject<{
                id: z.ZodNumber;
                account: z.ZodOptional<z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>>;
                repository_selection: z.ZodOptional<z.ZodEnum<["all", "selected"]>>;
                permissions: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            }, {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            }>>;
            sender: z.ZodObject<{
                id: z.ZodNumber;
                login: z.ZodString;
                node_id: z.ZodOptional<z.ZodString>;
                avatar_url: z.ZodOptional<z.ZodString>;
                type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }>;
        }, "strip", z.ZodTypeAny, {
            ref: string;
            repository: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            };
            sender: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            };
            before: string;
            after: string;
            deleted?: boolean | undefined;
            installation?: {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            } | undefined;
            created?: boolean | undefined;
            forced?: boolean | undefined;
            commits?: {
                message: string;
                id: string;
                author: {
                    name: string;
                    email: string;
                };
                added?: string[] | undefined;
                modified?: string[] | undefined;
                removed?: string[] | undefined;
            }[] | undefined;
        }, {
            ref: string;
            repository: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            };
            sender: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            };
            before: string;
            after: string;
            deleted?: boolean | undefined;
            installation?: {
                id: number;
                account?: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                } | undefined;
                repository_selection?: "all" | "selected" | undefined;
                permissions?: Record<string, string> | undefined;
            } | undefined;
            created?: boolean | undefined;
            forced?: boolean | undefined;
            commits?: {
                message: string;
                id: string;
                author: {
                    name: string;
                    email: string;
                };
                added?: string[] | undefined;
                modified?: string[] | undefined;
                removed?: string[] | undefined;
            }[] | undefined;
        }>;
        PingPayloadSchema: z.ZodObject<{
            zen: z.ZodString;
            hook_id: z.ZodNumber;
            hook: z.ZodOptional<z.ZodObject<{
                id: z.ZodNumber;
                type: z.ZodString;
                events: z.ZodArray<z.ZodString, "many">;
                active: z.ZodBoolean;
            }, "strip", z.ZodTypeAny, {
                type: string;
                id: number;
                events: string[];
                active: boolean;
            }, {
                type: string;
                id: number;
                events: string[];
                active: boolean;
            }>>;
            repository: z.ZodOptional<z.ZodObject<{
                id: z.ZodNumber;
                node_id: z.ZodOptional<z.ZodString>;
                name: z.ZodString;
                full_name: z.ZodString;
                private: z.ZodBoolean;
                owner: z.ZodObject<{
                    id: z.ZodNumber;
                    login: z.ZodString;
                    node_id: z.ZodOptional<z.ZodString>;
                    avatar_url: z.ZodOptional<z.ZodString>;
                    type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
                }, "strip", z.ZodTypeAny, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }, {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                }>;
                default_branch: z.ZodOptional<z.ZodString>;
                language: z.ZodOptional<z.ZodNullable<z.ZodString>>;
            }, "strip", z.ZodTypeAny, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }, {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            }>>;
            sender: z.ZodOptional<z.ZodObject<{
                id: z.ZodNumber;
                login: z.ZodString;
                node_id: z.ZodOptional<z.ZodString>;
                avatar_url: z.ZodOptional<z.ZodString>;
                type: z.ZodOptional<z.ZodEnum<["User", "Bot", "Organization"]>>;
            }, "strip", z.ZodTypeAny, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }, {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            zen: string;
            hook_id: number;
            repository?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | undefined;
            sender?: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            } | undefined;
            hook?: {
                type: string;
                id: number;
                events: string[];
                active: boolean;
            } | undefined;
        }, {
            zen: string;
            hook_id: number;
            repository?: {
                owner: {
                    id: number;
                    login: string;
                    type?: "Bot" | "User" | "Organization" | undefined;
                    node_id?: string | undefined;
                    avatar_url?: string | undefined;
                };
                id: number;
                name: string;
                private: boolean;
                full_name: string;
                node_id?: string | undefined;
                default_branch?: string | undefined;
                language?: string | null | undefined;
            } | undefined;
            sender?: {
                id: number;
                login: string;
                type?: "Bot" | "User" | "Organization" | undefined;
                node_id?: string | undefined;
                avatar_url?: string | undefined;
            } | undefined;
            hook?: {
                type: string;
                id: number;
                events: string[];
                active: boolean;
            } | undefined;
        }>;
        JobQuerySchema: z.ZodObject<{
            status: z.ZodOptional<z.ZodEnum<["pending", "processing", "completed", "failed"]>>;
            limit: z.ZodDefault<z.ZodOptional<z.ZodPipeline<z.ZodEffects<z.ZodString, number, string>, z.ZodNumber>>>;
            offset: z.ZodDefault<z.ZodOptional<z.ZodPipeline<z.ZodEffects<z.ZodString, number, string>, z.ZodNumber>>>;
        }, "strip", z.ZodTypeAny, {
            limit: number;
            offset: number;
            status?: "pending" | "processing" | "completed" | "failed" | undefined;
        }, {
            status?: "pending" | "processing" | "completed" | "failed" | undefined;
            limit?: string | undefined;
            offset?: string | undefined;
        }>;
        JobIdParamSchema: z.ZodObject<{
            id: z.ZodString;
        }, "strip", z.ZodTypeAny, {
            id: string;
        }, {
            id: string;
        }>;
    };
};
export default validation;
//# sourceMappingURL=validation.d.ts.map