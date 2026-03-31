import { NexeresConfig } from "./config.js";
export declare class NexeresClient {
    private config;
    constructor(config: NexeresConfig);
    /** Make a request to the Nexeres Admin API
     * @param endpoint The API endpoint to call, e.g. "/api/admin/login"
     * @param adminToken The admin's ephemeral token for authentication
     * @param options Fetch options, e.g. method, headers, body, etc.
     * @returns A promise that resolves to the response data or an error object.
    */
    private adminRequest;
    adminGet<T>(endpoint: string, adminToken: string, options?: RequestInit): NexeresAdminResponse<T>;
    adminPost<T>(endpoint: string, adminToken: string, body: any, options?: RequestInit): NexeresAdminResponse<T>;
    adminPut<T>(endpoint: string, adminToken: string, body: any, options?: RequestInit): NexeresAdminResponse<T>;
    adminDelete<T>(endpoint: string, adminToken: string, options?: RequestInit): NexeresAdminResponse<T>;
    /** Make a request to the Nexeres API.
     * @param endpoint The API endpoint to call, e.g. "/auth/login"
     * @param options Fetch options, e.g. method, headers, body, etc.
     * @returns A promise that resolves to the response data or an error object.
     */
    private request;
    get<T>(endpoint: string, options?: RequestInit): NexeresResponse<T>;
    post<T>(endpoint: string, body: any, options?: RequestInit): NexeresResponse<T>;
    put<T>(endpoint: string, body: any, options?: RequestInit): NexeresResponse<T>;
    delete<T>(endpoint: string, options?: RequestInit): NexeresResponse<T>;
}
/** Response for general API calls */
export type NexeresResponse<T> = Promise<{
    code: number;
    result?: T;
    error?: NexeresError;
}>;
/** Response for admin-related API calls */
export type NexeresAdminResponse<T> = Promise<{
    /** HTTP Status Code returned by the server: Non 200 code represents error, while a 200 code represents a successfull result */
    code: number;
    result?: T;
    error?: NexeresError;
    /** The expiry, in seconds from now, of the admin token. */
    adminTokenExpiry: string;
}>;
/** Error response from the Nexeres API */
export type NexeresError = {
    /** Error message
     *
     * You can safely display this message to end users.
    */
    message: string;
    /** Debug information
     *
     * This is meant for debugging purposes only and should not be displayed to end users.
     *
     * This won't be present unless the `debug` flag is enabled in your Nexeres backend config.
    */
    debug?: string | undefined;
};
//# sourceMappingURL=client.d.ts.map