package scripts

func GetAll() map[string]string {
	return map[string]string{
		"V1.0__eventstore.sql":                   V10Eventstore,
		"V1.1__databases.sql":                    V11Databases,
		"V1.2__views.sql":                        V12Views,
		"V1.3__notification_user_loginnames.sql": V13NotificationUserLoginnames,
		"V1.4__usergrant_grantid.sql":            V14UsergranGrantid,
	}
}
