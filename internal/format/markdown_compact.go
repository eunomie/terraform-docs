package format

import (
	gotemplate "text/template"

	"github.com/terraform-docs/terraform-docs/internal/print"
	"github.com/terraform-docs/terraform-docs/internal/template"
	"github.com/terraform-docs/terraform-docs/internal/terraform"
)

const (
	compactHeaderTpl = `
	{{- if .Settings.ShowHeader -}}
		{{- with .Module.Header -}}
			{{ sanitizeHeader . }}
			{{ printf "\n" }}
		{{- end -}}
	{{ end -}}
	`
	compactResourcesTpl = `
	{{- if .Settings.ShowResources -}}
		{{ indent 0 "#" }} Resources
		{{ if not .Module.Resources }}
			No resources.
		{{ else }}
			| Name |
			|------|
			{{- range .Module.Resources }}
			{{ if eq (len .URL) 0 }}
				| {{ .FullType }}
			{{- else -}}
				| [{{ .FullType }}]({{ .URL }}) |
			{{- end }}
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	compactRequirementsTpl = `
	{{- if .Settings.ShowRequirements -}}
		{{ indent 0 "#" }} Requirements
		{{ if not .Module.Requirements }}
			No requirements.
		{{ else }}
			| Name | Version |
			|------|---------|
			{{- range .Module.Requirements }}
				| {{ name .Name }} | {{ tostring .Version | default "n/a" }} |
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	compactProvidersTpl = `
	{{- if .Settings.ShowProviders -}}
		{{ indent 0 "#" }} Providers
		{{ if not .Module.Providers }}
			No provider.
		{{ else }}
			| Name | Version |
			|------|---------|
			{{- range .Module.Providers }}
				| {{ name .FullName }} | {{ tostring .Version | default "n/a" }} |
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	compactInputsTpl = `
	{{- if .Settings.ShowInputs -}}
		{{ indent 0 "#" }} Argument Reference
		{{ if not .Module.Inputs }}
			No argument.
		{{ else }}
			{{- if .Settings.ShowRequired -}}
				{{ if .Module.RequiredInputs }}
					The following arguments are required:
					{{- range .Module.RequiredInputs }}
						{{ template "input" . }}
					{{- end }}
				{{- end }}
				{{ if .Module.OptionalInputs }}
					The following arguments are optional:
					{{- range .Module.OptionalInputs }}
						{{ template "input" . }}
					{{- end }}
				{{ end }}
			{{ else -}}
				The following input variables are supported:
				{{- range .Module.Inputs }}
					{{ template "input" . }}
				{{- end }}
			{{- end }}
		{{- end }}
	{{ end -}}
	`

	compactInputTpl = `
	{{ printf "\n" }}
	- {{ .Name | inlineType }} - {{ .HasDefault | requiredFlag }} {{ tostring .Description | sanitizeTbl }}

	{{ indent 1 " " }} Type: {{ tostring .Type | type }}
	{{- if or .HasDefault (not isRequired) }}

		{{ indent 1 " " }} Default: {{ default "n/a" .GetValue | value }}
	{{- end }}
	`

	compactOutputsTpl = `
	{{- if .Settings.ShowOutputs -}}
		{{ indent 0 "#" }} Attributes Reference
		{{ if not .Module.Outputs }}
			No output.
		{{ else }}
			In addition to arguments above, the following attributes are exported:
			{{- range .Module.Outputs }}

				- {{ .Name | inlineOutputType }} - {{ tostring .Description | sanitizeTbl }}

				{{ if $.Settings.OutputValues }}
					{{- $sensitive := ternary .Sensitive "<sensitive>" .GetValue -}}
					{{ indent 1 " " }} Value: {{ value $sensitive | value }}

					{{ if $.Settings.ShowSensitivity -}}
						{{ indent 1 " " }} Sensitive: {{ ternary (.Sensitive) "yes" "no" }}
					{{- end }}
				{{ end }}
			{{ end }}
		{{ end }}
	{{ end -}}
	`

	compactModulecallsTpl = `
	{{- if .Settings.ShowModuleCalls -}}
		{{ indent 0 "#" }} Modules
		{{ if not .Module.ModuleCalls }}
			No Modules.
		{{ else }}
			| Name | Source | Version |
			|------|--------|---------|
			{{- range .Module.ModuleCalls }}
				| {{ .Name }} | {{ .Source }} | {{ .Version }} |
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	compactTpl = `
	{{- template "header" . -}}
	{{- template "requirements" . -}}
	{{- template "providers" . -}}
	{{- template "modulecalls" . -}}
	{{- template "resources" . -}}
	{{- template "inputs" . -}}
	{{- template "outputs" . -}}
	`
)

// MarkdownCompact represents Markdown Compact format.
type MarkdownCompact struct {
	template *template.Template
}

// NewMarkdownCompact returns new instance of Compact.
func NewMarkdownCompact(settings *print.Settings) print.Engine {
	tt := template.New(settings, &template.Item{
		Name: "document",
		Text: compactTpl,
	}, &template.Item{
		Name: "header",
		Text: compactHeaderTpl,
	}, &template.Item{
		Name: "requirements",
		Text: compactRequirementsTpl,
	}, &template.Item{
		Name: "providers",
		Text: compactProvidersTpl,
	}, &template.Item{
		Name: "resources",
		Text: compactResourcesTpl,
	}, &template.Item{
		Name: "inputs",
		Text: compactInputsTpl,
	}, &template.Item{
		Name: "input",
		Text: compactInputTpl,
	}, &template.Item{
		Name: "outputs",
		Text: compactOutputsTpl,
	}, &template.Item{
		Name: "modulecalls",
		Text: compactModulecallsTpl,
	})
	tt.CustomFunc(gotemplate.FuncMap{
		"type": func(t string) string {
			return printFencedCodeBlockWithIndent(t, "hcl")
		},
		"inlineType": func(t string) string {
			result, _ := printFencedCodeBlock(t, "hcl")
			return "<a id='input_" + t + "' href='#" + t + "'>" + result + "</a>"
		},
		"inlineOutputType": func(t string) string {
			result, _ := printFencedCodeBlock(t, "hcl")
			return "<a id='output_" + t + "' href='#" + t + "'>" + result + "</a>"
		},
		"value": func(v string) string {
			if v == "n/a" {
				return v
			}
			return printFencedCodeBlockWithIndent(v, "json")
		},
		"isRequired": func() bool {
			return settings.ShowRequired
		},
		"requiredFlag": func(r bool) string {
			if r {
				return "(Optional)"
			}
			return "(Required)"
		},
	})
	return &MarkdownCompact{
		template: tt,
	}
}

// Print a Terraform module as Markdown compact.
func (c *MarkdownCompact) Print(module *terraform.Module, settings *print.Settings) (string, error) {
	rendered, err := c.template.Render(module)
	if err != nil {
		return "", err
	}
	return sanitize(rendered), nil
}

func init() {
	register(map[string]initializerFn{
		"markdown compact": NewMarkdownCompact,
		"md compact":       NewMarkdownCompact,
	})
}
