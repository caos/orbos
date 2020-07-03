package scripts

func GetAll() map[string]string {
	return map[string]string{
		"V1.0__eventstore.sql":                V10Eventstore,
		"V1.1__management.sql":                V11Management,
		"V1.2__management_project_view.sql":   V12ManagementProjectView,
		"V1.3__management_user_view.sql":      V13ManagementUserView,
		"V1.4__admin_user_grants.sql":         V14AdminUserGrants,
		"V1.5__auth.sql":                      V15Auth,
		"V1.6__management_user_view.sql":      V16ManagementUserView,
		"V1.7__notification.sql":              V17Notification,
		"V1.8__admin.sql":                     V18Admin,
		"V1.9__admin_grant.sql":               V19AdminGrant,
		"V1.10__mgmt_orgs.sql":                V110MgmtOrgs,
		"V1.11__auth_oidc.sql":                V111AuthOidc,
		"V1.12__auth_user_grant_view.sql":     V112AuthUserGrantView,
		"V1.13__auth_org_view.sql":            V113AuthOrgView,
		"V1.14__authz.sql":                    V114Authz,
		"V1.15__management_project_view.sql":  V115ManagementProjectView,
		"V1.16__login_names.sql":              V116LoginNames,
		"V1.17__org_domains.sql":              V117OrgDomains,
		"V1.18__user_view.sql":                V118UserView,
		"V1.19__usersession_names.sql":        V119UsersessionNames,
		"V1.20__notification_passwordset.sql": V120NotificationPasswordset,
		"V1.21__project_grant_view.sql":       V121ProjectGrantView,
		"V1.22__admin_view.sql":               V122AdminView,
		"V1.23__admin_iam_members.sql":        V123AdminIamMembers,
		"V1.24__failed_events.sql":            V124FailedEvents,
	}
}
