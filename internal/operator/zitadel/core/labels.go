package core

type CurrentLabels struct {
	labels map[string]string
}

func (c *CurrentLabels) SetLabel(k string, v string) {
	if c.labels == nil {
		c.labels = map[string]string{k: v}
	} else {
		c.labels[k] = v
	}
}

func (c *CurrentLabels) ListLabels() map[string]string {
	return c.labels
}

func (c *CurrentLabels) GetLabel(k string) string {
	label, ok := c.labels[k]
	if !ok {
		return ""
	}
	return label
}
