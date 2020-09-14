package expression

import "strings"

type Expression struct {
	exp           string
	additionalExp []*AdditionalExpression
}

type AdditionalExpression struct {
	operator string
	exp      string
}

func New(exp string) *Expression {
	return &Expression{
		exp:           exp,
		additionalExp: []*AdditionalExpression{},
	}
}

func (e *Expression) And(expression *Expression) {
	if e.exp == "" {
		e.exp = expression.ToString()
		return
	}

	if expression.exp == "" {
		return
	}

	e.additionalExp = append(e.additionalExp, &AdditionalExpression{
		operator: "and",
		exp:      expression.ToString(),
	})
}

func (e *Expression) Or(expression *Expression) {
	if e.exp == "" {
		e.exp = AddBraces(expression.ToString())
		return
	}

	if expression.exp == "" {
		return
	}

	e.additionalExp = append(e.additionalExp, &AdditionalExpression{
		operator: "or",
		exp:      AddBraces(expression.ToString()),
	})
}

func (e *Expression) ToString() string {
	ret := e.exp

	if e.additionalExp == nil || len(e.additionalExp) == 0 {
		return ret
	}

	for _, addExp := range e.additionalExp {
		ret = Join(ret, " ", addExp.operator, " ", addExp.exp)
	}

	return ret
}

func TrimBraces(exp string) string {
	internalExp := strings.TrimPrefix(exp, "(")
	return strings.TrimSuffix(internalExp, ")")
}

func AddBraces(exp string) string {
	return strings.Join([]string{"(", exp, ")"}, "")
}

func AddNecessaryBraces(exp string) string {
	return AddBraces(TrimBraces(exp))
}

func Join(list ...string) string {
	return strings.Join(list, "")
}
