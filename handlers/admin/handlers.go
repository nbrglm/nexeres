package admin

import "github.com/nbrglm/nexeres/internal/interfaces"

func GetAdminHandlers() []interfaces.Handler {
	return []interfaces.Handler{
		NewAdminCreateDomainHandler(),
		NewAdminCreateRoleHandler(),
		NewAdminDeleteDomainHandler(),
		NewAdminDeleteRoleHandler(),
		NewAdminGetDomainVerifyCodeHandler(),
		NewAdminUpdateDomainHandler(),
		NewAdminUpdateRoleHandler(),
		NewAdminUpdateOrgHandler(),
		NewAdminVerifyDomainHandler(),
	}
}
