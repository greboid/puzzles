<ul>
{{ range $key, $value := . }}
    <li>{{$value}}</li>
{{ end }}
<ul/>