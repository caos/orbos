package cloudflare

import (
	"context"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type FirewallRule struct {
	ID          string      `json:"id,omitempty"`
	Paused      bool        `json:"paused"`
	Description string      `json:"description"`
	Action      string      `json:"action"`
	Priority    interface{} `json:"priority"`
	Filter      *Filter     `json:"filter"`
	Products    []string    `json:"products,omitempty"`
	CreatedOn   time.Time   `json:"created_on,omitempty"`
	ModifiedOn  time.Time   `json:"modified_on,omitempty"`
}

func (c *Cloudflare) CreateFirewallRules(ctx context.Context, domain string, rules []*FirewallRule) ([]*FirewallRule, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	createdRules, err := c.api.CreateFirewallRules(ctx, id, internalRulesToRules(rules))

	if err != nil {
		return nil, err
	}

	return rulesToInternalRules(createdRules), err
}

func (c *Cloudflare) UpdateFirewallRules(ctx context.Context, domain string, rules []*FirewallRule) ([]*FirewallRule, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	updatedRules, err := c.api.UpdateFirewallRules(ctx, id, internalRulesToRules(rules))
	if err != nil {
		return nil, err
	}

	return rulesToInternalRules(updatedRules), err
}

func (c *Cloudflare) DeleteFirewallRules(ctx context.Context, domain string, ruleIDs []string) error {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return err
	}

	for _, ruleID := range ruleIDs {
		if err := c.api.DeleteFirewallRule(ctx, id, ruleID); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cloudflare) GetFirewallRules(ctx context.Context, domain string) ([]*FirewallRule, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	rules, err := c.api.FirewallRules(ctx, id, pageOpts)
	if err != nil {
		return nil, err
	}
	return rulesToInternalRules(rules), nil
}

func rulesToInternalRules(rules []cloudflare.FirewallRule) []*FirewallRule {
	retRules := make([]*FirewallRule, 0)
	for _, rule := range rules {
		retRules = append(retRules, ruleToInternalRule(rule))
	}
	return retRules
}

func ruleToInternalRule(rule cloudflare.FirewallRule) *FirewallRule {
	priority := 1
	priorityInt, err := rule.Priority.(int)
	if err {
		priority = priorityInt
	}
	return &FirewallRule{
		ID:          rule.ID,
		Paused:      rule.Paused,
		Description: rule.Description,
		Action:      rule.Action,
		Priority:    priority,
		Filter: &Filter{
			ID: rule.Filter.ID,
		},
		Products:   rule.Products,
		CreatedOn:  rule.CreatedOn,
		ModifiedOn: rule.ModifiedOn,
	}
}

func internalRulesToRules(rules []*FirewallRule) []cloudflare.FirewallRule {
	retRules := make([]cloudflare.FirewallRule, 0)
	for _, rule := range rules {
		retRules = append(retRules, internalRuleToRule(rule))
	}
	return retRules
}

func internalRuleToRule(rule *FirewallRule) cloudflare.FirewallRule {

	return cloudflare.FirewallRule{
		ID:          rule.ID,
		Paused:      rule.Paused,
		Description: rule.Description,
		Action:      rule.Action,
		Priority:    rule.Priority,
		Filter: cloudflare.Filter{
			ID: rule.Filter.ID,
		},
		Products:   rule.Products,
		CreatedOn:  rule.CreatedOn,
		ModifiedOn: rule.ModifiedOn,
	}
}
