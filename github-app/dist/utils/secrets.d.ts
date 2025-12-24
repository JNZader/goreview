/**
 * Secure secret handling utilities
 * Provides encryption, masking, and secure comparison functions
 */
export interface EncryptedData {
    encrypted: string;
    iv: string;
    authTag: string;
    salt: string;
}
export interface SecretMaskOptions {
    showFirst?: number;
    showLast?: number;
    maskChar?: string;
    minLength?: number;
}
/**
 * Encrypts a secret using AES-256-GCM
 * @param plaintext - The secret to encrypt
 * @param masterPassword - Optional master password (defaults to env var)
 * @returns Encrypted data with IV, auth tag, and salt
 */
export declare function encryptSecret(plaintext: string, masterPassword?: string): EncryptedData;
/**
 * Decrypts a secret encrypted with encryptSecret
 * @param encryptedData - The encrypted data object
 * @param masterPassword - Optional master password (defaults to env var)
 * @returns Decrypted plaintext
 */
export declare function decryptSecret(encryptedData: EncryptedData, masterPassword?: string): string;
/**
 * Masks a secret string, showing only first/last characters
 * @param secret - The secret to mask
 * @param options - Masking options
 * @returns Masked string
 */
export declare function maskSecret(secret: string, options?: SecretMaskOptions): string;
/**
 * Scans text for sensitive data and masks it
 * @param text - The text to scan and mask
 * @returns Text with sensitive data masked
 */
export declare function maskSecrets(text: string): string;
/**
 * Safely stringifies an object, masking sensitive values
 * @param obj - Object to stringify
 * @param sensitiveKeys - Keys to mask (default: common sensitive keys)
 * @returns JSON string with masked values
 */
export declare function safeStringify(obj: unknown, sensitiveKeys?: string[]): string;
/**
 * Performs timing-safe string comparison to prevent timing attacks
 * @param a - First string
 * @param b - Second string
 * @returns True if strings are equal
 */
export declare function secureCompare(a: string, b: string): boolean;
/**
 * Generates a cryptographically secure random token
 * @param length - Token length in bytes (output will be hex, so 2x chars)
 * @returns Hex-encoded random token
 */
export declare function generateSecureToken(length?: number): string;
/**
 * Generates a URL-safe random token
 * @param length - Desired output length
 * @returns Base64url-encoded random token
 */
export declare function generateUrlSafeToken(length?: number): string;
/**
 * Creates a SHA-256 hash of a value
 * @param value - Value to hash
 * @returns Hex-encoded hash
 */
export declare function hashValue(value: string): string;
/**
 * Creates a HMAC-SHA256 signature
 * @param value - Value to sign
 * @param secret - Secret key
 * @returns Hex-encoded HMAC
 */
export declare function createHmac(value: string, secret: string): string;
/**
 * Verifies a HMAC-SHA256 signature using timing-safe comparison
 * @param value - Original value
 * @param signature - Signature to verify
 * @param secret - Secret key
 * @returns True if signature is valid
 */
export declare function verifyHmac(value: string, signature: string, secret: string): boolean;
export declare const secrets: {
    encryptSecret: typeof encryptSecret;
    decryptSecret: typeof decryptSecret;
    maskSecret: typeof maskSecret;
    maskSecrets: typeof maskSecrets;
    safeStringify: typeof safeStringify;
    secureCompare: typeof secureCompare;
    generateSecureToken: typeof generateSecureToken;
    generateUrlSafeToken: typeof generateUrlSafeToken;
    hashValue: typeof hashValue;
    createHmac: typeof createHmac;
    verifyHmac: typeof verifyHmac;
};
export default secrets;
//# sourceMappingURL=secrets.d.ts.map