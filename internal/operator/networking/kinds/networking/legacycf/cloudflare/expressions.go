package cloudflare

import (
	"strings"

	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/cloudflare/expression"
)

func EmptyExpression() *expression.Expression {
	return expression.New("")
}

func IPExpressionIsIn(list []string) *expression.Expression {
	containStr := strings.Join(list, " ")
	containStr = strings.Join([]string{"{", containStr, "}"}, "")

	exp := strings.Join([]string{"ip.src", "in", containStr}, " ")

	return expression.New(exp)
}

func IPExpressionEquals(IP string) *expression.Expression {
	exp := strings.Join([]string{"ip.src", "eq", IP}, " ")

	return expression.New(exp)
}

func HostnameExpressionIsIn(hostnames []string) *expression.Expression {

	containStr := strings.Join(hostnames, " ")
	containStr = strings.Join([]string{"{", containStr, "}"}, "")

	exp := strings.Join([]string{"http.host", "in", containStr}, " ")

	return expression.New(exp)
}

func HostnameExpressionContains(hostname string) *expression.Expression {
	exp := strings.Join([]string{"http.host", "contains", hostname}, " ")

	return expression.New(exp)
}

func HostnameExpressionEquals(hostname string) *expression.Expression {
	exp := strings.Join([]string{"http.host", "equals", hostname}, " ")

	return expression.New(exp)
}

func SSLExpression() *expression.Expression {
	return expression.New("ssl")
}
func NotSSLExpression() *expression.Expression {
	return expression.New("not ssl")
}
