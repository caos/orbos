import { createLogger } from "winston";
import { File } from "winston/lib/winston/transports";

export const logger = createLogger({
    transports: new File({
        filename: "e2e-tests.log"
    })
});
