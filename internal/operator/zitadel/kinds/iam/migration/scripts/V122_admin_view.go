package scripts

const V122AdminView = `BEGIN;

GRANT SELECT, INSERT, UPDATE, DROP, DELETE ON DATABASE auth TO admin_api;
GRANT SELECT, INSERT, UPDATE, DROP, DELETE ON TABLE auth.* TO admin_api;

GRANT SELECT, INSERT, UPDATE, DROP, DELETE ON DATABASE authz TO admin_api;
GRANT SELECT, INSERT, UPDATE, DROP, DELETE ON TABLE authz.* TO admin_api;

GRANT SELECT, INSERT, UPDATE, DROP, DELETE ON DATABASE management TO admin_api;
GRANT SELECT, INSERT, UPDATE, DROP, DELETE ON TABLE management.* TO admin_api;

GRANT SELECT, INSERT, UPDATE, DROP, DELETE ON DATABASE notification TO admin_api;
GRANT SELECT, INSERT, UPDATE, DROP, DELETE ON TABLE notification.* TO admin_api;

COMMIT;`