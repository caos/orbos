package mocklabels

import "github.com/caos/orbos/pkg/labels"

var (
	productKey         = "app.kubernetes.io/part-of"
	productVal         = "MOCKING"
	operatorKey        = "app.kubernetes.io/managed-by"
	operatorVal        = "test.caos.ch"
	operatorVersionKey = "app.kubernetes.io/version"
	operatorVersionVal = "v987.654.3210"
	apiKindKey         = "caos.ch/kind"
	apiKindVal         = "MockedLabels"
	apiVersionKey      = "caos.ch/apiversion"
	apiVersionVal      = "v9876"
	componentKey       = "app.kubernetes.io/component"
	componentVal       = "mocked-component"
	nameKey            = "app.kubernetes.io/name"
	NameVal            = "mocked-name"
	selectableKey      = "orbos.ch/selectable"
	selectableVal      = "yes"

	operator           = labels.MustForOperator(productVal, operatorVal, operatorVersionVal)
	api                = labels.MustForAPI(operator, apiKindVal, apiVersionVal)
	Component          = labels.MustForComponent(api, componentVal)
	Name               = labels.MustForName(Component, NameVal)
	ClosedNameSelector = labels.DeriveNameSelector(Name, false)
	Selectable         = labels.AsSelectable(Name)

	NameMap = map[string]string{
		nameKey:            NameVal,
		componentKey:       componentVal,
		apiKindKey:         apiKindVal,
		apiVersionKey:      apiVersionVal,
		operatorKey:        operatorVal,
		operatorVersionKey: operatorVersionVal,
		productKey:         productVal,
	}
	ClosedNameSelectorMap = map[string]string{
		selectableKey: selectableVal,
		componentKey:  componentVal,
		nameKey:       NameVal,
		operatorKey:   operatorVal,
		productKey:    productVal,
	}
	SelectableMap = map[string]string{
		nameKey:            NameVal,
		componentKey:       componentVal,
		apiKindKey:         apiKindVal,
		apiVersionKey:      apiVersionVal,
		operatorKey:        operatorVal,
		operatorVersionKey: operatorVersionVal,
		productKey:         productVal,
		selectableKey:      selectableVal,
	}
)
