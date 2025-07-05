{{- $alias := .Aliases.Table .Table.Name -}}
{{- $orig_tbl_name := .Table.Name -}}

//go:generate go run  ../../../cmd/generator/repoinit --fields={{- range $col := .Table.Columns -}} {{ $alias.Column $col.Name }}, {{- end }} {{$alias.UpSingular}}

// You can edit this as you like.

// {{$alias.UpSingular}} is an object representing the database table.
type {{$alias.UpSingular}} struct {
	{{- range $column := .Table.Columns -}}
	{{- $colAlias := $alias.Column $column.Name -}}
	{{- $orig_col_name := $column.Name -}}
	{{if ignore $orig_tbl_name $orig_col_name $.TagIgnore -}}
	{{$colAlias}} {{$column.Type}}
	{{else if eq $.StructTagCasing "title" -}}
	{{$colAlias}} {{$column.Type}}
	{{else if eq $.StructTagCasing "camel" -}}
	{{$colAlias}} {{$column.Type}}
	{{else if eq $.StructTagCasing "alias" -}}
	{{$colAlias}} {{$column.Type}}
	{{else -}}
	{{$colAlias}} {{$column.Type}} {{ if ne $column.Comment ""}} // {{ $column.Comment }} {{ end }}
	{{end -}}
	{{end -}}
}


func New{{$alias.UpSingular}}( {{ printf "\n" }}
        {{- range $column := .Table.Columns -}}
        {{- $colAlias := $alias.Column $column.Name -}}
        {{- $orig_col_name := $column.Name -}}
        {{- if ignore $orig_tbl_name $orig_col_name $.TagIgnore -}}
        {{- else -}}
        {{$colAlias}} {{$column.Type}}, {{ printf "\n" }}
        {{- end -}}
        {{- end -}}
) (*{{$alias.UpSingular}}) {
    do := &{{$alias.UpSingular}}{ {{ printf "\n" }}
{{- range $column := .Table.Columns -}}
        {{- $colAlias := $alias.Column $column.Name -}}
        {{- $orig_col_name := $column.Name -}}
        {{- if ignore $orig_tbl_name $orig_col_name $.TagIgnore -}}
        {{- else -}}
        {{$colAlias}}: {{$colAlias}}, {{ printf "\n" }}
        {{- end -}}
        {{- end -}}
    }
    return do
}