import { createLogger, format } from "winston";
import { File, Console } from "winston/lib/winston/transports";

export const logger = createLogger({
    format: format.combine(
        format.timestamp(),
        format.errors({stack: true}),
        format.printf(info => `e2e_ts="${info.timestamp}" e2e_level="${info.level}" e2e_origin="${info.origin}" ${info.message}`),
    ),
    transports: [
/*        new File({
            dirname: "./log",
            filename: "e2e.log"
        }),*/
        new Console()
    ]
});
