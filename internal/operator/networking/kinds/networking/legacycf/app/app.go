package app

import (
	"context"
	"strings"

	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/labels"

	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/cloudflare"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/cloudflare/expression"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/config"
)

type App struct {
	cloudflare     *cloudflare.Cloudflare
	groups         map[string][]string
	internalPrefix string
}

func New(user string, key string, userServiceKey string, groups map[string][]string, internalPrefix string) (*App, error) {
	api, err := cloudflare.New(user, key, userServiceKey)
	if err != nil {
		return nil, err
	}

	return &App{
		cloudflare:     api,
		groups:         groups,
		internalPrefix: internalPrefix,
	}, nil
}

func (a *App) TrimInternalPrefix(desc string) string {
	return strings.TrimPrefix(desc, a.internalPrefix)
}

func (a *App) AddInternalPrefix(desc string) string {
	return strings.Join([]string{a.internalPrefix, desc}, " ")
}

func (a *App) Ensure(ctx context.Context, k8sClient kubernetes.ClientInt, namespace string, domain string, subdomains []*config.Subdomain, rules []*config.Rule, originCALabels *labels.Name) error {
	firewallRulesInt := make([]*cloudflare.FirewallRule, 0)
	filtersInt := make([]*cloudflare.Filter, 0)
	recordsInt := make([]*cloudflare.DNSRecord, 0)

	for _, record := range subdomains {

		if record.Subdomain == "@" {
			record.Subdomain = domain
		}

		name := domain
		if record.Subdomain != domain {
			name = strings.Join([]string{record.Subdomain, domain}, ".")
		}
		ttl := record.TTL
		if ttl == 0 {
			ttl = 1
		}

		recordsInt = append(recordsInt, &cloudflare.DNSRecord{
			Type:     record.Type,
			Name:     name,
			Content:  record.IP,
			Proxied:  record.Proxied,
			TTL:      ttl,
			Priority: record.Priority,
		})
	}

	err := a.EnsureDNSRecords(ctx, domain, recordsInt)
	if err != nil {
		return err
	}

	if rules != nil {
		for _, rule := range rules {
			filterExp := cloudflare.EmptyExpression()
			for _, filter := range rule.Filters {
				filterExpAdd := cloudflare.EmptyExpression()

				// get all targets
				addContainsTargetsFromList(domain, filter.ContainsTargets, filterExpAdd)
				a.addContainsTargetGroupsFromList(domain, filter.ContainsTargetsGroups, filterExpAdd)

				// get all targets
				addTargetsFromList(domain, filter.Targets, filterExpAdd)
				a.addTargetGroupsFromList(domain, filter.TargetGroups, filterExpAdd)

				// get all sources
				addSourcesFromList(filter.Sources, filterExpAdd)
				a.addSourceGroupsFromList(filter.SourceGroups, filterExpAdd)

				if filter.SSL == "true" {
					filterExpAdd.And(cloudflare.SSLExpression())
				} else if filter.SSL == "false" {
					filterExpAdd.And(cloudflare.NotSSLExpression())
				}

				// add expression as or-element
				filterExp.Or(filterExpAdd)
			}

			filterInt := &cloudflare.Filter{
				Description: a.AddInternalPrefix(rule.Description),
				Expression:  filterExp.ToString(),
				Paused:      false,
			}
			filtersInt = append(filtersInt, filterInt)
		}

	}
	filters, deleteFiltersFunc, err := a.EnsureFilters(ctx, domain, filtersInt)
	if err != nil {
		return err
	}

	if rules != nil {
		for _, rule := range rules {
			for _, filter := range filters {
				descInt := a.AddInternalPrefix(rule.Description)
				if filter.Description == descInt {
					firewallRule := &cloudflare.FirewallRule{
						Paused:      false,
						Description: descInt,
						Action:      rule.Action,
						Filter:      filter,
						Priority:    rule.Priority,
					}
					firewallRulesInt = append(firewallRulesInt, firewallRule)
				}
			}
		}
	}

	if err := a.EnsureFirewallRules(ctx, domain, firewallRulesInt); err != nil {
		return err
	}

	// filters can only be deleted after there is no use left in the firewall rules
	if err := deleteFiltersFunc(); err != nil {
		return err
	}

	return a.EnsureOriginCACertificate(ctx, k8sClient, namespace, originCALabels, domain)
}

func addSourcesFromList(subList []string, exp *expression.Expression) {
	if subList != nil && len(subList) > 0 {
		exp.And(cloudflare.IPExpressionIsIn(subList))
	}
}

func (a *App) addSourceGroupsFromList(groupList []string, exp *expression.Expression) {
	if groupList != nil && len(groupList) > 0 {
		for _, groupName := range groupList {
			group, found := a.groups[groupName]
			if found {
				addSourcesFromList(group, exp)
			}
		}
	}
}

func addContainsTargetsFromList(domain string, subList []string, exp *expression.Expression) {
	if subList != nil && len(subList) > 0 {
		for _, sub := range subList {
			target := strings.Join([]string{"\"", sub, ".", domain, "\""}, "")

			exp.And(cloudflare.HostnameExpressionContains(target))
		}
	}
}

func (a *App) addContainsTargetGroupsFromList(domain string, groupList []string, exp *expression.Expression) {
	if groupList != nil && len(groupList) > 0 {
		for _, groupname := range groupList {
			group, found := a.groups[groupname]
			if found {
				addContainsTargetsFromList(domain, group, exp)
			}
		}
	}
}

func addTargetsFromList(domain string, list []string, exp *expression.Expression) {
	if list != nil && len(list) > 0 {
		targets := make([]string, 0)
		for _, sub := range list {
			targets = append(targets, strings.Join([]string{"\"", sub, ".", domain, "\""}, ""))
		}
		exp.And(cloudflare.HostnameExpressionIsIn(targets))
	}
}

func (a *App) addTargetGroupsFromList(domain string, groupList []string, exp *expression.Expression) {
	if groupList != nil && len(groupList) > 0 {
		for _, groupname := range groupList {
			group, found := a.groups[groupname]
			if found {
				addTargetsFromList(domain, group, exp)
			}
		}
	}
}
