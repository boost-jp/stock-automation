{{- $alias := .Aliases.Table .Table.Name }}

func ({{$alias.UpSingular}}) GetColumns() []string {
  return {{$alias.DownSingular}}AllColumns
}

func ({{$alias.UpSingular}}) GetPKs() []string {
  return {{$alias.DownSingular}}PrimaryKeyColumns
}

func ({{$alias.DownSingular}}Query) GetColumns() []string {
  return {{$alias.DownSingular}}AllColumns
}

func ({{$alias.DownSingular}}Query) GetPKs() []string {
  return {{$alias.DownSingular}}PrimaryKeyColumns
}