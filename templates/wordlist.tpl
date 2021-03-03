<ul>
{{ range $key, $value := . }}
    <li>{{$key}}<ul>
    {{ range $subkey, $subvalue := $value }}
        <li>{{$subvalue}}</li>
    {{ end }}
    </ul></li>
{{ end }}
</ul>