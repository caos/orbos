package scripts

func GetAll() map[string]string {
	return map[string]string{
		"V1.0_eventstore.sql":               V10Eventstore,
		"V1.1_management.sql":               V11Management,
		"V1.2_management_project_view.sql":  V12ManagementProjectView,
		"V1.3_management_user_view.sql":     V13ManagementUserView,
		"V1.4_admin_user_grants.sql":        V14AdminUserGrants,
		"V1.5_auth.sql":                     V15Auth,
		"V1.6_management_user_view.sql":     V16ManagementUserView,
		"V1.7_notification.sql":             V17Notification,
		"V1.8_admin.sql":                    V18Admin,
		"V1.9_admin_grant.sql":              V19AdminGrant,
		"V1.10_mgmt_orgs.sql":               V110MgmtOrgs,
		"V1.11_auth_oidc.sql":               V111AuthOidc,
		"V1.12_auth_user_grant_view.sql":    V112AuthUserGrantView,
		"V1.13_auth_org_view.sql":           V113AuthOrgView,
		"V1.14_authz":                       V114Authz,
		"V1.15_management_project_view.sql": V115ManagementProjectView,
		"V1.16_login_names.sql":             V116LoginNames,
		"V1.17_org_domains":                 V117OrgDomains,
		"V1.18_user_view":                   V118UserView,
		"V1.19_usersession_names":           V119UsersessionNames,
		"V1.20_notification_passwordset":    V120NotificationPasswordset,
		"V1.21_project_grant_view":          V121ProjectGrantView,
		"V1.22_admin_view":                  V122AdminView,
		"V1.23_admin_iam_members":           V123AdminIamMembers,
		"V1.24_failed_events":               V124FailedEvents,
	}
}
