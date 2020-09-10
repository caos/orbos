package app

import (
	"reflect"

	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf/cloudflare"
)

func (a *App) EnsureFirewallRules(domain string, rules []*cloudflare.FirewallRule) error {
	currentRules, err := a.cloudflare.GetFirewallRules(domain)
	if err != nil {
		return err
	}

	deleteRules := getFirewallRulesToDelete(currentRules, rules)
	if deleteRules != nil && len(deleteRules) > 0 {
		if err := a.cloudflare.DeleteFirewallRules(domain, deleteRules); err != nil {
			return err
		}
	}

	createRules := getFirewallRulesToCreate(currentRules, rules)
	if createRules != nil && len(createRules) > 0 {
		_, err := a.cloudflare.CreateFirewallRules(domain, createRules)
		if err != nil {
			return err
		}
	}

	updateRules := getFirewallRulesToUpdate(currentRules, rules)
	if updateRules != nil && len(updateRules) > 0 {
		_, err := a.cloudflare.UpdateFirewallRules(domain, updateRules)
		if err != nil {
			return err
		}

	}

	return nil
}

func getFirewallRulesToDelete(currentRules []*cloudflare.FirewallRule, rules []*cloudflare.FirewallRule) []string {
	deleteRules := make([]string, 0)

	for _, currentRule := range currentRules {
		found := false
		if rules != nil {
			for _, rule := range rules {
				if currentRule.Description == rule.Description {
					found = true
				}
			}
		}

		if found == false {
			deleteRules = append(deleteRules, currentRule.ID)
		}
	}

	return deleteRules
}

func getFirewallRulesToCreate(currentRules []*cloudflare.FirewallRule, rules []*cloudflare.FirewallRule) []*cloudflare.FirewallRule {
	createRules := make([]*cloudflare.FirewallRule, 0)

	if rules != nil {
		for _, rule := range rules {
			found := false
			for _, currentRule := range currentRules {
				if currentRule.Description == rule.Description {
					found = true
					break
				}
			}
			if found == false {
				createRules = append(createRules, rule)
			}
		}
	}

	return createRules
}

func getFirewallRulesToUpdate(currentRules []*cloudflare.FirewallRule, rules []*cloudflare.FirewallRule) []*cloudflare.FirewallRule {
	updateRules := make([]*cloudflare.FirewallRule, 0)

	if rules != nil {
		for _, rule := range rules {
			for _, currentRule := range currentRules {
				if currentRule.Description == rule.Description &&
					!reflect.DeepEqual(currentRule, rule) {
					rule.ID = currentRule.ID
					updateRules = append(updateRules, rule)
				}
			}
		}
	}

	return updateRules
}
