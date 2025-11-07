;
export function NewNexeresOrg(data) {
    data.createdAt = data.createdAt instanceof Date ? data.createdAt : new Date(data.createdAt);
    data.updatedAt = data.updatedAt instanceof Date ? data.updatedAt : new Date(data.updatedAt);
    data.deletedAt = data.deletedAt ? (data.deletedAt instanceof Date ? data.deletedAt : new Date(data.deletedAt)) : undefined;
    return data;
}
;
;
;
export function NewNexeresOrgRole(data) {
    data.createdAt = data.createdAt instanceof Date ? data.createdAt : new Date(data.createdAt);
    data.updatedAt = data.updatedAt instanceof Date ? data.updatedAt : new Date(data.updatedAt);
    return data;
}
;
export function NewNexeresOrgDomain(data) {
    data.createdAt = data.createdAt instanceof Date ? data.createdAt : new Date(data.createdAt);
    data.updatedAt = data.updatedAt instanceof Date ? data.updatedAt : new Date(data.updatedAt);
    data.verifiedAt = data.verifiedAt ? (data.verifiedAt instanceof Date ? data.verifiedAt : new Date(data.verifiedAt)) : undefined;
    return data;
}
//# sourceMappingURL=org.js.map