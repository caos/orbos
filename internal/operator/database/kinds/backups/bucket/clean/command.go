package clean

import (
	"strconv"
	"strings"
)

func getCommand(
	databases []string,
	dbURL string,
	dbPort int32,
) string {
	backupCommands := make([]string, 0)
	for _, database := range databases {
		backupCommands = append(backupCommands,
			strings.Join([]string{
				"cockroach",
				"sql",
				"--certs-dir=" + certPath,
				"--host=" + dbURL,
				"--port=" + strconv.Itoa(int(dbPort)),
				"-e",
				"\"DROP DATABASE IF EXISTS " + database + " CASCADE;\"",
			}, " "))
	}
	/*
		backupCommands = append(backupCommands,
			strings.Join([]string{
				"cockroach",
				"sql",
				"--certs-dir=" + certPath,
				"--host=" + dbURL,
				"--port=" + strconv.Itoa(int(dbPort)),
				"-e",
				"\"TRUNCATE defaultdb.flyway_schema_history;\"",
			}, " "))
	*/
	return strings.Join(backupCommands, " && ")
}
