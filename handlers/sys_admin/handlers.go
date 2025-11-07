package sys_admin

import "github.com/nbrglm/nexeres/internal/interfaces"

func GetSysAdminHandlers() []interfaces.Handler {
	return []interfaces.Handler{
		NewSysAdminCreateOrgHandler(),
		NewSysAdminDeleteOrgHandler(),
		NewSysAdminGetConfigHandler(),
		NewSysAdminGetOrgDetailsHandler(),
		NewSysAdminListOrgsHandler(),
	}
}
