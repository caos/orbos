package scripts

const V120NotificationPasswordset = `BEGIN;

ALTER TABLE notification.notify_users ADD COLUMN password_set BOOLEAN;

COMMIT;`
