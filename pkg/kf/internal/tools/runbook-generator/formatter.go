// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runbookgenerator

import (
	"context"
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/google/kf/v2/pkg/kf/doctor/troubleshooter"
	"github.com/russross/blackfriday"
)

// GenTroubleshooterRunbook generates a support runbook from a given doctor
// troubleshooter.
func GenTroubleshooterRunbook(ctx context.Context, w io.Writer, component troubleshooter.Component, docVersion string) {

	componentPlural := component.Type.FriendlyName() + "s"

	// Generate page frontmatter.
	fmt.Fprintln(w, `{# This page is autogenerated by Kf. File a bug under componentid:590084 rather than changing it directly. #}`)
	fmt.Fprintf(w, `{%% extends "migrate/kf/docs/%s/_base.html" %%}`, docVersion)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "{% block custom_heading %}")
	fmt.Fprintf(w, `{%% block page_title %%} Troubleshoot {{kf_product_name_short}} %s {%% endblock %%}`, componentPlural)
	fmt.Fprintln(w)
	fmt.Fprintf(w, `<meta name="description" content="Resolve issues with {{kf_product_name_short}} %s">`, componentPlural)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `<meta name="gtm_var" data-key="docType" data-value="troubleshooting">`)
	fmt.Fprintln(w, "{% endblock %}")

	fmt.Fprintln(w, "{% block body %}")
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Use these steps to troubleshoot various issues that can occur when using {{kf_product_name_short}} %s.", componentPlural)
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	for problemIdx, problem := range component.Problems {

		fmt.Fprintf(w, "## %s {:#problem-%d}", problem.Description, problemIdx)
		fmt.Fprintln(w)
		fmt.Fprintln(w)

		fmt.Fprintln(w, "Run the following command to get the resource information, then check for the causes listed below:")

		fmt.Fprintln(w)
		fmt.Fprintln(w, "```posix-terminal")

		gvr := component.Type.GroupVersionResource(ctx)

		nsFlag := ""
		spaceFlag := ""
		if component.Type.Namespaced() {
			nsFlag = "-n {{ '<var>' }}SPACE_NAME{{ '</var>' }} "
			spaceFlag = "--space {{ '<var>' }}SPACE_NAME{{ '</var>' }} "
		}
		fmt.Fprintf(w,
			"kubectl get %s.%s %s{{ '<var>' }}%s_NAME{{ '</var>' }} -o yaml",
			gvr.Resource,
			gvr.Group,
			nsFlag,
			strings.ToUpper(component.Type.FriendlyName()),
		)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "```")
		fmt.Fprintln(w)

		fmt.Fprintln(w, "The `kf` CLI can help check for some of the issues:")
		fmt.Fprintln(w)

		fmt.Fprintln(w, "```posix-terminal")
		fmt.Fprintf(w,
			"kf doctor %s%s/{{ '<var>' }}%s_NAME{{ '</var>' }}",
			spaceFlag,
			strings.ToLower(component.Type.FriendlyName()),
			strings.ToUpper(component.Type.FriendlyName()),
		)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "```")
		fmt.Fprintln(w)

		fmt.Fprintln(w, "<table>")

		fmt.Fprintln(w, "<thead>")
		fmt.Fprintln(w, "<tr><th>Possible Cause</th><th>Solution</th></tr>")
		fmt.Fprintln(w, "</thead>")

		fmt.Fprintln(w, "<tbody>")
		for _, cause := range problem.Causes {
			fmt.Fprintln(w, "<tr>")

			fmt.Fprintln(w, "<td>")
			fmt.Fprintln(w, html.EscapeString(cause.Description))
			fmt.Fprintln(w, "</td>")

			fmt.Fprintln(w, "<td>")
			formatted := blackfriday.MarkdownCommon([]byte(heredoc.Doc(cause.Recommendation)))
			fmt.Fprintln(w, string(formatted))
			fmt.Fprintln(w, "</td>")

			fmt.Fprintln(w, "</tr>")
		}
		fmt.Fprintln(w, "</tbody>")

		fmt.Fprintln(w, "</table>")
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, `{% endblock %}`)
	fmt.Fprintln(w)
}
