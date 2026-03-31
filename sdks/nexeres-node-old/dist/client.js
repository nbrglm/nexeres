export class NexeresClient {
    config;
    constructor(config) {
        this.config = config;
    }
    /** Make a request to the Nexeres Admin API
     * @param endpoint The API endpoint to call, e.g. "/api/admin/login"
     * @param adminToken The admin's ephemeral token for authentication
     * @param options Fetch options, e.g. method, headers, body, etc.
     * @returns A promise that resolves to the response data or an error object.
    */
    async adminRequest(endpoint, adminToken, options = {}) {
        const res = await fetch(`${this.config.baseUrl}${endpoint}`, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                'X-NEXERES-API-Key': this.config.apiKey,
                'X-NEXERES-Admin-Token': adminToken,
                ...(options.headers || {}),
            }
        });
        const adminTokenExpiry = res.headers.get('X-NEXERES-Admin-Token-Expiry') || (new Date()).toString();
        if (!res.ok) {
            const error = await res.json();
            return { error, code: res.status, adminTokenExpiry };
        }
        return { result: await res.json(), code: res.status, adminTokenExpiry };
    }
    adminGet(endpoint, adminToken, options = {}) {
        return this.adminRequest(endpoint, adminToken, { method: 'GET', ...options });
    }
    adminPost(endpoint, adminToken, body, options = {}) {
        return this.adminRequest(endpoint, adminToken, { method: 'POST', body: JSON.stringify(body), ...options });
    }
    adminPut(endpoint, adminToken, body, options = {}) {
        return this.adminRequest(endpoint, adminToken, { method: 'PUT', body: JSON.stringify(body), ...options });
    }
    adminDelete(endpoint, adminToken, options = {}) {
        return this.adminRequest(endpoint, adminToken, { method: 'DELETE', ...options });
    }
    /** Make a request to the Nexeres API.
     * @param endpoint The API endpoint to call, e.g. "/auth/login"
     * @param options Fetch options, e.g. method, headers, body, etc.
     * @returns A promise that resolves to the response data or an error object.
     */
    async request(endpoint, options = {}) {
        const res = await fetch(`${this.config.baseUrl}${endpoint}`, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                'X-NEXERES-API-Key': this.config.apiKey,
                ...(options.headers || {}),
            }
        });
        if (!res.ok) {
            const error = await res.json();
            return { error, code: res.status };
        }
        return { result: await res.json(), code: res.status };
    }
    get(endpoint, options = {}) {
        return this.request(endpoint, { method: 'GET', ...options });
    }
    post(endpoint, body, options = {}) {
        return this.request(endpoint, { method: 'POST', body: JSON.stringify(body), ...options });
    }
    put(endpoint, body, options = {}) {
        return this.request(endpoint, { method: 'PUT', body: JSON.stringify(body), ...options });
    }
    delete(endpoint, options = {}) {
        return this.request(endpoint, { method: 'DELETE', ...options });
    }
}
//# sourceMappingURL=client.js.map