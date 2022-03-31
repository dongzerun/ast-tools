package pkg

import (
	"bytes"
	"fmt"
)

type report struct {
	fieldName string
	ignore    bool
	reason    string
	reports   []*report
}

func (c *convert) appendReport(fieldName string, r *report) {
	if _, exists := c.reported[fieldName]; exists {
		return
	}
	c.reported[fieldName] = struct{}{}
	c.reports = append(c.reports, r)
}

func (c *convert) report(buf *bytes.Buffer, steps int) {
	ident := ""
	for i := 0; i < steps; i++ {
		ident = ident + "    "
	}
	buf.WriteString(ident + "{\n")
	for i := range c.reports {
		buf.WriteString(ident + "    ")
		child, exists := c.children[c.reports[i].fieldName]

		var output string
		if c.reports[i].ignore {
			output = fmt.Sprintf("%s False %s\n", c.reports[i].fieldName, c.reports[i].reason)
		} else {
			if !exists {
				output = fmt.Sprintf("%s OK\n", c.reports[i].fieldName)
			} else {
				output = fmt.Sprintf("%s\n", c.reports[i].fieldName)
			}
		}

		buf.WriteString(output)
		if c.reports[i].ignore {
			continue
		}

		if !exists {
			continue
		}
		child.report(buf, steps+1)
	}
	buf.WriteString(ident + "}\n")
}
