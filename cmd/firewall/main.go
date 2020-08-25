package main

import (
	"context"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func main() {
	key := []byte(
		`{
	"type": "service_account",
	"project_id": "caos-240809",
	"private_key_id": "1cad1be6a700c21d02328903b5699beb0f0dae60",
	"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCxgLn0ts+bzBjj\nLjvX95ukXi8sbe0/dYFSenaaaGk4smRxOXfDT41yuAC9sZ0wt+mogbcD8I4rNPON\nD+RHO2LpEertxzUNRd1awlodahpdTs7dHY2Inn8FuoTIwsKqkp0OBqDU/yv+58IK\n0BRnvvJyIRAHmI4l5CGUxex2s6G4P8JHAh6A2+4BkKVZrdPBRDSb/fM04dRcOw17\nPdWrYxdUsWiHE4RB0ykQSNwhSiswK1Fwppw2FRV4XtpV/E4zFZAgORZ1fd4LC+jm\nSGCa6LMMb1ePBMdq22Sw6qJkllob7Dqjj6dKPACQNxRsioWjxkZvXnBuoJKzQyVf\nV7ezaYWxAgMBAAECggEAATgiHt6VTYuw4ixpp0WmK6v+ppwaZV8SSTdTr2Lj+FdO\nXHikwrvq4j5S3/qzbeSEUTH56G4+tcIee0OzWyaa56dMOsxzyBBvEUrOoAJzP0L7\n/zwVaK2CoxupW/0tfxea4IZ2YWj5NWr3YK5i+eVcTweJdgayREU2+PAFAq0Gp+Ho\ntFaLRqDdqTibaEc5AuYgkYQjWwVGVrU2x6JlNHzwmoh/iG3CYG4YQTVswKIN/qLm\neXjEizyydkGoBgB5k0cIlXpvEXv+cMqlprEeirhnnGO4c196Ph/JjEnygpjq0apw\nCZ/vhu23tzm5QBYMzgv2YBEAWHt77MFt4b+aPfopWQKBgQDhDC/nDQYAAwkViyL4\nM6V/rpnkp46P2xgXy3PPeUYrPHbXcf2rPXnPDqMvVkmU3bMFswXJSt+1ostNlB1q\neKkMHr6aMXCJU0uwKk7XeM1HP69VoAXcADKW8evS2DfgpQEwwP/3LNhs4OaOhXY4\n/JqjybxfBcPEfxZIPZiSmaTn2QKBgQDJ6oMM1RF9Sq2W5STjIfMAeoQo9jLkEXok\nuwx3AB2jRyDsw8UxflgxyROMqKaUbBdcO9kyyxzAZ/93SRKF+2OfrdE81bWZiOnz\npb4P8RqmXJOg7C3XCwye5sMZ55igLAj2fIPNQDyyjZzhZU8iGrMfRLPnMf9j/yxR\nBliFHhZ9mQKBgA51zQIomREZINVMimOuVdz9aBAEICnoJwUoYnmbTkHq8avoPCdr\nnM8MHrok7jdtg1pDZYTIldVC75M9iCJWPG5170NTF+sK+hsIrOY1ceM5GVgEHzxC\nmv2N79wtXnHFyGzMieXk8McWMFpKAw2oVXtetAbbBPg0PkdIBeytiKYJAoGAOCNQ\nbkfrBee2Xaa128R7mF130x+oRIqrZ/ztWUSZ+OR0vf8sGzeic60RF2FodwmacRVe\nrOWVx9TiTRru4HtlVmbwLrbIN7i+OvSQ5EPHggtpLCueDxTOXHuSMOiYIag8kbNK\nvc0nUwlWXcBaAQRlWsMyNYxMElRG0PwvrksQO7kCgYAiVnjoy/9v83P/pZxuQyiG\nOfGzqMQv9ONbplm3NBjjhFbB11OGn1azB2AJehJeFyPDIQlHYoW8c6cftiPDyXhF\nXUDoj7QL+uVMOyV7nWOc0wnrHwMbKwo5P/ClQ7qLcaP5ukCmcWC85SwwET/cpbcf\n/Ajwk5ubueZQfSke87KfiQ==\n-----END PRIVATE KEY-----\n",
	"client_email": "orbos-benz@caos-240809.iam.gserviceaccount.com",
	"client_id": "113549904366732153355",
	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	"token_uri": "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/orbos-benz%40caos-240809.iam.gserviceaccount.com"
}`)

	svc, err := compute.NewService(context.Background(), option.WithCredentialsJSON(key))
	if err != nil {
		panic(err.Error())
	}

	gceFirewalls, err := svc.Firewalls.
		List("caos-240809").
		//Filter(fmt.Sprintf(`network : "*"`)).
		Fields("items(network,name,allowed,targetTags,sourceRanges)").
		Do()
	if err != nil {
		panic(err)
	}

	if len(gceFirewalls.Items) == 0 {
		panic("No firewalls listed")
	}
}
